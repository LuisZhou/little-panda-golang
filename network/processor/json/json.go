package json

import (
	"encoding/json"
	_ "errors"
	"fmt"
	"github.com/LuisZhou/lpge/network"
	"reflect"
)

type JsonProcessor struct {
	msgMap map[uint16]reflect.Type
}

func NewJsonProcessor() network.Processor {
	p := new(JsonProcessor)
	p.msgMap = make(map[uint16]reflect.Type)
	return p
}

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

func (p *JsonProcessor) Unmarshal(cmd uint16, data []byte) (interface{}, error) {
	if _, ok := p.msgMap[cmd]; !ok {
		return nil, fmt.Errorf("message %d can not handle", cmd)
	}

	i := p.msgMap[cmd]

	msg := reflect.New(i).Interface()

	return msg, json.Unmarshal(data, msg)
}

func (p *JsonProcessor) Marshal(cmd uint16, msg interface{}) ([]byte, error) {
	data, err := json.Marshal(msg)
	return data, err
}
