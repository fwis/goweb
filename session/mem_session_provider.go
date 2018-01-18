package session

import (
	"container/list"
	"errors"
	"sync"
	"time"
)

var mempder = &MemSessionProvider{list: list.New(), sessions: make(map[string]*list.Element)}

type MemSessionProvider struct {
	lock           sync.RWMutex             //用来锁
	sessions       map[string]*list.Element //用来存储在内存
	list           *list.List               //用来做gc
	timeoutseconds int64
	savePath       string
}

func NewMemSessionProvider(timeoutseconds int64) *MemSessionProvider {
	provider := &MemSessionProvider{}
	provider.list = &list.List{}
	provider.timeoutseconds = timeoutseconds
	return provider
}

func (pder *MemSessionProvider) TimeoutSeconds() int64 {
	return pder.timeoutseconds
}

func (pder *MemSessionProvider) SessionInit() error {
	return nil
}

//not change the last-access-time
func (pder *MemSessionProvider) HasSession(sid string) (bool, error) {
	pder.lock.RLock()
	defer pder.lock.RUnlock()

	_, ok := pder.sessions[sid]
	return ok, nil
}

//will update the last-access-time
func (pder *MemSessionProvider) GetSession(sid string) (Session, error) {
	pder.lock.RLock()
	defer pder.lock.RUnlock()

	if element, ok := pder.sessions[sid]; ok {
		sw := element.Value.(Session)
		if sw == nil {
			return nil, nil
		}
		attributes := sw.Attributes().(*MemSessionAttributes)
		if attributes != nil {
			attributes.timeAccessed = time.Now()
		}
		return sw, nil
	} else {
		return nil, nil
	}
}

func (pder *MemSessionProvider) AddNewSession(sw Session) error {
	pder.lock.Lock()
	defer pder.lock.Unlock()
	sid := sw.SessionID()
	if sid == "" {
		return errors.New("can not create session with empty sid")
	}

	if _, ok := pder.sessions[sid]; ok {
		return errors.New("session with sid=" + sid + " exist")
	}
	attributes := NewMemSessionAttributes()
	sw.SetAttributes(attributes)
	element := pder.list.PushBack(sw)
	pder.sessions[sid] = element
	return nil
}

func (pder *MemSessionProvider) RemoveSession(sid string) error {
	pder.lock.Lock()
	defer pder.lock.Unlock()
	if element, ok := pder.sessions[sid]; ok {
		delete(pder.sessions, sid)
		pder.list.Remove(element)
		return nil
	}
	return nil
}

func (pder *MemSessionProvider) RemoveExpired() {
	pder.lock.RLock()
	for {
		element := pder.list.Back()
		if element == nil {
			break
		}
		sxn := element.Value.(Session)
		if (sxn.Attributes().(*MemSessionAttributes).timeAccessed.Unix() + pder.TimeoutSeconds()) < time.Now().Unix() {
			pder.lock.RUnlock()
			pder.lock.Lock()
			pder.list.Remove(element)
			delete(pder.sessions, sxn.SessionID())
			pder.lock.Unlock()
			pder.lock.RLock()
		} else {
			break
		}
	}
	pder.lock.RUnlock()
}

func (pder *MemSessionProvider) SessionAccess(sid string) error {
	pder.lock.Lock()
	defer pder.lock.Unlock()
	if element, ok := pder.sessions[sid]; ok {
		sxn := element.Value.(Session)
		sxn.Attributes().(*MemSessionAttributes).timeAccessed = time.Now()
		pder.list.MoveToFront(element)
		return nil
	}
	return nil
}

func init() {
	Register("memory", mempder)
}
