package builtin

import (
	. "github.com/yatlabs/bigger"
)


var (
	StatusFound		= Bigger.Status(1, ".found", "不存在", false)
	StatusError		= Bigger.Status(2, ".error", "系统错误", false)
	StatusFailed	= Bigger.Status(3, ".error", "系统错误", false)
	StatusDenied	= Bigger.Status(4, ".denied", "拒绝访问", false)

	StatusArgsEmpty	= Bigger.Status(11, ".args.empty", "参数[%s]不可为空", false)
	StatusArgsError	= Bigger.Status(12, ".args.error", "参数[%s]不是有效的值", false)

	StatusItemEmpty	= Bigger.Status(21, ".item.empty", "%s记录不存在", false)
	StatusItemError	= Bigger.Status(22, ".item.error", "%s记录不存在", false)

	StatusDataEmpty	= Bigger.Status(31, ".data.empty", "数据[%s]不可为空", false)
	StatusDataError	= Bigger.Status(32, ".data.error", "数据[%s]不是有效的值", false)
)
