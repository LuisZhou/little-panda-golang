package network

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
)

// ---------------------
// | Size | Cmd | Data |
// ---------------------

const PACKET_HEAD_SIZE uint16 = 4

type Head struct {
	Size uint16
	Cmd  uint16
}

type MsgParser struct {
	lenMsgLen int
	minMsgLen uint16
	maxMsgLen uint16
	endian    binary.ByteOrder
}

func NewMsgParser() *MsgParser {
	p := new(MsgParser)
	p.lenMsgLen = 2
	p.minMsgLen = 1
	p.maxMsgLen = 4096
	p.endian = binary.BigEndian

	return p
}

// It's dangerous to call the method on reading or writing
func (p *MsgParser) SetMsgLen(lenMsgLen int, minMsgLen uint16, maxMsgLen uint16) {
	// fix to short now.
	p.lenMsgLen = 2

	if minMsgLen != 0 {
		p.minMsgLen = minMsgLen
	}
	if maxMsgLen != 0 {
		p.maxMsgLen = maxMsgLen
	}

	var max uint16 = math.MaxUint16
	// switch p.lenMsgLen {
	// case 1:
	// 	max = math.MaxUint8
	// case 2:
	// 	max = math.MaxUint16
	// case 4:
	// 	max = math.MaxUint32
	// }
	if p.minMsgLen > max {
		p.minMsgLen = max
	}
	if p.maxMsgLen > max {
		p.maxMsgLen = max
	}
}

// It's dangerous to call the method on reading or writing
func (p *MsgParser) SetByteOrder(littleEndian bool) {
	if littleEndian {
		p.endian = binary.LittleEndian
	} else {
		p.endian = binary.BigEndian
	}
}

// ParseHead
func (p *MsgParser) ParseHead(data []byte) (*Head, error) {
	if len(data) != 4 {
		return nil, fmt.Errorf("HEAD ERROR")
	}
	var head Head
	buf_read := bytes.NewBuffer(data[:PACKET_HEAD_SIZE])

	err := binary.Read(buf_read, p.endian, &head)
	if err != nil {
		return nil, err
	}
	return &head, nil
}

// goroutine safe
func (p *MsgParser) Read(conn io.Reader) ([]byte, error) {
	bufMsgLen := make([]byte, PACKET_HEAD_SIZE)

	// read len
	if _, err := io.ReadFull(conn, bufMsgLen); err != nil {
		return nil, err
	}

	header, parse_err := p.ParseHead(bufMsgLen)
	if parse_err != nil {
		return nil, parse_err
	}

	// check len
	if header.Size > p.maxMsgLen {
		return nil, errors.New("message too long")
	} else if header.Size < p.minMsgLen {
		return nil, errors.New("message too short")
	}

	// data
	msgData := make([]byte, header.Size)
	if _, err := io.ReadFull(conn, msgData); err != nil {
		return nil, err
	}

	return msgData, nil
}

func (p *MsgParser) Pack(cmd uint16, data []byte) (ret []byte, err error) {
	head := &Head{}
	head.Cmd = cmd
	head.Size = uint16(len(data))

	// check len
	if head.Size > p.maxMsgLen {
		return nil, errors.New("message too long")
	} else if head.Size < p.minMsgLen {
		return nil, errors.New("message too short")
	}

	//msg := make([]byte, PACKET_HEAD_SIZE+head.Size)
	msg := new(bytes.Buffer)

	binary.Write(msg, p.endian, head)

	// byte order is not care when writing byte.
	msg.Write(data)

	return msg.Bytes(), nil
}

// goroutine safe
func (p *MsgParser) Write(conn io.Writer, cmd uint16, data []byte) error {
	msg, err := p.Pack(cmd, data)
	if err != nil {
		return err
	}
	conn.Write(msg)
	return nil
}
