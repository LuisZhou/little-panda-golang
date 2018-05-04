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
const RECOMMENDED_MAX_DATA_LEN = 32768

type Head struct {
	Size uint16
	Cmd  uint16
}

type MsgParser struct {
	minMsgLen uint16
	maxMsgLen uint16
	endian    binary.ByteOrder
}

func NewMsgParser() *MsgParser {
	p := new(MsgParser)
	p.minMsgLen = 0
	p.maxMsgLen = RECOMMENDED_MAX_DATA_LEN
	p.endian = binary.LittleEndian
	return p
}

// It's dangerous to call the method on reading or writing
func (p *MsgParser) SetMsgLen(minMsgLen uint16, maxMsgLen uint16) {
	if minMsgLen != 0 {
		p.minMsgLen = minMsgLen
	}
	if maxMsgLen != 0 {
		p.maxMsgLen = maxMsgLen
	}

	var max uint16 = math.MaxUint16
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
func (p *MsgParser) Read(conn io.Reader) (uint16, []byte, error) {
	bufMsgLen := make([]byte, PACKET_HEAD_SIZE)

	if _, err := io.ReadFull(conn, bufMsgLen); err != nil {
		return 0, nil, err
	}

	header, parse_err := p.ParseHead(bufMsgLen)
	if parse_err != nil {
		return 0, nil, parse_err
	}

	if header.Size > p.maxMsgLen {
		return 0, nil, errors.New("message too long")
	} else if header.Size < p.minMsgLen {
		return 0, nil, errors.New("message too short")
	}

	// data
	msgData := make([]byte, header.Size)
	if _, err := io.ReadFull(conn, msgData); err != nil {
		return 0, nil, err
	}

	return header.Cmd, msgData, nil
}

func (p *MsgParser) Pack(cmd uint16, data []byte) (ret []byte, err error) {
	head := &Head{}
	head.Cmd = cmd
	head.Size = uint16(len(data))

	if head.Size > p.maxMsgLen {
		return nil, errors.New("message too long")
	} else if head.Size < p.minMsgLen {
		return nil, errors.New("message too short")
	}

	msg := new(bytes.Buffer)
	binary.Write(msg, p.endian, head)
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
