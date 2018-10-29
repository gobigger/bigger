package bigger

import (
	"net/url"
	"strings"
	"net/http"
	"fmt"
	"time"
	"encoding/json"
	"encoding/xml"
	"crypto/sha1"
	"io/ioutil"
	"io"
	"os"
	"path"
	"github.com/gobigger/bigger/toml"
)


// http driver begin
type (
	//事件驱动
	HttpDriver interface {
		Connect(HttpConfig) (HttpConnect,*Error)
	}

	//事件连接
	HttpConnect interface {
		Open() *Error
        Health() (*HttpHealth,*Error)
		Close() *Error

		Accept(HttpHandler) *Error
		Register(name string, config HttpRegister) *Error

		//开始
		Start() *Error
		//开始TLS
		StartTLS(certFile, keyFile string) *Error
	}

	//事件处理器
	HttpHandler		func(*HttpRequest, HttpResponse)
	HttpRegister	struct {
		Site		string
		Uris		[]string
		Methods		[]string
		Hosts		[]string
	}

	HttpHeaders		map[string]string
	HttpCookies		map[string]http.Cookie


    HttpHealth struct {
        Workload    int64
    }

	//事件请求实体
	HttpRequest struct {
		Id		string
		Name	string
		Site	string
		Params	Map

		Reader *http.Request
		Writer http.ResponseWriter
	}
	//事件响应接口
	HttpResponse interface {
		Finish(*HttpRequest) *Error
	}

	//跳转
	httpGotoBody struct {
		url		string
	}
	httpTextBody struct {
		text	string
	}
	httpHtmlBody struct {
		html	string
	}
	httpScriptBody struct {
		script	string
	}
	httpJsonBody struct {
		json	Any
	}
	httpJsonpBody struct {
		json		Any
		callback	string
	}
	httpApiBody struct {
		code	int
		text	string
		data	Map
	}
	httpXmlBody struct {
		xml		Any
	}
	httpFileBody struct {
		file	string
		name	string
	}
	httpDownBody struct {
		bytes	[]byte
		name	string
	}
	httpBufferBody struct {
		buffer	io.ReadCloser
		name	string
	}
	httpViewBody struct {
		view	string
		model	Any
	}

)



type (
    httpModule struct {
		driver		coreBranch
		router		coreBranch
		filter		coreBranch
		handler		coreBranch
		
		config		HttpConfig
		connect		HttpConnect
	}


	httpGroup struct {
		http	*httpModule
		name	string
		root	string
	}

)

func (module *httpModule) newGroup(name string, roots ...string) (*httpGroup) {
	root := ""
	if len(roots) > 0 {
		root = strings.TrimRight(roots[0], "/")
	}
	return &httpGroup{ module, name, root }
}
//HTTP分组注册
func (group *httpGroup) Router(name string, config Map, overrides ...bool) {
	realName := fmt.Sprintf("%s.%s", group.name, name)
	if group.root != "" {
		if uri,ok := config[kURI].(string); ok {
			config[kURI] = group.root + uri
		} else if uris,ok := config[kURIS].([]string); ok {
			for i,uri := range uris {
				uris[i] = group.root + uri
			}
			config[kURIS] = uris
		}
	}
	group.http.Router(realName, config, overrides...)
}


func (module *httpModule) Driver(name string, driver HttpDriver, overrides ...bool) {
    if driver == nil {
        panic("[HTTP]驱动不可为空")
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




//因为router.key要返回，所以不能用数组
func (module *httpModule) Routers(sites ...string) ([]KVPair) {
	prefixs := []string{}
	if len(sites) > 0 {
		prefixs = append(prefixs, sites[0] + ".")
	}
	chunks := module.router.chunks(prefixs...)
	routers := []KVPair{}
	for _,chunk := range chunks {
		routers = append(routers, KVPair{chunk.name, chunk.data})
	}
	return routers
}
func (module *httpModule) Router(name string, config Map, overrides ...bool) {

    override := true
    if len(overrides) > 0 {
        override = overrides[0]
	}
	
	//直接的时候直接拆分成目标格式
	objects := map[string]Map{}
	if strings.HasPrefix(name, "*.") {
		//全站点
		for site,_ := range Bigger.Config.Site {
			siteName := strings.Replace(name, "*", site, 1)
			siteConfig := Map{}
			
			//复制配置
			for k,v := range config {
				siteConfig[k] = v
			}
			//站点名
			siteConfig[kSITE] = site

			//先记录下
			objects[siteName] = siteConfig
		}
	} else {
		//单站点
		objects[name] = config
	}




	//处理对方是单方法，还是多方法
	routers := map[string]Map{}
	for routerName,routerConfig := range objects {

		if routeConfig, ok := routerConfig[kROUTE].(Map); ok {
			//多method版本
			for method,vvvv := range routeConfig {
				if methodConfig,ok := vvvv.(Map); ok {

					realName := fmt.Sprintf("%s.%s", routerName, method)
					realConfig := Map{}

					//复制全局的定义
					for k,v := range routerConfig {
						if k != kROUTE {
							realConfig[k] = v
						}
					}

					//复制子级的定义
					//注册,args, auth, item等
					for k,v := range methodConfig {
						if lllMap,ok := v.(Map); ok && (k==kARGS || k==kAUTH || k==kITEM) {
							if gggMap,ok := realConfig[k].(Map); ok {

								newMap := Map{}
								//复制全局
								for gk,gv := range gggMap {
									newMap[gk] = gv
								}
								//复制方法级
								for lk,lv := range lllMap {
									newMap[lk] = lv
								}

								realConfig[k] = newMap

							} else {
								realConfig[k] = v
							}
						} else {
							realConfig[k] = v
						}
					}

					//相关参数
					realConfig[kMETHOD] = method

					//加入列表
					routers[realName] = realConfig
				}
			}

		} else {

			//单方法版本
			realName := routerName
			realConfig := Map{}

			//复制定义	
			for k,v := range routerConfig {
				realConfig[k] = v
			}

			//加入列表
			routers[realName] = realConfig
		}
	}

	//这里才是真的注册
	for k,v := range routers {
		if override {
			module.router.chunking(k, v)
		} else {
			if module.router.chunkdata(k) == nil {
				module.router.chunking(k, v)
			}
		}
	}
}
func (module *httpModule) Filter(name string, config Map, overrides ...bool) {
    override := true
    if len(overrides) > 0 {
        override = overrides[0]
    }

	//直接的时候直接拆分成目标格式
	filters := map[string]Map{}
	if strings.HasPrefix(name, "*.") {
		//全站点
		for site,_ := range Bigger.Config.Site {
			siteName := strings.Replace(name, "*", site, 1)
			siteConfig := Map{}
			
			//复制配置
			for k,v := range config {
				siteConfig[k] = v
			}
			//站点名
			siteConfig[kSITE] = site

			//先记录下
			filters[siteName] = siteConfig
		}
	} else {
		//单站点
		filters[name] = config
	}

	//这里才是真的注册
	for k,v := range filters {
		if override {
			module.filter.chunking(k, v)
		} else {
			if module.filter.chunkdata(k) == nil {
				module.filter.chunking(k, v)
			}
		}
	}
}
func (module *httpModule) Handler(name string, config Map, overrides ...bool) {
    override := true
    if len(overrides) > 0 {
        override = overrides[0]
	}
	
	//直接的时候直接拆分成目标格式
	handlers := map[string]Map{}
	if strings.HasPrefix(name, "*.") {
		//全站点
		for site,_ := range Bigger.Config.Site {
			siteName := strings.Replace(name, "*", site, 1)
			siteConfig := Map{}

			//复制配置
			for k,v := range config {
				siteConfig[k] = v
			}
			//站点名
			siteConfig[kSITE] = site

			//先记录下
			handlers[siteName] = siteConfig
		}
	} else {
		//单站点
		handlers[name] = config
	}

	//这里才是真的注册
	for k,v := range handlers {
		if override {
			module.handler.chunking(k, v)
		} else {
			if module.handler.chunkdata(k) == nil {
				module.handler.chunking(k, v)
			}
		}
	}
}








func (module *httpModule) connecting(config HttpConfig) (HttpConnect,*Error) {
    if driver,ok := module.driver.chunkdata(config.Driver).(HttpDriver); ok {
        return driver.Connect(config)
    }
    panic("[HTTP]不支持的驱动：" + config.Driver)
}
func (module *httpModule) initing() {

    //连接
    config := Bigger.Config.Http
    connect,err := module.connecting(config)
    if err != nil {
        panic("[HTTP]连接失败：" + err.Error())
    }
    
    //打开连接
    err = connect.Open()
    if err != nil {
        panic("[HTTP]打开失败：" + err.Error())
    }


	//遍历站点去注册路由
	for site,_ := range Bigger.Config.Site {

		//注册路由
		locals := module.router.chunks(fmt.Sprintf("%s.", site))
		for _,v := range locals {
			if vv,ok := v.data.(Map); ok {
				regis := module.registering(site, vv)
				err := connect.Register(v.name, regis)
				if err != nil {
					panic("[HTTP]注册失败：" + err.Error())
				}
			}
		}
	}

	//绑定回调
	connect.Accept(module.serve)

	err = connect.Start()
	if err != nil {
		panic("[HTTP]启用失败：" + err.Error())
	}

    //保存连接
    module.config = config
    module.connect = connect
	
}



func (module *httpModule) registering(site string, config Map) (HttpRegister) {

	//Uris
	uris := []string{}
	if vv,ok := config[kURI].(string); ok && vv != "" {
		uris = append(uris, vv)
	}
	if vv,ok := config[kURIS].([]string); ok {
		uris = append(uris, vv...)
	}
	//方法
	methods := []string{}
	if vv,ok := config[kMETHOD].(string); ok && vv != "" {
		methods = append(methods, vv)
	}
	if vv,ok := config[kMETHODS].([]string); ok {
		methods = append(methods, vv...)
	}

	regis := HttpRegister{ Site: site, Uris: uris, Methods: methods }

	if cfg,ok := Bigger.Config.Site[site]; ok {
		regis.Hosts = cfg.Hosts
	}

	return regis
}


//退出
func (module *httpModule) exiting() {
    if module.connect != nil {
		module.connect.Close()
	}
}


func (module *httpModule) newbie(newbie coreNewbie) (*Error) {
	if newbie.branch != bHTTPROUTER {
		return nil
	}

	//拿到站点名称, 再去拿站点配置，要拿域名列表
	keys := strings.Split(newbie.block, ".")
	site := keys[0]

	if config,ok := module.router.chunkdata(newbie.block).(Map); ok {
		name := newbie.block
		regis := module.registering(site, config)
		err := module.connect.Register(name, regis)
		if err != nil {
			return err
		}
	}
	
	return nil
}









//filter定义的时候，已经去掉*了
func (module *httpModule) requestFilterActions(site string) ([]Funcing) {
	return module.filter.funcings(kREQUEST, site+".")
}
func (module *httpModule) executeFilterActions(site string) ([]Funcing) {
	return module.filter.funcings(kEXECUTE, site+".")
}
func (module *httpModule) responseFilterActions(site string) ([]Funcing) {
	return module.filter.funcings(kRESPONSE, site+".")
}
//handler定义的时候，已经去掉*了
func (module *httpModule) foundHandlerActions(site string) ([]Funcing) {
	return module.handler.funcings(kFOUND, site+".")
}
func (module *httpModule) errorHandlerActions(site string) ([]Funcing) {
	return module.handler.funcings(kERROR, site+".")
}
func (module *httpModule) failedHandlerActions(site string) ([]Funcing) {
	return module.handler.funcings(kFAILED, site+".")
}
func (module *httpModule) deniedHandlerActions(site string) ([]Funcing) {
	return module.handler.funcings(kDENIED, site+".")
}









//事件Http  请求开始
func (module *httpModule) serve(req *HttpRequest, res HttpResponse) {

	ctx := newHttpContext(req, res)
	if config,ok := module.router.chunkdata(ctx.Name).(Map); ok {
		ctx.Config = config
	}

	//request拦截器，加入调用列表
	requestFilters := module.requestFilterActions(ctx.Site)
	ctx.next(requestFilters...)

	ctx.next(module.request)
	ctx.next(module.execute)

	ctx.Next()
}


func (module *httpModule) request(ctx *Context) {

	//请求id
	ctx.Id = ctx.Cookie(ctx.site.Cookie)
	if ctx.Id == "" {
		ctx.Id = Bigger.Unique()
		ctx.Cookie(ctx.site.Cookie, ctx.Id)
	}

	//请求的一开始，主要是SESSION处理
	if ctx.sessional(true) {
		mmm,eee := mSESSION.Read(ctx.Id)
		if eee == nil {
			for k,v := range mmm {
				//待处理session要不要写入value，好让args可以处理
				ctx.Session[k] = v
			}
		}
	}

	//404么
	if ctx.Name == "" || ctx.Config == nil {

		//路由不存在， 找静态文件

		//静态文件放在这里处理
		file := ""
		sitePath := path.Join(Bigger.Config.Path.Static, ctx.Site, ctx.Uri)
		if fi,err := os.Stat(sitePath); err == nil && fi.IsDir() == false {
			file = sitePath
		} else {
			sharedPath := path.Join(Bigger.Config.Path.Static, Bigger.Config.Path.Shared, ctx.Uri)
			if fi,err := os.Stat(sharedPath); err == nil && fi.IsDir() == false {
				file = sharedPath
			}	
		}

		if file != "" {
			ctx.File(file, "")
		} else {
			module.found(ctx)
		}

	} else {
		//表单这里处理，这样会在 requestFilter之前处理好
		if err := module.formHandler(ctx); err != nil {
			ctx.lastError = err
			module.failed(ctx)
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
						//往下走，到execute
						ctx.Next()
					}
				}
			}
		}
	}

	//session写回去
	if ctx.sessional(true) {
		//待处理，SESSION如果没有任何变化， 就不写session
		//这样节省SESSION的资源
		if ctx.site.Expiry != "" {
			td,err := Bigger.Timing(ctx.site.Expiry)
			if err == nil {
				mSESSION.Write(ctx.Id, ctx.Session, td)
			} else {
				mSESSION.Write(ctx.Id, ctx.Session)
			}
		} else {
			mSESSION.Write(ctx.Id, ctx.Session)
		}
	}


	//响应前清空执行线
	ctx.clear()

	//response拦截器，加入调用列表
	filters := module.responseFilterActions(ctx.Site)
	ctx.next(filters...)

	//最终的body处理
	ctx.next(module.body)

	ctx.Next()
}



//事件执行，调用action的地方
func (module *httpModule) execute(ctx *Context) {
	ctx.clear()

	//executeFilters
	filters := module.executeFilterActions(ctx.Site)
	ctx.next(filters...)

	//actions
	actions := ctx.funcing(kACTION)
	ctx.next(actions...)

	ctx.Next()
}



func (module *httpModule) formHandler(ctx *Context) (*Error) {
	var req = ctx.http.req.Reader


	//URL中的参数
	for k,v := range ctx.http.req.Params {
		ctx.Param[k] = v
		ctx.Value[k] = v
	}

	//urlquery
	for k,v := range req.URL.Query() {
		if len(v) == 1 {
			ctx.Query[k] = v[0]
			ctx.Value[k] = v[0]
		} else if len(v) > 1 {
			ctx.Query[k] = v
			ctx.Value[k] = v
		}
	}

	
	


	//是否AJAX请求
	if ctx.Header("X-Requested-With") == "XMLHttpRequest" {
		ctx.Ajax = true
	} else if ctx.Header("Ajax") != "" {
		ctx.Ajax = true
	} else {
		ctx.Ajax = false
	}

	//客户端的默认语言
	if al := ctx.Header("Accept-Language"); al != "" {
		accepts := strings.Split(al, ",")
		if len(accepts) > 0 {
			llll:
			for _,accept := range accepts {
				if i := strings.Index(accept, ";"); i > 0 {
					accept = accept[0:i]
				}
				//遍历匹配
				for lang,config := range Bigger.Config.Lang {
					for _,acccc := range config.Accepts {
						if strings.ToLower(acccc) == strings.ToLower(accept) {
							ctx.Lang = lang
							break llll
						}
					}
				}
			}
		}
	}

	
	if ctx.Method == "POST" || ctx.Method == "PUT" || ctx.Method == "PATCH" {
		//根据content-type来处理
		ctype := ctx.Header("Content-Type")
		if strings.Contains(ctype, "json") {
			body, err := ioutil.ReadAll(req.Body)
			if err == nil {
				ctx.Body = string(body)

				m := Map{}
				err := json.Unmarshal(body, &m)
				if err == nil {
					//遍历JSON对象
					for k,v := range m {
						ctx.Form[k] = v
						ctx.Value[k] = v
					}
				}
			}
		} else if strings.Contains(ctype, "xml") {
			body, err := ioutil.ReadAll(req.Body)
			if err == nil {
				ctx.Body = string(body)
				
				m := Map{}
				err := xml.Unmarshal(body, &m)
				if err == nil {
					//遍历XML对象
					for k,v := range m {
						ctx.Form[k] = v
						ctx.Value[k] = v
					}
				}
			}
		} else {

			// if ctype=="application/x-www-form-urlencoded" || ctype=="multipart/form-data" {
			// }

			err := req.ParseMultipartForm(32 << 20)
			if err != nil {
				//表单解析有问题，就处理成原始STRING
				body, err := ioutil.ReadAll(req.Body)
				if err == nil {
					ctx.Body = string(body)
				}

			}



			names := []string{}
			values := url.Values{}
			uploads := map[string][]Map{}

			if req.MultipartForm != nil {

				//处理表单，这里是否应该直接写入ctx.Form比较好？
				for k,v := range req.MultipartForm.Value {
					//有个问题，当type=file时候，又不选文件的时候，value里会存在一个空字串的value
					//如果同一个form name 有多条记录，这时候会变成一个[]string，的空串数组
					//这时候，mapping解析文件的时候[file]就会出问题，会判断文件类型，这时候是[]string就出问题了
					// ctx.Form[k] = v
					names = append(names, k)
					values[k] = v
				}
				
				//FILE可能要弄成JSON，文件保存后，MIME相关的东西，都要自己处理一下
				for k,v := range req.MultipartForm.File {
					//这里应该保存为数组
					files := []Map{}

					//处理多个文件
					for _,f := range v {

						if f.Size <= 0 || f.Filename == "" {
							continue
						}

						hash := ""
						filename := f.Filename
						mimetype := f.Header.Get("Content-Type")
						extension := strings.ToLower(path.Ext(filename))
						if extension != "" {
							extension = extension[1:]	//去掉点.
						}
						
						var tempfile string
						var length int64 = f.Size

						//先计算hash
						if file, err := f.Open(); err == nil {

							h := sha1.New()
							if _, err := io.Copy(h, file); err == nil {
								
								hash = fmt.Sprintf("%x", h.Sum(nil))

								//保存临时文件
								tempfile = path.Join(Bigger.Config.Path.Upload, fmt.Sprintf("%s_%s", Bigger.Name, hash))
								if extension != "" {
									tempfile = fmt.Sprintf("%s.%s", tempfile, extension)
								}

								//重新定位
								file.Seek(0, 0)

								if save, err := os.OpenFile(tempfile, os.O_WRONLY|os.O_CREATE, 0777); err == nil {
									io.Copy(save, file)	//保存文件
									save.Close()

									msg := Map{
										"hash": hash,
										"filename": filename,
										"extension": extension,
										"mimetype": mimetype,
										"length": length,
										"tempfile": tempfile,
									}
			
									files = append(files, msg)
								}
							}

							//最后关闭文件
							file.Close()
						}

						uploads[k] = files
					}
				}

			} else if req.PostForm != nil {
				for k,v := range req.PostForm {
					names = append(names, k)
					values[k] = v
				}

			} else if req.Form != nil {
				for k,v := range req.Form {
					names = append(names, k)
					values[k] = v
				}
			}

			tomlroot := map[string][]string{}
			tomldata := map[string]map[string][]string{}

			//顺序很重要
			tomlexist := map[string]bool{}
			tomlnames := []string{}

			//统一解析
			for _,k := range names {
				v := values[k]

				//写入form
				if len(v) == 1 {
					ctx.Form[k] = v[0]
				} else if len(v) > 1 {
					ctx.Form[k] = v
				}

				// key := fmt.Sprintf("value[%s]", k)
				// forms[k] = v


				if strings.Contains(k, ".") {

					//以最后一个.分割，前为key，后为field
					i := strings.LastIndex(k, ".")
					key := k[:i]
					field := k[i+1:]


					if vv,ok := tomldata[key]; ok {
						vv[field] = v
					} else {
						tomldata[key] = map[string][]string{
							field: v,
						}
					}

					if _,ok := tomlexist[key]; ok==false {
						tomlexist[key] = true
						tomlnames = append(tomlnames, key)
					}

				} else {
					tomlroot[k] = v
				}


				//这里不写入， 解析完了才
				// ctx.Value[k] = v
			}


			lines := []string{}
			for kk,vv := range tomlroot {
				if len(vv) > 1 {
					lines = append(lines, fmt.Sprintf(`%s = ["""%s"""]`, kk, strings.Join(vv, `""","""`)))
				} else {
					lines = append(lines, fmt.Sprintf(`%s = """%s"""`, kk, vv[0]))
				}
			}
			for _,kk := range tomlnames {
				vv := tomldata[kk]

				//普通版
				// lines = append(lines, fmt.Sprintf("[%s]", kk))
				// for ff,fv := range vv {
				// 	if len(fv) > 1 {
				// 		lines = append(lines, fmt.Sprintf(`%s = ["%s"]`, ff, strings.Join(fv, `","`)))
				// 	} else {
				// 		lines = append(lines, fmt.Sprintf(`%s = "%s"`, ff, fv[0]))
				// 	}
				// }

				//数组版
				//先判断一下，是不是map数组
				length := 0
				for _,fv := range vv {
					if length == 0 {
						length = len(fv)
					} else {
						if length != len(fv) {
							length = -1
							break
						}
					}
				}

				//如果length>1是数组，并且相等
				if length > 1 {
					for i:=0;i<length;i++ {
						lines = append(lines, fmt.Sprintf("[[%s]]", kk))
						for ff,fv := range vv {
							lines = append(lines, fmt.Sprintf(`%s = """%s"""`, ff, fv[i]))
						}
					}

				} else {
					lines = append(lines, fmt.Sprintf("[%s]", kk))
					for ff,fv := range vv {
						if len(fv) > 1 {
							lines = append(lines, fmt.Sprintf(`%s = ["""%s"""]`, ff, strings.Join(fv, `""","""`)))
						} else {
							lines = append(lines, fmt.Sprintf(`%s = """%s"""`, ff, fv[0]))
						}
					}
				}
			}

			value := Map{}
			_,err = toml.Decode(strings.Join(lines, "\n"), &value)
			if err == nil {
				for k,v := range value {
					ctx.Value[k] = v
				}
			} else {
				for k,v := range values {
					if len(v) == 1 {
						ctx.Value[k] = v[0]
					} else if len(v) > 1 {
						ctx.Value[k] = v
					}
				}
			}

			for k,v := range uploads {
				if len(v) == 1 {
					ctx.Value[k] = v[0]
					ctx.Upload[k] = v[0]
				} else if len(v) > 1 {
					ctx.Value[k] = v
					ctx.Upload[k] = v
				}
			}


		}
	}

	return nil
}



















//最终响应
func (module *httpModule) body(ctx *Context) {
	if ctx.Code == 0 {
		ctx.Code = http.StatusOK
	}

	//设置cookies, headers

	//cookie超时时间
	//为了极致的性能，可以在启动的时候先解析好
	var maxage time.Duration
	if ctx.site.MaxAge != "" {
		td,err := Bigger.Timing(ctx.site.MaxAge)
		if err == nil {
			maxage = td
		}
	}

	res := ctx.http.req.Writer
	for _,v := range ctx.cookies {

		if ctx.domain != "" {
			v.Domain = ctx.domain
		}
		if ctx.site.MaxAge != "" {
			v.MaxAge = int(maxage.Seconds())
		}

		//加密cookie
		v.Value = Bigger.Encrypt(v.Value)

		http.SetCookie(res, &v)
	}
	for k,v := range ctx.headers {
		res.Header().Set(k, v)
	}

	switch body := ctx.Body.(type) {
	case httpGotoBody:
		module.bodyGoto(ctx, body)
	case httpTextBody:
		module.bodyText(ctx, body)
	case httpHtmlBody:
		module.bodyHtml(ctx, body)
	case httpScriptBody:
		module.bodyScript(ctx, body)
	case httpJsonBody:
		module.bodyJson(ctx, body)
	case httpJsonpBody:
		module.bodyJsonp(ctx, body)
	case httpApiBody:
		module.bodyApi(ctx, body)
	case httpXmlBody:
		module.bodyXml(ctx, body)
	case httpFileBody:
		module.bodyFile(ctx, body)
	case httpDownBody:
		module.bodyDown(ctx, body)
	case httpBufferBody:
		module.bodyBuffer(ctx, body)
	case httpViewBody:
		module.bodyView(ctx, body)
	default:
		module.bodyDefault(ctx)
	}


	//最终响应前做清理工作
	ctx.final()
}
func (module *httpModule) bodyDefault(ctx *Context) {
	ctx.Code = http.StatusNotFound
	http.NotFound(ctx.http.req.Writer, ctx.http.req.Reader)
	ctx.http.res.Finish(ctx.http.req)
}
func (module *httpModule) bodyGoto(ctx *Context, body httpGotoBody) {
	http.Redirect(ctx.http.req.Writer, ctx.http.req.Reader, body.url, http.StatusFound)
	ctx.http.res.Finish(ctx.http.req)
}
func (module *httpModule) bodyText(ctx *Context, body httpTextBody) {
	res := ctx.http.req.Writer

	if ctx.Type == "" {
		ctx.Type = "text"
	}

	ctx.Type = mCONST.MimeType(ctx.Type, "text/explain")
	res.Header().Set("Content-Type", fmt.Sprintf("%v; charset=%v", ctx.Type, ctx.Charset))

	res.WriteHeader(ctx.Code)
	fmt.Fprint(res, body.text)

	ctx.http.res.Finish(ctx.http.req)
}
func (module *httpModule) bodyHtml(ctx *Context, body httpHtmlBody) {
	res := ctx.http.req.Writer

	if ctx.Type == "" {
		ctx.Type = "html"
	}

	ctx.Type = mCONST.MimeType(ctx.Type, "text/html")
	res.Header().Set("Content-Type", fmt.Sprintf("%v; charset=%v", ctx.Type, ctx.Charset))

	res.WriteHeader(ctx.Code)
	fmt.Fprint(res, body.html)

	ctx.http.res.Finish(ctx.http.req)
}
func (module *httpModule) bodyScript(ctx *Context, body httpScriptBody) {
	res := ctx.http.req.Writer

	ctx.Type = mCONST.MimeType(ctx.Type, "application/script")
	res.Header().Set("Content-Type", fmt.Sprintf("%v; charset=%v", ctx.Type, ctx.Charset))

	res.WriteHeader(ctx.Code)
	fmt.Fprint(res, body.script)
	ctx.http.res.Finish(ctx.http.req)
}
func (module *httpModule) bodyJson(ctx *Context, body httpJsonBody) {
	res := ctx.http.req.Writer

	bytes, err := json.Marshal(body.json)
	if err != nil {
		//要不要发到统一的错误ctx.Error那里？再走一遍
		http.Error(res, err.Error(), http.StatusInternalServerError)
	} else {

		ctx.Type = mCONST.MimeType(ctx.Type, "text/json")
		res.Header().Set("Content-Type", fmt.Sprintf("%v; charset=%v", ctx.Type, ctx.Charset))

		res.WriteHeader(ctx.Code)
		fmt.Fprint(res, string(bytes))
	}
	ctx.http.res.Finish(ctx.http.req)
}
func (module *httpModule) bodyJsonp(ctx *Context, body httpJsonpBody) {
	res := ctx.http.req.Writer

	bytes, err := json.Marshal(body.json)
	if err != nil {
		//要不要发到统一的错误ctx.Error那里？再走一遍
		http.Error(res, err.Error(), http.StatusInternalServerError)
	} else {

		ctx.Type = mCONST.MimeType(ctx.Type, "application/script")
		res.Header().Set("Content-Type", fmt.Sprintf("%v; charset=%v", ctx.Type, ctx.Charset))

		res.WriteHeader(ctx.Code)
		fmt.Fprint(res, fmt.Sprintf("%s(%s);", body.callback, string(bytes)))
	}
	ctx.http.res.Finish(ctx.http.req)
}
func (module *httpModule) bodyXml(ctx *Context, body httpXmlBody) {
	res := ctx.http.req.Writer

	bytes, err := xml.Marshal(body.xml)
	if err != nil {
		//要不要发到统一的错误ctx.Error那里？再走一遍
		http.Error(res, err.Error(), http.StatusInternalServerError)
	} else {
		ctx.Type = mCONST.MimeType(ctx.Type, "text/xml")
		res.Header().Set("Content-Type", fmt.Sprintf("%v; charset=%v", ctx.Type, ctx.Charset))

		res.WriteHeader(ctx.Code)
		fmt.Fprint(res, string(bytes))
	}
	ctx.http.res.Finish(ctx.http.req)
}
func (module *httpModule) bodyApi(ctx *Context, body httpApiBody) {

	json := Map{
		"code": body.code,
		"time": time.Now().Unix(),
	}

	if body.text != "" {
		json["text"] = body.text
	}

	if body.data != nil {

		crypto := ctx.site.Crypto
		if vv,ok := ctx.Config["crypto"].(bool); ok && vv == false {
			crypto = ""
		}
		if vv,ok := ctx.Config["plain"].(bool); ok && vv {
			crypto = ""
		}

		// if crypto != "" {
		// 	json["crypto"] = crypto
		// }

		tempConfig := Map{
			"data": Map{
				"type": "json", "must": true, "encode": crypto,
			},
		}
		tempData := Map{
			"data": body.data,
		}

		//有自定义返回数据类型
		if vv,ok := ctx.Config["data"].(Map); ok {
			tempConfig = Map{
				"data": Map{
					"type": "json", "must": true, "encode": crypto,
					"json": vv,
				},
			}
		}

		val := Map{}
		err := mMAPPING.Parse(tempConfig, tempData, val, false, false, ctx)
		if err != nil {
			err.status = strings.Replace(err.status, ".mapping.", ".data.", 1)

			json["code"] = err.Code()
			json["text"] = err.Lang(ctx.Lang).String()
		} else {
			//处理后的data
			json["data"] = val["data"]
		}

	}

	//转到jsonbody去处理
	module.bodyJson(ctx, httpJsonBody{json})
}









func (module *httpModule) bodyFile(ctx *Context, body httpFileBody) {
	req, res := ctx.http.req.Reader, ctx.http.req.Writer

	//文件类型
	if ctx.Type != "file" {
		ctx.Type = mCONST.MimeType(ctx.Type, "application/octet-stream")
		res.Header().Set("Content-Type", fmt.Sprintf("%v; charset=%v", ctx.Type, ctx.Charset))
	}
	//加入自定义文件名
	if body.name != "" {
		res.Header().Set("Content-Disposition", fmt.Sprintf("attachment;filename=%v;", body.name))
	}

	http.ServeFile(res, req, body.file)
	ctx.http.res.Finish(ctx.http.req)
}
func (module *httpModule) bodyDown(ctx *Context, body httpDownBody) {
	res := ctx.http.req.Writer
	
	if ctx.Type == "" {
		ctx.Type = "file"
	}

	ctx.Type = mCONST.MimeType(ctx.Type, "application/octet-stream")
	res.Header().Set("Content-Type", fmt.Sprintf("%v; charset=%v", ctx.Type, ctx.Charset))
	//加入自定义文件名
	if body.name != "" {
		res.Header().Set("Content-Disposition", fmt.Sprintf("attachment;filename=%v;", body.name))
	}

	res.WriteHeader(ctx.Code)
	res.Write(body.bytes)
	
	ctx.http.res.Finish(ctx.http.req)
}
func (module *httpModule) bodyBuffer(ctx *Context, body httpBufferBody) {
	res := ctx.http.req.Writer
	
	if ctx.Type == "" {
		ctx.Type = "file"
	}
	
	ctx.Type = mCONST.MimeType(ctx.Type, "application/octet-stream")
	res.Header().Set("Content-Type", fmt.Sprintf("%v; charset=%v", ctx.Type, ctx.Charset))
	//加入自定义文件名
	if body.name != "" {
		res.Header().Set("Content-Disposition", fmt.Sprintf("attachment;filename=%v;", body.name))
	}

	res.WriteHeader(ctx.Code)
	io.Copy(res, body.buffer)
	body.buffer.Close()
	
	ctx.http.res.Finish(ctx.http.req)
}
func (module *httpModule) bodyView(ctx *Context, body httpViewBody) {
	res := ctx.http.req.Writer

	viewdata := Map{
		kARGS: ctx.Args, kAUTH: ctx.Auth,
		kSETTING: Bigger.Setting, kLOCAL: ctx.Local,
		kDATA:	ctx.Data, kMODEL: body.model,
	}

	//系统内置的helper
	helpers := Map{
		"route": ctx.Url.Route,
		"browse": ctx.Url.Browse,
		"preview": ctx.Url.Preview,
		"download": ctx.Url.Download,
		"backurl": ctx.Url.Back,
		"lasturl": ctx.Url.Last,
		"siteurl": func(name string, paths ...string) string {
			path := "/"
			if len(paths) > 0 {
				path = paths[0]
			}
			return ctx.Url.Site(name, path)
		},

		"lang": func() string {
			return ctx.Lang
		},
		"zone": func() *time.Location {
			return ctx.Zone
		},
		"timezone": func() (string) {
			return ctx.String(ctx.Zone.String())
		},
		"format": func(format string, args ...interface{}) (string) {
			//支持一下显示时间
			if len(args) == 1 {
				if args[0] == nil {
					return format
				} else if ttt,ok := args[0].(time.Time); ok {
					zoneTime := ttt.In(ctx.Zone)
					return zoneTime.Format(format)
				} else if ttt,ok := args[0].(int64); ok {
					//时间戳是大于1971年是, 千万级, 2016年就是10亿级了
					if ttt >= int64(31507200) && ttt <= int64(31507200000) {
						ttt := time.Unix(ttt, 0)
						zoneTime := ttt.In(ctx.Zone)
						sss := zoneTime.Format(format)
						if strings.HasPrefix(sss, "%")==false || format != sss {
							return sss
						}
					}
				}
			}
			return fmt.Sprintf(format, args...)
		},


		"signed": func(key string) bool {
			return ctx.Signed(key)
		},
		"signal": func(key string) string {
			return ctx.Signal(key)
		},
		"signer": func(key string) string {
			return ctx.Signer(key)
		},
		"string": func(key string, args ...Any) string {
			return ctx.String(key, args...)
		},
		"enum": func(name, field string,v Any) (string) {
			value := fmt.Sprintf("%v", v)
			//多语言支持
			//key=enum.name.file.value
			langkey := fmt.Sprintf("enum.%s.%s.%s", name, field, value)
			langval := ctx.String(langkey)
			if langkey != langval {
				return langval
			} else {
				enums := Bigger.Enums(name, field)
				if vv,ok := enums[value].(string); ok {
					return vv
				}
				return value
			}
		},
	}

	vhelpers := mVIEW.helperActions()
	for k,v := range vhelpers {
		helpers[k] = v
	}

	html,err := mVIEW.parse(ctx, ViewBody{
		Root:		Bigger.Config.Path.View,
		Shared:		Bigger.Config.Path.Shared,
		View:		body.view,
		Data: 		viewdata,
		Helpers:	helpers,
	})

	if err != nil {
		http.Error(res, err.Lang(ctx.Lang).String(), 500)
	} else {
		ctx.Type = mCONST.MimeType(ctx.Type, "text/html")
		res.Header().Set("Content-Type", fmt.Sprintf("%v; charset=%v", ctx.Type, ctx.Charset))
		res.WriteHeader(ctx.Code)
		fmt.Fprint(res, html)
	}

	ctx.http.res.Finish(ctx.http.req)
}










// func (module *httpModule) sessionKey(ctx *Context) string {
// 	format := "http_%s"
// 	if vv,ok := CONFIG.Session.Format[bHTTP].(string); ok && vv != "" {
// 		format = vv
// 	}
// 	return fmt.Sprintf(format, ctx.Id)
// }



//事件handler,找不到
func (module *httpModule) found(ctx *Context) {
	ctx.clear()

	ctx.Code = http.StatusNotFound

	//如果有自定义的错误处理，加入调用列表
	funcs := ctx.funcing(kFOUND)
	ctx.next(funcs...)


	//把处理器加入调用列表
	handlers := module.foundHandlerActions(ctx.Site)
	ctx.next(handlers...)

	//加入默认的错误处理
	ctx.next(module.foundDefaultHandler)
	ctx.Next()
}
//最终还是由response处理
func (module *httpModule) foundDefaultHandler(ctx *Context) {
	//如果是ajax访问，返回JSON对应，要不然返回页面
	if ctx.Ajax {
		ctx.Result(Bigger.Erring(_kFOUND))
	} else {
		ctx.View(kFOUND)
	}
}

//事件handler,错误的处理
func (module *httpModule) error(ctx *Context) {
	ctx.clear()

	//如果有自定义的错误处理，加入调用列表
	funcs := ctx.funcing(kERROR)
	ctx.next(funcs...)

	//把错误处理器加入调用列表
	handlers := module.errorHandlerActions(ctx.Site)
	ctx.next(handlers...)

	//加入默认的错误处理
	ctx.next(module.errorDefaultHandler)
	ctx.Next()
}

//最终还是由response处理
func (module *httpModule) errorDefaultHandler(ctx *Context) {
	error := Bigger.Erring(_kERROR)
	if err := ctx.Erred(); err != nil {
		error = err
	}
	if ctx.Ajax {
		ctx.Result(error)
	} else {
		ctx.Data[kERROR] = Map{
			"code": error.Code(),
			"text": error.Lang(ctx.Lang).String(),
		}
		ctx.View(kERROR)
	}
}


//事件handler,失败处理，主要是args失败
func (module *httpModule) failed(ctx *Context) {
	ctx.clear()

	//如果有自定义的失败处理，加入调用列表
	funcs := ctx.funcing(kFAILED)
	ctx.next(funcs...)


	//把失败处理器加入调用列表
	handlers := module.failedHandlerActions(ctx.Site)
	ctx.next(handlers...)

	//加入默认的错误处理
	ctx.next(module.failedDefaultHandler)
	ctx.Next()
}
//最终还是由response处理
func (module *httpModule) failedDefaultHandler(ctx *Context) {
	failed := Bigger.Erring(_kFAILED)
	if err := ctx.Erred(); err != nil {
		failed = err
	}

	if ctx.Ajax {
		ctx.Result(failed)
	} else {
		ctx.Alert(failed)
	}
}



//事件handler,失败处理，主要是args失败
func (module *httpModule) denied(ctx *Context) {
	ctx.clear()

	//如果有自定义的失败处理，加入调用列表
	funcs := ctx.funcing(kDENIED)
	ctx.next(funcs...)

	//把失败处理器加入调用列表
	handlers := module.deniedHandlerActions(ctx.Site)
	ctx.next(handlers...)

	//加入默认的错误处理
	ctx.next(module.deniedDefaultHandler)
	ctx.Next()
}
//最终还是由response处理
//如果是ajax。返回拒绝
//如果不是， 返回一个脚本提示
func (module *httpModule) deniedDefaultHandler(ctx *Context) {
	denied := Bigger.Erring(_kDENIED)
	if err := ctx.Erred(); err != nil {
		denied = err
	}

	if ctx.Ajax {
		ctx.Result(denied)
	} else {
		ctx.Alert(denied)
	}
}



