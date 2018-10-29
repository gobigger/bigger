package bigger

import (
	"time"
	"github.com/gobigger/bigger/hashring"
)


// queue driver begin
type (
	//队列驱动
	QueueDriver interface {
		Connect(string,QueueConfig) (QueueConnect,*Error)
	}

	//队列连接
	QueueConnect interface {
		Open() *Error
        Health() (*QueueHealth,*Error)
		Close() *Error

		Accept(QueueHandler) *Error
		Register(name string, config QueueRegister) *Error
		
		Start() *Error

		Produce(name string, value Map) *Error
		DeferredProduce(name string, delay time.Duration, value Map) *Error
	}

	//队列处理器
	QueueHandler func(*QueueRequest, QueueResponse)

	QueueRegister struct {
		Lines	int
	}


    QueueHealth struct {
        Workload    int64
    }

	//队列请求实体
	QueueRequest struct {
		Id		string
		Name	string
		Value	Map
	}
	//队列响应接口
	QueueResponse interface {
		//完成
		Finish(*QueueRequest) *Error
		//重新开始
		Delay(*QueueRequest, time.Duration) *Error
	}

)



type (
    queueModule struct {
		driver		coreBranch
		router		coreBranch
		filter		coreBranch
		handler		coreBranch
        
        //队列配置，队列连接
		connects    map[string]QueueConnect
		hashring	*hashring.HashRing
    }
)


func (module *queueModule) Driver(name string, driver QueueDriver, overrides ...bool) {
    if driver == nil {
        panic("[队列]驱动不可为空")
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




//liner只是更新现有配置中的lines字段
func (module *queueModule) Liner(name string, lines int, overrides ...bool) {
	if config,ok := module.router.chunkdata(name).(Map); ok {
		config[kLINES] = lines
		module.Router(name, config, overrides...)
	}
}
func (module *queueModule) Router(name string, config Map, overrides ...bool) {
    override := true
    if len(overrides) > 0 {
        override = overrides[0]
	}
	
	if config[kLINES] == nil && config[kLINE] == nil {
		config[kLINES] = 1
	}

    if override {
		module.router.chunking(name, config)
    } else {
        if module.router.chunkdata(name) == nil {
			module.router.chunking(name, config)
        }
    }
}
func (module *queueModule) Filter(name string, config Map, overrides ...bool) {
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
func (module *queueModule) Handler(name string, config Map, overrides ...bool) {
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


func (module *queueModule) Produce(name string, value Map, bases ...string) (*Error) {
	base := DEFAULT
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
		return conn.Produce(name, value)
	}

	return Bigger.Erring("[队列]无效连接")
}

func (module *queueModule) DeferredProduce(name string, delay time.Duration, value Map, bases ...string) *Error {
	base := DEFAULT
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
		return conn.DeferredProduce(name, delay, value)
	}

	return Bigger.Erring("[队列]无效连接")
}





func (module *queueModule) connecting(name string, config QueueConfig) (QueueConnect,*Error) {
    if driver,ok := module.driver.chunkdata(config.Driver).(QueueDriver); ok {
        return driver.Connect(name, config)
    }
    panic("[队列]不支持的驱动：" + config.Driver)
}
func (module *queueModule) initing() {

	weights := make(map[string]int)
    for name,config := range Bigger.Config.Queue {
		if config.Weight > 0 {
			weights[name] = config.Weight
		}

		connect,err := module.connecting(name, config)
		if err != nil {
			panic("[队列]连接失败：" + err.Error())
		}
		err = connect.Open()
		if err != nil {
			panic("[队列]打开失败：" + err.Error())
		}

		//所有的也开始订阅者
		err = connect.Accept(module.serve)
		if err != nil {
			panic("[队列]注册回调失败：" + err.Error())
		}

		//注册队列
		routers := module.router.chunks()
		for _,val := range routers {
			if vcfg,ok := val.data.(Map); ok {
				
				regis := module.registering(vcfg)

				//如果配置文件中有定义线程数
				if vv,ok := config.Liner[val.name]; ok {
					regis.Lines = vv
				}

				err := connect.Register(val.name, regis)
				if err != nil {
					panic("[队列]注册订阅失败：" + err.Error())
				}
			}
		}

		err = connect.Start()
		if err != nil {
			panic("[队列]订阅失败：" + err.Error())
		}

		module.connects[name] = connect
	}
	
	module.hashring = hashring.New(weights)
}
func (module *queueModule) registering(config Map) (QueueRegister) {

	lines := 1
	if vv,ok := config[kLINE].(int); ok && vv > 0 {
		lines = vv
	}
	if vv,ok := config[kLINE].(int64); ok && vv > 0 {
		lines = int(vv)
	}
	if vv,ok := config[kLINES].(int); ok && vv > 0 {
		lines = vv
	}
	if vv,ok := config[kLINES].(int64); ok && vv > 0 {
		lines = int(vv)
	}
	if lines <= 0 {
		lines = 1
	}

	return QueueRegister{ Lines: lines }
}

func (module *queueModule) exiting() {
    for _,connect := range module.connects {
        connect.Close()
    }
}



//动态注册新来的
func (module *queueModule) newbie(newbie coreNewbie) (*Error) {
	if newbie.branch != bQUEUEROUTER {
		return nil
	}

	for _,connect := range module.connects {
		if config,ok := module.router.chunkdata(newbie.block).(Map); ok {
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




func (module *queueModule) requestFilterActions() ([]Funcing) {
	return module.filter.funcings(kREQUEST)
}
func (module *queueModule) executeFilterActions() ([]Funcing) {
	return module.filter.funcings(kEXECUTE)
}
func (module *queueModule) responseFilterActions() ([]Funcing) {
	return module.filter.funcings(kRESPONSE)
}


func (module *queueModule) foundHandlerActions() ([]Funcing) {
	return module.handler.funcings(kFOUND)
}
func (module *queueModule) errorHandlerActions() ([]Funcing) {
	return module.handler.funcings(kERROR)
}
func (module *queueModule) failedHandlerActions() ([]Funcing) {
	return module.handler.funcings(kFAILED)
}
func (module *queueModule) deniedHandlerActions() ([]Funcing) {
	return module.handler.funcings(kDENIED)
}










//队列Queue  请求开始
func (module *queueModule) serve(req *QueueRequest, res QueueResponse) {
	ctx := newQueueContext(req, res)

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


//队列响应
//先执行下面
//再清理执行线
//再把
func (module *queueModule) request(ctx *Context) {

	if ctx.Id == "" {
		ctx.Id = Bigger.Unique()
	}

	//请求的一开始，主要是SESSION处理
	if ctx.sessional() {
		//待处理，如果SESSION没有动过，就不写SESSION
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


//队列执行，调用action的地方
func (module *queueModule) execute(ctx *Context) {
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
func (module *queueModule) body(ctx *Context) {

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
func (module *queueModule) bodyDefault(ctx *Context) {
	ctx.queue.res.Finish(ctx.queue.req)
}
func (module *queueModule) bodyFinish(ctx *Context, body finishBody) {
	ctx.queue.res.Finish(ctx.queue.req)
}
func (module *queueModule) bodyDelay(ctx *Context, body delayBody) {
	ctx.queue.res.Delay(ctx.queue.req, body.Delay)
}













//队列handler,找不到
func (module *queueModule) found(ctx *Context) {
	ctx.clear()

	//如果有自定义的错误处理，加入调用列表
	funcs := ctx.funcing(kFOUND)
	ctx.next(funcs...)

	//把处理器加入调用列表
	handlers := module.foundHandlerActions()
	ctx.next(handlers...)

	//加入默认的错误处理
	ctx.next(module.foundDefaultHandler)
	ctx.Next()
}
//最终还是由response处理
func (module *queueModule) foundDefaultHandler(ctx *Context) {
	ctx.Finish()
}

//队列handler,错误处理
func (module *queueModule) error(ctx *Context) {
	ctx.clear()

	//如果有自定义的错误处理，加入调用列表
	funcs := ctx.funcing(kERROR)
	ctx.next(funcs...)

	//把错误处理器加入调用列表
	handlers := module.errorHandlerActions()
	ctx.next(handlers...)

	//加入默认的错误处理
	ctx.next(module.errorDefaultHandler)
	ctx.Next()
}
//最终还是由response处理
func (module *queueModule) errorDefaultHandler(ctx *Context) {
	ctx.Finish()
}


//队列handler,失败处理，主要是args失败
func (module *queueModule) failed(ctx *Context) {
	ctx.clear()

	//如果有自定义的失败处理，加入调用列表
	funcs := ctx.funcing(kFAILED)
	ctx.next(funcs...)

	//把失败处理器加入调用列表
	handlers := module.failedHandlerActions()
	ctx.next(handlers...)

	//加入默认的错误处理
	ctx.next(module.failedDefaultHandler)
	ctx.Next()
}
//最终还是由response处理
func (module *queueModule) failedDefaultHandler(ctx *Context) {
	ctx.Finish()
}



//队列handler,失败处理，主要是args失败
func (module *queueModule) denied(ctx *Context) {
	ctx.clear()

	//如果有自定义的失败处理，加入调用列表
	funcs := ctx.funcing(kDENIED)
	ctx.next(funcs...)

	//把失败处理器加入调用列表
	handlers := module.deniedHandlerActions()
	ctx.next(handlers...)

	//加入默认的错误处理
	ctx.next(module.deniedDefaultHandler)
	ctx.Next()
}
//最终还是由response处理
func (module *queueModule) deniedDefaultHandler(ctx *Context) {
	ctx.Finish()
}



