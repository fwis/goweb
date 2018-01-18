package render

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	. "github.com/fwis/goweb/sweb/context"
	. "github.com/fwis/goweb/sweb/errs"
	"github.com/fwis/goweb/sweb/zip"
)

type htmlRenderer struct {
	errHandle    ErrorHandle
	interceptors []Interceptor
	engine       *HTMLRenderEngine
	option       *HTMLRenderOption
}

func NewDefaultHTMLRenderer(tplRootDir string, f template.FuncMap) HTMLRenderer {
	htmlRenderOption := NewDefaultHTMLRenderOption()
	htmlRenderEngine := NewDefaultHTMLRenderEngine(tplRootDir, f)
	return NewHTMLRenderer(htmlRenderEngine, htmlRenderOption)
}

func NewHTMLRenderer(engine *HTMLRenderEngine, option *HTMLRenderOption) HTMLRenderer {
	return &htmlRenderer{
		engine: engine,
		option: option,
	}
}

func (m *htmlRenderer) Error(ctx *Context, status int, msg string) {
	if m.errHandle != nil {
		m.errHandle.Error(ctx, status, msg)
		return
	}

	if status < http.StatusOK {
		status = http.StatusInternalServerError
	}

	http.Error(ctx.W, msg, status)
}

func (m *htmlRenderer) Errorv(ctx *Context, status int, err error) {
	if m.errHandle != nil {
		m.errHandle.Errorv(ctx, status, err)
		return
	}
	HandErrorv(m, ctx, status, err)
}

func (m *htmlRenderer) SetErrorHandle(errHandle ErrorHandle) {
	m.errHandle = errHandle
}

func (m *htmlRenderer) AddInterceptor(interceptor Interceptor) {
	m.interceptors = append(m.interceptors, interceptor)
}

func (m *htmlRenderer) Intercept(context *Context, xable interface{}) bool {
	//fmt.Printf("htmlRenderer, Intercept, len(m.interceptors)=%d\n", len(m.interceptors))
	for _, interceptor := range m.interceptors {
		if !interceptor.Intercept(context, xable) {
			return false
		}
	}
	return true
}

func (m *htmlRenderer) RenderRaw(context *Context, raw []byte) {
	if m.option.Before != nil {
		m.option.Before(context.W, context.R)
	}
	context.W.Header().Set("Content-Type", "text/html; charset=utf-8")
	written := false
	if m.option.Zip && len(raw) > m.option.ZipThreshold {
		encoding := zip.GetZipAcceptEncoding(context.R)
		if encoding != "" && zip.CanZip(context.W, context.R) {
			if encoding == zip.ENCODING_GZIP {
				_, err := zip.GzipWrite(context.W, m.option.GzipLevel, raw)
				if err != nil {
					fmt.Printf("htmlRenderer GzipWrite err=%v\n", err)
				}
				written = true
			} else if encoding == zip.ENCODING_DEFLATE {
				_, err := zip.DeflateWrite(context.W, m.option.DeflateLevel, raw)
				if err != nil {
					fmt.Printf("htmlRenderer DeflateWrite err=%v\n", err)
				}
				written = true
			}
		}
	}

	if !written {
		context.W.Header().Set("Content-Length", strconv.Itoa(len(raw)))
		context.W.Write(raw)
	}
}

func (m *htmlRenderer) Render(context *Context, xable interface{}, tplName string, tplLayout string, data interface{}) {
	icontent, err := m.engine.Render(tplName, tplLayout, data)
	if err != nil {
		fmt.Printf("htmlRenderer Render %s err=%v\n", tplName, err)
		http.Error(context.W, "System render html template error", http.StatusInternalServerError)
		return
	}

	if m.option.Before != nil {
		m.option.Before(context.W, context.R)
	}

	context.W.Header().Set("Content-Type", "text/html; charset=utf-8")
	written := false
	if m.option.Zip && len(icontent) > m.option.ZipThreshold {
		encoding := zip.GetZipAcceptEncoding(context.R)
		if encoding != "" && zip.CanZip(context.W, context.R) {
			if encoding == zip.ENCODING_GZIP {
				_, err := zip.GzipWrite(context.W, m.option.GzipLevel, icontent)
				if err != nil {
					fmt.Printf("htmlRenderer GzipWrite err=%v\n", err)
				}
				written = true
			} else if encoding == zip.ENCODING_DEFLATE {
				_, err := zip.DeflateWrite(context.W, m.option.DeflateLevel, icontent)
				if err != nil {
					fmt.Printf("htmlRenderer DeflateWrite err=%v\n", err)
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
