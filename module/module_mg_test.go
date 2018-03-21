package module

import (
	"github.com/LuisZhou/lpge/chanrpc"
	"github.com/LuisZhou/lpge/module"
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
	s := &module.Skeleton{
		GoLen:              10,
		TimerDispatcherLen: 10,
		AsynCallLen:        10,
		ChanRPCServer:      chanrpc.NewServer(10, 0),
	}
	s.Init()

	m := &TestModule{}
	m.Skeleton = s

	module.Register(m, "test")

	f, e := module.Search("test")
	t.Log(f, e)
}
