package session

import (
	"bytes"
	"encoding/gob"
)

func init() {
	gob.Register(map[string]interface{}{})
}

// EncodeGob encode the obj to gob
func EncodeGob(obj map[string]interface{}) ([]byte, error) {
	for _, v := range obj {
		gob.Register(v)
	}
	buf := bytes.NewBuffer(nil)
	enc := gob.NewEncoder(buf)
	err := enc.Encode(obj)
	if err != nil {
		return []byte(""), err
	}
	return buf.Bytes(), nil
}

// DecodeGob decode data to map
func DecodeGob(encoded []byte) (map[string]interface{}, error) {
	buf := bytes.NewBuffer(encoded)
	dec := gob.NewDecoder(buf)
	var out map[string]interface{}
	err := dec.Decode(&out)
	if err != nil {
		return nil, err
	}
	return out, nil
}
