package session

import (
	"sync"
	"time"
)

type MemSessionAttributes struct {
	timeAccessed time.Time              //最后访问时间
	kv           map[string]interface{} //session store
	lock         sync.RWMutex
	fp           *SessionFilePersistence
	sid          string
}

func NewMemSessionAttributes(sid string, fp *SessionFilePersistence) *MemSessionAttributes {
	sxn := &MemSessionAttributes{}
	sxn.kv = make(map[string]interface{})
	sxn.timeAccessed = time.Now()
	sxn.sid = sid
	sxn.fp = fp
	return sxn
}

func (st *MemSessionAttributes) SessionID() string {
	return st.sid
}

func (st *MemSessionAttributes) TimeAccessed() time.Time {
	return st.timeAccessed
}

func (st *MemSessionAttributes) SetTimeAccessed(t time.Time) {
	st.timeAccessed = t
}

func (st *MemSessionAttributes) Set(key string, value interface{}) error {
	st.lock.Lock()
	defer st.lock.Unlock()

	if vv, ok := st.kv[key]; !ok || vv != value {
		st.kv[key] = value
		err := st.fp.Save(st)
		if err != nil {
			return err
		}
	}
	return nil
}

func (st *MemSessionAttributes) Get(key string) interface{} {
	st.lock.RLock()
	defer st.lock.RUnlock()
	if v, ok := st.kv[key]; ok {
		return v
	}
	return nil
}

func (st *MemSessionAttributes) Delete(key string) error {
	st.lock.Lock()
	defer st.lock.Unlock()
	if _, ok := st.kv[key]; ok {
		delete(st.kv, key)
		st.fp.Save(st)
	}
	return nil
}

func (st *MemSessionAttributes) Clear() error {
	st.lock.Lock()
	defer st.lock.Unlock()
	st.kv = make(map[string]interface{})
	st.fp.Clear(st.SessionID())
	return nil
}

func (st *MemSessionAttributes) Release() {
	st.lock.Lock()
	defer st.lock.Unlock()
	st.kv = nil
	st.fp = nil
}

func (st *MemSessionAttributes) Encode() ([]byte, error) {
	return EncodeGob(st.kv)
}

func (st *MemSessionAttributes) Decode(encoded []byte) error {
	kv, err := DecodeGob(encoded)
	if err != nil {
		return err
	}

	st.kv = kv
	if st.kv == nil {
		st.kv = make(map[string]interface{})
	}
	return nil
}
