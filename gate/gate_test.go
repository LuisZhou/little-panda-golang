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

var (
	wg        sync.WaitGroup
	processor network.Processor
	person    *protobuf.Person
)

func init() {
	person = &protobuf.Person{}
	person.Name = "abc"

	processor = protobuf.NewProtobufProcessor()
	processor.Register(1, protobuf.Person{})
}

func newWsAgent(conn network.Conn, gate *gate.Gate) network.Agent {
	a := &NewAgent{}
	a.Init(conn, gate)
	a.Processor = processor
	wg.Done()
	return a
}

func newTcpAgent(conn network.Conn, gate *gate.Gate) network.Agent {
	a := &NewAgent{}
	a.Init(conn, gate)
	a.Skeleton.RegisterChanRPC(uint16(1), func(args []interface{}) (ret interface{}, err error) {
		wg.Done()
		return nil, nil
	})

	a.Processor = processor
	wg.Done()
	return a
}

func TestNewGate(t *testing.T) {

	gateInstance := &gate.Gate{
		MaxConnNum:      2,
		PendingWriteNum: 20,
		MaxMsgLen:       4096,
		HTTPTimeout:     10 * time.Second,
		TCPAddr:         "127.0.0.1:3563",
		LittleEndian:    true,
		NewWsAgent:      newWsAgent,
		NewTcpAgent:     newTcpAgent,
	}

	gateInstance.OnInit()

	wg.Add(1)

	go func() {
		gateInstance.Run(make(chan bool))
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

	buf, err1 := processor.Marshal(1, person)
	t.Log(buf, err1)

	parser_l.Write(buffer_l, 1, buf)
	_, err_for_write := conn.Write(buffer_l.Bytes())
	_ = err_for_write

	wg.Wait()
	//time.Sleep(time.Second * 3)
}
