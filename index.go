package bigger


import (
	"time"
    "plugin"
    "os"
    "fmt"
    "encoding/base64"
    "github.com/yatlabs/bigger/fastid"
    "github.com/yatlabs/bigger/hashid"
    "github.com/yatlabs/bigger/toml"
)

var (
    kernel		*coreKernel
    Bigger      *bigger

    mLOGGER      *loggerModule

	mCONST   	*constModule
    mMAPPING	*mappingModule
    
    mMutex      *mutexModule
    mSESSION    *sessionModule

    mCACHE      *cacheModule
    mDATA       *dataModule
    mFILE       *fileModule

    mPLAN       *planModule
    mEVENT      *eventModule
    mQUEUE      *queueModule

    mHTTP       *httpModule
    mSOCKET     *socketModule
    mVIEW       *viewModule

    mSERVICE    *serviceModule
)


func init() {
	initCore()
    initConfig()

    initBigger()
    initLogger()

	initConst()
    initMapping()
    
    initMutex()
    initSession()

    initCache()
    initData()
    initFile()

    initPlan()
    initEvent()
    initQueue()

    initHttp()
    initSocket()
    initView()

    initService()

    Bigger.ready()
}

func initCore() {
	kernel = &coreKernel{
        blocks:	make(map[string]*coreBlock, 0),
        indexs: make([]*coreBlock, 0),
        newbies: make([]coreNewbie, 0),
    }
    
	Bigger = &bigger{
        Mode:           Developing,
        plugins:        map[string]*plugin.Plugin{},
        hosts:          make(map[string]string),
        url:            &contextUrl{},
    }
}





//加载配置
func initConfig() {
	config,err := loadConfig()
    if err != nil {
        panic("加载配置文件失败：" + err.Error())
    }

    switch config.Mode {
    case "d","dev","develop","development","developing":
        Bigger.Mode = Developing
    case "t","test","testing":
        Bigger.Mode = Testing
    case "p","pro","prod","product","production":
        Bigger.Mode = Production
    default:
        Bigger.Mode = Developing
    }

    if config.Charset == "" {
        config.Charset = vUTF8
    }

    if config.Node.Id == 0 {
        config.Node.Id = 1
    }
    if config.Node.Name == "" {
        config.Node.Name = fmt.Sprintf("%s%d", config.Name, config.Node.Id)
    }
    if config.Node.Port <= 0 || config.Node.Port > 65535 {
        config.Node.Port = 80
    }


    Bigger.Id = config.Node.Id
    Bigger.Name = config.Node.Name

    if config.serial.Start == "" {
        t,e := time.Parse("2006-01-02", config.serial.Start)
        if e == nil {
            config.serial.begin = t.UnixNano()
        } else {
            config.serial.begin = time.Date(2018, 10, 1, 0, 0, 0, 0, time.Local).UnixNano()
        }
    } else {
        config.serial.begin = time.Date(2018, 10, 1, 0, 0, 0, 0, time.Local).UnixNano()
    }
    if config.serial.Time <= 0 {
        config.serial.Time = 43
    }
    if config.serial.Node <= 0 {
        config.serial.Node = 7
    }
    if config.serial.Seq <= 0 {
        config.serial.Seq = 13
    }


    //默认logger驱动
    if config.Logger.Driver == "" {
        config.Logger.Driver = kDEFAULT
    }
    if config.Logger.Flag == "" {
        config.Logger.Flag = config.Node.Name
    }

    //默认plan驱动
    if config.Plan.Driver == "" {
        config.Plan.Driver = kDEFAULT
    }
    if config.Plan.Prefix == "" {
        config.Plan.Prefix = config.Name
    }

    //默认锁驱动
    if config.Mutex.Driver == "" {
        config.Mutex.Driver = kDEFAULT
    }
    if config.Mutex.Prefix == "" {
        config.Mutex.Prefix = config.Name
    }

    //默认session驱动
    if config.Session.Driver == "" {
        config.Session.Driver = kDEFAULT
    }
    if config.Session.Prefix == "" {
        config.Session.Prefix = config.Name
    }

    //默认HTTP驱动
    if config.Http.Driver == "" {
        config.Http.Driver = kDEFAULT
    }
    if config.Http.Port <= 0 || config.Http.Port > 65535 {
        config.Http.Port = config.Node.Port
    }

    //默认view驱动
    if config.View.Driver == "" {
        config.View.Driver = kDEFAULT
    }



    //默认lang驱动
    if config.Lang == nil {
        config.Lang = map[string]langConfig {
            "zh-CN": langConfig {
                Name: "简体中文",
                Accepts: []string{ "zh", "cn", "zh-CN", "zhCN" },
            },
            "zh-TW": langConfig {
                Name: "繁體中文",
                Accepts: []string{ "zh-TW", "zhTW", "tw" },
            },
            "en-US": langConfig {
                Name: "English",
                Accepts: []string{ "en", "en-US" },
            },
        }
    }


    //默认file驱动
    if config.File == nil {
        config.File = map[string]FileConfig {
            kDEFAULT: FileConfig {
                Driver: kDEFAULT, Setting: Map{
                    "storage": "store/storage", "thumbnail": "store/thumbnail",
                },
            },
        }
    }

    //默认cache驱动
    if config.Cache == nil {
        config.Cache = map[string]CacheConfig {
            kDEFAULT: CacheConfig {
                Driver: kDEFAULT,
                Prefix: config.Name,
            },
        }
    } else {
        for k,v := range config.Cache {
            if v.Prefix == "" {
                v.Prefix = config.Name
            }
            config.Cache[k] = v
        }
    }

    //默认event驱动
    if config.Event == nil {
        config.Event = map[string]EventConfig{
            kDEFAULT: EventConfig {
                Driver: kDEFAULT,
                Prefix: config.Name,
            },
        }
    } else {
        for k,v := range config.Event {
            if v.Prefix == "" {
                v.Prefix = config.Name
            }
            config.Event[k] = v
        }
    }

    //默认queue驱动
    if config.Queue == nil {
        config.Queue = map[string]QueueConfig{
            kDEFAULT: QueueConfig {
                Driver: kDEFAULT,
                Prefix: config.Name,
            },
        }
    } else {
        for k,v := range config.Queue {
            if v.Prefix == "" {
                v.Prefix = config.Name
            }
            config.Queue[k] = v
        }
    }


    //默认socket驱动
    if config.Socket == nil {
        config.Socket = map[string]SocketConfig{
            kDEFAULT: SocketConfig {
                Driver: kDEFAULT,
                Prefix: config.Name,
            },
        }
    } else {
        for k,v := range config.Socket {
            if v.Prefix == "" {
                v.Prefix = config.Name
            }
            config.Socket[k] = v
        }
    }




    //http默认驱动
    //此处改没有用，因为http定义不是指针
    for k,v := range config.Site {
        if v.Charset == "" {
            v.Charset = config.Charset
        }
        if v.Domain == "" {
            v.Domain = config.Domain
        }
        if v.Cookie == "" {
            v.Cookie = config.Name
        }
        if v.Expiry == "" {
            v.Expiry = config.Http.Expiry
        }
        if v.MaxAge == "" {
            v.MaxAge = config.Http.MaxAge
        }
        if v.Hosts == nil {
            v.Hosts = []string{}
        }
        if v.Host != "" {
            v.Hosts = append(v.Hosts, v.Host)
            v.Weights = append(v.Weights, 1)
        }
        if v.Weights==nil || len(v.Weights) == 0 {
            v.Weights = []int{}
            for range v.Hosts {
                v.Weights = append(v.Weights, 1)
            }
        }

        //还没有设置域名，自动来一波
        if len(v.Hosts) == 0 && v.Domain != "" {
            v.Hosts = append(v.Hosts, k+"."+v.Domain)
            v.Weights = append(v.Weights, 1)
        }

        //记录http的所有域名
        for _,host := range v.Hosts {
            Bigger.hosts[host] = k
        }

        config.Site[k] = v
    }

    //隐藏的空站点，不接域名
    config.Site[""] = SiteConfig{}



    if config.Path.Node == "" {
        config.Path.Node = "node"
    }
    if config.Path.Lang == "" {
        config.Path.Lang = "asset/langs"
    }
    if config.Path.Plugin == "" {
        config.Path.Plugin = "plugin"
    }
    if config.Path.View == "" {
        config.Path.View = "asset/views"
    }
    if config.Path.Static == "" {
        config.Path.Static = "asset/statics"
    }
    if config.Path.Upload == "" {
        config.Path.Upload = os.TempDir()
    }
    if config.Path.Shared == "" {
        config.Path.Shared = "shared"
    }



    //设置
    setting := make(Map)
    for k,v := range config.setting {
        setting[k] = v
    }

    //几个hash的复制
    if vv,ok := setting["encodeTextAlphabet"].(string); ok && vv != "" {
        encodeTextAlphabet = vv
    }
    if vv,ok := setting["encodeDigitAlphabet"].(string); ok && vv != "" {
        encodeDigitAlphabet = vv
    }
    if vv,ok := setting["encodeDigitSalt"].(string); ok && vv != "" {
        encodeDigitSalt = vv
    }
    if vv,ok := setting["encodeDigitLength"].(int64); ok {
        encodeDigitLength = int(vv)
    }

    Bigger.Config = config
    Bigger.Setting = setting

}
func loadConfig() (*configConfig,error){
    cfgFile := "config.toml"
    if len(os.Args) >= 2 {
        cfgFile = os.Args[1]
    }
    var config configConfig
    _,err := toml.DecodeFile(cfgFile, &config)
    if err != nil {
        return nil,err
    }

    return &config,nil
}




func initBigger() {

    Bigger.fastid = fastid.NewFastIDWithConfig(Bigger.Config.serial.Time, Bigger.Config.serial.Node, Bigger.Config.serial.Seq, Bigger.Config.serial.begin, Bigger.Id)

    Bigger.textCoder = base64.NewEncoding(encodeTextAlphabet)

	hd := hashid.NewData()
	hd.Alphabet = encodeDigitAlphabet
	hd.Salt = encodeDigitSalt
    if encodeDigitLength > 0 {
        hd.MinLength = encodeDigitLength
    } 
    coder,err := hashid.NewWithData(hd)
    if err == nil {
        Bigger.digitCoder = coder
    }

}



func initLogger() {
    mLOGGER = &loggerModule{
        driver:  coreBranch{kernel, bLoggerDriver},
    }
}





//初始化Const模块
func initConst() {
    mCONST = &constModule{
        mime:   coreBranch{kernel, bCONSTMIME},
        status:  coreBranch{kernel, bCONSTSTATUS},
        regular: coreBranch{kernel, bCONSTREGULAR},
        lang: coreBranch{kernel, bCONSTLANG},
    }

    //加载语言包
    strs,err := loadLang(fmt.Sprintf("%v/%v.toml", Bigger.Config.Path.Lang, kDEFAULT))
    if err == nil {
        mCONST.Lang(kDEFAULT, strs)
    }
    for lang,_ := range Bigger.Config.Lang {
        strs,err := loadLang(fmt.Sprintf("%v/%v.toml", Bigger.Config.Path.Lang, lang))
        if err == nil {
            mCONST.Lang(lang, strs)
        }
    }
}
func loadLang(file string) (Map,error) {
    var config Map
    _,err := toml.DecodeFile(file, &config)
    if err != nil {
        return nil,err
    }
    return config,nil
}




//初始化Mapping分支
func initMapping() {
    mMAPPING = &mappingModule{
        tttt:   coreBranch{kernel, bMAPPINGTYPE},
        crypto:  coreBranch{kernel, bMAPPINGCRYPTO},
    }
}




func initMutex() {
    mMutex = &mutexModule{
        driver:  coreBranch{kernel, bMutexDriver},
    }
}
func initSession() {
    mSESSION = &sessionModule{
        driver:  coreBranch{kernel, bSessionDriver},
    }
}




func initCache() {
    mCACHE = &cacheModule{
        driver:  coreBranch{kernel, bCacheDriver},
        connects: map[string]CacheConnect{},
    }
}




func initData() {
    mDATA = &dataModule{
        driver:  coreBranch{kernel, bDataDriver},
        table:  coreBranch{kernel, bDATATABLE},
        view:  coreBranch{kernel, bDATAVIEW},
        model:  coreBranch{kernel, bDATAMODEL},
        connects: map[string]DataConnect{},
    }
}



func initFile() {
    mFILE = &fileModule{
        driver:  coreBranch{kernel, bFileDriver},
        connects: map[string]FileConnect{},
    }
}

func initPlan() {
    mPLAN = &planModule{
        driver:      coreBranch{kernel, bPlanDriver},
        router:    coreBranch{kernel, bPLANROUTER},
        filter:    coreBranch{kernel, bPLANFILTER},
        handler:   coreBranch{kernel, bPLANHANDLER},
    }
}




func initEvent() {
    mEVENT = &eventModule{
        driver:      coreBranch{kernel, bEventDriver},
        router:     coreBranch{kernel, bEVENTROUTER},
        filter:     coreBranch{kernel, bEVENTFILTER},
        handler:    coreBranch{kernel, bEVENTHANDLER},
        connects:   map[string]EventConnect{},
    }
}









func initQueue() {
    mQUEUE = &queueModule{
        driver:      coreBranch{kernel, bQueueDriver},
        router:     coreBranch{kernel, bQUEUEROUTER},
        filter:     coreBranch{kernel, bQUEUEFILTER},
        handler:    coreBranch{kernel, bQUEUEHANDLER},
        connects:   map[string]QueueConnect{},
    }
}



func initHttp() {
    mHTTP = &httpModule{
        driver:       coreBranch{kernel, bHttpDriver},
        router:     coreBranch{kernel, bHTTPROUTER},
        filter:     coreBranch{kernel, bHTTPFILTER},
        handler:    coreBranch{kernel, bHTTPHANDLER},
    }
}



func initSocket() {
    mSOCKET = &socketModule{
        driver:     coreBranch{kernel, bSocketDriver},
        router:     coreBranch{kernel, bSOCKETROUTER},
        filter:     coreBranch{kernel, bSOCKETFILTER},
        handler:    coreBranch{kernel, bSOCKETHANDLER},
        command:    coreBranch{kernel, bSocketCommand},
        connects:   map[string]SocketConnect{},
    }
}



func initView() {
    mVIEW = &viewModule{
        driver:       coreBranch{kernel, bVIEW},
        helper:     coreBranch{kernel, bVIEWHELPER},
    }
}



func initService() {
    mSERVICE = &serviceModule{
        service:      coreBranch{kernel, bSERVICE},
    }
}

