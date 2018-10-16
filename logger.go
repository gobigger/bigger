package bigger

import (
	"strings"
	"time"
    "fmt"
)


type (
    //日志驱动
    LoggerDriver interface {
        Connect(config LoggerConfig) (LoggerConnect,*Error)
    }
    //日志连接
    LoggerConnect interface {
        //打开、关闭
        Open() *Error
        Health() (*LoggerHealth,*Error)
        Close() *Error

        Debug(string,...Any)
        Trace(string,...Any)
        Info(string,...Any)
        Warning(string,...Any)
        Error(string,...Any)
    }

    LoggerHealth struct {
        Workload    int64
    }

    loggerModule struct {
        driver	    coreBranch
        
        //日志配置，日志连接
        config      LoggerConfig
        connect     LoggerConnect
    }
)


//注册日志驱动
func (module *loggerModule) Driver(name string, driver LoggerDriver, overrides ...bool) {
    if driver == nil {
        panic("[日志]驱动不可为空")
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





func (module *loggerModule) connecting(config LoggerConfig) (LoggerConnect,*Error) {
    if driver,ok := module.driver.chunk(config.Driver).(LoggerDriver); ok {
        return driver.Connect(config)
    }
    panic("[日志]不支持的驱动：" + config.Driver)
}


//初始化
func (module *loggerModule) initing() {

    //连接会话
    config := Bigger.Config.Logger
    connect,err := module.connecting(config)
    if err != nil {
        panic("[日志]连接失败：" + err.Error())
    }
    
    //打开连接
    err = connect.Open()
    if err != nil {
        panic("[日志]打开失败：" + err.Error())
    }

    //保存连接
    module.config = config
    module.connect = connect

}

//退出
func (module *loggerModule) exiting() {
    if module.connect != nil {
        module.connect.Close()
    }
}





//调试
func (module *loggerModule) Debug(body string, args ...Any) {
    if module.connect != nil {
        go module.connect.Debug(body, args...)
    } else {
        args2 := []Any{
            time.Now().Format("2006-01-02 15:04:05.999"),
        }

        if len(args)>0 && strings.Count(body, "%")==len(args) {
            args2 = append(args2, fmt.Sprintf(body, args...))
        } else {
            args2 = append(args2, body)
            args2 = append(args2, args...)
        }

        go fmt.Println(args2...)
    }
}
//信息
func (module *loggerModule) Trace(body string, args ...Any) {
    if module.connect != nil {
        go module.connect.Trace(body, args...)
    } else {
        args2 := []Any{
            time.Now().Format("2006/01/02 15:04:05"),
        }

        if len(args)>0 && strings.Count(body, "%")==len(args) {
            args2 = append(args2, fmt.Sprintf(body, args...))
        } else {
            args2 = append(args2, body)
            args2 = append(args2, args...)
        }
        
        go fmt.Println(args2...)
    }
}
//信息
func (module *loggerModule) Info(body string, args ...Any) {
    if module.connect != nil {
        go module.connect.Info(body, args...)
    } else {
        args2 := []Any{
            time.Now().Format("2006/01/02 15:04:05"),
        }

        if len(args)>0 && strings.Count(body, "%")==len(args) {
            args2 = append(args2, fmt.Sprintf(body, args...))
        } else {
            args2 = append(args2, body)
            args2 = append(args2, args...)
        }
        
        go fmt.Println(args2...)
    }
}
//警告
func (module *loggerModule) Warning(body string, args ...Any) {
    if module.connect != nil {
        go module.connect.Warning(body, args...)
    } else {
        args2 := []Any{
            time.Now().Format("2006/01/02 15:04:05"),
        }

        if len(args)>0 && strings.Count(body, "%")==len(args) {
            args2 = append(args2, fmt.Sprintf(body, args...))
        } else {
            args2 = append(args2, body)
            args2 = append(args2, args...)
        }
        
        go fmt.Println(args2...)
    }
}
//错误
func (module *loggerModule) Error(body string, args ...Any) {
    if module.connect != nil {
        go module.connect.Error(body, args...)
    } else {
        args2 := []Any{
            time.Now().Format("2006/01/02 15:04:05"),
        }

        if len(args)>0 && strings.Count(body, "%")==len(args) {
            args2 = append(args2, fmt.Sprintf(body, args...))
        } else {
            args2 = append(args2, body)
            args2 = append(args2, args...)
        }
        
        go fmt.Println(args2...)
    }
}

