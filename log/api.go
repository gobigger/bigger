package log

import (
	"fmt"
	"strconv"
	"github.com/yatlabs/bigger/util"
	"errors"
	"reflect"
)

const API_ADAPTER_NAME = "api"

// adapter api
type AdapterApi struct {
	config *ApiConfig
}

// api config
type ApiConfig struct {

	// request url adddress
	Url string

	// request method
	// GET, POST
	Method string

	// request headers
	Headers map[string]string

	// is verify response code
	IsVerify bool

	// verify response http code
	VerifyCode int
}

func (ac *ApiConfig) Name() string {
	return API_ADAPTER_NAME
}

func NewAdapterApi() LoggerAbstract {
	return &AdapterApi{}
}

func (adapterApi *AdapterApi) Init(apiConfig Config) error {

	if apiConfig.Name() != API_ADAPTER_NAME {
		return errors.New("logger api adapter init error, config must ApiConfig")
	}

	vc := reflect.ValueOf(apiConfig)
	ac := vc.Interface().(*ApiConfig)
	adapterApi.config = ac

	if adapterApi.config.Url == "" {
		return errors.New("config Url cannot be empty!")
	}
	if adapterApi.config.Method != "GET" && adapterApi.config.Method != "POST" {
		return errors.New("config Method must one of the 'GET', 'POST'!")
	}
	if adapterApi.config.IsVerify && (adapterApi.config.VerifyCode == 0) {
		return errors.New("config if IsVerify is true, VerifyCode cannot be 0!")
	}
	return nil
}

func (adapterApi *AdapterApi) Write(loggerMsg *loggerMessage) error {

	url :=  adapterApi.config.Url
	method :=  adapterApi.config.Method
	isVerify :=  adapterApi.config.IsVerify
	verifyCode :=  adapterApi.config.VerifyCode
	headers :=  adapterApi.config.Headers

	loggerMap := map[string]string {
		"nano": strconv.FormatInt(loggerMsg.Nano, 10),
		"time": loggerMsg.Time,
		"level": strconv.Itoa(loggerMsg.Level),
		"type": loggerMsg.Type,
		"body": loggerMsg.Body,
		"file": loggerMsg.File,
		"line": strconv.Itoa(loggerMsg.Line),
		"func": loggerMsg.Func,
	}

	var err error
	var code int
	if method == "GET" {
		_, code, err = util.NewMisc().HttpGet(url, loggerMap, headers, 0)
	}else {
		_, code, err = util.NewMisc().HttpPost(url, loggerMap, headers, 0)
	}
	if err != nil {
		return err
	}
	if(isVerify && (code != verifyCode)) {
		return fmt.Errorf("%s", "request "+ url +" faild, code=" + strconv.Itoa(code))
	}

	return nil
}

func (adapterApi *AdapterApi) Flush() {

}

func (adapterApi *AdapterApi)Name() string {
	return API_ADAPTER_NAME
}

func init()  {
	Register(API_ADAPTER_NAME, NewAdapterApi)
}





/*

	logger := blog.NewLogger()

	apiConfig := &blog.ApiConfig{
		Url: "http://127.0.0.1:8081/index.php",
		Method: "GET",
		Headers: map[string]string{},
		IsVerify: false,
		VerifyCode: 0,
	}
	logger.Attach("api", blog.LEVEL_DEBUG, apiConfig)
	logger.SetAsync()

	logger.Error("this is a error log!")
	logger.Info("this is a alert log!")

	logger.Flush()

*/