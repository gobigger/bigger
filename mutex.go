package bigger

import (
    "strings"
)



type (
    //互斥驱动
    MutexDriver interface {
        Connect(config MutexConfig) (MutexConnect,*Error)
    }
    //互斥连接
    MutexConnect interface {
        //打开、关闭
        Open() *Error
        Health() (*MutexHealth,*Error)
        Close() *Error

        Lock(key string) (bool)
        Unlock(key string) (*Error)
	}
    

    MutexHealth struct {
        Workload    int64
    }

    mutexModule struct {
        driver      coreBranch
        
        //互斥配置，互斥连接
        config      MutexConfig
        connect     MutexConnect
    }
)


func (module *mutexModule) Driver(name string, driver MutexDriver, overrides ...bool) {
    if driver == nil {
        panic("[互斥]驱动不可为空")
    }
    
    override := true
    if len(overrides) > 0 {
        override = overrides[0]
    }

    if override {
        module.driver.chunking(name, driver)
    } else {
        if module.driver.chunk(name) == nil {
            module.driver.chunking(name, driver)
        }
    }
}


func (module *mutexModule) connecting(config MutexConfig) (MutexConnect,*Error) {
    if driver,ok := module.driver.chunk(config.Driver).(MutexDriver); ok {
        return driver.Connect(config)
    }
    panic("[互斥]不支持的驱动：" + config.Driver)
}


//初始化
func (module *mutexModule) initing() {

    //连接互斥
    config := Bigger.Config.Mutex
    connect,err := module.connecting(config)
    if err != nil {
        panic("[互斥]连接失败：" + err.Error())
    }
    
    //打开连接
    err = connect.Open()
    if err != nil {
        panic("[互斥]打开失败：" + err.Error())
    }

    //保存连接
    module.config = config
    module.connect = connect

}

//退出
func (module *mutexModule) exiting() {
    if module.connect != nil {
        module.connect.Close()
    }
}


func (module *mutexModule) keying(args ...Any) string {
    keys := []string{}
    for _,v := range args {
        keys = append(keys, Bigger.ToString(v))
    }
    return strings.Join(keys, "_")
}

func (module *mutexModule) Lock(args ...Any) (bool) {
    if module.connect == nil {
        return false
    }
    key := module.keying(args...)
    return module.connect.Lock(key)
}
func (module *mutexModule) Unlock(args ...Any) (*Error) {
    if module.connect == nil {
        return Bigger.Erring("[互斥]无效连接")
    }
    key := module.keying(args...)
    return module.connect.Unlock(key)
}

