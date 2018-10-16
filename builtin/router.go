package builtin

import (
	. "github.com/yatlabs/bigger"
)

func init() {

	Bigger.Router("*._doc_", Map{
		"name": "系统文档", "text": "系统文档",
		"action": func(ctx *Context){
			ctx.Data["doc"] = Bigger.Document(ctx.Site)
			ctx.View("_doc_")
		},
	})

}


