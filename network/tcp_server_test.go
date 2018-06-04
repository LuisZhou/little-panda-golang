package network_test

import (
	_ "bytes"
	"fmt"
	"github.com/LuisZhou/lpge/log"
	"github.com/LuisZhou/lpge/network"
	"net"
	"sync"
	"testing"
	"time"
)

var wg sync.WaitGroup

type TestAgent struct {
	conn network.Conn
}

// is a echo server/client
func (a *TestAgent) Run() {
	for {
		cmd, data, err := a.conn.ReadMsg()
		if err != nil {
			log.Debug("read message: %v", err) // read EOF
			break
		}

		fmt.Println("", cmd, len(data), data, string(data))
		wg.Done()

		// echo the msg.
		a.conn.WriteMsg(cmd, data)
	}
	wg.Done()
}

func (a *TestAgent) OnClose() {
	log.Debug("agent close")
	wg.Done()
}

func newClient(addr string) *network.TCPClient {
	tcpClient := &network.TCPClient{
		Addr:            addr,
		ConnNum:         1,
		ConnectInterval: 1 * time.Second,
		PendingWriteNum: 100,
		NewAgent: func(conn *network.TCPConn) network.Agent {
			a := &TestAgent{conn: conn}
			wg.Done()
			return a
		},
		MinMsgLen:    0,
		MaxMsgLen:    4096,
		LittleEndian: true,
	}
	tcpClient.Start()
	return tcpClient
}

func newServer() *network.TCPServer {
	tcpServer := new(network.TCPServer)
	tcpServer.Addr = "localhost:6001"
	tcpServer.MaxConnNum = 100
	tcpServer.PendingWriteNum = 100
	tcpServer.MaxMsgLen = 0
	tcpServer.LittleEndian = true
	tcpServer.NewAgent = func(conn *network.TCPConn) network.Agent {
		fmt.Println("new client come")
		a := &TestAgent{conn: conn}
		wg.Done()
		return a
	}
	tcpServer.Start()
	return tcpServer
}

func TestNewTcpServer(t *testing.T) {
	// start server
	wg.Add(1)

	tcpServer := newServer()

	// client

	tcpAddr, _ := net.ResolveTCPAddr("tcp", tcpServer.Addr)
	conn, _ := net.DialTCP("tcp", nil, tcpAddr)

	// one for read, one for write
	wg.Add(2)

	//buffer_l := new(bytes.Buffer)

	// msg parser.
	msgParse := network.NewMsgParser()
	msgParse.SetByteOrder(true)

	// write (blocking oper)
	// msgParse.Write(buffer_l, 1, []byte("测试"))
	// _, err := conn.Write(buffer_l.Bytes())
	// _ = err
	msgParse.Write(conn, 1, []byte("测试"))

	// read (blocking oper)
	cmd, ret, err2 := msgParse.Read(conn)
	t.Log(cmd, ret, string(ret), err2)
	wg.Done()

	wg.Wait()

	// one for Run() exist, one for OnClose()
	wg.Add(2)
	conn.Close()
	wg.Wait()

	tcpServer.Close()

	time.Sleep(1 * time.Second)
}

func TestNewTcpClient(t *testing.T) {
	wg.Add(2)

	tcpServer := newServer()
	client := newClient(tcpServer.Addr)

	wg.Wait()

	wg.Add(2)

	// danger!
	for conn := range client.Conns() {
		conn.WriteMsg(1, []byte("测试"))
	}
	wg.Wait()
}
