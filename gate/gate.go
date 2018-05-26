/* Package gate stars TCP and websocket server internal, and create new agent when new connect happen. */
package gate

import (
	"github.com/LuisZhou/lpge/conf"
	_ "github.com/LuisZhou/lpge/log"
	"github.com/LuisZhou/lpge/module"
	"github.com/LuisZhou/lpge/network"
	"time"
)

type NewAgent func(conn network.Conn, gate *Gate) network.Agent

// Gate for ws and tcp connection.
type Gate struct {
	*module.Skeleton               // implement of module.
	MaxConnNum       int           // max conn of both tcp and ws connect.
	PendingWriteNum  int           // write channel buffer number, per agent, for both tcp and ws connect.
	MaxMsgLen        uint16        // max Msg Len of MsgParser of server, for both tcp and ws connect.
	WSAddr           string        // websocket server address.
	HTTPTimeout      time.Duration // websocket http timeout.
	CertFile         string        // websocket http cert file.
	KeyFile          string        // websocket http key file.
	NewWsAgent       NewAgent      // websocket creator for new agent.
	TCPAddr          string        // tcp server address.
	LittleEndian     bool          // tcp little endian or not of tcp connection.
	NewTcpAgent      NewAgent      // tcp creator for new agent.
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
			gate.Skeleton.GoRpc("NewAgent", a)
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
			gate.Skeleton.GoRpc("NewAgent", a)
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
		ChanRPCLen:         conf.GateConfig.ChanRPCLen,
		TimeoutAsynRet:     conf.GateConfig.TimeoutAsynRet,
	}
	s.Init()
	gate.Skeleton = s
}

func (gate *Gate) OnDestroy() {}
