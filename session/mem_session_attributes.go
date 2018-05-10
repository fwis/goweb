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

func (st *MemSessionAttributes) Set(key string, value interface{}) error {
	//fmt.Printf("session.attribute, set key=%v,value=%v\n", key, value)
	st.kv.Store(key, value)
	return nil
}

func (st *MemSessionAttributes) Get(key string) interface{} {
	v, ok := st.kv.Load(key)
	//fmt.Printf("session.attribute, get key=%v,v=%v,ok=%v\n", key, v, ok)
	if !ok {
		return nil
	}
	return v
}

func (st *MemSessionAttributes) Delete(key string) error {
	//fmt.Printf("session.attribute, del key=%v\n", key)
	st.kv.Delete(key)
	return nil
}

func (st *MemSessionAttributes) Clear() error {
	//fmt.Printf("session.attribute, clear\n")
	st.kv = &sync.Map{}
	return nil
}

func (st *MemSessionAttributes) Release() {
	//fmt.Printf("session.attribute, release\n")
	st.kv = nil
}
