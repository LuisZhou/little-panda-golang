package gate

import (
	"github.com/LuisZhou/lpge/gate"
	"github.com/LuisZhou/lpge/network"
	"net"
	"sync"
	"testing"
	"time"
)

// todo
// add

type NewAgent struct {
	gate.AgentTemplate
}

func TestNewGate(t *testing.T) {
	var wg sync.WaitGroup

	gateInstance := &gate.Gate{
		MaxConnNum:      2,
		PendingWriteNum: 20,
		MaxMsgLen:       4096,
		HTTPTimeout:     10 * time.Second,
		TCPAddr:         "127.0.0.1:3563",
		LittleEndian:    true,
	}

	wg.Add(1)

	go func() {
		gateInstance.Run(make(chan bool), func(conn network.Conn) network.Agent {
			t.Log("test")
			a := &NewAgent{}
			a.Init(conn, gateInstance)
			wg.Done()
			return a
		}, func(conn network.Conn) network.Agent {
			t.Log("test")
			a := &NewAgent{}
			a.Init(conn, gateInstance)
			wg.Done()
			return a
		})
	}()

	time.Sleep(time.Second)
	tcpAddr, _ := net.ResolveTCPAddr("tcp4", "127.0.0.1:3563")
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	t.Log(conn, err)

	wg.Wait()
}
