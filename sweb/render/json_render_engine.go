package render

import (
	"bytes"
	"container/list"
	"encoding/json"
	"strconv"

	"github.com/fwis/goweb/sweb/pagination"
)

/*
type JsonMeta struct {
	M struct {
		OK bool
		Status  int
		Msg     string
	}
	D interface{}
	P struct {
		Ps int //page size
		Pn int //page number
		Pt int64 //total page count, can be 0 means unknow
		Draw int64
		Offset int64
	}
}
*/

type JSONRenderEngine struct {
	Indent       bool
	UnEscapeHTML bool
}

func NewDefaultJSONRenderEngine() *JSONRenderEngine {
	return &JSONRenderEngine{Indent: false, UnEscapeHTML: true}
}

var (
	ok_json_content = []byte(`{"M":{"OK":true,"Status":0,"Msg":""}}`)
)

func (m *JSONRenderEngine) OK() []byte {
	return ok_json_content
}

func (m *JSONRenderEngine) Error(status int, msg string) []byte {
	buf := bytes.NewBuffer([]byte{})
	buf.WriteString(`{"M":{"OK":false`)
	buf.WriteString(`,"Status":`)
	buf.WriteString(strconv.Itoa(status))
	buf.WriteString(`,"Msg":"`)
	buf.WriteString(msg)
	buf.WriteString(`"}}`)
	return buf.Bytes()
}

/*
func (m *JSONRenderEngine) DataJson(datajson []byte) []byte {
	buf := bytes.NewBuffer([]byte{})
	buf.WriteString(`{"M":{"OK":true,"Status":0,"Msg":""},"D":`)
	buf.Write(datajson)
	buf.WriteString("}")
	return buf.Bytes()
}
*/

func (m *JSONRenderEngine) DataJson(p *pagination.Pagination, datajson []byte) []byte {
	buf := bytes.NewBuffer([]byte{})
	buf.WriteString(`{"M":{"OK":true,"Status":0,"Msg":""}`)
	if p != nil {
		buf.WriteString(`,"P":`)
		p.WriteToBufAsJson(buf)
	}
	buf.WriteString(`,"D":`)
	buf.Write(datajson)
	buf.WriteString("}")
	return buf.Bytes()
}

/*
func (m *JSONRenderEngine) DataObj(data interface{}) ([]byte, error) {
	var result []byte
	var err error

	if m.Indent {
		result, err = json.MarshalIndent(data, "", "  ")
		result = append(result, '\n')
	} else {
		result, err = json.Marshal(data)
	}
	if err != nil {
		return nil, err
	}

	// Unescape HTML if needed.
	if m.UnEscapeHTML {
		result = bytes.Replace(result, []byte("\\u003c"), []byte("<"), -1)
		result = bytes.Replace(result, []byte("\\u003e"), []byte(">"), -1)
		result = bytes.Replace(result, []byte("\\u0026"), []byte("&"), -1)
	}
	return m.DataJson(result), nil
}
*/

func (m *JSONRenderEngine) DataObjs(p *pagination.Pagination, objs interface{}) ([]byte, error) {
	var result []byte
	var err error

	if m.Indent {
		result, err = json.MarshalIndent(objs, "", "  ")
		result = append(result, '\n')
	} else {
		result, err = json.Marshal(objs)
	}
	if err != nil {
		return nil, err
	}

	// Unescape HTML if needed.
	if m.UnEscapeHTML {
		result = bytes.Replace(result, []byte("\\u003c"), []byte("<"), -1)
		result = bytes.Replace(result, []byte("\\u003e"), []byte(">"), -1)
		result = bytes.Replace(result, []byte("\\u0026"), []byte("&"), -1)
	}
	return m.DataJson(p, result), nil
}

func (m *JSONRenderEngine) DataList(p *pagination.Pagination, list *list.List) ([]byte, error) {
	buf := bytes.Buffer{}
	//var err error
	buf.WriteByte('[')
	i := 0
	for element := list.Front(); element != nil; element = element.Next() {
		if i > 0 {
			buf.WriteByte(',')
		}
		if m.Indent {
			if element_json, err := json.MarshalIndent(element.Value, "", "  "); err == nil {
				buf.Write(element_json)
			} else {
				return nil, err
			}
		} else {
			if element_json, err := json.Marshal(element.Value); err == nil {
				buf.Write(element_json)
			} else {
				return nil, err
			}
		}

		i++
	}
	buf.WriteByte(']')

	result := buf.Bytes()

	// Unescape HTML if needed.
	if m.UnEscapeHTML {
		result = bytes.Replace(result, []byte("\\u003c"), []byte("<"), -1)
		result = bytes.Replace(result, []byte("\\u003e"), []byte(">"), -1)
		result = bytes.Replace(result, []byte("\\u0026"), []byte("&"), -1)
	}
	return m.DataJson(p, result), nil
}
