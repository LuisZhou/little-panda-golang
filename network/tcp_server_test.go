package network_test

import (
	"github.com/LuisZhou/lpge/network"
	"testing"
)

// todo
// + need client.
// + there is a log in Start(), check why no here.

type TestAgent struct {
}

func (a *TestAgent) Run() {

}

func (a *TestAgent) OnClose() {

}

func TestNewTcpServer(t *testing.T) {
	tcpServer := new(network.TCPServer)
	tcpServer.Addr = "localhost:6001"
	tcpServer.MaxConnNum = 100
	tcpServer.PendingWriteNum = 100
	tcpServer.MaxMsgLen = 0
	tcpServer.LittleEndian = true
	tcpServer.NewAgent = func(conn *network.TCPConn) network.Agent {
		a := &TestAgent{}
		return a
	}
	tcpServer.Start()
	t.Log("test")
}
