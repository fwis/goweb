package render

import (
	. "github.com/smithfox/goweb/sweb/context"
	. "github.com/smithfox/goweb/sweb/errs"
)

type HTMLRenderer interface {
	SetErrorHandle(errHandle ErrorHandle)
	AddInterceptor(interceptor Interceptor)
	Intercept(context *Context, xable interface{}) bool
	RenderRaw(context *Context, raw []byte)
	Render(context *Context,
		xable interface{},
		tplName string,
		tplLayout string,
		data interface{},
	)
}
