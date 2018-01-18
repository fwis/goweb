package errs

import (
	"net/http"

	. "github.com/fwis/goweb/sweb/context"
)

type ErrorHandle interface {
	Error(context *Context, status int, desc string)
	Errorv(context *Context, status int, err error)
}

type defaultErrorHandler struct {
}

func NewDefaultErrorHandle() ErrorHandle {
	return &defaultErrorHandler{}
}

func (m *defaultErrorHandler) Error(context *Context, status int, desc string) {
	if status < http.StatusOK {
		status = http.StatusInternalServerError
	}
	http.Error(context.W, desc, status)
}

func (m *defaultErrorHandler) Errorv(context *Context, status int, err error) {
	HandErrorv(m, context, status, err)
}

func HandErrorv(errhandle ErrorHandle, context *Context, status int, err error) {
	if _, ok := err.(*PubError); ok {
		errhandle.Error(context, status, err.Error())
	} else {
		errhandle.Error(context, status, "系统异常")
	}
}
