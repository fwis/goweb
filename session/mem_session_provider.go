package session

import (
	"container/list"
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sync"
	"time"
)

type MemSessionProvider struct {
	lock           sync.RWMutex             //用来锁
	sessions       map[string]*list.Element //用来存储在内存
	list           *list.List               //用来做gc
	timeoutseconds int64
	fp             *SessionFilePersistence
}

func NewMemSessionProvider(timeoutseconds int64, savepath string) *MemSessionProvider {
	provider := &MemSessionProvider{}
	provider.list = &list.List{}
	provider.timeoutseconds = timeoutseconds
	provider.sessions = make(map[string]*list.Element)
	provider.fp = NewSessionFilePersistence(savepath)
	return provider
}

func (pder *MemSessionProvider) TimeoutSeconds() int64 {
	return pder.timeoutseconds
}

func (pder *MemSessionProvider) SessionInit() error {
	return nil
}

func (pder *MemSessionProvider) LoadSessions(f func(sid string, attr SessionAttributes) Session) error {
	if pder.fp == nil {
		return nil
	}
	fileInfos, err := ioutil.ReadDir(pder.fp.savePath)
	if err != nil {
		return err
	}
	for _, fileInfo := range fileInfos {
		if fileInfo.IsDir() {
			continue
		}
		sidFilePath := filepath.Join(pder.fp.savePath, fileInfo.Name())

		sid, attributes, err := LoadSessionAttributesFromFile(pder.fp, sidFilePath)
		if err != nil {
			fmt.Printf("ignore! fail to load session from session file=%v\n", sidFilePath)
			//return err
			continue
		}

		session := f(sid, attributes)
		if !pder.IsExpired(session) {
			pder.AddNewSession(session)
		} else {
			if pder.fp != nil {
				pder.fp.Remove(sid)
			}
		}
	}
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
			//fmt.Printf("because pder.sesssion[sid].value is not Session??!!, sid=%v\n", sid)
			return nil, nil
		}

		if sw.Attributes() != nil {
			sw.Attributes().SetTimeAccessed(time.Now())
			pder.list.MoveToFront(element)
			//fmt.Printf("SetTimeAccessed, sid=%v\n", sid)
		}
		return sw, nil
	} else {
		//fmt.Printf("not find session in provider, sid=%v\n", sid)
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

	//attributes := NewMemSessionAttributes()
	//sw.SetAttributes(attributes)
	element := pder.list.PushBack(sw)
	//fmt.Printf("list pushback\n")
	pder.sessions[sid] = element
	return nil
}

func (pder *MemSessionProvider) RemoveSession(sid string) error {
	if sid == "" {
		return nil
	}
	pder.lock.Lock()
	defer pder.lock.Unlock()
	if element, ok := pder.sessions[sid]; ok {
		//fmt.Printf("RemoveSession, sid=%v\n", sid)
		if pder.sessions != nil {
			delete(pder.sessions, sid)
		}
		if pder.list != nil {
			pder.list.Remove(element)
		}
		if pder.fp != nil {
			pder.fp.Remove(sid)
		}
		return nil
	}
	return nil
}

func (pder *MemSessionProvider) IsExpired(session Session) bool {
	//fmt.Printf("timeaccessed=%v, timeout=%v, now=%v\n", session.Attributes().TimeAccessed(), pder.TimeoutSeconds(), time.Now())
	return (session.Attributes().TimeAccessed().Unix() + pder.TimeoutSeconds()) < time.Now().Unix()
}

func (pder *MemSessionProvider) RemoveExpired() {
	pder.lock.RLock()

	for element := pder.list.Back(); element != nil; element = element.Prev() {
		sxn := element.Value.(Session)
		if sxn == nil {
			continue
		}
		if pder.IsExpired(sxn) {
			//fmt.Printf("RemoveExpired, sid=%v\n", sxn.SessionID())
			pder.lock.RUnlock()
			pder.lock.Lock()
			if pder.list != nil {
				pder.list.Remove(element)
			}
			if pder.sessions != nil {
				delete(pder.sessions, sxn.SessionID())
			}
			if pder.fp != nil {
				pder.fp.Remove(sxn.SessionID())
			}
			pder.lock.Unlock()
			pder.lock.RLock()
		} else {
			break
		}
	}

	pder.lock.RUnlock()
}

/*
func (pder *MemSessionProvider) SessionAccess(sid string) error {
	pder.lock.Lock()
	defer pder.lock.Unlock()
	if element, ok := pder.sessions[sid]; ok {
		sxn := element.Value.(Session)
		sxn.Attributes().SetTimeAccessed(time.Now())
		pder.list.MoveToFront(element)
		return nil
	}
	return nil
}
*/

func (pder *MemSessionProvider) NewSessionAttributes(sid string) SessionAttributes {
	return NewMemSessionAttributes(sid, pder.fp)
}

func (pder *MemSessionProvider) PersistSessions() {
	if pder.fp == nil {
		return
	}
	tmpList := list.New()
	tmpList.PushBackList(pder.list)

	for e := tmpList.Front(); e != nil; e = e.Next() {
		sxn := e.Value.(Session)
		if sxn != nil {
			err := pder.fp.Save(sxn.Attributes())
			if err != nil {
				fmt.Printf("ignore! fail to save session, sid=%v, err=%v\n", sxn.SessionID(), err)
			}
		}
	}
}
