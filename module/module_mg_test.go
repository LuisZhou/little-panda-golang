package module

import (
	"github.com/LuisZhou/lpge/module"
	"strconv"
	"sync"
	"testing"
)

type TestModule struct {
	*module.Skeleton
}

func (m *TestModule) OnInit() {

}

func (m *TestModule) OnDestroy() {
}

func TestModuleMg(t *testing.T) {
	for i := 0; i < 2; i++ {
		s := &module.Skeleton{
			GoLen:              10,
			TimerDispatcherLen: 10,
			AsynCallLen:        10,
		}
		s.Init()

		m := &TestModule{}
		m.Skeleton = s

		module.Register(m, "test"+strconv.Itoa(i))
	}

	m1, _, e1 := module.Search("test0")
	t.Log(m1, e1)

	var wg sync.WaitGroup

	m2, _, e2 := module.Search("test1")
	t.Log(m2, e2)
	m2.RegisterChanRPC(uint16(1), func(args []interface{}) (interface{}, error) {
		t.Log("what?", args)
		wg.Done()
		return nil, nil
	})

	wg.Add(1)
	e3 := m1.AsynCall("test1", uint16(1), "test")
	t.Log(e3)
	wg.Wait()

	wg.Add(1)
	m1.SynCall("test1", uint16(1), "test")
}
