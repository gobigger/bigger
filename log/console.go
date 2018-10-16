package log

import (
	"sync"
	"io"
	"os"
	"reflect"
	"errors"
)

const CONSOLE_ADAPTER_NAME  = "console"

// var levelColors = map[int] color.Attribute {
// 	LEVEL_ERROR:     color.FgRed,    //red
// 	LEVEL_WARNING:   color.FgYellow, //yellow
// 	LEVEL_INFO:      color.FgWhite,   //blue
// 	LEVEL_DEBUG:     color.BgBlue,   //background blue
// }

// adapter console
type AdapterConsole struct {
	write *ConsoleWriter
	config *ConsoleConfig
}

// console writer
type ConsoleWriter struct {
	lock sync.Mutex
	writer io.Writer
}

// console config
type ConsoleConfig struct {
	// console text is show color
	// Color bool

	// is json format
	Json bool

	// jsonFormat is false, please input format string
	// if format is empty, default format "%millisecond_format% [%level_string%] %body%"
	//
	//  Timestamp "%timestamp%"
	//	TimestampFormat "%timestamp_format%"
	//	Millisecond "%millisecond%"
	//	MillisecondFormat "%millisecond_format%"
	//	Level int "%level%"
	//	LevelString "%level_string%"
	//	Body string "%body%"
	//	File string "%file%"
	//	Line int "%line%"
	//	Function "%function%"
	//
	// example: format = "%millisecond_format% [%level_string%] %body%"
	Format string
}

func (cc *ConsoleConfig) Name() string {
	return CONSOLE_ADAPTER_NAME
}

func NewAdapterConsole() LoggerAbstract {
	consoleWrite := &ConsoleWriter{
		writer: os.Stdout,
	}
	config := &ConsoleConfig{}
	return &AdapterConsole{
		write: consoleWrite,
		config : config,
	}
}

func (adapterConsole *AdapterConsole) Init(consoleConfig Config) error {
	if consoleConfig.Name() != CONSOLE_ADAPTER_NAME {
		return errors.New("logger console adapter init error, config must ConsoleConfig")
	}

	vc := reflect.ValueOf(consoleConfig)
	cc := vc.Interface().(*ConsoleConfig)
	adapterConsole.config = cc

	if cc.Json == false && cc.Format == "" {
		cc.Format = defaultLoggerMessageFormat
	}

	return nil
}

func (adapterConsole *AdapterConsole) Write(loggerMsg *loggerMessage) error {

	msg := ""
	if adapterConsole.config.Json == true  {
		//jsonByte, _ := json.Marshal(loggerMsg)
		jsonByte, _ := loggerMsg.MarshalJSON()
		msg = string(jsonByte)
	}else {
		msg = loggerMessageFormat(adapterConsole.config.Format, loggerMsg)
	}
	consoleWriter := adapterConsole.write

	consoleWriter.lock.Lock()
	consoleWriter.writer.Write([]byte(msg + "\n"))
	consoleWriter.lock.Unlock()

	return nil
}

func (adapterConsole *AdapterConsole) Name() string {
	return CONSOLE_ADAPTER_NAME
}

func (adapterConsole *AdapterConsole) Flush() {

}


func init()  {
	Register(CONSOLE_ADAPTER_NAME, NewAdapterConsole)
}






/*

	logger := blog.NewLogger()
	//default attach console, detach console
	logger.Detach("console")

	consoleConfig := &blog.ConsoleConfig{
		Color: true,
		Json: false,
		Format: "%millisecond_format% [%level_string%] [%file%:%line%] %body%",
	}

	logger.Attach("console", blog.LEVEL_DEBUG, consoleConfig)

	logger.SetAsync()
	
	logger.Infof("this is a info %s log!", "format")
	logger.Debugf("this is a debug %s log!", "format")

	logger.Flush()

*/