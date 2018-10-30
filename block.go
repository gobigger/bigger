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
    name = block.sign(name)
    return Bigger.Table(name, configs...)
}
func (block *biggerBlock) View(name string, configs ...Map) (Map) {
    name = block.sign(name)
    return Bigger.View(name, configs...)
}
func (block *biggerBlock) Model(name string, configs ...Map) (Map) {
    name = block.sign(name)
    return Bigger.Model(name, configs...)
}



func (block *biggerBlock) Event(name string, config Map, overrides ...bool) {
    name = block.sign(name)
    Bigger.Queue(name, config, overrides...)
}
func (block *biggerBlock) Queue(name string, config Map, overrides ...bool) {
    name = block.sign(name)
    Bigger.Queue(name, config, overrides...)
}

func (block *biggerBlock) Register(name string, config Map, overrides ...bool) {
    name = block.sign(name)
    Bigger.Register(name, config, overrides...)
}


