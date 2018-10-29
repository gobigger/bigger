package bigger

import (
)

type (
    Error struct {
		lang	string
		status	string
		args	[]Any
	}
)

/*
	创建错误对象
*/
func newError(status string, args ...Any) (*Error) {
    return &Error{ DEFAULT, status, args }
}


func (e *Error) Code() int {
	return mCONST.StatusCode(e.status)
}
func (e *Error) Lang(lang string) *Error {
	e.lang = lang
	return e
}
func (e *Error) Args(args ...Any) *Error {
	e.args = args
	return e
}
func (e *Error) Error() string {
	return mCONST.LangString(e.lang, e.status, e.args...)
}
func (e *Error) String() string {
	return mCONST.LangString(e.lang, e.status, e.args...)
}

