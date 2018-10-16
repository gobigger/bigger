package bigger

import (
	"path"
	"time"
	"strings"
	"fmt"
	"io"
	"github.com/yatlabs/bigger/hashring"
)


type (
    FileDriver interface {
        Connect(name string, config FileConfig) (FileConnect,*Error)
    }
    FileConnect interface {
        Open() (*Error)
        Health() (*FileHealth,*Error)
		Close() (*Error)
		
        Base() (FileBase)
    }

    FileBase interface {
        Close() *Error
		Erred() *Error

		Assign(name string, metadata Map) (string)
		Storage(code string, reader io.Reader) (int64)
		Download(code string) (io.ReadCloser,*FileCoding)
		Thumbnail(code string, width, height, time int64) (io.ReadCloser,*FileCoding)
		Delete(code string) (*FileCoding)

		Browse(code string, name string, args Map, expires ...time.Duration) string
		Preview(code string, width,height,time int64, args Map, expires ...time.Duration) string
	}

	FileCoding struct {
		Base		string
		Name		string	//简化， 注意，名称包含扩展名，mimetype使用扩展名拿到，无需自定义
		File		string	//fid，有些存储系统会有
		Type		string	//不参与en/decode，方便读
	}


    FileHealth struct {
        Workload    int64
    }

)

type (
    fileModule struct {
        driver     coreBranch
        
        //文件配置，文件连接
		connects    map[string]FileConnect
		hashring	*hashring.HashRing
    }
)



func (module *fileModule) Encode(base, name, file string) (string) {
	
	ext := path.Ext(name)
	name = strings.TrimSuffix(name, ext)
	if strings.HasPrefix(ext, ".") {
		ext = ext[1:]
	}

    return Bigger.Encrypt(fmt.Sprintf("%s\n%s\n%s\n%s", base, name, ext, file))
}

func (module *fileModule) Decode(code string) (*FileCoding) {
	str := Bigger.Decrypt(code)
	if str == "" {
		return nil
	}
	args := strings.Split(str, "\n")
	if len(args) != 4 {
		return nil
	}

	return &FileCoding{
		Base: args[0],
		Name: args[1],
		Type: args[2],
		File: args[3],
	}
}


func (module *fileModule) Assign(name string, metadata Map, bases ...string) (string,*Error) {
	base := kDEFAULT
	if len(bases) > 0 {
		base = bases[0]
	} else if node := module.hashring.Locate(name); node != ""{
		//按权重随机分散文件存储
		base = node
	} else {
		base = kDEFAULT
	}

	if metadata==nil {
		metadata = Map{}
	}

	fb := module.Base(base); defer fb.Close()
	code := fb.Assign(name, metadata)
	return code, fb.Erred()
}

func (module *fileModule) Storage(code string, reader io.Reader) (int64,*Error) {
	data := module.Decode(code)
	if data == nil {
		return int64(0), Bigger.Erring("无效数据")
	}

	fb := module.Base(data.Base); defer fb.Close()
	size := fb.Storage(code, reader)
	return size,fb.Erred()
}
func (module *fileModule) Download(code string) (io.ReadCloser, *FileCoding, *Error) {
	data := module.Decode(code)
	if data == nil {
		return nil,nil,Bigger.Erring("无效数据")
	}

	fb := module.Base(data.Base); defer fb.Close()
	reader,data := fb.Download(code)
	return reader, data, fb.Erred()
}
func (module *fileModule) Thumbnail(code string, width,height,tttt int64) (io.ReadCloser, *FileCoding, *Error) {
	data := module.Decode(code)
	if data == nil {
		return nil,nil,Bigger.Erring("无效数据")
	}

	fb := module.Base(data.Base); defer fb.Close()
	reader,data := fb.Thumbnail(code, width,height,tttt)
	return reader, data, fb.Erred()
}



func (module *fileModule) Remove(code string) *Error {
	data := module.Decode(code)
	if data == nil {
		return Bigger.Erring("无效数据")
	}
	
	fb := module.Base(data.Base); defer fb.Close()
	fb.Delete(code)
	return fb.Erred()
}


func (module *fileModule) Browse(code string, name string, args Map, expires ...time.Duration) (string,*Error) {
	data := module.Decode(code)
	if data == nil {
		return "",Bigger.Erring("无效数据")
	}

	fb := module.Base(data.Base); defer fb.Close()
	url :=  fb.Browse(code, name, args, expires...)
	return url,fb.Erred()
}

func (module *fileModule) Preview(code string, width,height,time int64, args Map, expires ...time.Duration) (string,*Error) {
	data := module.Decode(code)
	if data == nil {
		return "",Bigger.Erring("无效数据")
	}

	fb := module.Base(data.Base); defer fb.Close()
	url :=  fb.Preview(code, width, height, time, args, expires...)
	return url,fb.Erred()
}



func (data *FileCoding) Mime() string {
	return mCONST.MimeType(data.Type)
}
func (data *FileCoding) Full() string {
	if data != nil {
		if data.Type != "" {
			return fmt.Sprintf("%s.%s", data.Name, data.Type)
		}
		return data.Name
	}
	return ""
}







//返回文件Base对象
func (module *fileModule) Base(names ...string) (FileBase) {
	name := kDEFAULT
	if len(names) > 0 {
		name = names[0]
	} else {
		for key,_ := range module.connects {
			name = key
			break
		}
	}

    if connect,ok := module.connects[name]; ok {
        return connect.Base()
    }
    panic("[文件]无效文件连接")
}




func (module *fileModule) Driver(name string, driver FileDriver, overrides ...bool) {
    if driver == nil {
        panic("[文件]驱动不可为空")
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














func (module *fileModule) connecting(name string, config FileConfig) (FileConnect, *Error) {
    if driver,ok := module.driver.chunk(config.Driver).(FileDriver); ok {
        return driver.Connect(name, config)
    }
    panic("[文件]不支持的驱动：" + config.Driver)
}




//初始化
func (module *fileModule) initing() {
	
	weights := make(map[string]int)
    for name,config := range Bigger.Config.File {
		if config.Weight > 0 {
			weights[name] = config.Weight
		}

		//连接
		connect,err := module.connecting(name, config)
		if err != nil {
			panic("[文件]连接失败：" + err.Error())
		}

		//打开连接
		err = connect.Open()
		if err != nil {
			panic("[文件]打开失败：" + err.Error())
		}

		module.connects[name] = connect
	}

	module.hashring = hashring.New(weights)
}



//退出
func (module *fileModule) exiting() {
    for _,connect := range module.connects {
        connect.Close()
    }
}
