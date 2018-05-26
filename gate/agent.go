package gate

import (
	"github.com/LuisZhou/lpge/chanrpc"
	"github.com/LuisZhou/lpge/conf"
	"github.com/LuisZhou/lpge/log"
	"github.com/LuisZhou/lpge/module"
	"github.com/LuisZhou/lpge/network"
	"net"
	"reflect"
)

type Agent interface {
	WriteMsg(cmd uint16, msg interface{})
	LocalAddr() net.Addr
	RemoteAddr() net.Addr
	Close()
	UserData() interface{}
	SetUserData(data interface{})
	GoRpc(id interface{}, args ...interface{})
}

type AgentTemplate struct {
	conn      network.Conn
	gate      *Gate
	userData  interface{}
	Processor network.Processor
	closeChan chan bool
	*module.Skeleton
}

func (a *AgentTemplate) Init(conn network.Conn, gate *Gate) {
	a.conn = conn
	a.gate = gate
	a.closeChan = make(chan bool, 1)
	s := &module.Skeleton{
		GoLen:              conf.AgentConfig.GoLen,
		TimerDispatcherLen: conf.AgentConfig.TimerDispatcherLen,
		AsynCallLen:        conf.AgentConfig.AsynCallLen,
		ChanRPCLen:         conf.AgentConfig.ChanRPCLen,
		TimeoutAsynRet:     conf.AgentConfig.TimeoutAsynRet,
	}
	s.Init()
	go s.Run(a.closeChan)
	a.Skeleton = s
}

func (a *AgentTemplate) Run() {
	for {
		cmd, data, err := a.conn.ReadMsg()
		if err != nil {
			log.Debug("read message: %v", err)
			break
		}

		if a.Processor != nil {
			msg, err := a.Processor.Unmarshal(cmd, data)
			if err != nil {
				log.Debug("unmarshal message error: %v", err)
				break
			}
			a.GoRpc(cmd, msg)
		}
	}
}

func (a *AgentTemplate) OnClose() {
	_, err := chanrpc.SynCall(a.gate.Skeleton.GetChanrpcServer(), "CloseAgent", a)
	if err != nil {
		log.Error("chanrpc error: %v", err)
	}

	a.closeChan <- true
}

func (a *AgentTemplate) WriteMsg(cmd uint16, msg interface{}) {
	if a.Processor != nil {
		data, err := a.Processor.Marshal(cmd, msg)
		if err != nil {
			log.Error("marshal message %v error: %v", reflect.TypeOf(msg), err)
			return
		}
		err = a.conn.WriteMsg(cmd, data)
		if err != nil {
			log.Error("write message %v error: %v", reflect.TypeOf(msg), err)
		}
	}
}

func (a *AgentTemplate) LocalAddr() net.Addr {
	return a.conn.LocalAddr()
}

func (a *AgentTemplate) RemoteAddr() net.Addr {
	return a.conn.RemoteAddr()
}

func (a *AgentTemplate) Close() {
	a.conn.Close()
}

func (a *AgentTemplate) UserData() interface{} {
	return a.userData
}

func (a *AgentTemplate) SetUserData(data interface{}) {
	a.userData = data
}
