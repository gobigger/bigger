package bigger

import (
	"fmt"
	"strings"
    "regexp"
)

type (
    constModule struct {
		mime	coreBranch
		status	coreBranch
		regular	coreBranch
		lang	coreBranch
    }
)






func (module *constModule) Mime(config Map, overrides ...bool) {
    override := true
    if len(overrides) > 0 {
        override = overrides[0]
    }
	
	for k,v := range config {
		if vv,ok := v.(string); ok {
			if override {
				module.mime.chunking(k, vv)
			} else {
				if module.mime.chunkdata(k) == nil {
					module.mime.chunking(k, vv)
				}
			}
		}
	}
}
func (module *constModule) MimeType(name string, defs ...string) string {

    if strings.Contains(name, "/") {
        return name
	}

	if strings.Contains(name, ".") {
		name = strings.Replace(name,".", "", -1)
	}

	if vv,ok := module.mime.chunkdata(name).(string); ok {
		return vv
	}
	if vv,ok := module.mime.chunkdata("*").(string); ok {
		return vv
	}
	if len(defs) > 0 {
		return defs[0]
	}
	
	return "application/octet-stream"
}
func (module *constModule) TypeMime(mime string, defs ...string) string {
    if strings.Contains(mime, "/") == false {
        return mime
    }
	mimes := module.mime.chunks()
    for _,v := range mimes {
        if vs,ok := v.data.(string); ok && vs==mime {
            return v.name
        }
    }
	if len(defs) > 0 {
		return defs[0]
	}
    return ""
}










func (module *constModule) Status(config Map, overrides ...bool) {
    override := true
    if len(overrides) > 0 {
        override = overrides[0]
    }
	
	for k,v := range config {
		if vv,ok := v.(int); ok {
			if override {
				module.status.chunking(k, vv)
			} else {
				if module.status.chunkdata(k) == nil {
					module.status.chunking(k, vv)
				}
			}
		}
	}
}
func (module *constModule) StatusCode(name string, defs ...int) int {
	if vv,ok := module.status.chunkdata(name).(int); ok {
		return vv
	}
	if len(defs) > 0 {
		return defs[0]
	}
	return -1
}
func (module *constModule) CodeStatus(code int, defs ...string) string {
	statuses := module.status.chunks()
    for _,v := range statuses {
        if vs,ok := v.data.(int); ok && vs==code {
            return v.name
        }
    }
	if len(defs) > 0 {
		return defs[0]
	}
	return ""
}

func (module *constModule) Results() (Map) {
	chunks := module.status.chunks()
	codes := Map{}
	for _,chunk := range chunks {
		code := fmt.Sprintf("%v", chunk.data)
		codes[code] = module.LangString(DEFAULT, chunk.name)
	}
	return codes
}



func (module *constModule) Regular(config Map, overrides ...bool) {
	
    override := true
    if len(overrides) > 0 {
        override = overrides[0]
    }
	
	for k,v := range config {
		vvs := []string{}

		switch vv := v.(type) {
		case string:
			vvs = append(vvs, vv)
		case []string:
			vvs = append(vvs, vv...)
		}

		if len(vvs) > 0 {
			if override {
				module.regular.chunking(k, vvs)
			} else {
				if module.regular.chunkdata(k) == nil {
					module.regular.chunking(k, vvs)
				}
			}
		}
	}
}
func (module *constModule) RegularExpress(name string, defs ...string) ([]string) {
	data := module.regular.chunkdata(name)
	if vv,ok := data.(string); ok {
		return []string{ vv }
	} else if vv,ok := data.([]string); ok {
		return vv
	}
	return defs
}





//lang做为前缀，加.和key分开
func (module *constModule) Lang(lang string, config Map, overrides ...bool) {
    override := true
    if len(overrides) > 0 {
        override = overrides[0]
    }

	for k,v := range config {
		if vv,ok := v.(string); ok {
			key := fmt.Sprintf("%v.%v", lang, k)
			if override {
				module.lang.chunking(key, vv)
			} else {
				if module.lang.chunkdata(key) == nil {
					module.lang.chunking(key, vv)
				}
			}
		}
	}
}

func (module *constModule) LangString(lang, name string, args ...Any) string {
	if lang == "" {
		lang = DEFAULT
	}


	defaultKey := fmt.Sprintf("%v.%v", DEFAULT, name)
	langKey := fmt.Sprintf("%v.%v", lang, name)

	langStr := ""

	if vv,ok := module.lang.chunkdata(langKey).(string); ok && vv != "" {
		langStr = vv
	} else if vv,ok := module.lang.chunkdata(defaultKey).(string); ok && vv != "" {
		langStr = vv
	} else {
		langStr = name
	}
	
	if len(args) > 0 {
		return fmt.Sprintf(langStr, args...)
	}
	return langStr
}





func (module *constModule) RegularMatch(value, regular string) bool {
    matchs := module.RegularExpress(regular)
    for _,v := range matchs {
        regx := regexp.MustCompile(v)
        if regx.MatchString(value) {
            return true
        }
    }
    return false
}


