package protobuf

import (
	"github.com/LuisZhou/lpge/network/processor/protobuf"
	"reflect"
	"testing"
)

func TestProtocol(t *testing.T) {
	person := &protobuf.Person{}
	person.Name = "abc"
	t.Log(person)

	p := protobuf.NewProtobufProcessor()

	p.Register(1, protobuf.Person{})

	buf, err1 := p.Marshal(1, person)
	t.Log(buf, err1)

	ret, err2 := p.Unmarshal(1, buf)
	t.Log(ret, err2)

	c := ret.(*protobuf.Person)
	t.Log(c.Name, reflect.TypeOf(ret), reflect.ValueOf(ret).Elem().CanSet())
}
