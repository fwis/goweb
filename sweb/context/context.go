package context

import (
	"net/http"
    "net/url"
	"strings"
	// "strconv"
)

type IRouteParamSetter interface {
    SetRouteParamters(url.Values)
}

//坚决杜绝往Context加很多功能
type Context struct {
	R *http.Request
	W http.ResponseWriter
}

//301
//永久重定向,告诉客户端以后应从新地址访问,会影响SEO
func (ctx *Context) RedirectPermanently(redirectURL string) {
	http.Redirect(ctx.W, ctx.R, redirectURL, http.StatusMovedPermanently)
}

//302
//作为HTTP1.0的标准,以前叫做Moved Temporarily ,现在叫Found. 现在使用只是为了兼容性的处理
//HTTP 1.1 有303 和307作为详细的补充,其实是对302的细化
func (ctx *Context) RedirectFound(redirectURL string) {
	http.Redirect(ctx.W, ctx.R, redirectURL, http.StatusFound)
}

//303
//对于POST请求，它表示请求已经被服务端处理，客户端可以接着使用GET方法去请求Location里的URI
func (ctx *Context) RedirectSeeOther(redirectURL string) {
	http.Redirect(ctx.W, ctx.R, redirectURL, http.StatusSeeOther)
}

//307
//对于POST请求，表示请求还没有被处理，客户端会向Location里的URI重新发起POST请求
//意味着 POST 请求会被再次发送到服务端
func (ctx *Context) RedirectTemporary(redirectURL string) {
	http.Redirect(ctx.W, ctx.R, redirectURL, http.StatusTemporaryRedirect)
}

func (ctx *Context) GetHeader(key string) string {
	return ctx.R.Header.Get(key)
}

func (ctx *Context) IsAjax() bool {
	return ctx.GetHeader("X-Requested-With") == "XMLHttpRequest"
}

func (ctx *Context) IsWebsocket() bool {
	return ctx.GetHeader("Upgrade") == "websocket"
}

func (ctx *Context) Proxy() []string {
	var ips string = ""
	ips = ctx.GetHeader("X-Real-IP")
	if len(ips) < 7 {
		ips = ctx.GetHeader("X-Forwarded-For")
		if len(ips) < 7 {
			ips = ctx.GetHeader("HTTP_X_FORWARDED_FOR")
			if len(ips) < 7 {
				ips = ctx.GetHeader("Proxy-Client-IP")
				if len(ips) < 7 {
					ips = ctx.GetHeader("WL-Proxy-Client-IP")
				}
			}
		}
	}
	if ips != "" {
		return strings.Split(ips, ",")
	}
	return []string{}
}

func (ctx *Context) IP() string {
	ips := ctx.Proxy()
	if len(ips) > 0 && ips[0] != "" {
		return ips[0]
	}
	ips = strings.Split(ctx.R.RemoteAddr, ":")
	if len(ips) == 2 {
		return ips[0]
	} else if len(ips) == 1 {
		return ips[0]
	}
	return "127.0.0.1"
}
