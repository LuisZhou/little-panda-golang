package processor

import (
	"github.com/golang/protobuf/proto"
)

// https://stackoverflow.com/questions/7850140/how-do-you-create-a-new-instance-of-a-struct-from-its-type-at-run-time-in-go

type protobuf struct {
}

func (p *protobuf) Unmarshal(cmd uint16, data []byte) (interface{}, error) {

}

func (p *protobuf) Marshal(cmd uint16, msg interface{}) ([]byte, error) {
	// data
	data, err := proto.Marshal(msg.(proto.Message))
	return data, err
}
