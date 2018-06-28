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
	SessionID() string
	TimeAccessed() time.Time
	SetTimeAccessed(t time.Time)
	Set(key string, value interface{}) error //set session value
	Get(key string) interface{}              //get session value
	Delete(key string) error                 //delete session value
	Release()                                //release the resource
	Clear() error                            //delete all data
	Encode() ([]byte, error)
	Decode(encoded []byte) error
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

	NewSessionAttributes(sid string) SessionAttributes

	RemoveExpired()

	PersistSessions()
}

func init() {
	mrand.Seed(time.Now().UnixNano())
}
