package bigger

import (
	"fmt"

)

type (
	biggerBlock struct {
		name	string
	}
)





func (block *biggerBlock) sign(name string) (string) {
	return fmt.Sprintf("%s.%s", block.name, name)
}

func (block *biggerBlock) Table(name string, configs ...Map) (Map) {
    return mDATA.Table(block.sign(name), configs...)
}
func (block *biggerBlock) View(name string, configs ...Map) (Map) {
    return mDATA.View(block.sign(name), configs...)
}
func (block *biggerBlock) Model(name string, configs ...Map) (Map) {
    return mDATA.Model(block.sign(name), configs...)
}



func (block *biggerBlock) Event(name string, config Map, overrides ...bool) {
    if config[kACTION] != nil {
        mEVENT.Router(name, config, overrides...)
    }
    if config[kREQUEST] != nil || config[kEXECUTE] != nil || config[kRESPONSE] != nil {
        mEVENT.Filter(name, config, overrides...)
    }
    if config[kFOUND] != nil || config[kERROR] != nil || config[kFAILED] != nil || config[kDENIED] != nil {
        mEVENT.Handler(name, config, overrides...)
    }
}
func (block *biggerBlock) Queue(name string, config Map, overrides ...bool) {
    if config[kACTION] != nil {
        mQUEUE.Router(name, config, overrides...)
    }
    if config[kREQUEST] != nil || config[kEXECUTE] != nil || config[kRESPONSE] != nil {
        mQUEUE.Filter(name, config, overrides...)
    }
    if config[kFOUND] != nil || config[kERROR] != nil || config[kFAILED] != nil || config[kDENIED] != nil {
        mQUEUE.Handler(name, config, overrides...)
    }
}





