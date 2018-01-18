package session

import (
	"sync"
	"time"
)

type MemSessionAttributes struct {
	timeAccessed time.Time //最后访问时间
	kv           *sync.Map
}

func NewMemSessionAttributes() *MemSessionAttributes {
	sxn := &MemSessionAttributes{}
	sxn.kv = &sync.Map{}
	sxn.timeAccessed = time.Now()
	return sxn
}

func (st *MemSessionAttributes) Set(key, value interface{}) error {
	st.kv.Store(key, value)
	return nil
}

func (st *MemSessionAttributes) Get(key interface{}) interface{} {
	v, ok := st.kv.Load(key)
	if !ok {
		return nil
	}
	return v
}

func (st *MemSessionAttributes) Delete(key interface{}) error {
	st.kv.Delete(key)
	return nil
}

func (st *MemSessionAttributes) Clear() error {
	st.kv = &sync.Map{}
	return nil
}

func (st *MemSessionAttributes) Release() {
	st.kv = nil
}
