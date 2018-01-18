package render

import (
	. "github.com/fwis/goweb/sweb/context"
)

//return false means has written http writer, break other processing.
type InterceptFunc func(context *Context, xable interface{}) bool

func (f InterceptFunc) Intercept(context *Context, xable interface{}) bool {
	return f(context, xable)
}

type Interceptor interface {
	Intercept(context *Context, xable interface{}) bool
}
