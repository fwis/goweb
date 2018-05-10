package session

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

type SessionMgrUsingCookie struct {
	cookieName    string //private cookiename
	extCookieName string //cookie session token extend info: IP, UID
	provider      SessionProvider
	httpOnly      bool
	domain        string
	Secure        bool   //为true时,只有https才传递到服务器端。http是不会传递的
	HashFuncName  string //support md5 & sha1
	HashKey       string //
	MaxAge        int64  //0表示不设置、-1表示立即删除、其他表示多少秒
	lock          sync.RWMutex
}

func isLocalHost(host string) bool {
	return strings.HasPrefix(host, "localhost") || strings.HasPrefix(host, "127.0.0.1") || strings.HasPrefix(host, "192.168.")
}

func normalizeCookieDomain(domain string) string {
	var cookieDomain string = ""
	if strings.HasPrefix(domain, "www.") {
		cookieDomain = domain[3:]
	} else if strings.HasPrefix(domain, ".") {
		cookieDomain = domain
	} else if isLocalHost(domain) {
		cookieDomain = ""
	} else {
		cookieDomain = "." + domain
	}
	return cookieDomain
}

func NewSessionMgrUsingCookie(provideName string, cookieName string, maxage int64, domain string, disableJsAccess bool, onlyUseHttps bool) (*SessionMgrUsingCookie, error) {
	provider, ok := provides[provideName]
	if !ok {
		return nil, fmt.Errorf("session: unknown provide %q (forgotten import?)", provideName)
	}

	//provider.SessionInit(maxlifetime, "")

	return &SessionMgrUsingCookie{
		provider:   provider,
		cookieName: cookieName,
		domain:     normalizeCookieDomain(domain),
		httpOnly:   disableJsAccess,
		MaxAge:     maxage,
		Secure:     onlyUseHttps,
		//HashFuncName: "sha1",
		//HashKey:      "changethedefaultkey",
	}, nil
}

func (manager *SessionMgrUsingCookie) SetHttpOnly(f bool) {
	manager.httpOnly = f
}

func (manager *SessionMgrUsingCookie) SetCookieDomain(domain string) {
	manager.domain = normalizeCookieDomain(domain)
}

//get SessionCookie, is sid
func (manager *SessionMgrUsingCookie) GetSessionCookie(r *http.Request) (string, error) {
	//fmt.Printf("GetSessionCookie, manager.cookieName=%s\n", manager.cookieName)
	cookie, err := r.Cookie(manager.cookieName)
	if err != nil {
		return "", err
	}
	//fmt.Printf("GetSessionCookie, cookie=%#v\n", *cookie)

	return cookie.Value, nil
}

//set session cookie
func (manager *SessionMgrUsingCookie) SetSessionCookie(w http.ResponseWriter, sid string) {
	cookie := &http.Cookie{
		Name:     manager.cookieName,
		Value:    url.QueryEscape(sid),
		Path:     "/",
		Domain:   manager.domain,
		HttpOnly: manager.httpOnly, //是否禁止js读取cookie,
		Secure:   manager.Secure,   //是否只应用于 https
		MaxAge:   0,
	}

	if manager.MaxAge < (math.MaxInt32 - 2) {
		cookie.MaxAge = int(manager.MaxAge)
	}

	cookie.Expires = time.Now().Add(time.Duration(manager.MaxAge) * time.Second)

	http.SetCookie(w, cookie)
}

//delete session cookie
func (manager *SessionMgrUsingCookie) DeleteSessionCookie(w http.ResponseWriter) {
	expiration := time.Now().AddDate(-1, 0, 0)
	cookie := http.Cookie{
		Name:     manager.cookieName,
		Path:     "/",
		Domain:   manager.domain,
		HttpOnly: true,
		Expires:  expiration,
		MaxAge:   -1,
	}
	http.SetCookie(w, &cookie)
}

func (manager *SessionMgrUsingCookie) GetSessionExtCookie(r *http.Request, cookieName string) (string, error) {
	cookie, err := r.Cookie(cookieName)
	if err != nil {
		return "", err
	}
	return url.QueryUnescape(cookie.Value)
}

func (manager *SessionMgrUsingCookie) SetSessionExtCookie(w http.ResponseWriter, cookieName string, value string) {
	cookie := &http.Cookie{
		Name:     cookieName,
		Value:    url.QueryEscape(value),
		Path:     "/",
		Domain:   manager.domain,
		HttpOnly: true,
		Secure:   manager.Secure}

	if manager.MaxAge < (math.MaxInt32 - 2) {
		cookie.MaxAge = int(manager.MaxAge)
	}

	cookie.Expires = time.Now().Add(time.Duration(int64(manager.MaxAge) * int64(time.Second)))

	http.SetCookie(w, cookie)
}

func (manager *SessionMgrUsingCookie) DeleteSessionExtCookie(w http.ResponseWriter, cookieName string) {
	expiration := time.Now().AddDate(-1, 0, 0)
	cookie := http.Cookie{
		Name:     cookieName,
		Path:     "/",
		Domain:   manager.domain,
		HttpOnly: true,
		Expires:  expiration,
		MaxAge:   -1}
	http.SetCookie(w, &cookie)
}

//get Session By sessionId
func (manager *SessionMgrUsingCookie) GetSession(sid string) (Session, error) {
	return manager.provider.GetSession(sid)
}

//get Session By sessionId
func (manager *SessionMgrUsingCookie) AddNewSession(sw Session) error {
	return manager.provider.AddNewSession(sw)
}

func (manager *SessionMgrUsingCookie) RemoveSession(sid string) {
	manager.provider.RemoveSession(sid)
}

func (manager *SessionMgrUsingCookie) GC() {
	manager.provider.RemoveExpired()
	time.AfterFunc(time.Duration(manager.MaxAge)*time.Second, func() { manager.GC() })
}

//remote_addr cruunixnano randdata
func (manager *SessionMgrUsingCookie) NewSessionId(r *http.Request) string {
	const randlen int = 12
	randbb := make([]byte, randlen)
	if _, err := io.ReadFull(rand.Reader, randbb); err != nil {
		return ""
	}

	sig := fmt.Sprintf("%s%d", r.RemoteAddr, time.Now().UnixNano())
	signbytes := []byte(sig)
	signlen := len(signbytes)
	h := md5.New()
	n, _ := h.Write(signbytes)
	for n < signlen {
		//fmt.Printf("!!!why1")
		m, _ := h.Write(signbytes[n:])
		n += m
	}

	n, _ = h.Write(randbb)
	for n < randlen {
		//fmt.Printf("!!!why2")
		m, _ := h.Write(randbb[n:])
		n += m
	}

	return hex.EncodeToString(h.Sum(nil))
}
