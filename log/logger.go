package log

import (
	"fmt"
	"sync"
	"runtime"
	"path"
	"time"
	"os"
	"strconv"
	"strings"
)

var version = "v1.2"

const (
	LEVEL_ERROR		= iota
	LEVEL_WARNING
	LEVEL_INFO
	LEVEL_TRACE
	LEVEL_DEBUG
)

type adapterLoggerFunc func() LoggerAbstract

type LoggerAbstract interface {
	Name() string
	Init(config Config) error
	Write(loggerMsg *loggerMessage) error
	Flush()
}

var adapters = make(map[string]adapterLoggerFunc)

var levelStringMapping = map[int]string{
	LEVEL_ERROR:       "Error",
	LEVEL_WARNING:     "Warning",
	LEVEL_INFO:        "Info",
	LEVEL_TRACE:       "Trace",
	LEVEL_DEBUG:       "Debug",
}

var defaultLoggerMessageFormat = "%time% [%type%] %body%"

//Register logger adapter
func Register(adapterName string, newLog adapterLoggerFunc)  {
	if adapters[adapterName] != nil {
		panic("logger: logger adapter "+ adapterName +" already registered!")
	}
	if newLog == nil {
		panic("logger: logger adapter "+ adapterName +" is nil!")
	}

	adapters[adapterName] = newLog
}

type Logger struct {
	lock        sync.Mutex //sync lock
	outputs     []*outputLogger // outputs loggers
	msgChan     chan *loggerMessage // message channel
	synchronous bool // is sync
	wait        sync.WaitGroup // process wait
	signalChan  chan string
	flag		string
}

type outputLogger struct {
	Name string
	Level int
	LoggerAbstract
}

type loggerMessage struct {
	Flag string `json:"flag"`
	Nano int64 `json:"nano"`
	Time string `json:"time"`
	Level int `json:"level"`
	Type string `json:"type"`
	File string `json:"file"`
	Line int `json:"line"`
	Func string `json:"func"`
	Body string `json:"body"`
}

//new logger
//return logger
func NewLogger(flags ...string) *Logger {
	logger := &Logger{
		outputs:        []*outputLogger{},
		msgChan:        make(chan *loggerMessage, 10),
		synchronous:    true,
		wait:           sync.WaitGroup{},
		signalChan:     make(chan string, 1),
	}

	if len(flags) > 0 {
		logger.flag = flags[0]
	}

	//default adapter console
	// logger.attach("console", LEVEL_DEBUG, &ConsoleConfig{})

	return logger
}

//start attach a logger adapter
//param : adapterName console | file | database | ...
//return : error
func (logger *Logger) Attach(adapterName string, level int, config Config) error {
	logger.lock.Lock()
	defer logger.lock.Unlock()

	return logger.attach(adapterName, level, config)
}

//attach a logger adapter after lock
//param : adapterName console | file | database | ...
//return : error
func (logger *Logger) attach(adapterName string, level int, config Config) error {
	for _, output := range logger.outputs {
		if output.Name == adapterName {
			printError("logger: adapter " +adapterName+ "already attached!")
		}
	}
	logFun, ok := adapters[adapterName]
	if !ok {
		printError("logger: adapter " +adapterName+ "is nil!")
	}
	adapterLog := logFun()
	err := adapterLog.Init(config)
	if err != nil {
		printError("logger: adapter " +adapterName+ " init failed, error: " +err.Error())
	}

	output := &outputLogger {
		Name:adapterName,
		Level: level,
		LoggerAbstract: adapterLog,
	}

	logger.outputs = append(logger.outputs, output)
	return nil
}

//start attach a logger adapter
//param : adapterName console | file | database | ...
//return : error
func (logger *Logger) Detach(adapterName string) error {
	logger.lock.Lock()
	defer logger.lock.Unlock()

	return logger.detach(adapterName)
}

//detach a logger adapter after lock
//param : adapterName console | file | database | ...
//return : error
func (logger *Logger) detach(adapterName string) error {
	outputs := []*outputLogger{}
	for _, output := range logger.outputs {
		if output.Name == adapterName {
			continue
		}
		outputs = append(outputs, output)
	}
	logger.outputs = outputs
	return nil
}


func (logger *Logger) SetFlag(flag string) {
	logger.flag = flag
}

//set logger synchronous false
//params : sync bool
func (logger *Logger) SetAsync(data... int) {
	logger.lock.Lock()
	defer logger.lock.Unlock()
	logger.synchronous = false

	msgChanLen := 100
	if len(data) > 0 {
		msgChanLen = data[0]
	}

	logger.msgChan = make(chan *loggerMessage, msgChanLen)
	logger.signalChan = make(chan string, 1)

	if !logger.synchronous {
		go func() {
			defer func() {
				e := recover()
				if e != nil {
					fmt.Printf("%v", e)
				}
			}()
			logger.startAsyncWrite()
		}()
	}
}

//write log message
//params : level int, msg string
//return : error
func (logger *Logger) Writer(level int, msg string) error {
	funcName := "null"
	pc, file, line, ok := runtime.Caller(2)
	if !ok {
		file = "null"
		line = 0
	}else {
		funcName = runtime.FuncForPC(pc).Name()
	}
	_, filename := path.Split(file)

	if levelStringMapping[level] == "" {
		printError("logger: level " + strconv.Itoa(level) + " is illegal!")
	}

	now := time.Now()

	loggerMsg := &loggerMessage {
		Flag:	logger.flag,
		Nano : now.UnixNano(),
		Time : now.Format("2006-01-02 15:04:05.000"),
		Level :level,
		Type: levelStringMapping[level],
		Body: msg,
		File : filename,
		Line : line,
		Func: funcName,
	}

	if !logger.synchronous {
		logger.wait.Add(1)
		logger.msgChan <- loggerMsg
	}else {
		logger.writeToOutputs(loggerMsg)
	}

	return nil
}

//sync write message to loggerOutputs
//params : loggerMessage
func (logger *Logger) writeToOutputs(loggerMsg *loggerMessage)  {
	for _, loggerOutput := range logger.outputs {
		// write level
		if loggerOutput.Level >= loggerMsg.Level {
			err := loggerOutput.Write(loggerMsg)
			if err != nil {
				fmt.Fprintf(os.Stderr, "logger: unable write loggerMessage to adapter:%v, error: %v\n", loggerOutput.Name, err)
			}
		}
	}
}

//start async write by read logger.msgChan
func (logger *Logger) startAsyncWrite()  {
	for {
		select {
		case loggerMsg := <-logger.msgChan:
			logger.writeToOutputs(loggerMsg)
			logger.wait.Done()
		case signal := <-logger.signalChan:
			if signal == "flush" {
				logger.flush()
			}
		}
	}
}

//flush msgChan data
func (logger *Logger) flush() {
	if !logger.synchronous {
		for {
			if len(logger.msgChan) > 0 {
				loggerMsg := <-logger.msgChan
				logger.writeToOutputs(loggerMsg)
				logger.wait.Done()
				continue
			}
			break
		}
		for _, loggerOutput := range logger.outputs {
			loggerOutput.Flush()
		}
	}
}

//if SetAsync() or logger.synchronous is false, must call Flush() to flush msgChan data
func (logger *Logger) Flush()  {
	if !logger.synchronous {
		logger.signalChan <- "flush"
		logger.wait.Wait()
		return
	}
	logger.flush()
}

func LoggerLevel(levelStr string) int {
	levelStr = strings.ToUpper(levelStr)
	switch levelStr {
	case "ERROR":
		return LEVEL_ERROR
	case "WARNING":
		return LEVEL_WARNING
	case "INFO":
		return LEVEL_INFO
	case "TRACE":
		return LEVEL_TRACE
	case "DEBUG":
		return LEVEL_DEBUG
	default:
		return LEVEL_DEBUG
	}
}

func loggerMessageFormat(format string, loggerMsg *loggerMessage) string {
	message := strings.Replace(format, "%flag%", loggerMsg.Flag, 1)
	message = strings.Replace(message, "%nano%", strconv.FormatInt(loggerMsg.Nano, 10), 1)
	message = strings.Replace(message, "%time%", loggerMsg.Time, 1)
	message = strings.Replace(message, "%level%", strconv.Itoa(loggerMsg.Level), 1)
	message = strings.Replace(message, "%type%", strings.ToUpper(loggerMsg.Type), 1)
	message = strings.Replace(message, "%file%", loggerMsg.File, 1)
	message = strings.Replace(message, "%line%", strconv.Itoa(loggerMsg.Line), 1)
	message = strings.Replace(message, "%func%", loggerMsg.Func, 1)
	message = strings.Replace(message, "%body%", loggerMsg.Body, 1)

	return message
}


//log error level
func (logger *Logger) Error(msg string) {
	logger.Writer(LEVEL_ERROR, msg)
}

//log error format
func (logger *Logger) Errorf(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	logger.Writer(LEVEL_ERROR, msg)
}

//log warning level
func (logger *Logger) Warning(msg string) {
	logger.Writer(LEVEL_WARNING, msg)
}

//log warning format
func (logger *Logger) Warningf(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	logger.Writer(LEVEL_WARNING, msg)
}

//log info level
func (logger *Logger) Info(msg string) {
	logger.Writer(LEVEL_INFO, msg)
}

//log info format
func (logger *Logger) Infof(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	logger.Writer(LEVEL_INFO, msg)
}



//log trace level
func (logger *Logger) Trace(msg string) {
	logger.Writer(LEVEL_TRACE, msg)
}

//log trace format
func (logger *Logger) Tracef(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	logger.Writer(LEVEL_TRACE, msg)
}



//log debug level
func (logger *Logger) Debug(msg string) {
	logger.Writer(LEVEL_DEBUG, msg)
}

//log debug format
func (logger *Logger) Debugf(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	logger.Writer(LEVEL_DEBUG, msg)
}

func printError(message string) {
	fmt.Println(message)
	os.Exit(0)
}

