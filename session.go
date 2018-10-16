package bigger

import (
    "time"
)



type (
    //会话驱动
    SessionDriver interface {
        Connect(config SessionConfig) (SessionConnect,*Error)
    }
    //会话连接
    SessionConnect interface {
        //打开、关闭
        Open() *Error
        Health() (*SessionHealth,*Error)
        Close() *Error

        Read(id string) (Map,*Error)
        Write(id string, value Map, expires ...time.Duration) (*Error)
        Delete(id string) (*Error)
    }
    

    SessionHealth struct {
        Workload    int64
    }
    
    sessionModule struct {
        driver     coreBranch
        
        //会话配置，会话连接
        config      SessionConfig
        connect     SessionConnect
    }
)


//注册会话驱动
func (module *sessionModule) Driver(name string, driver SessionDriver, overrides ...bool) {
    if driver == nil {
        panic("[会话]驱动不可为空")
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


func (module *sessionModule) connecting(config SessionConfig) (SessionConnect,*Error) {
    if driver,ok := module.driver.chunk(config.Driver).(SessionDriver); ok {
        return driver.Connect(config)
    }
    panic("[会话]不支持的驱动：" + config.Driver)
}


//初始化
func (module *sessionModule) initing() {

    //连接会话
    config := Bigger.Config.Session
    connect,err := module.connecting(config)
    if err != nil {
        panic("[会话]连接失败：" + err.Error())
    }
    
    //打开连接
    err = connect.Open()
    if err != nil {
        panic("[会话]打开失败：" + err.Error())
    }

    //保存连接
    module.config = config
    module.connect = connect

}

//退出
func (module *sessionModule) exiting() {
    if module.connect != nil {
        module.connect.Close()
    }
}





func (module *sessionModule) Read(id string) (Map,*Error) {
    if module.connect == nil {
        return nil, Bigger.Erring("[会话]无效连接")
    }
    return module.connect.Read(id)
}


func (module *sessionModule) Write(id string, value Map, expires ...time.Duration) (*Error) {
    if module.connect == nil {
        return Bigger.Erring("[会话]无效连接")
    }
    return module.connect.Write(id, value, expires...)
}


func (module *sessionModule) Delete(id string) (*Error) {
    if module.connect == nil {
        return Bigger.Erring("[会话]无效连接")
    }
    return module.connect.Delete(id)
}







