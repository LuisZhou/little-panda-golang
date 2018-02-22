package network_test

import (
	"bytes"
	"github.com/LuisZhou/lpge/network"
	"testing"
)

// todo: I need benchmark test and example.

func TestReadWrite(t *testing.T) {
	buffer := new(bytes.Buffer) //bytes.NewBuffer()
	parser := network.NewMsgParser()
	parser.Write(buffer, 1, []byte{1, 2})
	t.Log(buffer.Bytes())
	msg, _ := parser.Read(buffer)
	t.Log(msg)
}

func TestLittleEndian(t *testing.T) {
	buffer_l := new(bytes.Buffer)
	parser_l := network.NewMsgParser()
	parser_l.SetByteOrder(true)
	parser_l.Write(buffer_l, 1, []byte{1, 2})
	t.Log(buffer_l.Bytes())
	msg_l, _ := parser_l.Read(buffer_l)
	t.Log(msg_l)
}
