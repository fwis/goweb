package route

import (
	"net/http"
	"net/url"
	"strings"
)

type FilterFunc func(http.ResponseWriter, *http.Request) bool

func (f FilterFunc) FilterHTTP(w http.ResponseWriter, r *http.Request) bool {
	return f(w, r)
}

type Filter interface {
	FilterHTTP(http.ResponseWriter, *http.Request) bool
}

type Router struct {
	// Configurable Handler to be used when no route matches.
	NotFoundVHandler VHandler
	//first filters, then items
	filters []Filter
	// Routes to be matched, in order.
	items           []*RouteItem
	httphost        string
	httpshost       string
	enable_to_https bool //是否允许重定向到 https
}

// NewRouter returns a new router instance.
func NewRouter() *Router {
	router := &Router{}
	return router
}

// NewRouter returns a new router instance.
func NewRouterWithHost(httphost, httpshost string, enable_to_https bool) *Router {
	router := &Router{}
	router.httphost = httphost
	router.httpshost = httpshost
	if router.httphost == router.httpshost {
		enable_to_https = false
	}
	router.enable_to_https = enable_to_https
	return router
}

func (r *Router) NewRouteItem() *RouteItem {
	routeitem := &RouteItem{Router: r}
	r.items = append(r.items, routeitem)
	return routeitem
}

func (r *Router) VGET(path string, handler VHandler) *RouteItem {
	return r.NewRouteItem().Method("GET").Path(path).VHandler(handler)
}

func (r *Router) VGETfunc(path string, f func(http.ResponseWriter, *http.Request, url.Values)) *RouteItem {
	return r.NewRouteItem().Method("GET").Path(path).VHandlerFunc(f)
}

func (r *Router) VPOST(path string, handler VHandler) *RouteItem {
	return r.NewRouteItem().Method("POST").Path(path).VHandler(handler)
}

func (r *Router) VPOSTfunc(path string, f func(http.ResponseWriter, *http.Request, url.Values)) *RouteItem {
	return r.NewRouteItem().Method("POST").Path(path).VHandlerFunc(f)
}

func (r *Router) VDELETE(path string, handler VHandler) *RouteItem {
	return r.NewRouteItem().Method("DELETE").Path(path).VHandler(handler)
}

func (r *Router) VDELETEfunc(path string, f func(http.ResponseWriter, *http.Request, url.Values)) *RouteItem {
	return r.NewRouteItem().Method("DELETE").Path(path).VHandlerFunc(f)
}

func (r *Router) VHandle(path string, handler VHandler) *RouteItem {
	return r.NewRouteItem().Path(path).VHandler(handler)
}

func (r *Router) HandleFunc(path string, f func(http.ResponseWriter, *http.Request, url.Values)) *RouteItem {
	return r.NewRouteItem().Path(path).VHandlerFunc(f)
}

func (r *Router) Filter(filter Filter) {
	r.filters = append(r.filters, filter)
}

func (r *Router) FilterFunc(f func(http.ResponseWriter, *http.Request) bool) {
	r.Filter(FilterFunc(f))
}

// Match matches registered items against the request.
func (r *Router) Match(req *http.Request) *RouteItem {
	for _, item := range r.items {
		if item.Match(req) {
			return item
		}
	}
	return nil
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	for _, filter := range r.filters {
		if ok := filter.FilterHTTP(w, req); !ok {
			return
		}
	}

	var handler VHandler
	var routeParameters url.Values
	routeitem := r.Match(req)
	if routeitem != nil {
		//fmt.Printf("router.ServHTTP, matched for url=%v\n", req.URL)
		//{{处理 https和 http 的redirect
		redirectURL := routeitem.GetSchemeRedirectURL(req)

		if redirectURL != "" {
			//fmt.Printf("Router ServeHTTP redirectURL=%s\n", redirectURL)
			http.Redirect(w, req, redirectURL, http.StatusSeeOther)
			return
		}
		//}}

		//{{处理 RedirectSlash
		if routeitem.GetSlashOption() == END_SLASH_REDIRECT {
			p1 := strings.HasSuffix(req.URL.Path, "/")
			n2 := routeitem.EndWithSlash()
			if n2 <= 1 {
				p2 := (n2 == 1)
				if p1 != p2 {
					u, _ := url.Parse(req.URL.String())
					if p1 {
						u.Path = u.Path[:len(u.Path)-1]
					} else {
						u.Path += "/"
					}
					//fmt.Printf("Router ServeHTTP RedirectSlash redirectURL=%s\n", u.String())
					//此处Redirect要用 307, 不能用301,302,303, 否则POST请求被改为了 GET
					http.Redirect(w, req, u.String(), http.StatusTemporaryRedirect)
					return
				}
			}
		}
		//}}

		handler = routeitem.CreateHandler(w, req)
		routeParameters = routeitem.GetRouteParams(req)
	}

	if handler == nil {
		if r.NotFoundVHandler == nil {
			r.NotFoundVHandler = WrapHandlerFuncAsV(http.NotFound)
			//r.NotFoundHandler = http.NotFoundHandler()
		}
		handler = r.NotFoundVHandler
	}

	handler.VServeHTTP(w, req, routeParameters)
}
