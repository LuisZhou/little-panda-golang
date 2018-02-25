package gate

import (
	"github.com/name5566/leaf/chanrpc"
	"github.com/name5566/leaf/log"
	"github.com/name5566/leaf/network"
	"net"
	"reflect"
	"time"
)

type NewAgent func(conn *network.Conn) network.Agent

type Gate struct {
	MaxConnNum      int
	PendingWriteNum int
	MaxMsgLen       uint32
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
	var wsServer *network.WSServer
	if gate.WSAddr != "" {
		wsServer = new(network.WSServer)
		wsServer.Addr = gate.WSAddr
		wsServer.MaxConnNum = gate.MaxConnNum
		wsServer.PendingWriteNum = gate.PendingWriteNum
		wsServer.MaxMsgLen = gate.MaxMsgLen
		wsServer.HTTPTimeout = gate.HTTPTimeout
		wsServer.CertFile = gate.CertFile
		wsServer.KeyFile = gate.KeyFile
		wsServer.NewAgent = newWsAgent
		// wsServer.NewAgent = func(conn *network.WSConn) network.Agent {
		// 	a := &agent{conn: conn, gate: gate}
		// 	if gate.EventListener != nil {
		// 		gate.EventListener.Go("NewAgent", a)
		// 	}
		// 	return a
		// }
	}

	var tcpServer *network.TCPServer
	if gate.TCPAddr != "" {
		tcpServer = new(network.TCPServer)
		tcpServer.Addr = gate.TCPAddr
		tcpServer.MaxConnNum = gate.MaxConnNum
		tcpServer.PendingWriteNum = gate.PendingWriteNum
		tcpServer.LenMsgLen = gate.LenMsgLen
		tcpServer.MaxMsgLen = gate.MaxMsgLen
		tcpServer.LittleEndian = gate.LittleEndian
		tcpServer.NewAgent = newTcpAgent
		// tcpServer.NewAgent = func(conn *network.TCPConn) network.Agent {
		// 	a := &agent{conn: conn, gate: gate}
		// 	if gate.EventListener != nil {
		// 		gate.EventListener.Go("NewAgent", a)
		// 	}
		// 	return a
		// }
	}

	if wsServer != nil {
		wsServer.Start()
	}
	if tcpServer != nil {
		tcpServer.Start()
	}
	<-closeSig
	if wsServer != nil {
		wsServer.Close()
	}
	if tcpServer != nil {
		tcpServer.Close()
	}
}

func (gate *Gate) OnDestroy() {}

type AgentTemplate struct {
	conn      network.Conn
	gate      *Gate
	userData  interface{}
	Processor network.Processor
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
		err := a.gate.EventListener.Call("CloseAgent", a)
		if err != nil {
			log.Error("chanrpc error: %v", err)
		}
	}
}

func (a *AgentTemplate) WriteMsg(cmd uint16, msg interface{}) {
	if a.gate.Processor != nil {
		data, err := a.gate.Processor.Marshal(cmd, msg)
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

func (a *AgentTemplate) Handler(cmd uint16, data []byte) {
	log.Debug("Handler got: %d %v", cmd, data)
}
