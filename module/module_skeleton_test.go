package module

import (
	"github.com/LuisZhou/lpge/chanrpc"
	"github.com/LuisZhou/lpge/module"
	"testing"
)

func TestSkeleton(t *testing.T) {
	s := &module.Skeleton{
		GoLen:              10,
		TimerDispatcherLen: 10,
		AsynCallLen:        10,
		ChanRPCServer:      chanrpc.NewServer(10, 0),
	}
	s.Init()
	s.Run(make(chan bool))
}
