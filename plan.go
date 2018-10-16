package bigger

import (
	"time"
	"github.com/yatlabs/bigger/raft"
)


// Plan driver begin
type (
	//计划驱动
	PlanDriver interface {
		Connect(PlanConfig) (PlanConnect,*Error)
	}

	//计划连接
	PlanConnect interface {
		Open() *Error
        Health() (*PlanHealth,*Error)
		Close() *Error

		Accept(PlanHandler) *Error
		Register(name string, config PlanRegister) (*Error)

		Start() *Error

		Execute(name string, value Map) *Error
		DeferredExecute(name string, delay time.Duration, value Map) *Error
	}

	//计划处理器
	PlanHandler func(*PlanRequest, PlanResponse)
	PlanRegister struct {
		Times	[]string
		Delay	bool		//是否可延期执行，默认true，只有可delay的，ctx.Delay(才有效)
	}



    PlanHealth struct {
        Workload    int64
    }

	//计划请求实体
	PlanRequest struct {
		Id			string
		Name		string
		Value		Map
		Manual		bool
		Delay		bool
	}
	//计划响应接口
	PlanResponse interface {
		//完成
		Finish(*PlanRequest) *Error
		//重新开始
		Delay(*PlanRequest, time.Duration) *Error
	}

)


type (
    planModule struct {
		driver		coreBranch
		router		coreBranch
		filter		coreBranch
		handler		coreBranch
        
        //计划配置，计划连接
        config      PlanConfig
		connect     PlanConnect
    }
)





func (module *planModule) Driver(name string, driver PlanDriver, overrides ...bool) {
    if driver == nil {
        panic("[计划]驱动不可为空")
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



//Timer也只是更新times字段，不再单独branch
func (module *planModule) Timer(name string, times []string, overrides ...bool) {
	if config,ok := module.router.chunkdata(name).(Map); ok {
		config[kTIMES] = times
		module.Router(name, config, overrides...)
	}
}

func (module *planModule) Router(name string, config Map, overrides ...bool) {
    override := true
    if len(overrides) > 0 {
        override = overrides[0]
    }

    if override {
		module.router.chunking(name, config)
    } else {
        if module.router.chunkdata(name) == nil {
			module.router.chunking(name, config)
        }
    }
}
func (module *planModule) Filter(name string, config Map, overrides ...bool) {
    override := true
    if len(overrides) > 0 {
        override = overrides[0]
    }

    if override {
		module.filter.chunking(name, config)
    } else {
        if module.filter.chunkdata(name) == nil {
			module.filter.chunking(name, config)
        }
    }
}
func (module *planModule) Handler(name string, config Map, overrides ...bool) {
    override := true
    if len(overrides) > 0 {
        override = overrides[0]
    }

    if override {
		module.handler.chunking(name, config)
    } else {
        if module.handler.chunkdata(name) == nil {
			module.handler.chunking(name, config)
        }
    }
}


func (module *planModule) Execute(name string, args ...Map) (*Error) {
    if module.connect == nil {
        return Bigger.Erring("[计划]无效连接")
    }

	value := Map{}
	if len(args) > 0 {
		value = args[0]
    }
    
    return module.connect.Execute(name, value)
}
func (module *planModule) DeferredExecute(name string, delay time.Duration, args ...Map) (*Error) {
    if module.connect == nil {
        return Bigger.Erring("[计划]无效连接")
    }

	value := Map{}
	if len(args) > 0 {
		value = args[0]
    }
    
    return module.connect.DeferredExecute(name, delay, value)
}








func (module *planModule) connecting(config PlanConfig) (PlanConnect,*Error) {
    if driver,ok := module.driver.chunkdata(config.Driver).(PlanDriver); ok {
        return driver.Connect(config)
    }
    panic("[计划]不支持的驱动：" + config.Driver)
}
func (module *planModule) initing() {
    
    //连接
    config := Bigger.Config.Plan
    connect,err := module.connecting(config)
    if err != nil {
        panic("[计划]连接失败：" + err.Error())
    }
    
    //打开连接
    err = connect.Open()
    if err != nil {
        panic("[计划]打开失败：" + err.Error())
    }

    //注册
		err = connect.Accept(module.serve)
		if err != nil {
			panic("[计划]]注册回调失败：" + err.Error())
		}

		//注册队列
		routers := module.router.chunks()
		for _,val := range routers {
			if cfg,ok := val.data.(Map); ok {

				regis:= module.registering(cfg)

				//自定义时间在配置文件中
				if vvs,ok := Bigger.Config.Plan.Timer[val.name]; ok {
					regis.Times = vvs
				}

				if len(regis.Times) > 0 {
					err := connect.Register(val.name, regis)
					if err != nil {
						panic("[计划]注册失败：" + err.Error())
					}
				}

			}
		}

		err = connect.Start()
		if err != nil {
			panic("[计划]开始失败：" + err.Error())
		}


    //保存连接
    module.config = config
    module.connect = connect

}
func (module *planModule) registering(config Map) (PlanRegister) {
	times := []string{}


	delay := true
	if vv,ok := config[kDELAY].(bool); ok {
		delay = vv
	}
	if vv,ok := config[kTIME].(string); ok && vv != "" {
		times = append(times, vv)
	}
	if vvs,ok := config[kTIMES].([]string); ok && len(vvs)>0 {
		times = append(times, vvs...)
	}
	
	return PlanRegister{ Times: times, Delay: delay }
}

func (module *planModule) exiting() {
    if module.connect != nil {
        module.connect.Close()
    }
}


//动态注册新来的
func (module *planModule) newbie(newbie coreNewbie) (*Error) {
    if module.connect == nil {
		return nil
	}
	if newbie.branch != bPLANROUTER {
		return nil
	}

	if config,ok := module.router.chunkdata(newbie.block).(Map); ok {
		name := newbie.block
		regis := module.registering(config)
		
		return module.connect.Register(name, regis)
	}

	return nil
}



func (module *planModule) requestFilterActions() ([]Funcing) {
	return module.filter.funcings(kREQUEST)
}
func (module *planModule) executeFilterActions() ([]Funcing) {
	return module.filter.funcings(kEXECUTE)
}
func (module *planModule) responseFilterActions() ([]Funcing) {
	return module.filter.funcings(kRESPONSE)
}


func (module *planModule) foundHandlerActions() ([]Funcing) {
	return module.handler.funcings(kFOUND)
}
func (module *planModule) errorHandlerActions() ([]Funcing) {
	return module.handler.funcings(kERROR)
}
func (module *planModule) failedHandlerActions() ([]Funcing) {
	return module.handler.funcings(kFAILED)
}
func (module *planModule) deniedHandlerActions() ([]Funcing) {
	return module.handler.funcings(kDENIED)
}



























//计划Plan  请求开始
func (module *planModule) serve(req *PlanRequest, res PlanResponse) {

	state := Bigger.raft.State()
	//只有leader才处理计划
	if req.Manual == false && state != raft.Leader {
		if state == raft.Candidate  {
			//如果是选举模式，则延期执行
			res.Delay(req, time.Second*5)
		} else {
			//否则直接完成
			res.Finish(req)
		}
		return
	}



	ctx := newPlanContext(req, res)

	//判断路由是否存在
	if config,ok := module.router.chunkdata(ctx.Name).(Map); ok {	
		ctx.Config = config
	}

	//request拦截器，加入调用列表
	requestFilters := module.requestFilterActions()
	ctx.next(requestFilters...)
	
	ctx.next(module.request)
	ctx.next(module.execute)

	ctx.Next()
}


//计划响应
//先执行下面
//再清理执行线
func (module *planModule) request(ctx *Context) {

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
		//待处理，如果SESSION没有动过
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


//计划执行，调用action的地方
func (module *planModule) execute(ctx *Context) {
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
func (module *planModule) body(ctx *Context) {

	switch body := ctx.Body.(type) {
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
func (module *planModule) bodyDefault(ctx *Context) {
	ctx.plan.res.Finish(ctx.plan.req)
}
func (module *planModule) bodyFinish(ctx *Context, body finishBody) {
	ctx.plan.res.Finish(ctx.plan.req)
}
func (module *planModule) bodyDelay(ctx *Context, body delayBody) {
	ctx.plan.res.Delay(ctx.plan.req, body.Delay)
}













//计划handler,找不到
func (module *planModule) found(ctx *Context) {
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
func (module *planModule) foundDefaultHandler(ctx *Context) {
	ctx.Finish()
}

// 计划handler,错误处理
func (module *planModule) error(ctx *Context) {
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
func (module *planModule) errorDefaultHandler(ctx *Context) {
	ctx.Finish()
}


//计划handler,失败处理，主要是args失败
func (module *planModule) failed(ctx *Context) {
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
func (module *planModule) failedDefaultHandler(ctx *Context) {
	ctx.Finish()
}



//计划handler,失败处理，主要是args失败
func (module *planModule) denied(ctx *Context) {
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
func (module *planModule) deniedDefaultHandler(ctx *Context) {
	ctx.Finish()
}













