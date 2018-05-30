// Package json is a processor that Marshal/Unmarshal from/to json-type data to/from stream of byte.
package json

import (
	"encoding/json"
	_ "errors"
	"fmt"
	"github.com/LuisZhou/lpge/network"
	"reflect"
)

// JsonProcessor is type of json processor.
type JsonProcessor struct {
	msgMap map[uint16]reflect.Type
}

// NewJsonProcessor create a new json processor.
func NewJsonProcessor() network.Processor {
	p := new(JsonProcessor)
	p.msgMap = make(map[uint16]reflect.Type)
	return p
}

// Register implements the Register of interface Processor.
func (p *JsonProcessor) Register(cmd uint16, msg interface{}) error {
	msgType := reflect.TypeOf(msg)

	if _, ok := p.msgMap[cmd]; ok {
		return fmt.Errorf("message %s is already registered", msgType)
	}

	if msgType.Kind() != reflect.Ptr {
		p.msgMap[cmd] = msgType
	} else {
		p.msgMap[cmd] = msgType.Elem()
	}

	return nil
}

// Unmarshal implements the Unmarshal of interface Processor.
func (p *JsonProcessor) Unmarshal(cmd uint16, data []byte) (interface{}, error) {
	if _, ok := p.msgMap[cmd]; !ok {
		return nil, fmt.Errorf("message %d can not handle", cmd)
	}

	i := p.msgMap[cmd]

	msg := reflect.New(i).Interface()

	return msg, json.Unmarshal(data, msg)
}

// Marshal implements the Marshal of interface Processor.
func (p *JsonProcessor) Marshal(cmd uint16, msg interface{}) ([]byte, error) {
	data, err := json.Marshal(msg)
	return data, err
}
