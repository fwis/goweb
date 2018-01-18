// Copyright 2012 The Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package route

import (
	"fmt"
	"net/http"
    "net/url"
	"strings"
)

//type RouteParams map[string]string

type EndSlashOption int

const (
	END_SLASH_FUZZY    = EndSlashOption(0)
	END_SLASH_REDIRECT = EndSlashOption(1)
	END_SLASH_EXACT    = EndSlashOption(2)
)

type VHandler interface {
	VServeHTTP(http.ResponseWriter, *http.Request, url.Values)
}

type VHandlerFunc func(http.ResponseWriter, *http.Request, url.Values)

func (f VHandlerFunc) VServeHTTP(w http.ResponseWriter, r *http.Request, v url.Values) {
	f(w, r, v)
}

type WrapHandlerFuncAsV func(http.ResponseWriter, *http.Request)

func (f WrapHandlerFuncAsV) VServeHTTP(w http.ResponseWriter, r *http.Request, v url.Values) {
	f(w, r)
}


//ExactEndSlash
//是否精确匹配URL的最后'/'
//  /url  和  /url/ 被看着两个不同的URL

//RedictEndSlash
//若模糊匹配URL的最后'/', 是否强制服从 route 的定义
// route:  /url      request:  /url/   ==> redirect到  /url
// route:  /url/     request:  /url    ==> redirect到  /url/
//Redirect要用 307, 不能用301,302,303, 否则POST请求被改为了 GET

//FuzzyEndSlash
//模糊匹配, 但不redirect
// RouteItem stores information to match a request and build URLs.
type RouteItem struct {
	*Router
	// Request handler for the route.
	handler VHandler

	// List of matchers.
	matchers []matcher
	// Manager for the variables from host and path.
	regexp *routeRegexpGroup

	// How to deal with the traviling slash
	// FuzzyEndSlash  = 0
	// RedictEndSlash = 1
	// ExactEndSlash  = 2
	endSlashOption EndSlashOption

	// Error resulted from building a route.
	err error

	// OnlyScheme=http,  https will force redirect to http
	// OnlyScheme=https, http will force redirect to https
	// otherwise, ignore
	onlyscheme string
}

// Match matches the route against the request.
func (r *RouteItem) Match(req *http.Request) bool {
	if r.err != nil {
		return false
	}
	// Match everything.
	for _, m := range r.matchers {
		if matched := m.Match(req); !matched {
			return false
		}
	}

	return true
}

// ----------------------------------------------------------------------------
// RouteItem attributes
// ----------------------------------------------------------------------------

// GetError returns an error resulted from building the route, if any.
func (r *RouteItem) GetError() error {
	return r.err
}

func (r *RouteItem) GetSchemeRedirectURL(req *http.Request) string {
	var is_https_req bool = (req.TLS != nil)
	if "https" == req.Header.Get("X-Scheme") {
		is_https_req = true
	}

	var redirectURL string = ""
	if r.onlyscheme == "http" && is_https_req {
		redirectURL = r.Router.httphost + req.RequestURI
	} else if r.onlyscheme == "https" && !is_https_req && r.Router.enable_to_https {
		redirectURL = r.Router.httpshost + req.RequestURI
	} else {
		redirectURL = ""
	}
	return redirectURL
}

func (r *RouteItem) SetSlashOption(option EndSlashOption) {
	r.endSlashOption = option
}

func (r *RouteItem) GetSlashOption() EndSlashOption {
	return r.endSlashOption
}

// Handler --------------------------------------------------------------------

// Handler sets a handler for the route.
func (r *RouteItem) VHandler(handler VHandler) *RouteItem {
	if r.err == nil {
		r.handler = handler
	}
	return r
}

// HandlerFunc sets a handler function for the route.
func (r *RouteItem) VHandlerFunc(f func(http.ResponseWriter, *http.Request, url.Values)) *RouteItem {
	return r.VHandler(VHandlerFunc(f))
}

// GetHandler returns the handler for the route, if any.
func (r *RouteItem) CreateHandler(w http.ResponseWriter, req *http.Request) VHandler {
	return r.handler
}

// ----------------------------------------------------------------------------
// Matchers
// ----------------------------------------------------------------------------

// matcher types try to match a request.
type matcher interface {
	Match(*http.Request) bool
}

// addMatcher adds a matcher to the route.
func (r *RouteItem) addMatcher(m matcher) *RouteItem {
	if r.err == nil {
		r.matchers = append(r.matchers, m)
	}
	return r
}

func (r *RouteItem) GetRouteParams(req *http.Request) url.Values {
    if r.regexp == nil || r.regexp.path == nil || len(r.regexp.path.varsN)==0{
        return nil
    }

    s := make(url.Values)
	//routeParams := make(RouteParams)
	r.regexp.fillURLParamters(req, s)

	return s
}

//0-no,1-yes, otherwise no path
func (r *RouteItem) EndWithSlash() int {
	if r.regexp == nil || r.regexp.path == nil {
		return 99
	}
	if strings.HasSuffix(r.regexp.path.template, "/") {
		return 1
	} else {
		return 0
	}
}

// addRegexpMatcher adds a host or path matcher and builder to a route.
func (r *RouteItem) addRegexpMatcher(tpl string, matchPrefix bool) error {
	if r.err != nil {
		return r.err
	}
	r.regexp = r.getRegexpGroup()
	// if !matchHost {
		if len(tpl) == 0 || tpl[0] != '/' {
			return fmt.Errorf("mux: path must start with a slash, got %q", tpl)
		}
		if r.regexp.path != nil {
			tpl = strings.TrimRight(r.regexp.path.template, "/") + tpl
		}
	// }

	rr, err := newRouteRegexp(tpl, matchPrefix, r.endSlashOption != END_SLASH_EXACT)
	if err != nil {
		return err
	}
    /*
	if matchHost {
		if r.regexp.path != nil {
			if err = uniqueVars(rr.varsN, r.regexp.path.varsN); err != nil {
				return err
			}
		}
		r.regexp.host = rr
	} else {
    
		if r.regexp.host != nil {
			if err = uniqueVars(rr.varsN, r.regexp.host.varsN); err != nil {
				return err
			}
		}
        
		r.regexp.path = rr
	}
    */
    r.regexp.path = rr
	r.addMatcher(rr)
	return nil
}

// Headers --------------------------------------------------------------------

// headerMatcher matches the request against header values.
type headerMatcher map[string]string

func (m headerMatcher) Match(r *http.Request) bool {
	return matchMap(m, r.Header, true)
}

// Headers adds a matcher for request header values.
// It accepts a sequence of key/value pairs to be matched. For example:
//
//     r := mux.NewRouter()
//     r.Headers("Content-Type", "application/json",
//               "X-Requested-With", "XMLHttpRequest")
//
// The above route will only match if both request header values match.
//
// It the value is an empty string, it will match any value if the key is set.
func (r *RouteItem) Headers(pairs ...string) *RouteItem {
	if r.err == nil {
		var headers map[string]string
		headers, r.err = mapFromPairs(pairs...)
		return r.addMatcher(headerMatcher(headers))
	}
	return r
}

// Host -----------------------------------------------------------------------

// Host adds a matcher for the URL host.
// It accepts a template with zero or more URL variables enclosed by {}.
// Variables can define an optional regexp pattern to me matched:
//
// - {name} matches anything until the next dot.
//
// - {name:pattern} matches the given regexp pattern.
//
// For example:
//
//     r := mux.NewRouter()
//     r.Host("www.domain.com")
//     r.Host("{subdomain}.domain.com")
//     r.Host("{subdomain:[a-z]+}.domain.com")
//
// Variable names must be unique in a given route. They can be retrieved
// calling mux.Vars(request).
// func (r *RouteItem) Host(tpl string) *RouteItem {
// 	r.err = r.addRegexpMatcher(tpl, true, false)
// 	return r
// }

// MatcherFunc ----------------------------------------------------------------

// MatcherFunc is the function signature used by custom matchers.
type MatcherFunc func(*http.Request) bool

func (m MatcherFunc) Match(r *http.Request) bool {
	return m(r)
}

// MatcherFunc adds a custom function to be used as request matcher.
func (r *RouteItem) MatcherFunc(f MatcherFunc) *RouteItem {
	return r.addMatcher(f)
}

// Methods --------------------------------------------------------------------

type methodMatcher string

func (m methodMatcher) Match(r *http.Request) bool {
	return string(m) == r.Method
}

func (r *RouteItem) Method(method string) *RouteItem {
	return r.addMatcher(methodMatcher(strings.ToUpper(method)))
}

// methodsMatcher matches the request against HTTP methods.
type methodsMatcher []string

func (m methodsMatcher) Match(r *http.Request) bool {
	return matchInArray(m, r.Method)
}

// Methods adds a matcher for HTTP methods.
// It accepts a sequence of one or more methods to be matched, e.g.:
// "GET", "POST", "PUT".
func (r *RouteItem) _methods(methods ...string) *RouteItem {
	for k, v := range methods {
		methods[k] = strings.ToUpper(v)
	}
	return r.addMatcher(methodsMatcher(methods))
}

// Path -----------------------------------------------------------------------

// Path adds a matcher for the URL path.
// It accepts a template with zero or more URL variables enclosed by {}.
// Variables can define an optional regexp pattern to me matched:
//
// - {name} matches anything until the next slash.
//
// - {name:pattern} matches the given regexp pattern.
//
// For example:
//
//     r := mux.NewRouter()
//     r.Path("/products/").Handler(ProductsHandler)
//     r.Path("/products/{key}").Handler(ProductsHandler)
//     r.Path("/articles/{category}/{id:[0-9]+}").
//       Handler(ArticleHandler)
//
// Variable names must be unique in a given route. They can be retrieved
// calling mux.Vars(request).
func (r *RouteItem) Path(tpl string) *RouteItem {
	r.err = r.addRegexpMatcher(tpl, false)
	return r
}

// PathPrefix -----------------------------------------------------------------

// PathPrefix adds a matcher for the URL path prefix.
func (r *RouteItem) PathPrefix(tpl string) *RouteItem {
	r.endSlashOption = END_SLASH_EXACT
	r.err = r.addRegexpMatcher(tpl, true)
	return r
}

// Query ----------------------------------------------------------------------

// queryMatcher matches the request against URL queries.
type queryMatcher map[string]string

func (m queryMatcher) Match(r *http.Request) bool {
	return matchMap(m, r.URL.Query(), false)
}

// Queries adds a matcher for URL query values.
// It accepts a sequence of key/value pairs. For example:
//
//     r := mux.NewRouter()
//     r.Queries("foo", "bar", "baz", "ding")
//
// The above route will only match if the URL contains the defined queries
// values, e.g.: ?foo=bar&baz=ding.
//
// It the value is an empty string, it will match any value if the key is set.
func (r *RouteItem) Queries(pairs ...string) *RouteItem {
	if r.err == nil {
		var queries map[string]string
		queries, r.err = mapFromPairs(pairs...)
		return r.addMatcher(queryMatcher(queries))
	}
	return r
}

// Schemes --------------------------------------------------------------------
func (r *RouteItem) OnlyScheme(scheme string) *RouteItem {
	r.onlyscheme = strings.ToLower(scheme)
	return r
}

// getRegexpGroup returns regexp definitions from this route.
func (r *RouteItem) getRegexpGroup() *routeRegexpGroup {
	if r.regexp == nil {
		r.regexp = new(routeRegexpGroup)
	}
	return r.regexp
}
