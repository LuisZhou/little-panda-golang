package gate

import (
	"github.com/LuisZhou/lpge/chanrpc"
	"github.com/LuisZhou/lpge/log"
	"github.com/LuisZhou/lpge/network"
	"net"
	"reflect"
	"time"
)

type NewAgent func(conn network.Conn) network.Agent

// Gate start ws server and tcp server base on its configure. And bind EventListener and NewAgent to it.
// NewAgent will exe when there is a new client here.
type Gate struct {
	MaxConnNum      int
	PendingWriteNum int
	MaxMsgLen       uint16
	//Processor       network.Processor
	EventListener *chanrpc.Server // do like a event listener

	// websocket
	WSAddr      string
	HTTPTimeout time.Duration
	CertFile    string
	KeyFile     string

	// tcp
	TCPAddr      string
	LenMsgLen    int
	LittleEndian bool
}

func (gate *Gate) Run(closeSig chan bool, newWsAgent NewAgent, newTcpAgent NewAgent) {
	// var wsServer *network.WSServer
	// if gate.WSAddr != "" {
	// 	wsServer = new(network.WSServer)
	// 	wsServer.Addr = gate.WSAddr
	// 	wsServer.MaxConnNum = gate.MaxConnNum
	// 	wsServer.PendingWriteNum = gate.PendingWriteNum
	// 	wsServer.MaxMsgLen = gate.MaxMsgLen
	// 	wsServer.HTTPTimeout = gate.HTTPTimeout
	// 	wsServer.CertFile = gate.CertFile
	// 	wsServer.KeyFile = gate.KeyFile
	// 	//wsServer.NewAgent = newWsAgent
	// 	wsServer.NewAgent = func(conn *network.WSConn) network.Agent {
	// 		a := newWsAgent(conn.(*network.Conn)) //&agent{conn: conn, gate: gate}
	// 		if gate.EventListener != nil {
	// 			gate.EventListener.Go("NewAgent", a)
	// 		}
	// 		return a
	// 	}
	// }

	var tcpServer *network.TCPServer
	if gate.TCPAddr != "" {
		tcpServer = new(network.TCPServer)
		tcpServer.Addr = gate.TCPAddr
		tcpServer.MaxConnNum = gate.MaxConnNum
		tcpServer.PendingWriteNum = gate.PendingWriteNum
		//tcpServer.LenMsgLen = gate.LenMsgLen
		tcpServer.MaxMsgLen = gate.MaxMsgLen
		tcpServer.LittleEndian = gate.LittleEndian
		//tcpServer.NewAgent = newTcpAgent
		tcpServer.NewAgent = func(conn *network.TCPConn) network.Agent {
			a := newTcpAgent(conn) //&agent{conn: conn, gate: gate}
			if gate.EventListener != nil {
				gate.EventListener.Go("NewAgent", a)
			}
			return a
		}
	}

	// if wsServer != nil {
	// 	wsServer.Start()
	// }
	if tcpServer != nil {
		tcpServer.Start()
	}
	<-closeSig
	// if wsServer != nil {
	// 	wsServer.Close()
	// }
	if tcpServer != nil {
		tcpServer.Close()
	}
}

func (gate *Gate) OnDestroy() {}

// todo: the client should also be a module.
type AgentTemplate struct {
	conn      network.Conn
	gate      *Gate
	userData  interface{}
	Processor network.Processor
	handlers  map[uint16]func(uint16, interface{})
}

func (a *AgentTemplate) Init(conn network.Conn, gate *Gate) {
	a.conn = conn
	a.gate = gate
	a.handlers = make(map[uint16]func(uint16, interface{}))
}

func (a *AgentTemplate) Run() {
	for {
		cmd, data, err := a.conn.ReadMsg()
		if err != nil {
			log.Debug("read message: %v", err)
			break
		}

		// why I decide not route the msg to module at once, but route to agent.
		// 1. Route to agent is more natural. Msg is first route agent, and the agent decide what to do next.
		// 2. One client map to one agent.
		if a.Processor != nil {
			msg, err := a.Processor.Unmarshal(cmd, data)
			if err != nil {
				log.Debug("unmarshal message error: %v", err)
				break
			}
			err = a.Handler(cmd, msg)
			if err != nil {
				log.Debug("route message error: %v", err)
				break
			}
		}
	}
}

func (a *AgentTemplate) OnClose() {
	if a.gate.EventListener != nil {
		_, err := a.gate.EventListener.Call("CloseAgent", a)
		if err != nil {
			log.Error("chanrpc error: %v", err)
		}
	}
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

func (a *AgentTemplate) Destroy() {
	a.conn.Destroy()
}

func (a *AgentTemplate) UserData() interface{} {
	return a.userData
}

func (a *AgentTemplate) SetUserData(data interface{}) {
	a.userData = data
}

func (a *AgentTemplate) Handler(cmd uint16, msg interface{}) (err error) {
	if f, ok := a.handlers[cmd]; ok {
		f(cmd, msg)
	} else {
		log.Debug("Can't handle: %d %v", cmd, msg)
	}

	return
}

func (a *AgentTemplate) Register(cmd uint16, f func(uint16, interface{})) (err error) {
	a.handlers[cmd] = f
	return
}
