package bigger

import (
)



type (
    //视图驱动
    ViewDriver interface {
        Connect(config ViewConfig) (ViewConnect,*Error)
    }
    //视图连接
    ViewConnect interface {
        //打开、关闭
        Open() *Error
        Health() (*ViewHealth, *Error)
        Close() *Error

        Parse(*Context, ViewBody) (string,*Error)
	}


	ViewBody struct {
        Root    string
        Shared  string
		View	string
		Data	Map
		Helpers	Map
    }
    

    ViewHealth struct {
        Workload    int64
    }

    viewModule struct {
        driver  coreBranch
        helper  coreBranch

        
        //视图配置，视图连接
        config      ViewConfig
        connect     ViewConnect
    }
)


//注册视图驱动
func (module *viewModule) Driver(name string, driver ViewDriver, overrides ...bool) {
    if driver == nil {
        panic("[视图]驱动不可为空")
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

func (module *viewModule) Helper(name string, config Map, overrides ...bool) {

    override := true
    if len(overrides) > 0 {
        override = overrides[0]
    }

    if override {
        module.helper.chunking(name, config)
    } else {
        if module.helper.chunkdata(name) == nil {
            module.helper.chunking(name, config)
        }
    }
    
}

func (module *viewModule) parse(ctx *Context, body ViewBody) (string,*Error) {
    if module.connect == nil {
        return "", Bigger.Erring("[会话]无效连接")
    }
    return module.connect.Parse(ctx, body)
}







func (module *viewModule) helperActions() (Map) {
    actions := Map{}
    
	for _,vv := range module.helper.chunks() {
		if config,ok := vv.data.(Map); ok {

            if action,ok := config[kACTION]; ok {
                actions[vv.name] = action
            }
		}
    }
    
    return actions
}








func (module *viewModule) connecting(config ViewConfig) (ViewConnect,*Error) {
    if driver,ok := module.driver.chunkdata(config.Driver).(ViewDriver); ok {
        return driver.Connect(config)
    }
    panic("[视图]不支持的驱动：" + config.Driver)
}


//初始化
func (module *viewModule) initing() {

    //连接视图
    config := Bigger.Config.View
    connect,err := module.connecting(config)
    if err != nil {
        panic("[视图]连接失败：" + err.Error())
    }
    
    //打开连接
    err = connect.Open()
    if err != nil {
        panic("[视图]打开失败：" + err.Error())
    }

    //保存连接
    module.config = config
    module.connect = connect

}

//退出
func (module *viewModule) exiting() {
    if module.connect != nil {
        module.connect.Close()
    }
}


