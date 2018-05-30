// Package protobuf is a processor that Marshal/Unmarshal from/to proto-type data to/from stream of byte.
package protobuf

import (
	"fmt"
	"github.com/LuisZhou/lpge/network"
	"github.com/golang/protobuf/proto"
	"reflect"
)

type Protobuf struct {
	msgMap map[uint16]reflect.Type
}

// NewProtobufProcessor create a new protobuf processor.
func NewProtobufProcessor() network.Processor {
	p := new(Protobuf)
	p.msgMap = make(map[uint16]reflect.Type)
	return p
}

// Register implements the Register of interface Processor.
func (p *Protobuf) Register(cmd uint16, msg interface{}) error {
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
func (p *Protobuf) Unmarshal(cmd uint16, data []byte) (interface{}, error) {
	if _, ok := p.msgMap[cmd]; !ok {
		return nil, fmt.Errorf("message %d can not handle", cmd)
	}

	i := p.msgMap[cmd]

	msg := reflect.New(i).Interface()

	return msg, proto.Unmarshal(data, msg.(proto.Message))
}

// Marshal implements the Marshal of interface Processor.
func (p *Protobuf) Marshal(cmd uint16, msg interface{}) ([]byte, error) {
	data, err := proto.Marshal(msg.(proto.Message))
	return data, err
}
