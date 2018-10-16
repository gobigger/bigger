package bigger

import (
	"time"
	"fmt"
)

type (
    serviceModule struct {
		service	coreBranch
	}
	

	serviceGroup struct {
		service	*serviceModule
		name	string
	}

	serviceLogic struct {
		service	*serviceModule
		ctx		*Context
		Name	string
		Setting	Map
	}

	Service struct {
		ctx		*Context
		Lang	string
		Zone	*time.Location

		Name	string
		Config	Map
		Setting	Map
		Params	Map
		Args	Map
	}
	serviceAction	= func(*Service)(Map,*Error)
)


func (module *serviceModule) newGroup(name string) (*serviceGroup) {
	return &serviceGroup{ module, name }
}




func (module *serviceModule) newLogic(ctx *Context, name string, settings ...Map) (*serviceLogic) {
	setting := Map{}
	if len(settings) > 0 {
		setting = settings[0]
	}
	return &serviceLogic{ module, ctx, name, setting }
}


//逻辑分组注册
func (group *serviceGroup) Name() string {
	return group.name
}
func (group *serviceGroup) Register(name string, config Map, overrides ...bool) {
	realName := fmt.Sprintf("%s.%s", group.name, name)
	group.service.Register(realName, config, overrides...)
}



//分组调用
func (client *serviceLogic) Invoke(name string, params Map) (Map,*Error) {
	return client.service.Invoke(client.ctx, name, client.Setting, params)
}

func (module *serviceModule) Register(name string, config Map, overrides ...bool) {
    override := true
    if len(overrides) > 0 {
        override = overrides[0]
	}

	if override {
		module.service.chunking(name, config)
	} else {
		if module.service.chunk(name) == nil {
			module.service.chunking(name, config)
		}
	}
}
//args, data都解析， 这样会牺牲一点性能，
//如果可以人为保证传的值和返回的值是OK的，其实就不需要解析了
func (module *serviceModule) Invoke(ctx *Context, name string, setting Map, params Map) (Map,*Error) {
	var config Map
	if vv,ok := module.service.chunk(name).(Map); ok {
		config = vv
	}

	if config == nil {
		return nil, Bigger.Erring("[服务]未注册")
	}

	if ctx == nil {
		ctx = newContext()
		defer ctx.final()
	}

	args := Map{}
	if arging,ok := config[kARGS].(Map); ok {
		err := mMAPPING.Parse(arging, params, args, false, false, ctx)
		if err != nil {
			return nil, Bigger.Erring("[服务]参数解析失败")
		}
	}

	if setting == nil {
		setting = Map{}
	}

	msv := &Service{
		ctx: ctx, Lang: ctx.Lang, Zone: ctx.Zone,
		Name: name, Config: config, Setting: setting,
		Params: params, Args: args,
	}

	if ff,ok := config[kACTION].(serviceAction); ok {
		result,err := ff(msv)
		if err != nil {
			return result, err
		}

		if dating,ok := config[kDATA].(Map); ok {
			out := Map{}
			err := mMAPPING.Parse(dating, result, out, false, false, ctx)
			if err == nil {
				return out, nil
			}
		}

		return result, nil
	}

	return nil,Bigger.Erring("[服务]调用失败")
}

//服务上下文，依赖Context
func (msv *Service) Erred() *Error {
	return msv.ctx.Erred()
}
func (msv *Service) File(bases ...string) (FileBase) {
	return msv.ctx.fileBase(bases...)
}
func (msv *Service) Data(bases ...string) (DataBase) {
	return msv.ctx.dataBase(bases...)
}
func (msv *Service) Cache(bases ...string) (CacheBase) {
	return msv.ctx.cacheBase(bases...)
}
func (msv *Service) Service(name string, settings ...Map) (*serviceLogic) {
	return msv.ctx.Service(name, settings...)
}
func (msv *Service) Storage(upload Map, named Any, metadata Map, bases ...string) (string) {
	return msv.ctx.Storage(upload, named, metadata, bases...)
}
func (msv *Service) Invoke(name string, params ...Map) (Map) {
	msv.ctx.lastError = nil

	param := Map{}
	if len(params) > 0 {
		param = params[0]
	}
	result,err := mSERVICE.Invoke(msv.ctx, name, msv.Setting, param)
	if err != nil {
		msv.ctx.lastError = err
	}
	return result
}

// func (msv *Service) Signed(key string) (bool) {
// 	return msv.ctx.Signed(key)
// }
// func (msv *Service) Signin(key string, id,name Any) {
// 	msv.ctx.Signin(key, id, name)
// }
// func (msv *Service) Signout(key string) {
// 	msv.ctx.Signout(key)
// }
// func (msv *Service) Signal(key string) Any {
// 	return msv.ctx.Signal(key)
// }
// func (msv *Service) Signer(key string) Any {
// 	return msv.ctx.Signer(key)
// }