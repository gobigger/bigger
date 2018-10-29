package bigger

import (
	"encoding/json"
	"os"
	"io"
	"strings"
	"fmt"
	"time"
	"net"
	"net/http"
	"net/url"
	"regexp"
)



type (
	Funcing func(*Context)
	Context struct {
		nexts	[]Funcing //方法列表
		index	int       //下一个索引

		lastError	*Error
	
		mode	mod

		plan	*planContext
		event	*eventContext
		queue	*queueContext
		http	*httpContext
		socket	*socketContext

		site	SiteConfig
		store	Map

		databases	map[string]DataBase
		cachebases	map[string]CacheBase
		filebases	map[string]FileBase

		headers		HttpHeaders
		cookies		HttpCookies

		Url	*contextUrl

		Id		string
		Site	string		//来自站点
		Name	string
		Config	Map

		Charset	string
		Lang	string
		Zone	*time.Location
		domain	string

		Local	Map
		Session	Map
		Value	Map
		Args	Map
		Auth	Map
		Item	Map

		//http专用
		Ajax	bool
		Method	string
		Host	string
		Uri		string

		Client	Map
		Param	Map
		Query	Map
		Form	Map
		Upload	Map
		Data	Map

		Code		int		//响应状态
		Type		string	//响应类型
		Body		Any		//

	}
	// triggerContext struct {
	// 	req *TriggerRequest
	// 	res TriggerResponse
	// }
	planContext struct {
		req *PlanRequest
		res PlanResponse
	}
	eventContext struct {
		req *EventRequest
		res EventResponse
	}
	queueContext struct {
		req *QueueRequest
		res QueueResponse
	}
	httpContext struct {
		req *HttpRequest
		res HttpResponse
	}
	socketContext struct {
		req *SocketRequest
		res SocketResponse
	}


	//响应完成
	finishBody struct {}
	//响应重新触发
	delayBody struct {
		Delay time.Duration
	}



	contextUrl struct {
		ctx		*Context
		req		*http.Request
	}
)


func newContext() *Context {
	ctx := &Context{
		index: 0, nexts: []Funcing{}, store: Map{},
		databases: map[string]DataBase{}, cachebases: map[string]CacheBase{}, filebases: map[string]FileBase{},
		Local: Map{}, Session: Map{}, Value: Map{}, Args: Map{}, Auth: Map{}, Item: Map{},
	}

	ctx.Url = &contextUrl{
		ctx: ctx,
	}

	ctx.Charset = vUTF8
	ctx.Lang = DEFAULT
	ctx.Zone = time.Local

	return ctx
}

func newPlanContext(req *PlanRequest, res PlanResponse) *Context {
	ctx := &Context{
		index: 0, nexts: []Funcing{}, store: Map{},
		mode: planMode, plan: &planContext{req, res},
		databases: map[string]DataBase{}, cachebases: map[string]CacheBase{}, filebases: map[string]FileBase{},
		Local: Map{}, Session: Map{}, Value: Map{}, Args: Map{}, Auth: Map{}, Item: Map{},
	}

	ctx.Id = req.Id
	ctx.Name = req.Name
	ctx.Value = req.Value

	ctx.Url = &contextUrl{
		ctx: ctx,
	}

	ctx.Charset = vUTF8
	ctx.Lang = DEFAULT
	ctx.Zone = time.Local

	return ctx
}
func newEventContext(req *EventRequest, res EventResponse) *Context {
	ctx := &Context{
		index: 0, nexts: []Funcing{}, store: Map{},
		mode: eventMode, event: &eventContext{req, res},
		databases: map[string]DataBase{}, cachebases: map[string]CacheBase{}, filebases: map[string]FileBase{},
		Local: Map{}, Session: Map{}, Value: Map{}, Args: Map{}, Auth: Map{}, Item: Map{},
	}

	ctx.Id = req.Id
	ctx.Name = req.Name
	ctx.Value = req.Value

	ctx.Url = &contextUrl{
		ctx: ctx,
	}

	ctx.Charset = vUTF8
	ctx.Lang = DEFAULT
	ctx.Zone = time.Local

	return ctx
}
func newQueueContext(req *QueueRequest, res QueueResponse) *Context {
	ctx := &Context{
		index: 0, nexts: []Funcing{}, store: Map{},
		mode: queueMode, queue: &queueContext{req, res},
		databases: map[string]DataBase{}, cachebases: map[string]CacheBase{}, filebases: map[string]FileBase{},
		Local: Map{}, Session: Map{}, Value: Map{}, Args: Map{}, Auth: Map{}, Item: Map{},
	}

	ctx.Id = req.Id
	ctx.Name = req.Name
	ctx.Value = req.Value

	ctx.Url = &contextUrl{
		ctx: ctx,
	}

	ctx.Charset = vUTF8
	ctx.Lang = DEFAULT
	ctx.Zone = time.Local

	return ctx
}


func newHttpContext(req *HttpRequest, res HttpResponse) *Context {
	ctx := &Context{
		index: 0, nexts: []Funcing{}, store: Map{},
		mode: httpMode, http: &httpContext{req, res},
		databases: map[string]DataBase{}, cachebases: map[string]CacheBase{}, filebases: map[string]FileBase{},
		headers: HttpHeaders{},  cookies: HttpCookies{},
		Local: Map{}, Session: Map{}, Value: Map{}, Args: Map{}, Auth: Map{}, Item: Map{},
		Client: Map{}, Param: Map{}, Query: Map{}, Form: Map{}, Upload: Map{}, Data: Map{},
	}

	ctx.Id = req.Id
	ctx.Name = req.Name
	ctx.Site = req.Site

	ctx.Charset = vUTF8
	ctx.Lang = DEFAULT
	ctx.Zone = time.Local

	ctx.Url = &contextUrl{
		ctx: ctx, req: req.Reader,
	}

	ctx.Method	= req.Reader.Method
	ctx.Uri 	= req.Reader.RequestURI

	//使用域名去找site
	ctx.Host = req.Reader.Host
	if strings.Contains(ctx.Host, ":") {
		hosts := strings.Split(ctx.Host, ":")
		if len(hosts) > 0 {
			ctx.Host = hosts[0]
		}
	}
	if ctx.Site == "" {
		if site,ok := Bigger.hosts[ctx.Host]; ok {
			ctx.Site = site
		}
	}

	if vvvv,ookk := Bigger.Config.Site[ctx.Site]; ookk {
		ctx.site = vvvv
	} else {
		ctx.site = SiteConfig{}
	}


	//获取根域名
	ctx.domain = req.Reader.Host
	parts := strings.Split(ctx.domain, ".")
	if len(parts) >= 2 {
		l := len(parts)
		ctx.domain = parts[l-2] + "." + parts[l-1]
	}

	return ctx
}

func newSocketContext(req *SocketRequest, res SocketResponse) *Context {
	ctx := &Context{
		index: 0, nexts: []Funcing{}, store: Map{},
		mode: socketMode, socket: &socketContext{req, res},
		databases: map[string]DataBase{}, cachebases: map[string]CacheBase{}, filebases: map[string]FileBase{},
		Session: Map{}, Value: Map{}, Local: Map{}, Auth: Map{}, Args: Map{},
	}

	ctx.Id = req.Id
	// ctx.Name = req.Name
	// ctx.Value = req.Value
	
	//解析数据，拿到name和value
	mmm := Map{}
	if err := json.Unmarshal(req.Data, &mmm); err == nil {
		if vv,ok := mmm[kNAME].(string); ok {
			ctx.Name = vv
		}
		if vv,ok := mmm[kTYPE].(string); ok {
			ctx.Name = vv
		}
		if vv,ok := mmm[kDATA].(Map); ok {
			ctx.Value = vv
		}
	}

	ctx.Url = &contextUrl{
		ctx: ctx,
	}

	ctx.Charset = vUTF8
	ctx.Lang = DEFAULT
	ctx.Zone = time.Local

	return ctx
}

//加入执行线
func (ctx *Context) clear() {
	ctx.index = 0
	ctx.nexts = make([]Funcing, 0)
}
func (ctx *Context) next(nexts ...Funcing) {
	ctx.nexts = append(ctx.nexts, nexts...)
}
func (ctx *Context) funcing(key string) ([]Funcing) {
	funcs := []Funcing{}

	if action,ok := ctx.Config[key]; ok && action != nil {
		switch actions:= action.(type) {
		case func(*Context):
			funcs = append(funcs, actions)
		case []func(*Context):
			for _,action := range actions {
				funcs = append(funcs, action)
			}
		case Funcing:
			funcs = append(funcs, actions)
		case []Funcing:
			funcs = append(funcs, actions...)
		default:
		}
	}

	return funcs
}
//最终的清理工作
func (ctx *Context) final() {
	for _,base := range ctx.databases {
		base.Close()
	}
	for _,base := range ctx.cachebases {
		base.Close()
	}
	for _,base := range ctx.filebases {
		base.Close()
	}
}
func (ctx *Context) Next() {
	if len(ctx.nexts) > ctx.index {
		next := ctx.nexts[ctx.index]
		ctx.index++
		if next != nil {
			next(ctx)
		} else {
			ctx.Next()
		}
	} else {
		//是否需要做执行完的处理
	}
}

//------------------------ 模块函数 begin ----------------------
//上下文获取数据库对象
//以保证同一个上下文，使用同一个数据连接，不多开
//ctx下禁止直接访问数据，只允许在服务里调用
func (ctx *Context) fileBase(bases ...string) (FileBase) {
	base := DEFAULT
	if len(bases) > 0 {
		base = bases[0]
	} else {
		for key,_ := range mFILE.connects {
			base = key
			break
		}
	}

	if _,ok := ctx.filebases[base]; ok==false {
		ctx.filebases[base] = mFILE.Base(base)
	}
	return ctx.filebases[base]
}
func (ctx *Context) dataBase(bases ...string) (DataBase) {
	base := DEFAULT
	if len(bases) > 0 {
		base = bases[0]
	} else {
		for key,_ := range mDATA.connects {
			base = key
			break
		}
	}

	if _,ok := ctx.databases[base]; ok==false {
		ctx.databases[base] = mDATA.Base(base)
	}
	return ctx.databases[base]
}
func (ctx *Context) cacheBase(bases ...string) (CacheBase) {
	base := DEFAULT
	if len(bases) > 0 {
		base = bases[0]
	} else {
		for key,_ := range mCACHE.connects {
			base = key
			break
		}
	}

	if _,ok := ctx.cachebases[base]; ok==false {
		ctx.cachebases[base] = mCACHE.Base(base)
	}
	return ctx.cachebases[base]
}
//------------------------ 模块函数 end ----------------------




//------------------------ 基础函数 begin ----------------------
//是否开启SESSION
func (ctx *Context) sessional(defs ...bool) bool {
	sessional := false
	if len(defs) > 0 {
		sessional = defs[0]
	}

	if vv,ok := ctx.Config[kSESSION].(bool); ok {
		sessional = vv
	}

	//如果有auth节，强制使用session
	if _,ok := ctx.Config[kAUTH]; ok {
		sessional = true
	}

	return sessional
}


//处理参数
func (ctx *Context) argsHandler() *Error {

	//argn表示参数都可为空
	argn := false
	if v,ok := ctx.Config["argn"].(bool); ok {
		argn = v
	}


	
	if argsConfig,ok := ctx.Config["args"].(Map); ok {
		argsValue := Map{}
		err := mMAPPING.Parse(argsConfig, ctx.Value, argsValue, argn, false)
		if err != nil {
			err.status = strings.Replace(err.status, ".mapping.", ".args.", 1)
			return err
		}

		for k,v := range argsValue {
			ctx.Args[k] = v
		}
	}

	return nil
}

//处理认证
func (ctx *Context) authHandler() *Error {
	
	if auths,ok := ctx.Config[kAUTH].(Map); ok {
		saveMap := Map{}

		for authKey,authMap := range auths {

			if _,ok := authMap.(Map); ok == false {
				continue
			}

			ohNo := false
			authConfig := authMap.(Map)

			if vv,ok := authConfig[kSIGN].(string); ok==false || vv=="" {
				continue
			}

			authSign := authConfig[kSIGN].(string)
			authMust := false
			// authName := authSign

			if vv,ok := authConfig[kMUST].(bool); ok {
				authMust = vv
			}
			// if vv,ok := authConfig[kNAME].(string); ok && vv!="" {
			// 	authName = vv
			// }

			//判断是否登录
			if ctx.Signed(authSign) {

				//支持两种方式
				//1. data=table 如： "auth": "db.user"
				//2. base=db, table=user

				base, table := "",""
				if authConfig["data"] != nil {
					if vv,ok := authConfig["data"].(string); ok {
						i := strings.Index(vv, ".")
						base = vv[:i]
						table = vv[i+1:]
					}
				} else if authConfig["base"]!=nil &&  authConfig["table"]!=nil {
					if vv,ok := authConfig["base"].(string); ok {
						base = vv
					}
					if vv,ok := authConfig["table"].(string); ok {
						table = vv
					}
				}

				if base!="" && table!="" {
					db := ctx.dataBase(base)
					id := ctx.Signal(authSign)
					item := db.Table(table).Entity(id)

					if item == nil {
						if authMust {
							if vv,ok := authConfig[kERROR].(*Error); ok {
								return vv
							} else {
								errKey := ".auth.error"
								if vv,ok := authConfig[kERROR].(string); ok {
									errKey = vv
								}
								return newError(errKey, authKey)
							}
						}
					} else {
						saveMap[authKey] = item
					}
				}

			} else {
				ohNo = true
			}

			//到这里是未登录的
			//而且是必须要登录，才显示错误
			if ohNo && authMust {
				if vv,ok := authConfig[kEMPTY].(*Error); ok {
					return vv
				} else {
					errKey := ".auth.empty"
					if vv,ok := authConfig[kEMPTY].(string); ok {
						errKey = vv
					}
					return newError(errKey, authKey)
				}
			}
		}

		//存入
		for k,v := range saveMap {
			ctx.Auth[k] = v
		}
	}

	

	return nil
}

//Entity实体处理
func (ctx *Context) itemHandler() (*Error) {
	
	if cfg,ok := ctx.Config[kITEM].(Map); ok {
		saveMap := Map{}

		for itemKey,v := range cfg {
			if config,ok := v.(Map); ok {
				
				//是否必须
				must := true
				if vv,ok := config[kMUST].(bool); ok {
					must = vv
				}

				itemName := itemKey
				if vv,ok := config[kNAME].(string); ok && vv != "" {
					itemName = vv
				}

				realKey := itemKey
				var realVal Any
				if vv,ok := config[kARGS].(string); ok {
					realKey = vv
					realVal = ctx.Args[realKey]
				} else if vv,ok := config[kPARAM].(string); ok {
					realKey = vv
					realVal = ctx.Param[realKey]
				} else if vv,ok := config[kQUERY].(string); ok {
					realKey = vv
					realVal = ctx.Query[realKey]
				} else if vv,ok := config[kVALUE].(string); ok {
					realKey = vv
					realVal = ctx.Value[realKey]
				} else if vv,ok := config[kKEY].(string); ok {
					realKey = vv
					realVal = ctx.Value[realKey]
				} else {
					realVal = nil
				}

				if realVal == nil && must {
					if vv,ok := config[kEMPTY].(*Error); ok {
						return vv
					} else {
						errKey := ".item.empty"
						if vv,ok := config[kEMPTY].(string); ok {
							errKey = vv
						}
						return newError(errKey, itemName)
					}
				} else {

					//支持两种方式
					//1. data=table 如： "auth": "db.user"
					//2. base=db, table=user

					base, table := "",""
					if config["data"] != nil {
						if vv,ok := config["data"].(string); ok {
							i := strings.Index(vv, ".")
							base = vv[:i]
							table = vv[i+1:]
						}
					} else if config["base"]!=nil &&  config["table"]!=nil {
						if vv,ok := config["base"].(string); ok {
							base = vv
						}
						if vv,ok := config["table"].(string); ok {
							table = vv
						}
					}

					//判断是否需要查询数据
					if base!="" && table!="" {
						//要查询库
						db := ctx.dataBase(base)
						item := db.Table(table).Entity(realVal)
						if must && (ctx.Erred() != nil || item == nil) {
							if vv,ok := config[kERROR].(*Error); ok {
								return vv
							} else {
								errKey := ".item.error"
								if vv,ok := config[kERROR].(string); ok {
									errKey = vv
								}
								return newError(errKey, itemName)
							}
						} else {
							saveMap[itemKey] = item
						}
					}

				}
			}
		}

		//存入
		for k,v := range saveMap {
			ctx.Item[k] = v
		}
	}
	return nil
}



//返回最后的错误信息
func (ctx *Context) Erred() *Error {
	err := ctx.lastError
	ctx.lastError = nil
	if err != nil {
		err.Lang(ctx.Lang)
	}
	return err
}


//------------------------ 基础函数 end ----------------------








//----------------------- 签名系统 begin ---------------------------------
func (ctx *Context) signKey(key string) string {
	return fmt.Sprintf("$.sign.%s", key)
}
func (ctx *Context) Signed(key string) (bool) {
	key = ctx.signKey(key)
	if _,ok := ctx.Session[key]; ok {
		return true
	}
	return false
}
func (ctx *Context) Signin(key string, id,name Any) {
	key = ctx.signKey(key)
	ctx.Session[key] = Map{
		kID: fmt.Sprintf("%v", id),
		kNAME: fmt.Sprintf("%v", name),
	}
}
func (ctx *Context) Signout(key string) {
	key = ctx.signKey(key)
	delete(ctx.Session, key)
}
func (ctx *Context) Signal(key string) string {
	key = ctx.signKey(key)
	if vv,ok := ctx.Session[key].(Map); ok {
		if id,ok := vv[kID].(string); ok {
			return id
		}
	}
	return ""
}
func (ctx *Context) Signer(key string) string {
	key = ctx.signKey(key)
	if vv,ok := ctx.Session[key].(Map); ok {
		if id,ok := vv[kNAME].(string); ok {
			return id
		}
	}
	return ""
}

//----------------------- 签名系统 end ---------------------------------



//-------------------- 模块调用 begin ----------------------
func (ctx *Context) Service(name string, settings ...Map) (*serviceLogic) {
	return mSERVICE.newLogic(ctx, name, settings...)
}
func (ctx *Context) Invoke(name string, params ...Map) (Map) {
	ctx.lastError = nil

	param := Map{}
	if len(params) > 0 {
		param = params[0]
	}

	result,err := mSERVICE.Invoke(ctx, name, Map{}, param)
	if err != nil {
		ctx.lastError = err
	}
	return result
}
func (ctx *Context) Invokes(name string, params ...Map) ([]Map) {
	result := ctx.Invoke(name, params...)
	if result == nil {
		return []Map{}
	}
	if results,ok := result["items"].([]Map); ok {
		return results
	}
	return []Map{ result }
}
func (ctx *Context) Invoking(name string, offset, limit int64, params ...Map) (int64,[]Map) {
	param := Map{}
	if len(params) > 0 {
		param = params[0]
	}
	param["offset"] = offset
	param["limit"] = limit

	result := ctx.Invoke(name, param)
	if result == nil {
		return 0, []Map{}
	}

	count,countOK := result["count"].(int64)
	items,itemsOK := result["items"].([]Map)

	if countOK && itemsOK {
		return count, items
	}

	return 0, []Map{ result }
}
func (ctx *Context) Invoker(name string, params ...Map) (Map,[]Map) {
	result := ctx.Invoke(name, params...)
	if result == nil {
		return nil, []Map{}
	}

	item,itemOK := result["item"].(Map)
	items,itemsOK := result["items"].([]Map)

	if itemOK && itemsOK {
		return item, items
	}

	return result, []Map{ result }
}

func (ctx *Context) Download(code string, names ...string) {
	reader, data, err := Bigger.Download(code)
	if err != nil {
		ctx.Error(err)
	} else {
		ctx.Buffer(reader, data.Type, names...)
	}
}
func (ctx *Context) Thumbnail(code string, width,height,tttt int64) {
	reader, data, err := Bigger.Thumbnail(code, width, height, tttt)
	if err != nil {
		ctx.Error(err)
	} else {
		ctx.Buffer(reader, data.Type)
	}
}
//保存文件，返回code
func (ctx *Context) Storage(upload Map, named Any, metadata Map, bases ...string) (string) {
	ctx.lastError = nil

	ext := ""
	if vve,ok := upload["extension"].(string); ok && vve!="" {
		ext = vve
	}

	name := fmt.Sprintf("%s.%s", upload["hash"], ext)
	if vvn := fmt.Sprintf("%v", named); vvn != "" && named != nil {
		name = vvn
		if strings.Contains(name, ".") == false && ext != "" {
			name = fmt.Sprintf("%s.%s", name, ext)
		}
	}

	code, err := mFILE.Assign(name, Map{}, bases...)
	if err != nil {
		ctx.lastError = err
		return ""
	}

	//打开文件
	tempfile := upload["tempfile"].(string)
	fff,eee := os.Open(tempfile)
	if eee != nil {
		ctx.lastError = Bigger.Erred(eee)
		return ""
	}
	defer fff.Close()

	//保存文件
	_,err = mFILE.Storage(code, fff)
	if err != nil {
		ctx.lastError = err
		return ""
	}

	return code
}


func (ctx *Context) Upgrading(flags ...string) string {
	if ctx.mode != httpMode {
		panic("[上下文]非HTTP上下文")
	}
	flag := ""
	if len(flags) > 0 {
		flag = flags[0]
	}

	

	code := mSOCKET.Assign(flag, ctx.Id)

	//临时存一下
	ctx.store["upgrading"] = code

	return code
}
func (ctx *Context) Upgrade(codes ...string) {
	if ctx.mode != httpMode {
		panic("[上下文]非HTTP上下文")
	}

	code := ""
	if len(codes) > 0 {
		code = codes[0]
	} else if vv,ok := ctx.store["upgrading"].(string); ok && vv != "" {
		code = vv
	} else {
		code = mSOCKET.Assign("", ctx.Id)
	}


	ctx.lastError = nil
	err := mSOCKET.Upgrade(code, ctx.http.req.Reader, ctx.http.req.Writer)
	if err != nil {
		ctx.lastError = err
	} else {
		ctx.Text("connected")
	}
}
func (ctx *Context) Degrade() {
	if ctx.mode != socketMode {
		panic("[上下文]非SOCKET上下文")
	}

	ctx.lastError = nil
	ctx.lastError = mSOCKET.Degrade(ctx.Id)
}
func (ctx *Context) Follow(channel string) {
	if ctx.mode != socketMode {
		panic("[上下文]非SOCKET上下文")
	}

	ctx.lastError = nil
	ctx.lastError = mSOCKET.Follow(ctx.Id, channel)
}
func (ctx *Context) Unfollow(channel string) {
	if ctx.mode != socketMode {
		panic("[上下文]非SOCKET上下文")
	}

	ctx.lastError = nil
	ctx.lastError = mSOCKET.Unfollow(ctx.Id, channel)
}
//-------------------- 模块调用 end ----------------------





//接入错误处理流程，和模块挂钩了
func (ctx *Context) Found() {
	switch ctx.mode {
	case planMode:
		mPLAN.found(ctx)
	case eventMode:
		mEVENT.found(ctx)
	case queueMode:
		mQUEUE.found(ctx)
	case httpMode:
		mHTTP.found(ctx)
	default:
	}
}
func (ctx *Context) Error(err *Error) {
	ctx.lastError = err

	switch ctx.mode {
	case planMode:
		mPLAN.error(ctx)
	case eventMode:
		mEVENT.error(ctx)
	case queueMode:
		mQUEUE.error(ctx)
	case httpMode:
		mHTTP.error(ctx)
	default:
	}
}
func (ctx *Context) Failed(err *Error) {
	ctx.lastError = err

	switch ctx.mode {
	case planMode:
		mPLAN.failed(ctx)
	case eventMode:
		mEVENT.failed(ctx)
	case queueMode:
		mQUEUE.failed(ctx)
	case httpMode:
		mHTTP.failed(ctx)
	default:
	}
}
func (ctx *Context) Denied(err *Error) {
	ctx.lastError = err

	switch ctx.mode {
	case planMode:
		mPLAN.denied(ctx)
	case eventMode:
		mEVENT.denied(ctx)
	case queueMode:
		mQUEUE.denied(ctx)
	case httpMode:
		mHTTP.denied(ctx)
	default:
	}
}







func (ctx *Context) Reply(msg string, datas ...Map) {
	if ctx.mode != socketMode {
		panic("非SOCKET下上文")
	}
	ctx.lastError = nil

	data := Map{}
	if len(datas) > 0 {
		data = datas[0]
	}

	if err := mSOCKET.Message(ctx.Id, msg, data); err != nil {
		ctx.lastError = err
	}
}


func (ctx *Context) Finish() {
	ctx.Body = finishBody{}
}
func (ctx *Context) Delay(delay time.Duration) {
	ctx.Body = delayBody{delay}
}


//获取langString
func (ctx *Context) String(key string, args ...Any) string {
	return mCONST.LangString(ctx.Lang, key, args...)
}




//通用方法
func (ctx *Context) Header(key string, vals ...string) string {
	if ctx.mode != httpMode {
		return ""
	}

	if len(vals) > 0 {
		ctx.headers[key] = vals[0]
		return vals[0]
	} else {
		//读header
		return ctx.http.req.Reader.Header.Get(key)
	}
}

//通用方法
func (ctx *Context) Cookie(key string, vals ...Any) string {
	if ctx.mode != httpMode {
		return ""
	}

	if len(vals) > 0 {
		
		//设置header
		switch val := vals[0].(type) {
			case http.Cookie:
				val.Value = url.QueryEscape(val.Value)
				ctx.cookies[key] = val
			case string:
				cookie := http.Cookie{ Name: key, Value: url.QueryEscape(val), Path: "/", HttpOnly: true }
				ctx.cookies[key] = cookie
			default:
				return ""
		}

	} else {
		//读cookie
		c,e := ctx.http.req.Reader.Cookie(key)
		if e == nil {
			//加密cookie
			return Bigger.Decrypt(c.Value)
		}
	}
	return ""
}






func (ctx *Context) Goto(url string) {
	if ctx.mode != httpMode {
		panic("[上下文]非HTTP上下文")
	}

	//如果已经存在了httpDownBody，那还要把原有的reader关闭
	//释放资源， 当然在file.base.close中也应该关闭已经打开的资源
	if vv,ok := ctx.Body.(httpBufferBody); ok {
		vv.buffer.Close()
	}

	ctx.Body = httpGotoBody{url}
}
func (ctx *Context) Goback() {
	url := ctx.Url.Back()
	ctx.Goto(url)
}
func (ctx *Context) Text(text Any, codes ...int) {
	if ctx.mode != httpMode {
		panic("[上下文]非HTTP上下文")
	}


	//如果已经存在了httpDownBody，那还要把原有的reader关闭
	//释放资源， 当然在file.base.close中也应该关闭已经打开的资源
	if vv,ok := ctx.Body.(httpBufferBody); ok {
		vv.buffer.Close()
	}

	if len(codes) > 0 {
		ctx.Code = codes[0]
	}
	ctx.Type = "text"
	ctx.Body = httpTextBody{fmt.Sprintf("%v", text)}
}
func (ctx *Context) Html(html string, codes ...int) {
	if ctx.mode != httpMode {
		panic("[上下文]非HTTP上下文")
	}


	//如果已经存在了httpDownBody，那还要把原有的reader关闭
	//释放资源， 当然在file.base.close中也应该关闭已经打开的资源
	if vv,ok := ctx.Body.(httpBufferBody); ok {
		vv.buffer.Close()
	}
	
	if len(codes) > 0 {
		ctx.Code = codes[0]
	}
	ctx.Type = "html"
	ctx.Body = httpHtmlBody{html}
}
func (ctx *Context) Script(script string, codes ...int) {
	if ctx.mode != httpMode {
		panic("[上下文]非HTTP上下文")
	}


	//如果已经存在了httpDownBody，那还要把原有的reader关闭
	//释放资源， 当然在file.base.close中也应该关闭已经打开的资源
	if vv,ok := ctx.Body.(httpBufferBody); ok {
		vv.buffer.Close()
	}


	
	if len(codes) > 0 {
		ctx.Code = codes[0]
	}
	ctx.Type = "script"
	ctx.Body = httpScriptBody{script}
}
func (ctx *Context) Json(json Any, codes ...int) {
	if ctx.mode != httpMode {
		panic("[上下文]非HTTP上下文")
	}


	//如果已经存在了httpDownBody，那还要把原有的reader关闭
	//释放资源， 当然在file.base.close中也应该关闭已经打开的资源
	if vv,ok := ctx.Body.(httpBufferBody); ok {
		vv.buffer.Close()
	}


	
	if len(codes) > 0 {
		ctx.Code = codes[0]
	}
	ctx.Type = "json"
	ctx.Body = httpJsonBody{json}
}
func (ctx *Context) Jsonp(callback string, json Any, codes ...int) {
	if ctx.mode != httpMode {
		panic("[上下文]非HTTP上下文")
	}


	//如果已经存在了httpDownBody，那还要把原有的reader关闭
	//释放资源， 当然在file.base.close中也应该关闭已经打开的资源
	if vv,ok := ctx.Body.(httpBufferBody); ok {
		vv.buffer.Close()
	}
	
	if len(codes) > 0 {
		ctx.Code = codes[0]
	}
	ctx.Type = "jsonp"
	ctx.Body = httpJsonpBody{json, callback}
}
func (ctx *Context) Xml(xml Any, codes ...int) {
	if ctx.mode != httpMode {
		panic("[上下文]非HTTP上下文")
	}

	//如果已经存在了httpDownBody，那还要把原有的reader关闭
	//释放资源， 当然在file.base.close中也应该关闭已经打开的资源
	if vv,ok := ctx.Body.(httpBufferBody); ok {
		vv.buffer.Close()
	}
	
	if len(codes) > 0 {
		ctx.Code = codes[0]
	}
	ctx.Type = "xml"
	ctx.Body = httpXmlBody{xml}
}


func (ctx *Context) File(file string, mimeType string, names ...string) {
	if ctx.mode != httpMode {
		panic("[上下文]非HTTP上下文")
	}
	

	//如果已经存在了httpDownBody，那还要把原有的reader关闭
	//释放资源， 当然在file.base.close中也应该关闭已经打开的资源
	if vv,ok := ctx.Body.(httpBufferBody); ok {
		vv.buffer.Close()
	}

	name := ""
	if len(names) > 0 {
		name = names[0]
	}
	if mimeType != "" {
		ctx.Type = mimeType
	} else {
		ctx.Type = "file"
	}
	ctx.Body = httpFileBody{file,name}
}


func (ctx *Context) Buffer(rd io.ReadCloser, mimeType string, names ...string) {
	if ctx.mode != httpMode {
		panic("[上下文]非HTTP上下文")
	}
	
	//如果已经存在了httpDownBody，那还要把原有的reader关闭
	//释放资源， 当然在file.base.close中也应该关闭已经打开的资源
	if vv,ok := ctx.Body.(httpBufferBody); ok {
		vv.buffer.Close()
	}

	name := ""
	if len(names) > 0 {
		name = names[0]
	}

	ctx.Code = http.StatusOK
	if mimeType != "" {
		ctx.Type = mimeType
	} else {
		ctx.Type = "file"
	}
	ctx.Body = httpBufferBody{rd, name}
}
func (ctx *Context) Down(bytes []byte, mimeType string, names ...string) {
	if ctx.mode != httpMode {
		panic("[上下文]非HTTP上下文")
	}
	if vv,ok := ctx.Body.(httpBufferBody); ok {
		vv.buffer.Close()
	}

	ctx.Code = http.StatusOK
	if mimeType != "" {
		ctx.Type = mimeType
	} else {
		ctx.Type = "file"
	}
	name := ""
	if len(names) > 0 {
		name = names[0]
	}
	ctx.Body = httpDownBody{bytes, name}
}

func (ctx *Context) View(view string, types ...string) {
	if ctx.mode != httpMode {
		panic("[上下文]非HTTP上下文")
	}

	//如果已经存在了httpDownBody，那还要把原有的reader关闭
	//释放资源， 当然在file.base.close中也应该关闭已经打开的资源
	if vv,ok := ctx.Body.(httpBufferBody); ok {
		vv.buffer.Close()
	}

	
	ctx.Type = "html"
	if len(types) > 0 {
		ctx.Type = types[0]
	}
	ctx.Body = httpViewBody{view, Map{}}
}


func (ctx *Context) Route(name string, values ...Map) {
	if ctx.mode != httpMode {
		panic("[上下文]非HTTP上下文")
	}
	url := ctx.Url.Route(name, values...)
	ctx.Redirect(url)
}

func (ctx *Context) Redirect(url string) {
	if ctx.mode != httpMode {
		panic("[上下文]非HTTP上下文")
	}

	//如果已经存在了httpDownBody，那还要把原有的reader关闭
	//释放资源， 当然在file.base.close中也应该关闭已经打开的资源
	if vv,ok := ctx.Body.(httpBufferBody); ok {
		vv.buffer.Close()
	}
	
	ctx.Goto(url)
}

func (ctx *Context) Alert(err *Error, urls ...string) {
	if ctx.mode != httpMode {
		panic("[上下文]非HTTP上下文")
	}

	//如果已经存在了httpDownBody，那还要把原有的reader关闭
	//释放资源， 当然在file.base.close中也应该关闭已经打开的资源
	if vv,ok := ctx.Body.(httpBufferBody); ok {
		vv.buffer.Close()
	}
	
	if err.Code() == 0 {
		ctx.Code = http.StatusOK
	} else {
		ctx.Code = http.StatusInternalServerError
	}
	text := err.Lang(ctx.Lang).String()

	if len(urls) > 0 {
		text = fmt.Sprintf(`<script type="text/javascript">alert("%s"); location.href="%s";</script>`, text, urls[0])
	} else {
		text = fmt.Sprintf(`<script type="text/javascript">alert("%s"); history.back();</script>`, text)
	}
	ctx.Script(text)
}
//展示通用的提示页面
func (ctx *Context) Show(err *Error, urls ...string) {
	if ctx.mode != httpMode {
		panic("[上下文]非HTTP上下文")
	}

	code := err.Code()
	text := err.Lang(ctx.Lang).String()

	if err.Code() == 0 {
		ctx.Code = http.StatusOK
	} else {
		ctx.Code = http.StatusInternalServerError
	}

	m := Map{
		"code": code,
		"text": text,
		"url": "",
	}
	if len(urls) > 0 {
		m["url"] = urls[0]
	}

	ctx.Data[kSHOW] = m
	ctx.View(kSHOW)
}


//返回操作结果，表示成功
//比如，登录，修改密码，等操作类的接口， 成功的时候，使用这个，
//args表示返回给客户端的data
//data 强制改为json格式，因为data有统一加密的可能
//所有数组都要加密。
func (ctx *Context) Result(err *Error, args ...Map) {
	if ctx.mode != httpMode {
		panic("[上下文]非HTTP上下文")
	}

	//如果已经存在了httpDownBody，那还要把原有的reader关闭
	//释放资源， 当然在file.base.close中也应该关闭已经打开的资源
	if vv,ok := ctx.Body.(httpBufferBody); ok {
		vv.buffer.Close()
	}

	code := 0
	text := ""

	if err != nil {
		code = err.Code()
		text = err.Lang(ctx.Lang).String()
		if code == -1 {
			code = 0
		}
	}

	if code == 0 {
		ctx.Code = http.StatusOK
	} else {
		ctx.Code = http.StatusInternalServerError
	}

	var data Map
	if len(args) > 0 {
		data = args[0]
	}

	ctx.Type = "json"
	ctx.Body = httpApiBody{code, text, data}
}

//返回数据，表示成功
//data必须为json，因为data节点可能统一加密
//如果在data同级返回其它数据，如page信息， 会有泄露数据风险
//所以这里强制data必须为json
func (ctx *Context) Answer(data Map) {
	if ctx.mode != httpMode {
		panic("[上下文]非HTTP上下文")
	}

	//如果已经存在了httpDownBody，那还要把原有的reader关闭
	//释放资源， 当然在file.base.close中也应该关闭已经打开的资源
	if vv,ok := ctx.Body.(httpBufferBody); ok {
		vv.buffer.Close()
	}

	ctx.Type = "json"
	ctx.Code = http.StatusOK
	ctx.Body = httpApiBody{0, "", data}
}



//通用方法
func (ctx *Context) UserAgent() string {
	return ctx.Header("User-Agent")
}
func (ctx *Context) Ip() string {
	ip := "127.0.0.1"

	if ctx.mode == httpMode {
		req := ctx.http.req.Reader

		if realIp := req.Header.Get("X-Real-IP"); realIp != "" {
			ip = realIp
		} else if forwarded := req.Header.Get("x-forwarded-for"); forwarded != "" {
			ip = forwarded
		} else {
	
			newip,_,err := net.SplitHostPort(req.RemoteAddr)
			if err == nil {
				ip = newip
			}
	
		}
	}


	return ip
}





//--------------------- httpUrl begin ----------------------------------


//可否智能判断是否跨站返回URL
func (url *contextUrl) Route(name string, values ...Map) string {

	if strings.HasPrefix(name, "http://") || strings.HasPrefix(name, "https://") ||
	strings.HasPrefix(name, "ws://") || strings.HasPrefix(name, "wss://") {
		return name
	}
	
	//当前站点
	currSite := ""
	if url.ctx != nil {
		currSite = url.ctx.Site
		if name == "" {
			name = url.ctx.Name
		}
	}


	params,querys,options := Map{},Map{},Map{}
	if len(values) > 0 {
		for k,v := range values[0] {
			if strings.HasPrefix(k, "{") && strings.HasSuffix(k, "}") {
				params[k] = v
			} else if strings.HasPrefix(k, "[") && strings.HasSuffix(k, "]") {
				options[k] = v
			} else {
				querys[k] = v
			}
		}
	}


	// justSite, justName := "", ""
	justSite := ""
	if strings.Contains(name, ".") {
		i := strings.Index(name, ".")
		justSite = name[:i]
		// justName = name[i+1:]
	}

	//如果是*.开头，因为在file.driver里可能定义*.xx.xxx.xx做为路由访问文件
	if justSite == "*" {
		if currSite != "" {
			justSite = currSite
		} else {
			//只能随机选一个站点了
			for site,_ := range Bigger.Config.Site {
				justSite = site
				break
			}
		}
		name = strings.Replace(name, "*", justSite, 1)
	}

	//如果是不同站点，强制带域名
	if justSite != currSite {
		options["[site]"] = justSite
	} else if options["[site]"] != nil {
		options["[site]"] = currSite
	}

	
	
	nameget := fmt.Sprintf("%s.get", name)
	namepost := fmt.Sprintf("%s.post", name)
	var config Map

	//搜索定义
	if vv,ok := mHTTP.router.chunkdata(name).(Map); ok {
		config = vv
	} else if vv,ok := mHTTP.router.chunkdata(nameget).(Map); ok {
		config = vv
	} else if vv,ok := mHTTP.router.chunkdata(namepost).(Map); ok {
		config = vv
	} else {
		//没有找到路由定义
		return name
	}


	if config["socket"] != nil {
		options["[socket]"] = true
	}

	uri := ""
	if vv,ok := config["uri"].(string); ok {
		uri = vv
	} else if vv,ok := config["uris"].([]string); ok && len(vv)>0 {
		uri = vv[0]
	} else {
		return "[no uri here]"
	}


	argsConfig := Map{}
	if c,ok := config["args"].(Map); ok {
		argsConfig = c
	}



		//选项处理
		if options["[back]"] != nil && url.ctx != nil {
			var url = url.Back()
			querys[BACKURL] = Bigger.Encrypt(url)
		}
		//选项处理
		if options["[last]"] != nil && url.ctx != nil {
			var url = url.Last()
			querys[BACKURL] = Bigger.Encrypt(url)
		}
		//选项处理
		if options["[current]"] != nil && url.ctx != nil {
			var url = url.Current()
			querys[BACKURL] = Bigger.Encrypt(url)
		}
		//自动携带原有的query信息
		if options["[query]"] != nil && url.ctx != nil {
			for k,v := range url.ctx.Query {
				querys[k] = v
			}
		}


		//所以，解析uri中的参数，值得分几类：
		//1传的值，2param值, 3默认值
		//其中主要问题就是，传的值，需要到args解析，用于加密，这个值和auto值完全重叠了，除非分2次解析
		//为了框架好用，真是操碎了心
		dataValues, paramValues, autoValues := Map{},Map{},Map{}

		//1. 处理传过来的值
		//从value中获取
		//如果route不定义args，这里是拿不到值的
		dataArgsValues, dataParseValues := Map{},Map{}
		for k,v := range params {
			if k[0:1] == "{" {
				k = strings.Replace(k, "{","",-1)
				k = strings.Replace(k, "}","",-1)
				dataArgsValues[k] = v
			} else {
				//这个也要？要不然指定的一些page啥的不行？
				dataArgsValues[k] = v
				//另外的是query的值
				querys[k] = v
			}
		}
		dataErr := mMAPPING.Parse(argsConfig, dataArgsValues, dataParseValues, false, true)
		if dataErr == nil {
			for k,v := range dataParseValues {

				//注意，这里能拿到的，还有非param，所以不能直接用加{}写入
				if _,ok := params[k]; ok {
					dataValues[k] = v
				} else if _,ok := params["{"+k+"}"]; ok {
					dataValues["{"+k+"}"] = v
				} else {
					//这里是默认值应该，就不需要了
				}
			}
		}


		//所以这里还得处理一次，如果route不定义args，parse就拿不到值，就直接用values中的值
		for k,v := range params {
			if k[0:1] == "{" && dataValues[k] == nil {
				dataValues[k] = v
			}
		}

		//2.params中的值
		//从params中来一下，直接参数解析
		if url.ctx != nil {
			for k,v := range url.ctx.Param {
				paramValues["{"+k+"}"] = v
			}
		}


		//3. 默认值
		//从value中获取
		autoArgsValues, autoParseValues := Map{},Map{}
		autoErr := mMAPPING.Parse(argsConfig, autoArgsValues, autoParseValues, false, true)
		if autoErr == nil {
			for k,v := range autoParseValues {
				autoValues["{"+k+"}"] = v
			}
		}

		//开始替换值
		regx := regexp.MustCompile(`\{[_\*A-Za-z0-9]+\}`)
		uri = regx.ReplaceAllStringFunc(uri, func(p string) string {
			key := strings.Replace(p, "*", "", -1)

			if v,ok := dataValues[key]; ok {
				//for query string encode/decode
				delete(dataValues, key)
				//先从传的值去取
				return fmt.Sprintf("%v", v)
			} else if v,ok := paramValues[key]; ok {
				//再从params中去取
				return fmt.Sprintf("%v", v)
			} else if v,ok := autoValues[key]; ok {
				//最后从默认值去取
				return fmt.Sprintf("%v", v)
			} else {
				//有参数没有值,
				return p
			}
		})


		//get参数，考虑一下走mapping，自动加密参数不？
		queryStrings := []string{}
		for k,v := range querys {
			sv := fmt.Sprintf("%v", v)
			if sv != "" {
				queryStrings = append(queryStrings, fmt.Sprintf("%v=%v", k, v))
			}
		}
		if len(queryStrings) > 0 {
			uri += "?" + strings.Join(queryStrings, "&")
		}

		if site,ok := options["[site]"].(string); ok && site != "" {
			uri = url.Site(site, uri, options)
		}

	return uri
}






















func (url *contextUrl) Site(name string, path string, options ...Map) string {
	option := Map{}
	if len(options) > 0 {
		option = options[0]
	}

	uuu := ""
	ssl, socket := false, false

	//如果有上下文，如果是当前站点，就使用当前域
	if url.ctx != nil && url.ctx.Site == name {
		uuu = url.ctx.Host
	} else if vv,ok := Bigger.Config.Site[name]; ok {
		ssl = vv.Ssl
		if len(vv.Hosts) > 0 {
			uuu = vv.Hosts[0]
		}
	} else {
		uuu = fmt.Sprintf("127.0.0.1:%v", Bigger.Config.Node.Port)
	}

	if option["[ssl]"] != nil {
		ssl = true
	}
	if option["[socket]"] != nil {
		socket = true
	}

	if socket {
		if ssl { uuu = "wss://" + uuu } else { uuu = "ws://" + uuu }
	} else {
		if ssl { uuu = "https://" + uuu } else { uuu = "http://" + uuu }
	}
	
	if path != "" {
		return fmt.Sprintf("%s%s", uuu, path)
	} else {
		return uuu
	}
}


func (url *contextUrl) Backing() bool {
	if url.req == nil {
		return false
	}

	if s,ok := url.ctx.Query[BACKURL]; ok && s != "" {
		return true
	} else if url.req.Referer() != "" {
		return true
	}
	return false
}


func (url *contextUrl) Back() string {
	if url.ctx == nil {
		return "/"
	}

	if s,ok := url.ctx.Query[BACKURL].(string); ok && s != "" {
		return Bigger.Decrypt(s)
	} else if url.ctx.Header("referer") != "" {
		return url.ctx.Header("referer")
	} else {
		//都没有，就是当前URL
		return url.Current()
	}
}




func (url *contextUrl) Last() string {
	if url.req == nil {
		return "/"
	}

	if ref := url.req.Referer(); ref != "" {
		return ref
	} else {
		//都没有，就是当前URL
		return url.Current()
	}
}



func (url *contextUrl) Current() string {
	if url.req == nil {
		return "/"
	}
	
	// return url.req.URL.String()

	// return fmt.Sprintf("%s://%s%s", url.req.URL., url.req.Host, url.req.URL.RequestURI())


	return url.Site(url.ctx.Site, url.req.URL.RequestURI())
}



//为了view友好，expires改成Any，支持duration解析
func (url *contextUrl) Download(target Any, name string, args ...Any) string {
	if url.ctx == nil {
		return ""
	}

	if coding,ok := target.(string); ok && coding != "" {

		if strings.HasPrefix("http://", coding) || strings.HasPrefix("https://", coding) || strings.HasPrefix("ftp://", coding){
			return coding
		}


		expires := []time.Duration{}
		if len(args) > 0 {
			switch vv := args[0].(type) {
			case int:
				expires = append(expires, time.Second*time.Duration(vv))
			case time.Duration:
				expires = append(expires, vv)
			case string:
				if dd,ee := Bigger.Timing(vv); ee == nil {
					expires = append(expires, dd)
				}
			}
		}

		url.ctx.lastError = nil
		aaaaa := Map{
			"id": url.ctx.Id,
			"ip": url.ctx.Ip(),
		}
		if uuu,err := mFILE.Browse(coding, name, aaaaa, expires...); err != nil {
			url.ctx.lastError = err
			return ""
		} else {
			return uuu
		}
	}

	return "[无效下载]"
}
func (url *contextUrl) Browse(target Any, args ...Any) string {
	if url.ctx == nil {
		return ""
	}

	if coding,ok := target.(string); ok && coding != "" {

		if strings.HasPrefix("http://", coding) || strings.HasPrefix("https://", coding) || strings.HasPrefix("ftp://", coding){
			return coding
		}

		expires := []time.Duration{}
		if len(args) > 0 {
			switch vv := args[0].(type) {
			case int:
				expires = append(expires, time.Second*time.Duration(vv))
			case time.Duration:
				expires = append(expires, vv)
			case string:
				if dd,ee := Bigger.Timing(vv); ee == nil {
					expires = append(expires, dd)
				}
			}
		}

		url.ctx.lastError = nil
		aaaaa := Map{
			"id": url.ctx.Id,
			"ip": url.ctx.Ip(),
		}
		if uuu,err := mFILE.Browse(coding, "", aaaaa, expires...); err != nil {
			url.ctx.lastError = err
			return ""
		} else {
			return uuu
		}
	}

	return "[无效文件]"
}
func (url *contextUrl) Preview(target Any, width,height,tttt int64, args ...Any) string {
	if url.ctx == nil {
		return ""
	}

	if coding,ok := target.(string); ok && coding != "" {

		if strings.HasPrefix("http://", coding) || strings.HasPrefix("https://", coding) || strings.HasPrefix("ftp://", coding){
			return coding
		}

		expires := []time.Duration{}
		if len(args) > 0 {
			switch vv := args[0].(type) {
			case int:
				expires = append(expires, time.Second*time.Duration(vv))
			case time.Duration:
				expires = append(expires, vv)
			case string:
				if dd,ee := Bigger.Timing(vv); ee == nil {
					expires = append(expires, dd)
				}
			}
		}

		url.ctx.lastError = nil
		aaaaa := Map{
			"id": url.ctx.Id,
			"ip": url.ctx.Ip(),
		}
		if uuu,err := mFILE.Preview(coding, width, height, tttt, aaaaa, expires...); err != nil {
			url.ctx.lastError = err
			return ""
		} else {
			return uuu
		}
	}

	return "/no.png"
}
//--------------------- httpUrl end ----------------------------------