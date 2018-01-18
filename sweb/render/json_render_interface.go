package render

import (
	"container/list"

	. "github.com/fwis/goweb/sweb/context"
	. "github.com/fwis/goweb/sweb/errs"
	. "github.com/fwis/goweb/sweb/pagination"
)

type JSONRenderer interface {
	ErrorHandle
	AddInterceptor(interceptor Interceptor)
	Intercept(context *Context, xable interface{}) bool
	RenderOK(context *Context)
	RenderJson(context *Context, p *Pagination, jsonbytes []byte)
	RenderData(context *Context, p *Pagination, data interface{})
	RenderList(context *Context, p *Pagination, list *list.List)
}
