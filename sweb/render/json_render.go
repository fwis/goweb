package render

import (
	"container/list"
	"fmt"
	"net/http"
	"strconv"

	. "github.com/fwis/goweb/sweb/context"
	. "github.com/fwis/goweb/sweb/errs"
	. "github.com/fwis/goweb/sweb/pagination"
	"github.com/fwis/goweb/sweb/zip"
)

type jsonRenderer struct {
	errHandle    ErrorHandle
	interceptors []Interceptor
	engine       *JSONRenderEngine
	option       *JSONRenderOption
}

func NewJSONRenderer(engine *JSONRenderEngine, option *JSONRenderOption) JSONRenderer {
	return &jsonRenderer{engine: engine, option: option}
}

func (m *jsonRenderer) _write(context *Context, icontent []byte) {
	/*
		ctx.W.Header().Set("Content-Length", strconv.Itoa(len(icontent)))
		ctx.W.Header().Set("Content-Type", "application/json;charset=UTF-8")
		ctx.W.Write(icontent)
	*/

	if m.option.Before != nil {
		m.option.Before(context.W, context.R)
	}

	context.W.Header().Set("Content-Type", "application/json;charset=UTF-8")
	written := false
	if m.option.Zip && len(icontent) > m.option.ZipThreshold {
		encoding := zip.GetZipAcceptEncoding(context.R)
		if encoding != "" && zip.CanZip(context.W, context.R) {
			if encoding == zip.ENCODING_GZIP {
				_, err := zip.GzipWrite(context.W, m.option.GzipLevel, icontent)
				if err != nil {
					fmt.Printf("jsonRenderer GzipWrite err=%v\n", err)
				}
				written = true
			} else if encoding == zip.ENCODING_DEFLATE {
				_, err := zip.DeflateWrite(context.W, m.option.DeflateLevel, icontent)
				if err != nil {
					fmt.Printf("jsonRenderer DeflateWrite err=%v\n", err)
				}
				written = true
			}
		}
	}

	if !written {
		context.W.Header().Set("Content-Length", strconv.Itoa(len(icontent)))
		context.W.Write(icontent)
	}
}

func (m *jsonRenderer) Error(ctx *Context, status int, msg string) {
	fmt.Printf("jsonRenderer Error, status=%d msg=%s\n", status, msg)
	if status == 0 {
		status = http.StatusInternalServerError
	}
	icontent := m.engine.Error(status, msg)
	m._write(ctx, icontent)
}

func (m *jsonRenderer) Errorv(ctx *Context, status int, err error) {
	fmt.Printf("jsonRenderer Errorv, status=%d err=%v\n", status, err)
	if status == 0 {
		status = http.StatusInternalServerError
	}
	HandErrorv(m, ctx, status, err)
}

func (m *jsonRenderer) AddInterceptor(interceptor Interceptor) {
	m.interceptors = append(m.interceptors, interceptor)
}

func (m *jsonRenderer) Intercept(context *Context, xable interface{}) bool {
	//fmt.Printf("jsonRenderer, Intercept, len(m.interceptors)=%d\n", len(m.interceptors))
	for _, interceptor := range m.interceptors {
		if !interceptor.Intercept(context, xable) {
			return false
		}
	}
	return true
}

func (m *jsonRenderer) RenderOK(ctx *Context) {
	m._write(ctx, m.engine.OK())
}

func (m *jsonRenderer) RenderJson(ctx *Context, p *Pagination, jsonbytes []byte) {
	m._write(ctx, m.engine.DataJson(p, jsonbytes))
}

func (m *jsonRenderer) RenderData(ctx *Context, p *Pagination, objs interface{}) {
	icontent, err := m.engine.DataObjs(p, objs)
	if err != nil {
		fmt.Printf("jsonRenderer RenderData, err=%v, objs=%v\n", err, objs)
		m.engine.Error(http.StatusInternalServerError, "系统异常")
	} else {
		m._write(ctx, icontent)
	}
}

func (m *jsonRenderer) RenderList(ctx *Context, p *Pagination, list *list.List) {
	icontent, err := m.engine.DataList(p, list)
	if err != nil {
		fmt.Printf("jsonRenderer RenderList, err=%v, list=%v\n", err, list)
		m.engine.Error(http.StatusInternalServerError, "系统异常")
	} else {
		m._write(ctx, icontent)
	}
}
