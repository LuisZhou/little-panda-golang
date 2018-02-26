package gate

import (
	"github.com/LuisZhou/lpge/gate"
	"github.com/LuisZhou/lpge/network"
	"net"
	"testing"
	"time"
)

func TestNewGate(t *testing.T) {
	gateInstance := &gate.Gate{
		MaxConnNum:      2,
		PendingWriteNum: 20,
		MaxMsgLen:       4096,
		HTTPTimeout:     10 * time.Second,
		TCPAddr:         "127.0.0.1:3563",
		LittleEndian:    true,
	}

	gateInstance.Run(make(chan bool), func(conn network.Conn) network.Agent {
		t.Log("test")
		a := &gate.AgentTemplate{}
		a.Init(conn, gateInstance)
		return a
	}, func(conn network.Conn) network.Agent {
		t.Log("test")
		a := &gate.AgentTemplate{}
		a.Init(conn, gateInstance)
		return a
	})

	tcpAddr, _ := net.ResolveTCPAddr("tcp4", "127.0.0.1:3563")
	conn, _ := net.DialTCP("tcp", nil, tcpAddr)
	_ = conn
}
