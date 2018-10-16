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
		Ctx		*Context
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
		Ctx: ctx, Lang: ctx.Lang, Zone: ctx.Zone,
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
func (sv *Service) Erred() *Error {
	return sv.Ctx.Erred()
}
func (sv *Service) File(bases ...string) (FileBase) {
	return sv.Ctx.fileBase(bases...)
}
func (sv *Service) Data(bases ...string) (DataBase) {
	return sv.Ctx.dataBase(bases...)
}
func (sv *Service) Cache(bases ...string) (CacheBase) {
	return sv.Ctx.cacheBase(bases...)
}
func (sv *Service) Service(name string, settings ...Map) (*serviceLogic) {
	return sv.Ctx.Service(name, settings...)
}
func (sv *Service) Storage(upload Map, named Any, metadata Map, bases ...string) (string) {
	return sv.Ctx.Storage(upload, named, metadata, bases...)
}
func (sv *Service) Invoke(name string, params ...Map) (Map) {
	sv.Ctx.lastError = nil

	param := Map{}
	if len(params) > 0 {
		param = params[0]
	}
	result,err := mSERVICE.Invoke(sv.Ctx, name, sv.Setting, param)
	if err != nil {
		sv.Ctx.lastError = err
	}
	return result
}

// func (sv *Service) Signed(key string) (bool) {
// 	return sv.Ctx.Signed(key)
// }
// func (sv *Service) Signin(key string, id,name Any) {
// 	sv.Ctx.Signin(key, id, name)
// }
// func (sv *Service) Signout(key string) {
// 	sv.Ctx.Signout(key)
// }
// func (sv *Service) Signal(key string) Any {
// 	return sv.Ctx.Signal(key)
// }
// func (sv *Service) Signer(key string) Any {
// 	return sv.Ctx.Signer(key)
// }