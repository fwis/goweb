package session

import (
	mrand "math/rand"
	"time"
)

type Session interface {
	Attributes() SessionAttributes
	SetAttributes(attrs SessionAttributes)
	SessionID() string
}

type SessionAttributes interface {
	Set(key, value interface{}) error //set session value
	Get(key interface{}) interface{}  //get session value
	Delete(key interface{}) error     //delete session value
	Release()                         //release the resource
	Clear() error                     //delete all data
}

type SessionProvider interface {
	SessionInit() error

	TimeoutSeconds() int64

	//not change the last-access-time
	HasSession(sid string) (bool, error)

	//will update the last-access-time
	GetSession(sid string) (Session, error)

	//if exist will return nil and an error
	AddNewSession(sw Session) error

	RemoveSession(sid string) error

	RemoveExpired()
}

func init() {
	mrand.Seed(time.Now().UnixNano())
}

var provides = make(map[string]SessionProvider)

// Register makes a session provide available by the provided name.
// If Register is called twice with the same name or if driver is nil,
// it panics.
func Register(name string, provide SessionProvider) {
	if provide == nil {
		panic("session: Register provide is nil")
	}
	if _, dup := provides[name]; dup {
		panic("session: Register called twice for provider " + name)
	}
	provides[name] = provide
}
