package module

import (
	"github.com/LuisZhou/lpge/chanrpc"
	"github.com/LuisZhou/lpge/console"
	"github.com/LuisZhou/lpge/go"
	"github.com/LuisZhou/lpge/timer"
	"time"
)

type Skeleton struct {
	GoLen              int
	TimerDispatcherLen int
	AsynCallLen        int
	ChanRPCLen         int
	TimeoutAsynRet     int
	g                  *g.Go
	dispatcher         *timer.Dispatcher
	client             *chanrpc.Client
	server             *chanrpc.Server
	commandServer      *chanrpc.Server
}

func (s *Skeleton) Init() {
	if s.GoLen <= 0 {
		s.GoLen = 0
	}
	if s.TimerDispatcherLen <= 0 {
		s.TimerDispatcherLen = 0
	}
	if s.ChanRPCLen <= 0 {
		s.ChanRPCLen = 1
	}
	if s.AsynCallLen <= 0 {
		s.AsynCallLen = 1
	}
	if s.TimeoutAsynRet <= 0 {
		s.TimeoutAsynRet = 10
	}
	s.g = g.New(s.GoLen)
	s.dispatcher = timer.NewDispatcher(s.TimerDispatcherLen)
	s.client = chanrpc.NewClient(s.AsynCallLen, 10)
	s.server = chanrpc.NewServer(s.ChanRPCLen, time.Duration(s.TimeoutAsynRet))
	s.commandServer = chanrpc.NewServer(s.ChanRPCLen, time.Duration(s.TimeoutAsynRet))
}

func (s *Skeleton) Run(closeSig chan bool) {
	for {
		select {
		case <-closeSig:
			s.commandServer.Close()
			s.server.Close()
			s.g.Close()
			s.client.Close()
			return
		case ri := <-s.client.ChanAsynRet:
			s.client.Cb(ri)
		case ci := <-s.server.ChanCall:
			s.server.Exec(ci)
		case ci := <-s.commandServer.ChanCall:
			s.commandServer.Exec(ci)
		case cb := <-s.g.ChanCb:
			s.g.Cb(cb)
		case t := <-s.dispatcher.ChanTimer:
			t.Cb()
		}
	}
}

func (s *Skeleton) AfterFunc(d time.Duration, cb func()) *timer.Timer {
	if s.TimerDispatcherLen == 0 {
		panic("invalid TimerDispatcherLen")
	}

	return s.dispatcher.AfterFunc(d, cb)
}

func (s *Skeleton) CronFunc(cronExpr *timer.CronExpr, cb func()) *timer.Cron {
	if s.TimerDispatcherLen == 0 {
		panic("invalid TimerDispatcherLen")
	}

	return s.dispatcher.CronFunc(cronExpr, cb)
}

func (s *Skeleton) Go(f func(), cb func()) {
	if s.GoLen == 0 {
		panic("invalid GoLen")
	}

	s.g.Go(f, cb)
}

func (s *Skeleton) NewLinearContext() *g.LinearContext {
	if s.GoLen == 0 {
		panic("invalid GoLen")
	}

	return s.g.NewLinearContext()
}

func (s *Skeleton) AsynCall(server *chanrpc.Server, id interface{}, args ...interface{}) error {
	if s.AsynCallLen == 0 {
		panic("invalid AsynCallLen")
	}

	return s.client.AsynCall(server, id, args...)
}

func (s *Skeleton) SynCall(server *chanrpc.Server, id interface{}, args ...interface{}) (interface{}, error) {
	return s.client.SynCall(server, id, args...)
}

func (s *Skeleton) RegisterChanRPC(id interface{}, f interface{}) {
	s.server.Register(id, f.(func([]interface{}) (interface{}, error)))
}

func (s *Skeleton) RegisterCommand(name string, help string, f interface{}) {
	console.Register(name, help, f, s.commandServer)
}

func (s *Skeleton) GoRpc(id interface{}, args ...interface{}) {
	s.AsynCall(s.server, id, args...)
}

func (s *Skeleton) GetChanrpcServer() *chanrpc.Server {
	return s.server
}
