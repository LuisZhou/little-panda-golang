/* Package gate stars TCP and websocket server internal, and create new agent when new connect happen. */
package gate

import (
	"github.com/LuisZhou/lpge/chanrpc"
	"github.com/LuisZhou/lpge/conf"
	"github.com/LuisZhou/lpge/log"
	"github.com/LuisZhou/lpge/module"
	"github.com/LuisZhou/lpge/network"
	"net"
	"reflect"
	"time"
)

type NewAgent func(conn network.Conn, gate *Gate) network.Agent

// Gate start ws server and tcp server base on its configure. And bind EventListener and NewAgent to it.
// NewAgent will exe when there is a new client here.
type Gate struct {
	MaxConnNum      int
	PendingWriteNum int
	MaxMsgLen       uint16
	//EventListener   *chanrpc.Server

	// websocket
	WSAddr      string
	HTTPTimeout time.Duration
	CertFile    string
	KeyFile     string

	// tcp
	TCPAddr      string
	LittleEndian bool

	*module.Skeleton
	NewWsAgent  NewAgent
	NewTcpAgent NewAgent
}

func (gate *Gate) Run(closeSig chan bool) {
	var wsServer *network.WSServer
	if gate.WSAddr != "" {
		wsServer = new(network.WSServer)
		wsServer.Addr = gate.WSAddr
		wsServer.MaxConnNum = gate.MaxConnNum
		wsServer.PendingWriteNum = gate.PendingWriteNum
		wsServer.MaxMsgLen = uint32(gate.MaxMsgLen) // todo: double check
		wsServer.HTTPTimeout = gate.HTTPTimeout
		wsServer.CertFile = gate.CertFile
		wsServer.KeyFile = gate.KeyFile
		wsServer.NewAgent = func(conn *network.WSConn) network.Agent {
			a := gate.NewWsAgent(conn, gate)
			if gate.Skeleton.ChanRPCServer != nil {
				gate.Skeleton.ChanRPCServer.Go("NewAgent", a)
			}
			return a
		}
	}

	// todo: need another server which is used to cluster.

	var tcpServer *network.TCPServer
	if gate.TCPAddr != "" {
		tcpServer = new(network.TCPServer)
		tcpServer.Addr = gate.TCPAddr
		tcpServer.MaxConnNum = gate.MaxConnNum
		tcpServer.PendingWriteNum = gate.PendingWriteNum
		tcpServer.MaxMsgLen = gate.MaxMsgLen
		tcpServer.LittleEndian = gate.LittleEndian
		tcpServer.NewAgent = func(conn *network.TCPConn) network.Agent {
			a := gate.NewTcpAgent(conn, gate)
			if gate.Skeleton.ChanRPCServer != nil {
				gate.Skeleton.ChanRPCServer.Go("NewAgent", a)
			}
			return a
		}
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

func (gate *Gate) OnInit() {
	if gate.NewTcpAgent == nil || gate.NewWsAgent == nil {
		panic("gate miss NewTcpAgent or NewWsAgent")
	}

	s := &module.Skeleton{
		GoLen:              conf.GateConfig.GoLen,
		TimerDispatcherLen: conf.GateConfig.TimerDispatcherLen,
		AsynCallLen:        conf.GateConfig.AsynCallLen,
		ChanRPCServer:      chanrpc.NewServer(conf.GateConfig.ChanRPCLen, time.Duration(conf.GateConfig.TimeoutAsynRet)),
	}
	s.Init()
	gate.Skeleton = s
}

func (gate *Gate) OnDestroy() {}

// define AgentTemplae

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
		ChanRPCServer:      chanrpc.NewServer(conf.AgentConfig.ChanRPCLen, time.Duration(conf.AgentConfig.TimeoutAsynRet)),
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
	if a.gate.Skeleton.ChanRPCServer != nil {
		_, err := a.gate.Skeleton.ChanRPCServer.Call("CloseAgent", a)
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
	a.closeChan <- true
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
