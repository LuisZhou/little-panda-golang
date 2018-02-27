package processor

import (
	"github.com/LuisZhou/lpge/network/processor"
	"testing"
)

func TestProtocol(t *testing.T) {
	person := &processor.Person{}
	person.Name = "abc"
	t.Log(person)

	p := processor.NewProtobufProcessor()

	p.Register(1, processor.Person{})

	buf, err1 := p.Marshal(1, person)
	t.Log(buf, err1)

	ret, err2 := p.Unmarshal(2, buf)
	t.Log(ret, err2)
}
