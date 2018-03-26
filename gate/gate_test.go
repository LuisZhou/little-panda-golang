package gate

import (
	"bytes"
	"github.com/LuisZhou/lpge/gate"
	"github.com/LuisZhou/lpge/network"
	"github.com/LuisZhou/lpge/network/processor/protobuf"
	"net"
	"sync"
	"testing"
	"time"
)

// todo
// add close(), desctory() test.

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

	person := &protobuf.Person{}
	person.Name = "abc"

	p := protobuf.NewProtobufProcessor()
	p.Register(1, protobuf.Person{})

	go func() {
		gateInstance.Run(make(chan bool), func(conn network.Conn) network.Agent {
			t.Log("test")
			a := &NewAgent{}
			a.Init(conn, gateInstance)
			a.Processor = p
			wg.Done()
			return a
		}, func(conn network.Conn) network.Agent {
			t.Log("test")
			a := &NewAgent{}
			a.Init(conn, gateInstance)
			a.Skeleton.RegisterChanRPC(uint16(1), func(args []interface{}) (ret interface{}, err error) {
				t.Log("what?", args)
				wg.Done()
				return nil, nil
			})

			a.Processor = p
			wg.Done()
			return a
		})
	}()

	time.Sleep(time.Second)
	tcpAddr, _ := net.ResolveTCPAddr("tcp4", "127.0.0.1:3563")
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	t.Log(conn, err)

	wg.Wait()

	wg.Add(1)

	buffer_l := new(bytes.Buffer)
	parser_l := network.NewMsgParser()
	parser_l.SetByteOrder(true)

	buf, err1 := p.Marshal(1, person)
	t.Log(buf, err1)

	parser_l.Write(buffer_l, 1, buf)
	_, err_for_write := conn.Write(buffer_l.Bytes())
	_ = err_for_write

	wg.Wait()
	//time.Sleep(time.Second * 3)
}
