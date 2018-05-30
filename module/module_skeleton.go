package module

import (
	"github.com/LuisZhou/lpge/chanrpc"
	"github.com/LuisZhou/lpge/console"
	"github.com/LuisZhou/lpge/go"
	"github.com/LuisZhou/lpge/timer"
	"time"
)

// Skeleton implements main function of module.
type Skeleton struct {
	GoLen              int               // len of call-channel of Go.
	TimerDispatcherLen int               // len of call-channel of Timer.
	AsynCallLen        int               // len of channel of return of async call of rpc.
	ChanRPCLen         int               // len of channel of called of rpc.
	TimeoutAsynRet     int               // timeout of wait for return of rpc.
	g                  *g.Go             // Go module.
	dispatcher         *timer.Dispatcher // Timer module.
	client             *chanrpc.Client   // Client of rpc.
	server             *chanrpc.Server   // Server of rpc.
	commandServer      *chanrpc.Server   // Command module.
}

// Init do init of internal module.
func (s *Skeleton) Init() {
	if s.GoLen <= 0 {
		s.GoLen = 1
	}
	if s.TimerDispatcherLen <= 0 {
		s.TimerDispatcherLen = 1
	}
	if s.ChanRPCLen <= 0 {
		s.ChanRPCLen = 1
	}
	if s.AsynCallLen <= 0 {
		s.AsynCallLen = 1
	}
	if s.TimeoutAsynRet < 0 {
		s.TimeoutAsynRet = 0
	}
	s.g = g.New(s.GoLen)
	s.dispatcher = timer.NewDispatcher(s.TimerDispatcherLen)
	s.client = chanrpc.NewClient(s.AsynCallLen, 10)
	s.server = chanrpc.NewServer(s.ChanRPCLen, time.Duration(s.TimeoutAsynRet))
	s.commandServer = chanrpc.NewServer(s.ChanRPCLen, time.Duration(s.TimeoutAsynRet))
}

// Run start running all module.
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

// AfterFunc call cb after d duration.
func (s *Skeleton) AfterFunc(d time.Duration, cb func()) *timer.Timer {
	return s.dispatcher.AfterFunc(d, cb)
}

// CronFunc do a cron job on time dispatcher.
func (s *Skeleton) CronFunc(cronExpr *timer.CronExpr, cb func()) *timer.Cron {
	return s.dispatcher.CronFunc(cronExpr, cb)
}

// Go do call to go module.
func (s *Skeleton) Go(f func(), cb func()) {
	s.g.Go(f, cb)
}

// NewLinearContext create new linear context.
func (s *Skeleton) NewLinearContext() *g.LinearContext {
	return s.g.NewLinearContext()
}

// RegisterCommand register command.
func (s *Skeleton) RegisterCommand(name string, help string, f func([]interface{}) (interface{}, error)) {
	console.Register(name, help, f, s.commandServer)
}

// AsynCall do a async call to rpc server.
func (s *Skeleton) AsynCall(server *chanrpc.Server, id interface{}, args ...interface{}) error {
	return s.client.AsynCall(server, id, args...)
}

// SynCall do a syn call to rpc server.
func (s *Skeleton) SynCall(server *chanrpc.Server, id interface{}, args ...interface{}) (interface{}, error) {
	return s.client.SynCall(server, id, args...)
}

// GoRpc do a async call to this Skeleton's rpc server, but no async return.
func (s *Skeleton) GoRpc(id interface{}, args ...interface{}) {
	s.AsynCall(s.server, id, args...)
}

// GetChanrpcServer return this Skeleton's rpc server
func (s *Skeleton) GetChanrpcServer() *chanrpc.Server {
	return s.server
}

// RegisterChanRPC register handler for id.
func (s *Skeleton) RegisterChanRPC(id interface{}, f func([]interface{}) (interface{}, error)) {
	s.server.Register(id, f)
}

// SetChanRPCHandle set rpc handlers.
func (s *Skeleton) SetChanRPCHandlers(m map[interface{}]interface{}) {
	s.server.SetHandlers(m)
}
