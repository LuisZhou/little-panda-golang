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

// Agent for client of ws or tcp.
type Agent interface {
	Run()                                      // run start running, normaly process the msg.
	WriteMsg(cmd uint16, msg interface{})      // write msg to message.
	LocalAddr() net.Addr                       // get local addr
	RemoteAddr() net.Addr                      // get remote addr
	Close()                                    // close agent.
	UserData() interface{}                     // get user data of the agent.
	SetUserData(data interface{})              // set user data of the agent.
	GoRpc(id interface{}, args ...interface{}) // GoRpc do a async call to this Skeleton's rpc server, but no async ret.
}

// Implement of Agent.
type AgentTemplate struct {
	*module.Skeleton                   // is a Skeleton.
	conn             network.Conn      // conn of this agent.
	gate             *Gate             // gate of this server.
	userData         interface{}       // user data.
	Processor        network.Processor // processor of msg.
	closeChan        chan bool         // close sig for internal module skeleton.
}

// Init do init agent.
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

// Run start process msg from agent. The server will go this func when new agent create.
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

// OnClose is called when the connection is destoried.
func (a *AgentTemplate) OnClose() {
	_, err := chanrpc.SynCall(a.gate.Skeleton.GetChanrpcServer(), "CloseAgent", a)
	if err != nil {
		log.Error("chanrpc error: %v", err)
	}
	a.closeChan <- true
}

// Write msg to the connection.
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

// LocalAddr return local addr.
func (a *AgentTemplate) LocalAddr() net.Addr {
	return a.conn.LocalAddr()
}

// RemoteAddr return remote addr.
func (a *AgentTemplate) RemoteAddr() net.Addr {
	return a.conn.RemoteAddr()
}

// Close close the connection.
func (a *AgentTemplate) Close() {
	a.conn.Close()
}

// UserData get user data.
func (a *AgentTemplate) UserData() interface{} {
	return a.userData
}

// SetUserData set user data.
func (a *AgentTemplate) SetUserData(data interface{}) {
	a.userData = data
}
