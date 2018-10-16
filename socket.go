package bigger

import (
	"time"
	"fmt"
	"strings"
	"encoding/json"
	"net/http"
	"github.com/yatlabs/bigger/hashring"
)

type (
	//套接驱动
	SocketDriver interface {
		Connect(string,SocketConfig) (SocketConnect,*Error)
	}

	//套接连接
	SocketConnect interface {
		Open() *Error
        Health() (*SocketHealth,*Error)
		Close() *Error

		Accept(SocketHandler) *Error
		Start() *Error

		Upgrade(string, *http.Request,http.ResponseWriter) *Error
		Degrade(string) *Error
		Follow(id string, channel string) *Error
		Unfollow(id string, channel string) *Error
		
		Message(string, []byte) *Error
		DeferredMessage(id string, delay time.Duration, bytes []byte) *Error
		Broadcast(channel string, bytes []byte) *Error
		DeferredBroadcast(channel string, delay time.Duration, bytes []byte) *Error
	}

	//套接处理器
	SocketHandler func(*SocketRequest, SocketResponse)


    SocketHealth struct {
        Workload    int64
    }

	//套接请求实体
	SocketRequest struct {
		Id		string
		Data	[]byte
	}
	//套接响应接口
	SocketResponse interface {
		//完成
		Finish(*SocketRequest) *Error
		//重新开始
		Delay(*SocketRequest, time.Duration) *Error
	}

)

type (
    socketModule struct {
		driver		coreBranch
		router		coreBranch
		filter		coreBranch
		handler		coreBranch
		command		coreBranch
        
		connects    map[string]SocketConnect
		hashring	*hashring.HashRing
    }
	SocketCoding struct {
		Base		string		//所在的库
		Flag		string		//自定义的flag
		Rand		string		//随机串
	}
)



func (module *socketModule) Encode(base, flag, rand string) (string) {
    return Bigger.Encrypt(fmt.Sprintf("%s\n%s\n%s", base, flag, rand))
}

func (module *socketModule) Decode(code string) (*SocketCoding) {
	str := Bigger.Decrypt(code)
	if str == "" {
		return nil
	}
	args := strings.Split(str, "\n")
	if len(args) != 3 {
		return nil
	}
	return &SocketCoding{
		Base: args[0],
		Flag: args[1],
		Rand: args[2],
	}
}





func (module *socketModule) Driver(name string, driver SocketDriver, overrides ...bool) {
    if driver == nil {
        panic("[套接]驱动不可为空")
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




func (module *socketModule) Router(name string, config Map, overrides ...bool) {
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
func (module *socketModule) Filter(name string, config Map, overrides ...bool) {
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
func (module *socketModule) Handler(name string, config Map, overrides ...bool) {
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
func (module *socketModule) Command(name string, config Map, overrides ...bool) {
    override := true
    if len(overrides) > 0 {
        override = overrides[0]
    }

    if override {
		module.command.chunking(name, config)
    } else {
        if module.command.chunk(name) == nil {
			module.command.chunking(name, config)
        }
    }
}


func (module *socketModule) Assign(flag, rand string, bases ...string) (string) {
	base := kDEFAULT
	if len(bases) > 0 {
		base = bases[0]
	} else if node := module.hashring.Locate(rand); node != "" {
		//按权重随机分散文件存储
		base = node
	} else {
		for n,_ := range module.connects {
			base = n
			break
		}
	}

	if rand == "" {
		rand = Bigger.Unique()
	}

	return module.Encode(base, flag, rand)
}




func (module *socketModule) Upgrade(code string, req *http.Request, res http.ResponseWriter) (*Error) {
	data := module.Decode(code)
	if data == nil {
		return Bigger.Erring("解析失败")
	}
	if conn,ok := module.connects[data.Base]; ok {
		return conn.Upgrade(code, req, res)
	}
	return Bigger.Erring("[套接]无效连接")
}
func (module *socketModule) Degrade(code string) (*Error) {
	data := module.Decode(code)
	if data == nil {
		return Bigger.Erring("解析失败")
	}
	if conn,ok := module.connects[data.Base]; ok {
		return conn.Degrade(code)
	}
	return Bigger.Erring("[套接]无效连接")
}
func (module *socketModule) Follow(code, channel string) (*Error) {
	data := module.Decode(code)
	if data == nil {
		return Bigger.Erring("解析失败")
	}
	if conn,ok := module.connects[data.Base]; ok {
		return conn.Follow(code, channel)
	}
	return Bigger.Erring("[套接]无效连接")
}
func (module *socketModule) Unfollow(code, channel string) (*Error) {
	data := module.Decode(code)
	if data == nil {
		return Bigger.Erring("解析失败")
	}
	if conn,ok := module.connects[data.Base]; ok {
		return conn.Unfollow(code, channel)
	}
	return Bigger.Erring("[套接]无效连接")
}




func (module *socketModule) Message(code string, cmd string, value Map) (*Error) {
	coding := module.Decode(code)
	if coding == nil {
		return Bigger.Erring("解析失败")
	}
	
	if value == nil {
		value = Map{}
	}

	if conn,ok := module.connects[coding.Base]; ok {
		if config,ok := module.command.chunk(cmd).(Map); ok {
			mmmm := Map{ kTYPE: cmd, kDATA: value }

			//包装数据
			if dataConfig,ok := config[kDATA].(Map); ok {
				data := Map{}
				err := Bigger.Mapping(dataConfig, value, data, false, false)
				if err != nil {
					return err
				}
				mmmm[kDATA] = data
			} else{
				mmmm[kDATA] = value
			}
			
			bytes,err := json.Marshal(&mmmm)
			if err != nil {
				return Bigger.Erred(err)
			}

			//发送消息
			return conn.Message(code, bytes)

		} else {
			return Bigger.Erring("未知消息")
		}
	} else {
		return Bigger.Erring("无效连接")
	}
}
func (module *socketModule) DeferredMessage(code string, cmd string, delay time.Duration, value Map) *Error {
	coding := module.Decode(code)
	if coding == nil {
		return Bigger.Erring("解析失败")
	}

	if value == nil {
		value = Map{}
	}

	if conn,ok := module.connects[coding.Base]; ok {
		if config,ok := module.command.chunk(cmd).(Map); ok {
			mmmm := Map{ kTYPE: cmd, kDATA: value }

			//包装数据
			if dataConfig,ok := config[kDATA].(Map); ok {
				data := Map{}
				err := Bigger.Mapping(dataConfig, value, data, false, false)
				if err != nil {
					return err
				}
				mmmm[kDATA] = data
			} else{
				mmmm[kDATA] = value
			}
			
			bytes,err := json.Marshal(&mmmm)
			if err != nil {
				return Bigger.Erred(err)
			}

			//发送消息
			return conn.DeferredMessage(code, delay, bytes)

		} else {
			return Bigger.Erring("未知消息")
		}
	} else {
		return Bigger.Erring("无效连接")
	}
}

func (module *socketModule) Broadcast(channel string, command string, value Map, bases ...string) (*Error) {
	base := kDEFAULT
	if len(bases) > 0 {
		base = bases[0]
	} else if node := module.hashring.Locate(channel); node != "" {
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
		if config,ok := module.command.chunk(command).(Map); ok {
			mmmm := Map{ kTYPE: command, kDATA: value }

			//包装数据
			if dataConfig,ok := config[kDATA].(Map); ok {
				data := Map{}
				err := Bigger.Mapping(dataConfig, value, data, false, false)
				if err != nil {
					return err
				}
				mmmm[kDATA] = data
			} else{
				mmmm[kDATA] = value
			}
			
			bytes,err := json.Marshal(&mmmm)
			if err != nil {
				return Bigger.Erred(err)
			}

			//发送广播
			return conn.Broadcast(channel, bytes)

		} else {
			return Bigger.Erring("未知消息")
		}
	} else {
		return Bigger.Erring("无效连接")
	}
}
func (module *socketModule) DeferredBroadcast(channel string, command string, delay time.Duration, value Map, bases ...string) *Error {
	base := kDEFAULT
	if len(bases) > 0 {
		base = bases[0]
	} else if node := module.hashring.Locate(channel); node != "" {
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
		if config,ok := module.command.chunk(command).(Map); ok {
			mmmm := Map{ kTYPE: command, kDATA: value }

			//包装数据
			if dataConfig,ok := config[kDATA].(Map); ok {
				data := Map{}
				err := Bigger.Mapping(dataConfig, value, data, false, false)
				if err != nil {
					return err
				}
				mmmm[kDATA] = data
			} else{
				mmmm[kDATA] = value
			}
			
			bytes,err := json.Marshal(&mmmm)
			if err != nil {
				return Bigger.Erred(err)
			}

			//发送广播
			return conn.DeferredBroadcast(channel, delay, bytes)

		} else {
			return Bigger.Erring("未知消息")
		}
	} else {
		return Bigger.Erring("无效连接")
	}
}













func (module *socketModule) connecting(name string, config SocketConfig) (SocketConnect,*Error) {
    if driver,ok := module.driver.chunk(config.Driver).(SocketDriver); ok {
        return driver.Connect(name, config)
    }
    panic("[套接]不支持的驱动：" + config.Driver)
}
func (module *socketModule) initing() {

	weights := make(map[string]int)
    for name,config := range Bigger.Config.Socket {
		if config.Weight > 0 {
			weights[name] = config.Weight
		}

		connect,err := module.connecting(name, config)
		if err != nil {
			panic("[套接]连接失败：" + err.Error())
		}
		err = connect.Open()
		if err != nil {
			panic("[套接]打开失败：" + err.Error())
		}

		//所有的也开始订阅者
		err = connect.Accept(module.serve)
		if err != nil {
			panic("[套接]注册回调失败：" + err.Error())
		}

		err = connect.Start()
		if err != nil {
			panic("[套接]开始失败：" + err.Error())
		}

		module.connects[name] = connect
	}
	
	module.hashring = hashring.New(weights)
}

func (module *socketModule) exiting() {
    for _,connect := range module.connects {
        connect.Close()
    }
}




func (module *socketModule) requestFilterActions() ([]Funcing) {
	return module.filter.funcings(kREQUEST)
}
func (module *socketModule) executeFilterActions() ([]Funcing) {
	return module.filter.funcings(kEXECUTE)
}
func (module *socketModule) responseFilterActions() ([]Funcing) {
	return module.filter.funcings(kRESPONSE)
}


func (module *socketModule) foundHandlerActions() ([]Funcing) {
	return module.handler.funcings(kFOUND)
}
func (module *socketModule) errorHandlerActions() ([]Funcing) {
	return module.handler.funcings(kERROR)
}
func (module *socketModule) failedHandlerActions() ([]Funcing) {
	return module.handler.funcings(kFAILED)
}
func (module *socketModule) deniedHandlerActions() ([]Funcing) {
	return module.handler.funcings(kDENIED)
}





//套接Socket  请求开始
func (module *socketModule) serve(req *SocketRequest, res SocketResponse) {
	ctx := newSocketContext(req, res)

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


//套接响应
//先执行下面
//再清理执行线
func (module *socketModule) request(ctx *Context) {

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



//套接执行，调用action的地方
func (module *socketModule) execute(ctx *Context) {
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
func (module *socketModule) body(ctx *Context) {
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
func (module *socketModule) bodyDefault(ctx *Context) {
	ctx.socket.res.Finish(ctx.socket.req)
}
func (module *socketModule) bodyFinish(ctx *Context, body finishBody) {
	ctx.socket.res.Finish(ctx.socket.req)
}
func (module *socketModule) bodyDelay(ctx *Context, body delayBody) {
	ctx.socket.res.Delay(ctx.socket.req, body.Delay)
}



//套接handler,找不到
func (module *socketModule) found(ctx *Context) {
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
func (module *socketModule) foundDefaultHandler(ctx *Context) {
	ctx.Finish()
}

//套接handler,错误处理
func (module *socketModule) error(ctx *Context) {
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
func (module *socketModule) errorDefaultHandler(ctx *Context) {
	ctx.Finish()
}


//套接handler,失败处理，主要是args失败
func (module *socketModule) failed(ctx *Context) {
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
func (module *socketModule) failedDefaultHandler(ctx *Context) {
	ctx.Finish()
}



//套接handler,失败处理，主要是args失败
func (module *socketModule) denied(ctx *Context) {
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
func (module *socketModule) deniedDefaultHandler(ctx *Context) {
	ctx.Finish()
}

