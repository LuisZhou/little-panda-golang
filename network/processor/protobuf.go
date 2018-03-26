package processor

import (
	"fmt"
	"github.com/LuisZhou/lpge/network"
	"github.com/golang/protobuf/proto"
	"reflect"
)

type protobuf struct {
	msgMap map[uint16]reflect.Type
}

func NewProtobufProcessor() network.Processor {
	p := new(protobuf)
	p.msgMap = make(map[uint16]reflect.Type)
	return p
}

func (p *protobuf) Register(cmd uint16, msg interface{}) error {
	msgType := reflect.TypeOf(msg)

	if _, ok := p.msgMap[cmd]; ok {
		return fmt.Errorf("message %s is already registered", msgType)
	}

	if msgType.Kind() != reflect.Ptr {
		p.msgMap[cmd] = msgType
	} else {
		// Elem returns a type's element type.
		// It panics if the type's Kind is not Array, Chan, Map, Ptr, or Slice.
		p.msgMap[cmd] = msgType.Elem()
	}

	return nil
}

func (p *protobuf) Unmarshal(cmd uint16, data []byte) (interface{}, error) {
	if _, ok := p.msgMap[cmd]; !ok {
		return nil, fmt.Errorf("message %d can not handle", cmd)
	}

	i := p.msgMap[cmd]

	// Interface returns v's current value as an interface{}
	msg := reflect.New(i).Interface()

	return msg, proto.Unmarshal(data, msg.(proto.Message))
}

func (p *protobuf) Marshal(cmd uint16, msg interface{}) ([]byte, error) {
	data, err := proto.Marshal(msg.(proto.Message))
	return data, err
}
