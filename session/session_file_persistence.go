package session

import (
	"errors"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sync"
)

type SessionFilePersistence struct {
	lock     sync.RWMutex
	savePath string
}

func NewSessionFilePersistence(savePath string) *SessionFilePersistence {
	fp := &SessionFilePersistence{}
	fp.savePath = savePath
	os.MkdirAll(savePath, 0777)
	return fp
}

func (fp *SessionFilePersistence) sidFilePath(sid string) string {
	return path.Join(fp.savePath, sid)
}

func ParseSidFromFilePath(sidFilePath string) string {
	_, fileName := filepath.Split(sidFilePath)
	ext := filepath.Ext(fileName)
	if ext == "" {
		return fileName
	} else if ext == ".s" && len(fileName) > 4 {
		return fileName[0 : len(fileName)-2]
	} else {
		return ""
	}
}

func LoadSessionAttributesFromFile(fp *SessionFilePersistence, sidFilePath string) (string, SessionAttributes, error) {
	fileInfo, err := os.Stat(sidFilePath)

	if err != nil {
		return "", nil, err
	} else if fileInfo.IsDir() {
		return "", nil, errors.New("Can't use dir as session store file")
	} else {
		sid := ParseSidFromFilePath(sidFilePath)
		if sid == "" {
			return "", nil, errors.New("Can't parse sid from store file name")
		}
		b, err := ioutil.ReadFile(sidFilePath)
		if err != nil {
			return "", nil, err
		}

		ss := &MemSessionAttributes{fp: fp, sid: sid, timeAccessed: fileInfo.ModTime()}

		if len(b) > 0 {
			err = ss.Decode(b)

			if err != nil {
				return "", nil, err
			}
		} else {
			ss.Clear()
		}

		return sid, ss, nil
	}
}

func (fp *SessionFilePersistence) Has(sid string) bool {
	_, err := os.Stat(path.Join(fp.savePath, sid))
	return err == nil
}

func (fp *SessionFilePersistence) Remove(sid string) {
	os.Remove(fp.sidFilePath(sid))
}

func (fp *SessionFilePersistence) Clear(sid string) {
	os.Truncate(fp.sidFilePath(sid), 0)
}

func (fp *SessionFilePersistence) Save(attr SessionAttributes) error {
	if attr == nil {
		return nil
	}
	encoded, err := attr.Encode()
	if err != nil {
		return err
	}

	fp.lock.Lock()
	defer fp.lock.Unlock()

	sidFilePath := fp.sidFilePath(attr.SessionID())

	os.MkdirAll(fp.savePath, 0777)

	err = ioutil.WriteFile(sidFilePath, encoded, 0777)

	if err != nil {
		return err
	}

	err = os.Chtimes(sidFilePath, attr.TimeAccessed(), attr.TimeAccessed())

	return err
}
