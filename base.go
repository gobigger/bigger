package bigger


type (
	Any	= interface{}
	Map	= map[string]Any
	env = int
	mod	= int
)
type (
	KVPair struct {
		Key	string
		Val	Any
	}
)

const (
    _ env = iota
    Developing
    Testing
    Production
)

const (
	_ mod = iota
	planMode
    eventMode
    queueMode
	httpMode
	socketMode
)






const (
	nBIGGER	= "bigger"
	BACKURL = "_back_"
)

const (

	URL_CURRENT	= "[current]"
	URL_BACK	= "[back]"
	URL_LAST	= "[last]"
	URL_QUERY	= "[query]"

	

	bCONSTMIME		= "const.mime"
	bCONSTSTATUS	= "const.status"
	bCONSTREGULAR	= "const.regular"
	bCONSTLANG		= "const.lang"

	bMAPPINGTYPE		= "mapping.type"
	bMAPPINGCRYPTO		= "mapping.crypto"


	bLoggerDriver	= "logger.driver"
	bMutexDriver	= "mutex.driver"
	bSessionDriver	= "session.driver"


	bCacheDriver	= "cache.driver"

	bDataDriver	= "data.driver"
	bDATATABLE	= "data.table"
	bDATAVIEW	= "data.view"
	bDATAMODEL	= "data.model"

	bFileDriver	= "file.driver"

	bPlanDriver		= "plan.driver"
	bPLANROUTER		= "plan.router"
	bPLANFILTER		= "plan.filter"
	bPLANHANDLER	= "plan.handler"


	bEventDriver	= "event.driver"
	bEVENTROUTER	= "event.router"
	bEVENTFILTER	= "event.filter"
	bEVENTHANDLER	= "event.handler"

	bQueueDriver	= "queue.driver"
	bQUEUEROUTER	= "queue.router"
	bQUEUEFILTER	= "queue.filter"
	bQUEUEHANDLER	= "queue.handler"

	bHttpDriver		= "http.driver"
	bHTTPROUTER		= "http.router"
	bHTTPFILTER		= "http.filter"
	bHTTPHANDLER	= "http.handler"


	bSocketDriver	= "socket.driver"
	bSOCKETROUTER	= "socket.router"
	bSOCKETFILTER	= "socket.filter"
	bSOCKETHANDLER	= "socket.handler"
	bSocketCommand	= "socket.command"

	bVIEW			= "view"
	bVIEWHELPER		= "view.helper"



	bSERVICE			= "service"


	kID			= "id"
	kNAME		= "name"


	kSHOW		= "show"

	


	kDEFAULT	= "default"
	kBUFFER		= "buffer"
	kTYPE		= "type"
	kTYPES		= "types"
	kVALID		= "valid"
	kVALUE		= "value"
	kVAL		= "val"
	kDECODE		= "decode"
	kENCODE		= "encode"
	kAUTO		= "auto"
	kJSON		= "json"
	kENUM		= "enum"
	kTEXT		= "text"
	kMUST		= "must"
	kEMPTY		= "empty"
	kACTION		= "action"
	kSESSION	= "session"
	kINVOKE		= "invoke"
	kTIME		= "time"
	kDELAY		= "delay"
	kTIMES		= "times"
	kLINE		= "line"
	kLINES		= "lines"

	kFIELDS		= "fields"

	kROUTE		= "route"
	kMETHOD		= "method"
	kDATA		= "data"
	kTABLE		= "table"
	kVIEW		= "view"
	kMODEL		= "model"

	kKEY		= "key"
	kARGS		= "args"
	kPARAM		= "param"
	kQUERY		= "query"
	kSIGN		= "sign"
	kSITE		= "site"
	kAUTH		= "auth"
	kBASE		= "base"
	kITEM		= "item"
	
	kFOUND		= "found"
	kERROR		= "error"
	kFAILED		= "failed"
	kDENIED		= "denied"

	_kFOUND		= ".found"
	_kERROR		= ".error"
	_kFAILED	= ".failed"
	_kDENIED	= ".denied"
	
	kSERVE		= "serve"
	kREQUEST	= "request"
	kEXECUTE	= "execute"
	kRESPONSE	= "response"

	kCRYPTO		= "crypto"
	kCRYPTOS	= "cryptos"
	kURI		= "uri"
	kURIS		= "uris"
	kMETHODS	= "methods"
	kDOMAIN		= "domain"
	kDOMAINS	= "domains"

	vUTF8		= "utf-8"
	kSETTING	= "setting"
	kLOCAL		= "local"

	kRESULTS	= "$results"
	kRESULTING	= "$resuling"


)






/*
    用于生成查询条件，示例如下：

    Map{
        "views":	Map{ GT: 100, LT: 500 },			解析成：	views>100 AND views<500
        "hits":		Map{ GTE: 100, LTE: 500 },			解析成	hits>=100 AND hits<=500
        "name": 	Map{ EQ: "noggo", NE: "nogio" },	解析成：	name='noggo' AND name!='nogio'
        "tags": 	Map{ ANY: "nog" },					解析成：	ANY(tags)='nog'
        "id":		Map{ IN: []int{ 1,2,3 } },			解析成：	id IN (1,2,3)
        "email": 	nil,								解析成：	email IS NULL
        "email": 	NIL,								解析成：	email IS NULL
        "email": 	NOL,								解析成：	email IS NOT NULL

        "id":		ASC									解析成：id ASC
        "id":		DESC								解析成：id DESC
    }

*/

const (
	REMOVED = "removed"

	DELIMS	= `"`	//字段以及表名边界符，自己实现数据驱动才需要处理这个，必须能启标识作用

	COUNT   = "COUNT"
	SUM     = "SUM"
	MAX     = "MAX"
	MIN     = "MIN"
	AVG     = "AVG"



	IS		= "="	//等于
	NOT 	= "!="	//不等于
	EQ		= "="	//等于
	NE		= "!="	//不等于
	NEQ		= "!="	//不等于

	//约等于	正则等于
	AE		= "~*"		//正则等于，约等于
	AEC		= "~"		//正则等于，区分大小写，
	RE		= "~*"		//正则等于，约等于
	REC		= "~"		//正则等于，区分大小写，
	REQ		= "~*"		//正则等于，约等于
	REQC	= "~"		//正则等于，区分大小写，

	NAE		= "!~*"		//正则不等于，
	NAEC	= "!~"		//正则不等于，区分大小写，
	NRE		= "!~*"		//正则不等于，
	NREC	= "!~"		//正则不等于，区分大小写，
	NREQ	= "!~*"		//正则不等于，
	NREQC	= "!~"		//正则不等于，区分大小写，

	//换位约等于，值在前，字段在后，用于黑名单查询
	EA		= "$$~*$$"		//正则等于，约等于
	EAC		= "$$~$$"		//正则等于，区分大小写，
	ER		= "$$~*$$"		//正则等于，约等于
	ERC		= "$$~$$"		//正则等于，区分大小写，
	EQR		= "$$~*$$"		//正则等于，约等于
	EQRC	= "$$~$$"		//正则等于，区分大小写，

	NEA		= "$$!~*$$"		//正则不等于，
	NEAC	= "$$!~$$"		//正则不等于，区分大小写，
	NER		= "$$!~*$$"		//正则不等于，
	NERC	= "$$!~$$"		//正则不等于，区分大小写，
	NEQR	= "$$!~*$$"		//正则不等于，
	NEQRC	= "$$!~$$"		//正则不等于，区分大小写，


	GT	= ">"	//大于
	GE	= ">="	//大于等于
	GTE	= ">="	//大于等于
	LT	= "<"	//小于
	LE	= "<="	//小于等于
	LTE	= "<="	//小于等于

	IN	= "$$IN$$"	//支持  WHERE id IN (1,2,3)			//这条还没支持
	NI  = "$$NOTIN$$"	//支持	WHERE id NOT IN(1,2,3)
	NIN = "$$NOTIN$$"	//支持	WHERE id NOT IN(1,2,3)
	ANY = "$$ANY$$"		//支持数组字段的

	LIKE = "$$full$$"		//like搜索
	FULL = "$$full$$"		//like搜索
	LEFT = "$$left$$"		//like left搜索
	RIGHT = "$$right$$"		//like right搜索

	INC	= "$$inc$$"	//累加，    UPDATE时用，解析成：views=views+value

	BYASC	= "asc"
	BYDESC	= "desc"

)

type (
	dataNil		struct {}
	dataNol		struct {}
	dataAsc		struct {}
	dataDesc	struct {}
)

var (
	NIL		dataNil		//为空	IS NULL
	NOL		dataNol		//不为空	IS NOT NULL
	NULL	dataNil		//为空	IS NULL
	NOLL	dataNol		//不为空	IS NOT NULL
	ASC		dataAsc		//正序	asc
	DESC	dataDesc	//倒序	desc
)




const (
	EventBiggerStart	= ".bigger.start"
	EventBiggerEnd		= ".bigger.end"

	EventDataCreate		= ".data.create"
	EventDataChange		= ".data.change"
	EventDataRemove		= ".data.remove"

	EventSocketUpgrade	= ".socket.upgrade"
	EventSocketDegrade	= ".socket.degrade"
	EventSocketFollow	= ".socket.follow"
	EventSocketUnfollow	= ".socket.unfollow"
)


var (
	BiggerUploadConfig	= Map{
		"filename": Map{
			"type": "string", "must": true, "name": "文件名", "text": "文件名",
		},
		"extension": Map{
			"type": "string", "must": true, "name": "扩展名", "text": "扩展名",
		},
		"mimetype": Map{
			"type": "string", "must": true, "name": "mime类型", "text": "mime类型",
		},
		"length": Map{
			"type": "int", "must": true, "name": "文件大小", "text": "文件大小",
		},
		"tempfile": Map{
			"type": "string", "must": true, "name": "文件路径", "text": "文件路径",
		},
	}
	BiggerInvokesConfig = Map{
		"items": Map{
			"type": "[map]", "must": true, "name": "列表", "text": "列表",
		},
	}
	BiggerInvokingConfig = Map{
		"count": Map{
			"type": "int", "must": true, "auto": int64(0), "name": "统计", "text": "统计",
		},
		"items": Map{
			"type": "[map]", "must": true, "auto": []Map{}, "name": "列表", "text": "列表",
		},
	}
	BiggerInvokerConfig = Map{
		"item": Map{
			"type": "map", "must": true, "auto": Map{}, "name": "单体", "text": "单体",
		},
		"items": Map{
			"type": "[map]", "must": true, "auto": []Map{}, "name": "列表", "text": "列表",
		},
	}
)



