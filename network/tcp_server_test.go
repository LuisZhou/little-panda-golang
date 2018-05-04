package network_test

import (
	"bytes"
	"fmt"
	"github.com/LuisZhou/lpge/log"
	"github.com/LuisZhou/lpge/network"
	"net"
	"sync"
	"testing"
	"time"
)

// todo:
// + add test the conn destory/close
// + test serer.Close.
// + test the goroutine exist when close and destory.

var wg sync.WaitGroup

type TestAgent struct {
	conn network.Conn
}

func (a *TestAgent) Run() {
	for {
		cmd, data, err := a.conn.ReadMsg()
		if err != nil {
			log.Debug("read message: %v", err) // read EOF
			break
		}

		fmt.Println("", cmd, len(data), data, string(data))
		wg.Done()

		a.conn.WriteMsg(2, []byte{1, 2})
	}
	wg.Done()
}

func (a *TestAgent) OnClose() {
	log.Debug("agent close")
	wg.Done()
}

func TestNewTcpServer(t *testing.T) {
	wg.Add(1)

	tcpServer := new(network.TCPServer)
	tcpServer.Addr = "localhost:6001"
	tcpServer.MaxConnNum = 100
	tcpServer.PendingWriteNum = 100
	tcpServer.MaxMsgLen = 0
	tcpServer.LittleEndian = true
	tcpServer.NewAgent = func(conn *network.TCPConn) network.Agent {
		a := &TestAgent{conn: conn}
		wg.Done()
		return a
	}
	tcpServer.Start()

	tcpAddr, _ := net.ResolveTCPAddr("tcp4", tcpServer.Addr)
	conn, _ := net.DialTCP("tcp", nil, tcpAddr)

	wg.Add(2) // one for read, one for write

	buffer_l := new(bytes.Buffer)
	parser_l := network.NewMsgParser()
	parser_l.SetByteOrder(true)
	//parser_l.Write(buffer_l, 1, []byte{1, 2})
	parser_l.Write(buffer_l, 1, []byte("测试"))
	_, err := conn.Write(buffer_l.Bytes())
	_ = err

	cmd, ret, err2 := parser_l.Read(conn)
	t.Log(cmd, ret, err2)
	wg.Done()

	wg.Wait()

	wg.Add(2)
	//conn.Close()
	tcpServer.Close()
	wg.Wait()

	time.Sleep(1 * time.Second)
}
