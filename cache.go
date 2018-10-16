package bigger



import (
	"time"
)




type (
    CacheDriver interface {
        Connect(name string, config CacheConfig) (CacheConnect,*Error)
    }
    CacheConnect interface {
        Open() (*Error)
        Health() (*CacheHealth,*Error)
		Close() (*Error)
		
        Base() (CacheBase)
    }
    CacheBase interface {
        Close() *Error
		Erred() *Error

        Read(string) (Any)
        Write(key string, val Any, exps ...time.Duration)
        Delete(key string)
        Serial(key string,nums ...int64) (int64)
        Keys(prefixs ...string) ([]string)
        Clear(prefixs ...string)
	}
	
    CacheHealth struct {
        Workload    int64
    }
    cacheModule struct {
        driver     coreBranch
        
        //缓存配置，缓存连接
		config      CacheConfig
		connects    map[string]CacheConnect
    }
)



func (module *cacheModule) Driver(name string, driver CacheDriver, overrides ...bool) {
    if driver == nil {
        panic("[缓存]驱动不可为空")
    }
    
    override := true
    if len(overrides) > 0 {
        override = overrides[0]
    }

    if override {
        module.driver.chunking(name, driver)
    } else {
        if module.driver.chunkdata(name) == nil {
            module.driver.chunking(name, driver)
        }
    }
}


func (module *cacheModule) connecting(name string, config CacheConfig) (CacheConnect, *Error) {
    if driver,ok := module.driver.chunkdata(config.Driver).(CacheDriver); ok {
        return driver.Connect(name, config)
    }
    panic("[缓存]不支持的驱动：" + config.Driver)
}





//初始化
func (module *cacheModule) initing() {

    for name,config := range Bigger.Config.Cache {

		//连接
		connect,err := module.connecting(name, config)
		if err != nil {
			panic("[缓存]连接失败：" + err.Error())
		}

		//打开连接
		err = connect.Open()
		if err != nil {
			panic("[缓存]打开失败：" + err.Error())
		}

		module.connects[name] = connect
	}
}

//退出
func (module *cacheModule) exiting() {
    for _,connect := range module.connects {
        connect.Close()
    }
}


//返回缓存Base对象
func (module *cacheModule)  Base(names ...string) (CacheBase) {
    name := kDEFAULT
	if len(names) > 0 {
		name = names[0]
	} else {
		for key,_ := range module.connects {
			name = key
			break
		}
    }

    if connect,ok := module.connects[name]; ok {
        return connect.Base()
    }

    panic("[缓存]无效缓存连接")
}

