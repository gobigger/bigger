package bigger

import (
	"strconv"
	"path"
	"runtime"
    "io/ioutil"
	"sync"
    "os"
    "io"
	"os/signal"
	"syscall"
    "plugin"
    "strings"
    "time"
    "fmt"
    "encoding/json"
    "encoding/base64"
    "github.com/yatlabs/bigger/hashid"
    "github.com/yatlabs/bigger/fastid"
)

var (
    encodeTextAlphabet      = "01234AaBbCcDdEeFfGgHhIiJjKkLlMmNnOoPpQqRrSsTtUuVvWwXxYyZz56789+/"
	encodeDigitAlphabet     = "abcdefghijkmnpqrstuvwxyz123456789ACDEFGHJKLMNPQRSTUVWXYZ"	//l,I,0,O,o,B
    encodeDigitSalt         = "bigger.digit.salt"
    encodeDigitLength       = 7
)

type (
	bigger	struct {
        running     bool
        Id          int64
        Name        string
        Mode    	env

        Config      *configConfig
        Setting     Map

        mutex	    sync.Mutex
        hosts       map[string]string
        plugins		map[string]*plugin.Plugin

        raft        *raftStore
        url         *contextUrl

        textCoder   *base64.Encoding
        digitCoder  *hashid.HashID
        fastid      *fastid.FastID
	}
)

func (bigger *bigger) initing() {
    if bigger.running {
        return
    }
    
    mLOGGER.initing()
    mMutex.initing()

    //最新加载插件
    bigger.loading()    //加载插件

    bigger.raftIniting()

    
    mSESSION.initing()
    mCACHE.initing()
    mDATA.initing()
    mFILE.initing()

    mPLAN.initing()
    mEVENT.initing()
    mQUEUE.initing()

    mHTTP.initing()
    mSOCKET.initing()
    mVIEW.initing()

    bigger.Info("bigger raft is binding on", bigger.Config.Node.Bind)
    Bigger.Info("bigger http is running at", bigger.Config.Node.Port)
    
    //开始触发器
    bigger.SyncTrigger(EventBiggerStart, Map{})

    bigger.running = true
}

func (bigger *bigger) exiting() {
    if bigger.running == false {
        return
    }
    
    Bigger.Info("bigger end")

    //结束触发器
    bigger.SyncTrigger(EventBiggerEnd, Map{})


    mVIEW.exiting()
    mSOCKET.exiting()
    mHTTP.exiting()

    mPLAN.exiting()
    mQUEUE.exiting()
    mEVENT.exiting()

    mFILE.exiting()
    mDATA.exiting()
    mCACHE.exiting()

    mSESSION.exiting()

    bigger.raftExiting()

    mMutex.exiting()
    mLOGGER.exiting()
}
func (bigger *bigger) wating() {
    exitChan := make(chan os.Signal, 1)
	signal.Notify(exitChan, os.Kill, os.Interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
    <-exitChan
}


//预加载，在所有init代码载入以前执行
//先预加载一部分so文件
func (bigger *bigger) preloading() {
    
    files := []string{}

    dirs,err := ioutil.ReadDir(path.Join(Bigger.Config.Path.Plugin))
    if err != nil {
        panic("[插件]加载目录失败："+err.Error())
    }

    for _,ff := range dirs {
        name := ff.Name()
        if !ff.IsDir() && strings.HasSuffix(name, "."+runtime.GOOS+".so") &&  strings.HasPrefix(name, ".") == true {
            files = append(files, path.Join(Bigger.Config.Path.Plugin, name))
        }
    }

    for _,n := range files {
        err := bigger.load(n)
        if err != nil {
            mLOGGER.Info("[插件]", n, "加载失败：" + err.Error())
        } else {
            mLOGGER.Info("[插件]", n, "加载成功")
        }
    }
}




//这个加载是在init的时候做的，代码早就注册好了
//在这里加载的so，会替换已经存在的
func (bigger *bigger) loading() {
    
    files := []string{}
    
    dirs,err := ioutil.ReadDir(path.Join(Bigger.Config.Path.Plugin))
    if err != nil {
        panic("[插件]加载目录失败："+err.Error())
    }

    for _,ff := range dirs {
        name := ff.Name()
        if !ff.IsDir() && strings.HasSuffix(name, "."+runtime.GOOS+".so") && strings.HasPrefix(name, ".") == false {
            files = append(files, path.Join(Bigger.Config.Path.Plugin, name))
        }
    }

    for _,n := range files {
        err := bigger.load(n)
        if err != nil {
            Bigger.Info("[插件]", n, "加载失败", err.Error())
        } else {
            Bigger.Info("[插件]", n, "加载成功")
        }
    }
}



//准备工作，主要是预加载
func (bigger *bigger) ready() {
    if bigger.running == false {
        //这里有个问题，bigger/builtin包，会在preloading之后加载
        //如果想在preload的时候注册builtin里有的内容，会被替换，所以builtin里要能不覆盖就好了
        bigger.builtin()
        bigger.preloading()
    }
}

func (bigger *bigger) Go() {
    if bigger.running == false {
        bigger.initing()
        bigger.wating() //等待退出信号
        bigger.exiting()
    }
}


func (bigger *bigger) raftIniting() {
	raft := newRaftStore(bigger.Config.Path.Node, bigger.Config.Node.Bind)
	if err := raft.Open(bigger.Config.Node.Join); err != nil {
		panic("[RAFT] 打开失败：" + err.Error())
    }
    bigger.raft = raft
}
func (bigger *bigger) raftExiting() {
    if bigger.raft != nil {
        bigger.raft.Close()
    }
}




func (bigger *bigger) load(files ...string) (*Error) {
	bigger.mutex.Lock()
	defer bigger.mutex.Unlock()

    for _,file := range files {

        //直接加载就行
        //如果是运行模式，系统会自动预加载，保存在临时的列表
        pg, err := plugin.Open(file)
        if err != nil {
            return Bigger.Erred(err)
        }

        //运行时动态注册
        if bigger.running {
            bigger.register()
        }

        // //获取插件标识：只要Lookup一个小写开头的，就会报错，会提示插件ID
        _,err = pg.Lookup("aaaaa")
        if err == nil {
            return Bigger.Erring("[插件]获取标识失败")
        }
        msg := err.Error()
        pps := "in plugin "
        if false == strings.Contains(msg, pps) {
            return Bigger.Erring("[插件]获取标识失败")
        }
        i := strings.Index(msg, pps) + len(pps)
        key := msg[i:]
        if key == "" {
            return Bigger.Erring("[插件]获取标识失败")
        }

        bigger.plugins[key] = pg
    }


	return nil
}







//把动态加载的对象，注册到对应的模块
func (bigger *bigger) register() (*Error) {
    newbies := kernel.lastNewbies()

    for _,newbie := range newbies {
        switch newbie.branch {
        case bPLANROUTER:
            mPLAN.newbie(newbie)
        case bEVENTROUTER:
            mEVENT.newbie(newbie)
        case bQUEUEROUTER:
            mQUEUE.newbie(newbie)
        case bHTTPROUTER:
            mHTTP.newbie(newbie)
        }
    }

    return nil
}



//加载，成功后复制到插件目录
func (bigger *bigger) Plugin(file Map) (*Error) {
    
    if file["extension"] != "so" {
        return newError("无效文件")
    }

    tempFile := file["tempfile"].(string)
    pluginFile := path.Join(bigger.Config.Path.Plugin, file["filename"].(string))
    if _,err := os.Stat(pluginFile); err == nil {
        return newError("插件已经存在")
    }

    if err := bigger.load(tempFile); err != nil {
        return err
    }

    bytes,err := ioutil.ReadFile(tempFile)
    if err != nil {
        return bigger.Erred(err)
    }

    err = ioutil.WriteFile(pluginFile, bytes, 0700)
    if err != nil {
        return bigger.Erred(err)
    }

    return nil
}




//------------------------------------- driver ----------------------------
func (bigger *bigger) Erring(state string, args ...Any) (*Error) {
	return newError(state, args...)
}
func (bigger *bigger) Erred(err error) (*Error) {
	return newError(err.Error())
}
func (bigger *bigger) Status(code int, status string, text string, overrides ...bool) (*Error) {
    if len(overrides) == 0 {
        //默认不替换
        overrides = append(overrides, false)
    }
    mCONST.Status(Map{ status: code }, overrides...)
    mCONST.Lang(kDEFAULT, Map{ status: text }, overrides...)
    return newError(status)
}

func (bigger *bigger) Driver(name string, obj Any) {
    switch driver := obj.(type) {
    case MutexDriver:
        mMutex.Driver(name, driver)
    case CacheDriver:
        mCACHE.Driver(name, driver)
    case DataDriver:
        mDATA.Driver(name, driver)
    case EventDriver:
        mEVENT.Driver(name, driver)
    case FileDriver:
        mFILE.Driver(name, driver)
    case HttpDriver:
        mHTTP.Driver(name, driver)
    case LoggerDriver:
        mLOGGER.Driver(name, driver)
    case PlanDriver:
        mPLAN.Driver(name, driver)
    case QueueDriver:
        mQUEUE.Driver(name, driver)
    case SessionDriver:
        mSESSION.Driver(name, driver)
    case ViewDriver:
        mVIEW.Driver(name, driver)
    case SocketDriver:
        mSOCKET.Driver(name, driver)
    }
}



//---------------------- logger --------------------------
func (bigger *bigger) Debug(body string, args ...Any) {
    mLOGGER.Debug(body, args...)
}
func (bigger *bigger) Trace(body string, args ...Any) {
    mLOGGER.Trace(body, args...)
}
func (bigger *bigger) Info(body string, args ...Any) {
    mLOGGER.Info(body, args...)
}
func (bigger *bigger) Warning(body string, args ...Any) {
    mLOGGER.Warning(body, args...)
}
func (bigger *bigger) Error(body string, args ...Any) {
    mLOGGER.Error(body, args...)
}
//---------------------- const --------------------------
func (bigger *bigger) Mime(config Map, overrides ...bool) {
    mCONST.Mime(config, overrides...)
}
func (bigger *bigger) String(lang, name string, args ...Any) string {
    return mCONST.LangString(lang, name, args...)
}
func (bigger *bigger) Code(status string, defs ...int) int {
    return mCONST.StatusCode(status, defs...)
}
func (bigger *bigger) Regular(config Map, overrides ...bool) {
    mCONST.Regular(config, overrides...)
}
func (bigger *bigger) Express(name string, defs ...string) ([]string) {
    return mCONST.RegularExpress(name, defs...)
}
func (bigger *bigger) Match(value, regular string) bool {
    return mCONST.RegularMatch(value, regular)
}
//---------------------- mapping --------------------------
func (bigger *bigger) Type(name string, config Map, overrides ...bool) {
    mMAPPING.Type(name, config, overrides...)
}
func (bigger *bigger) Crypto(name string, config Map, overrides ...bool) {
    mMAPPING.Crypto(name, config, overrides...)
}
func (bigger *bigger) Mapping(config Map, data Map, value Map, argn bool, pass bool, ctxs ...*Context) *Error {
    return mMAPPING.Parse(config, data, value, argn, pass, ctxs...)
}
//---------------------- cache --------------------------
func (bigger *bigger) Cache(names ...string) (CacheBase) {
    return mCACHE.Base(names...)
}
//---------------------- data --------------------------
func (bigger *bigger) Data(names ...string) (DataBase) {
    return mDATA.Base(names...)
}
func (bigger *bigger) Base(name string) (*dataGroup) {
    return mDATA.newGroup(name)
}
func (bigger *bigger) Table(name string, configs ...Map) (Map) {
    return mDATA.Table(name, configs...)
}
func (bigger *bigger) View(name string, configs ...Map) (Map) {
    return mDATA.View(name, configs...)
}
func (bigger *bigger) Model(name string, configs ...Map) (Map) {
    return mDATA.Model(name, configs...)
}
func (bigger *bigger) Fields(name string, keys []string, exts ...Map) (Map) {
    return mDATA.Fields(name, keys, exts...)
}
func (bigger *bigger) Enums(name string, field string) (Map) {
    return mDATA.Enums(name, field)
}
func (bigger *bigger) Query(args ...Any) (string,[]Any,string,*Error) {
    return mDATA.Parse(args...)
}
//---------------------- file --------------------------
func (bigger *bigger) File(names ...string) (FileBase) {
    return mFILE.Base(names...)
}
func (bigger *bigger) Encode(base, name, file string) (string) {
    return mFILE.Encode(base, name, file)
}
func (bigger *bigger) Decode(code string) (*FileCoding) {
    return mFILE.Decode(code)
}
func (bigger *bigger) Assign(name string, metadata Map, bases ...string) (string,*Error) {
    return mFILE.Assign(name, metadata, bases...)
}
func (bigger *bigger) Storage(code string, reader io.Reader) (int64,*Error) {
    return mFILE.Storage(code, reader)
}
func (bigger *bigger) Download(code string) (io.ReadCloser, *FileCoding, *Error) {
    return mFILE.Download(code)
}
func (bigger *bigger) Thumbnail(code string, width,height,tttt int64) (io.ReadCloser, *FileCoding, *Error) {
	return mFILE.Thumbnail(code, width, height, tttt)
}
func (bigger *bigger) Browse(code string, name string, args Map, expires ...time.Duration) (string,*Error) {
    return mFILE.Browse(code, name, args, expires...)
}
func (bigger *bigger) Preview(code string, width,height,time int64, args Map, expires ...time.Duration) (string,*Error) {
	return mFILE.Preview(code, width, height, time, args, expires...)
}
//---------------------- plan --------------------------
func (bigger *bigger) Plan(name string, config Map, overrides ...bool) {
    if config[kACTION] != nil {
        mPLAN.Router(name, config, overrides...)
    }
    if config[kREQUEST] != nil || config[kEXECUTE] != nil || config[kRESPONSE] != nil {
        mPLAN.Filter(name, config, overrides...)
    }
    if config[kFOUND] != nil || config[kERROR] != nil || config[kFAILED] != nil || config[kDENIED] != nil {
        mPLAN.Handler(name, config, overrides...)
    }
}
// func (bigger *bigger) Planning(name string, config Map, overrides ...bool) {
//     mPLAN.Filter(name, config, overrides...)
// }
// func (bigger *bigger) Planned(name string, config Map, overrides ...bool) {
//     mPLAN.Handler(name, config, overrides...)
// }
func (bigger *bigger) Timer(name string, times []string, config Map, overrides ...bool) {
    mPLAN.Timer(name, times, overrides...)
}
func (bigger *bigger) Execute(name string, value Map, delays ...time.Duration) (*Error) {
    if len(delays) > 0 {
        return mPLAN.DeferredExecute(name, delays[0], value)
    } else {
        return mPLAN.Execute(name, value)
    }
}
//---------------------- event --------------------------
func (bigger *bigger) Event(name string, config Map, overrides ...bool) {
    if config[kACTION] != nil {
        mEVENT.Router(name, config, overrides...)
    }
    if config[kREQUEST] != nil || config[kEXECUTE] != nil || config[kRESPONSE] != nil {
        mEVENT.Filter(name, config, overrides...)
    }
    if config[kFOUND] != nil || config[kERROR] != nil || config[kFAILED] != nil || config[kDENIED] != nil {
        mEVENT.Handler(name, config, overrides...)
    }
}
// func (bigger *bigger) Eventing(name string, config Map, overrides ...bool) {
//     mEVENT.Filter(name, config, overrides...)
// }
// func (bigger *bigger) Evented(name string, config Map, overrides ...bool) {
//     mEVENT.Handler(name, config, overrides...)
// }
func (bigger *bigger) Trigger(name string, value Map, bases ...string) (*Error) {
    return mEVENT.Trigger(name, value, bases...)
}
func (bigger *bigger) SyncTrigger(name string, value Map, bases ...string) (*Error) {
    return mEVENT.SyncTrigger(name, value, bases...)
}
func (bigger *bigger) Publish(name string, value Map, bases ...string) (*Error) {
    return mEVENT.Publish(name, value, bases...)
}
func (bigger *bigger) DeferredPublish(name string, delay time.Duration, value Map, bases ...string) (*Error) {
    return mEVENT.DeferredPublish(name, delay, value, bases...)
}
//---------------------- queue --------------------------
func (bigger *bigger) Queue(name string, config Map, overrides ...bool) {
    if config[kACTION] != nil {
        mQUEUE.Router(name, config, overrides...)
    }
    if config[kREQUEST] != nil || config[kEXECUTE] != nil || config[kRESPONSE] != nil {
        mQUEUE.Filter(name, config, overrides...)
    }
    if config[kFOUND] != nil || config[kERROR] != nil || config[kFAILED] != nil || config[kDENIED] != nil {
        mQUEUE.Handler(name, config, overrides...)
    }
}
// func (bigger *bigger) Queueing(name string, config Map, overrides ...bool) {
//     mQUEUE.Filter(name, config, overrides...)
// }
// func (bigger *bigger) Queued(name string, config Map, overrides ...bool) {
//     mQUEUE.Handler(name, config, overrides...)
// }
func (bigger *bigger) Liner(name string, lines int, overrides ...bool) {
    mQUEUE.Liner(name, lines, overrides...)
}
func (bigger *bigger) Produce(name string, value Map, bases ...string) (*Error) {
    return mQUEUE.Produce(name, value, bases...)
}
func (bigger *bigger) DeferredProduce(name string, delay time.Duration, value Map, bases ...string) (*Error) {
    return mQUEUE.DeferredProduce(name, delay, value, bases...)
}
//---------------------- http --------------------------
func (bigger *bigger) Http(name string, config Map, overrides ...bool) {
    if config[kACTION] != nil {
        mHTTP.Router(name, config, overrides...)
    }
    if config[kREQUEST] != nil || config[kEXECUTE] != nil || config[kRESPONSE] != nil {
        mHTTP.Filter(name, config, overrides...)
    }
    if config[kFOUND] != nil || config[kERROR] != nil || config[kFAILED] != nil || config[kDENIED] != nil {
        mHTTP.Handler(name, config, overrides...)
    }
}
func (bigger *bigger) Site(name string, roots ...string) (*httpGroup) {
    return mHTTP.newGroup(name, roots...)
}
func (bigger *bigger) Router(name string, config Map, overrides ...bool) {
    mHTTP.Router(name, config, overrides...)
}
func (bigger *bigger) Filter(name string, config Map, overrides ...bool) {
    mHTTP.Filter(name, config, overrides...)
}
func (bigger *bigger) Handler(name string, config Map, overrides ...bool) {
    mHTTP.Handler(name, config, overrides...)
}
//---------------------- view --------------------------
func (bigger *bigger) Helper(name string, config Map, overrides ...bool) {
    mVIEW.Helper(name, config, overrides...)
}
func (bigger *bigger) Route(name string, args... Map) string {
    return bigger.url.Route(name, args...)
}
//---------------------- socket --------------------------
func (bigger *bigger) Socket(name string, config Map, overrides ...bool) {
    if config[kACTION] != nil {
        mSOCKET.Router(name, config, overrides...)
    }
    if config[kREQUEST] != nil || config[kEXECUTE] != nil || config[kRESPONSE] != nil {
        mSOCKET.Filter(name, config, overrides...)
    }
    if config[kFOUND] != nil || config[kERROR] != nil || config[kFAILED] != nil || config[kDENIED] != nil {
        mSOCKET.Handler(name, config, overrides...)
    }
}
// func (bigger *bigger) Socketing(name string, config Map, overrides ...bool) {
//     mSOCKET.Filter(name, config, overrides...)
// }
// func (bigger *bigger) Socketed(name string, config Map, overrides ...bool) {
//     mSOCKET.Handler(name, config, overrides...)
// }
//是指服务端会主动下发到客户的message/broadcast定义
//客户端发给服务端的，是叫SocketRouter
func (bigger *bigger) Command(name string, config Map, overrides ...bool) {
    mSOCKET.Command(name, config, overrides...)
}
func (bigger *bigger) Degrade(code string) (*Error) {
    return mSOCKET.Degrade(code)
}
func (bigger *bigger) Message(code, message string, value Map) (*Error) {
    return mSOCKET.Message(code, message, value)
}
func (bigger *bigger) DeferredMessage(code, message string, delay time.Duration, value Map) (*Error) {
    return mSOCKET.DeferredMessage(code, message, delay, value)
}
func (bigger *bigger) Broadcast(channel, message string, value Map, bases ...string) (*Error) {
    return mSOCKET.Broadcast(channel, message, value, bases...)
}
func (bigger *bigger) DeferredBroadcast(channel, message string, delay time.Duration, value Map, bases ...string) (*Error) {
    return mSOCKET.DeferredBroadcast(channel, message, delay, value, bases...)
}
//---------------------- logic --------------------------
func (bigger *bigger) Service(name string) (*serviceGroup) {
    return mSERVICE.newGroup(name)
}
func (bigger *bigger) Register(name string, config Map) {
	mSERVICE.Register(name, config)
}
// func (bigger *bigger) Invoke(name string, args Map) (Map,*Error) {
// 	return mSERVICE.Invoke(nil, name, nil, args)
// }
//包装服务请求结果
func (bigger *bigger) Invoke(result Map, errs ...*Error) (Map,*Error) {
    var err *Error
    if len(errs) > 0 {
        err = errs[0]
    }
	return result, err
}
//包装服务请求结果
func (bigger *bigger) Invokes(results []Map, errs ...*Error) (Map,*Error) {
    var err *Error
    if len(errs) > 0 {
        err = errs[0]
    }
	return Map{
        "items": results,
    }, err
}
//包装服务请求结果
func (bigger *bigger) Invoker(item Map, items []Map, errs ...*Error) (Map,*Error) {
    var err *Error
    if len(errs) > 0 {
        err = errs[0]
    }
    return Map{
        "item":     item,
        "items":    items,
    }, err
}
//包装服务请求结果
func (bigger *bigger) Invoking(count int64, results []Map, errs ...*Error) (Map,*Error) {
    var err *Error
    if len(errs) > 0 {
        err = errs[0]
    }
    return Map{
        "count":    count,
        "items":    results,
    }, err
}
//---------------------- mutex --------------------------
func (bigger *bigger) Lock(args ...Any) (bool) {
    return mMutex.Lock(args...)
}
func (bigger *bigger) Unlock(args ...Any) {
    mMutex.Unlock(args...)
}
//----------------------- util -------------------------


func (bigger *bigger) ToString(val Any) string {
    sv := ""
    switch v:=val.(type) {
    case string:
        sv = v
    case int:
        sv = strconv.Itoa(v)
    case int64:
        sv = strconv.FormatInt(v, 10)
    case bool:
        sv = strconv.FormatBool(v)
    case Map:
        d,e := json.Marshal(v)
        if e == nil {
            sv = string(d)
        } else {
            sv = "{}"
        }
    case []Map:
        d,e := json.Marshal(v)
        if e == nil {
            sv = string(d)
        } else {
            sv = "[]"
        }
    case []int,[]int8,[]int16,[]int32,[]int64,[]float32,[]float64,[]string,[]bool,[]Any:
        d,e := json.Marshal(v)
        if e == nil {
            sv = string(d)
        } else {
            sv = "[]"
        }
    default:
        sv = fmt.Sprintf("%v", v)
    }

    return sv
}
func (bigger *bigger) Timing(s string) (time.Duration, *Error) {
    return parseDuration(s)
}
func (bigger *bigger) Sizing(s string) (int64) {
    return parseSize(s)
}

func (bigger *bigger) Encrypt(text string) (string) {
    return bigger.textCoder.EncodeToString([]byte(text))
}
func (bigger *bigger) Decrypt(code string) (string) {
    d, e := bigger.textCoder.DecodeString(code)
	if e == nil {
		return string(d)
	}
	return ""
}
func (bigger *bigger) Encrypts(texts []string) (string) {
    text := strings.Join(texts, "\n")
    return bigger.textCoder.EncodeToString([]byte(text))
}
func (bigger *bigger) Decrypts(code string) ([]string) {
    text, e := bigger.textCoder.DecodeString(code)
	if e == nil {
		return strings.Split(string(text), "\n")
	}
	return []string{}
}
func (bigger *bigger) Enhash(digit int64, lengths ...int) (string) {
    return bigger.Enhashs([]int64{ digit }, lengths...)
}
func (bigger *bigger) Dehash(code string, lengths ...int) (int64) {
    digits := bigger.Dehashs(code, lengths...)
	if len(digits) > 0 {
		return digits[0]
	} else {
		return int64(-1)
	}
}
//因为要自定义长度，所以动态创建对象
func (bigger *bigger) Enhashs(digits []int64, lengths ...int) (string) {
    coder := bigger.digitCoder

    if len(lengths) > 0 {
        length := lengths[0]

        hd := hashid.NewData()
        hd.Alphabet = encodeDigitAlphabet
        hd.Salt = encodeDigitSalt
        if length > 0 {
            hd.MinLength = length
        }

        coder,_ = hashid.NewWithData(hd)
    }

    if coder != nil {
        code,err := coder.EncodeInt64(digits)
        if err == nil {
            return code
        }
    }

	return ""
}
//因为要自定义长度，所以动态创建对象
func (bigger *bigger) Dehashs(code string, lengths ...int) ([]int64) {
    coder := bigger.digitCoder

    if len(lengths) > 0 {
        length := lengths[0]

        hd := hashid.NewData()
        hd.Alphabet = encodeDigitAlphabet
        hd.Salt = encodeDigitSalt
        if length > 0 {
            hd.MinLength = length
        } 

        coder,_ = hashid.NewWithData(hd)
    }

    if digits,err := coder.DecodeInt64WithError(code); err == nil {
		return digits
	}

	return []int64{}
}


func (bigger *bigger) Serial() int64 {
    return bigger.fastid.NextID()
}
func (bigger *bigger) Unique(prefixs ...string) string {
    id := bigger.fastid.NextID()
    if len(prefixs) > 0 {
        return fmt.Sprintf("%s%d", prefixs[0], id)
    } else {
        return bigger.Enhash(id)
    }
}

//生成文档
func (bigger *bigger) Document(sites ...string) Map {
    doc := Map{}



    return doc
}

func (bigger *bigger) Caller(skips ...int) (string,int,string,bool) {
    skip := 2
    if len(skips) > 0 {
        skip = skips[0]
    }

    funcName := "null"
    pc, file, line, ok := runtime.Caller(skip)
	if ok {
        funcName = runtime.FuncForPC(pc).Name()
	} else {
        file = "null"
		line = 0	
    }
    return file,line,funcName,ok
}