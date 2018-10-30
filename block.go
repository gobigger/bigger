package bigger

import (
	"fmt"

)

type (
	biggerBlock struct {
		name	string
	}
)





func (block *biggerBlock) Name(name string) (string) {
	return fmt.Sprintf("%s.%s", block.name, name)
}

func (block *biggerBlock) Table(name string, configs ...Map) (Map) {
    name = block.Name(name)
    return Bigger.Table(name, configs...)
}
func (block *biggerBlock) View(name string, configs ...Map) (Map) {
    name = block.Name(name)
    return Bigger.View(name, configs...)
}
func (block *biggerBlock) Model(name string, configs ...Map) (Map) {
    name = block.Name(name)
    return Bigger.Model(name, configs...)
}
func (block *biggerBlock) Fields(name string, keys []string, exts ...Map) (Map) {
    name = block.Name(name)
    return Bigger.Fields(name, keys, exts...)
}
func (block *biggerBlock) Enums(name string, field string) (Map) {
    name = block.Name(name)
    return Bigger.Enums(name, field)
}


func (block *biggerBlock) Event(name string, config Map, overrides ...bool) {
    name = block.Name(name)
    Bigger.Queue(name, config, overrides...)
}
func (block *biggerBlock) Queue(name string, config Map, overrides ...bool) {
    name = block.Name(name)
    Bigger.Queue(name, config, overrides...)
}

func (block *biggerBlock) Register(name string, config Map, overrides ...bool) {
    name = block.Name(name)
    Bigger.Register(name, config, overrides...)
}


