package network_test

import (
	"fmt"
	"github.com/LuisZhou/lpge/log"
	"github.com/LuisZhou/lpge/network"
	"sync"
	"testing"
)

var wg sync.WaitGroup

type TestAgent struct {
	conn network.Conn
}

func (a *TestAgent) Run() {
	fmt.Println("run")

	for {
		cmd, data, err := a.conn.ReadMsg()
		if err != nil {
			fmt.Println("read message: %v", err) // conn will close.
			break
		}

		fmt.Println("read message: %v", cmd, data)

		// fmt.Println("", cmd, len(data), data, string(data))
		// wg.Done()

		// a.conn.WriteMsg(2, []byte{1, 2})
	}
}

func (a *TestAgent) OnClose() {
	log.Debug("agent close")
	//wg.Done()
}

func TestNewTcpServer(t *testing.T) {
	wg.Add(1)

	tcpServer := new(network.WSServer)
	tcpServer.Addr = "0.0.0.0:6001"
	tcpServer.MaxConnNum = 100
	tcpServer.PendingWriteNum = 100
	tcpServer.MaxMsgLen = 0
	tcpServer.NewAgent = func(conn *network.WSConn) network.Agent {
		fmt.Println("new client")
		a := &TestAgent{conn: conn}
		//wg.Done()
		return a
	}
	tcpServer.Start()

	wg.Wait()
}
