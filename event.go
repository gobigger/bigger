package bigger

import (
	"time"
	"github.com/yatlabs/bigger/hashring"
)




// event driver begin
type (
	//事件驱动
	EventDriver interface {
		Connect(string,EventConfig) (EventConnect,*Error)
	}

	//事件连接
	EventConnect interface {
		Open() *Error
        Health() (*EventHealth,*Error)
		Close() *Error

		Accept(EventHandler) *Error
		Register(name string, config EventRegister) *Error
		
		Start() *Error

		Trigger(name string, value Map) *Error
		SyncTrigger(name string, value Map) *Error
		Publish(name string, value Map) *Error
		DeferredPublish(name string, delay time.Duration, value Map) *Error
	}



	//事件处理器
	EventHandler func(*EventRequest, EventResponse)

	//事件请求实体
	EventRegister struct {
	}

	//事件请求实体
	EventRequest struct {
		Id		string
		Name	string
		Value	Map
	}
	//事件响应接口
	EventResponse interface {
		//完成
		Finish(*EventRequest) *Error
		//重新开始
		Delay(*EventRequest, time.Duration) *Error
	}

    EventHealth struct {
        Workload    int64
    }
)



type (
    eventModule struct {
		driver		coreBranch
		router		coreBranch
		filter		coreBranch
		handler		coreBranch
        
        //缓存配置，缓存连接
		connects    map[string]EventConnect
		hashring	*hashring.HashRing
    }
)





func (module *eventModule) Driver(name string, driver EventDriver, overrides ...bool) {
    if driver == nil {
        panic("[事件]驱动不可为空")
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




func (module *eventModule) Router(name string, config Map, overrides ...bool) {
    override := true
    if len(overrides) > 0 {
        override = overrides[0]
    }

    if override {
		module.router.chunking(name, config)
    } else {
        if module.router.chunk(name) == nil {
			module.router.chunking(name, config)
        }
    }
}
func (module *eventModule) Filter(name string, config Map, overrides ...bool) {
    override := true
    if len(overrides) > 0 {
        override = overrides[0]
    }

    if override {
		module.filter.chunking(name, config)
    } else {
        if module.filter.chunk(name) == nil {
			module.filter.chunking(name, config)
        }
    }
}
func (module *eventModule) Handler(name string, config Map, overrides ...bool) {
    override := true
    if len(overrides) > 0 {
        override = overrides[0]
    }

    if override {
		module.handler.chunking(name, config)
    } else {
        if module.handler.chunk(name) == nil {
			module.handler.chunking(name, config)
        }
    }
}


//触发器只在当前进程有效
func (module *eventModule) Trigger(name string, value Map, bases ...string) (*Error) {
	base := kDEFAULT
	if len(bases) > 0 {
		base = bases[0]
	} else if node := module.hashring.Locate(name); node != "" {
		//按权重随机分散文件存储
		base = node
	} else {
		for n,_ := range module.connects {
			base = n
			break
		}
	}

	if value == nil {
		value = Map{}
	}

	if conn,ok := module.connects[base]; ok {
		return conn.Trigger(name, value)
	}

	return Bigger.Erring("[事件]无效连接")
}
func (module *eventModule) SyncTrigger(name string, value Map, bases ...string) (*Error) {
	base := kDEFAULT
	if len(bases) > 0 {
		base = bases[0]
	} else if node := module.hashring.Locate(name); node != "" {
		//按权重随机分散文件存储
		base = node
	} else {
		for n,_ := range module.connects {
			base = n
			break
		}
	}

	if value == nil {
		value = Map{}
	}

	if conn,ok := module.connects[base]; ok {
		return conn.SyncTrigger(name, value)
	}

	return Bigger.Erring("[事件]无效连接")
}
func (module *eventModule) Publish(name string, value Map, bases ...string) (*Error) {
	base := kDEFAULT
	if len(bases) > 0 {
		base = bases[0]
	} else if node := module.hashring.Locate(name); node != "" {
		//按权重随机分散文件存储
		base = node
	} else {
		for n,_ := range module.connects {
			base = n
			break
		}
	}

	if value == nil {
		value = Map{}
	}

	if conn,ok := module.connects[base]; ok {
		return conn.Publish(name, value)
	}

	return Bigger.Erring("[事件]无效连接")
}
func (module *eventModule) DeferredPublish(name string, delay time.Duration, value Map, bases ...string) *Error {
	base := kDEFAULT
	if len(bases) > 0 {
		base = bases[0]
	} else if node := module.hashring.Locate(name); node != "" {
		//按权重随机分散文件存储
		base = node
	} else {
		for n,_ := range module.connects {
			base = n
			break
		}
	}

	if conn,ok := module.connects[base]; ok {
		return conn.DeferredPublish(name, delay, value)
	}

	return Bigger.Erring("[事件]无效连接")
}




func (module *eventModule) connecting(name string, config EventConfig) (EventConnect,*Error) {
    if driver,ok := module.driver.chunk(config.Driver).(EventDriver); ok {
        return driver.Connect(name, config)
    }
    panic("[事件]不支持的驱动：" + config.Driver)
}
func (module *eventModule) initing() {

	weights := make(map[string]int)
    for name,config := range Bigger.Config.Event {
		if config.Weight > 0 {
			weights[name] = config.Weight
		}

		connect,err := module.connecting(name, config)
		if err != nil {
			panic("[事件]连接失败：" + err.Error())
		}
		err = connect.Open()
		if err != nil {
			panic("[事件]打开失败：" + err.Error())
		}

		//所有的也开始订阅者
		err = connect.Accept(module.serve)
		if err != nil {
			panic("[事件]注册回调失败：" + err.Error())
		}

		//注册事件
		routers := module.router.chunks()
		for name,val := range routers {
			if config,ok := val.(Map); ok {
				
				regis := module.registering(config)

				err := connect.Register(name, regis)
				if err != nil {
					panic("[事件]注册订阅失败：" + err.Error())
				}
			}
		}

		err = connect.Start()
		if err != nil {
			panic("[事件]订阅失败：" + err.Error())
		}

		module.connects[name] = connect
	}
	
	module.hashring = hashring.New(weights)
}

func (module *eventModule) registering(config Map) (EventRegister) {
	return EventRegister{}
}

func (module *eventModule) exiting() {
    for _,connect := range module.connects {
        connect.Close()
    }
}


//动态注册新来的
func (module *eventModule) newbie(newbie coreNewbie) (*Error) {
	if newbie.branch != bEVENTROUTER {
		return nil
	}

	//所有库
	for _,connect := range module.connects {
		if config,ok := module.router.chunk(newbie.block).(Map); ok {
			name := newbie.block
			regis := module.registering(config)
			err := connect.Register(name, regis)
			if err != nil {
				return err
			}
		}
	}

	return nil
}







func (module *eventModule) requestFilterActions() ([]Funcing) {
	return module.filter.funcings(kREQUEST)
}
func (module *eventModule) executeFilterActions() ([]Funcing) {
	return module.filter.funcings(kEXECUTE)
}
func (module *eventModule) responseFilterActions() ([]Funcing) {
	return module.filter.funcings(kRESPONSE)
}


func (module *eventModule) foundHandlerActions() ([]Funcing) {
	return module.handler.funcings(kFOUND)
}
func (module *eventModule) errorHandlerActions() ([]Funcing) {
	return module.handler.funcings(kERROR)
}
func (module *eventModule) failedHandlerActions() ([]Funcing) {
	return module.handler.funcings(kFAILED)
}
func (module *eventModule) deniedHandlerActions() ([]Funcing) {
	return module.handler.funcings(kDENIED)
}





//事件Event  请求开始
func (module *eventModule) serve(req *EventRequest, res EventResponse) {
	ctx := newEventContext(req, res)

	if config,ok := module.router.chunk(ctx.Name).(Map); ok {
		ctx.Config = config
	}
	
	//request拦截器，加入调用列表
	requestFilters := module.requestFilterActions()
	ctx.next(requestFilters...)
	
	ctx.next(module.request)
	ctx.next(module.execute)
	ctx.Next()
}


//事件响应
//先执行下面
//再清理执行线
func (module *eventModule) request(ctx *Context) {

	if ctx.Id == "" {
		ctx.Id = Bigger.Unique()
	}

	//请求的一开始，主要是SESSION处理
	if ctx.sessional() {
		mmm,eee := mSESSION.Read(ctx.Id)
		if eee == nil {
			for k,v := range mmm {
				ctx.Session[k] = v
			}
		}
	}

	if ctx.Name=="" || ctx.Config == nil {
		module.found(ctx)
	} else {
		if err := ctx.argsHandler(); err != nil {
			ctx.lastError = err
			module.failed(ctx)
		} else {
			if err := ctx.authHandler(); err != nil {
				ctx.lastError = err
				module.denied(ctx)
			} else {
				if err := ctx.itemHandler(); err != nil {
					ctx.lastError = err
					module.failed(ctx)
				} else {
					//往下走，再做响应
					ctx.Next()
				}
			}
		}
	}

	//session写回去
	if ctx.sessional() {
		//待处理，如果SESSION没有动过，就不写SESSION
		mSESSION.Write(ctx.Id, ctx.Session)
	}

	//响应前清空执行线
	ctx.clear()

	//response拦截器，加入调用列表
	filters := module.responseFilterActions()
	ctx.next(filters...)

	//最终的body处理
	ctx.next(module.body)

	ctx.Next()
}



//事件执行，调用action的地方
func (module *eventModule) execute(ctx *Context) {
	ctx.clear()

	//executeFilters
	filters := module.executeFilterActions()
	ctx.next(filters...)

	//actions
	funcs := ctx.funcing(kACTION)
	ctx.next(funcs...)

	ctx.Next()
}




//最终响应
func (module *eventModule) body(ctx *Context) {
	switch body:= ctx.Body.(type) {
	case finishBody:
		module.bodyFinish(ctx, body)
	case delayBody:
		module.bodyDelay(ctx, body)
	default:
		module.bodyDefault(ctx)
	}

	//最终响应前做清理工作
	ctx.final()
}
func (module *eventModule) bodyDefault(ctx *Context) {
	ctx.event.res.Finish(ctx.event.req)
}
func (module *eventModule) bodyFinish(ctx *Context, body finishBody) {
	ctx.event.res.Finish(ctx.event.req)
}
func (module *eventModule) bodyDelay(ctx *Context, body delayBody) {
	ctx.event.res.Delay(ctx.event.req, body.Delay)
}













//事件handler,找不到
func (module *eventModule) found(ctx *Context) {
	ctx.clear()

	//如果有自定义的错误处理，加入调用列表
	founds := ctx.funcing(kFOUND)
	ctx.next(founds...)

	//把处理器加入调用列表
	handlers := module.foundHandlerActions()
	ctx.next(handlers...)

	//加入默认的错误处理
	ctx.next(module.foundDefaultHandler)
	ctx.Next()
}
//最终还是由response处理
func (module *eventModule) foundDefaultHandler(ctx *Context) {
	ctx.Finish()
}

//事件handler,错误处理
func (module *eventModule) error(ctx *Context) {
	ctx.clear()

	//如果有自定义的错误处理，加入调用列表
	errors := ctx.funcing(kERROR)
	ctx.next(errors...)

	//把错误处理器加入调用列表
	handlers := module.errorHandlerActions()
	ctx.next(handlers...)

	//加入默认的错误处理
	ctx.next(module.errorDefaultHandler)
	ctx.Next()
}
//最终还是由response处理
func (module *eventModule) errorDefaultHandler(ctx *Context) {
	ctx.Finish()
}


//事件handler,失败处理，主要是args失败
func (module *eventModule) failed(ctx *Context) {
	ctx.clear()

	//如果有自定义的失败处理，加入调用列表
	faileds := ctx.funcing(kFAILED)
	ctx.next(faileds...)

	//把失败处理器加入调用列表
	handlers := module.failedHandlerActions()
	ctx.next(handlers...)

	//加入默认的错误处理
	ctx.next(module.failedDefaultHandler)
	ctx.Next()
}
//最终还是由response处理
func (module *eventModule) failedDefaultHandler(ctx *Context) {
	ctx.Finish()
}



//事件handler,失败处理，主要是args失败
func (module *eventModule) denied(ctx *Context) {
	ctx.clear()

	//如果有自定义的失败处理，加入调用列表
	denieds := ctx.funcing(kDENIED)
	ctx.next(denieds...)

	//把失败处理器加入调用列表
	handlers := module.deniedHandlerActions()
	ctx.next(handlers...)

	//加入默认的错误处理
	ctx.next(module.deniedDefaultHandler)
	ctx.Next()
}
//最终还是由response处理
func (module *eventModule) deniedDefaultHandler(ctx *Context) {
	ctx.Finish()
}

