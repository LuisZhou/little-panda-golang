package json

import (
	"github.com/LuisZhou/lpge/network/processor/json"
	_ "reflect"
	"testing"
)

type Person struct {
	Name string
}

func TestProtocol(t *testing.T) {
	person := &Person{}
	person.Name = "abc"
	t.Log(person)

	p := json.NewJsonProcessor()
	_ = p

	p.Register(1, Person{})

	buf, err1 := p.Marshal(1, person)
	t.Log(buf, err1)
	t.Log(string(buf))

	ret, err2 := p.Unmarshal(1, buf)
	t.Log(ret, err2)
}
