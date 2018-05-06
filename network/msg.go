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

// Size of msg header.
const PACKET_HEAD_SIZE uint16 = 4

// Recommend max size of msg.
const RECOMMENDED_MAX_DATA_LEN = 32768

// Head of msg.
type Head struct {
	Size uint16 // size of payload
	Cmd  uint16 // cmd of msg
}

// MsgParser is a msg parser.
type MsgParser struct {
	minMsgLen uint16
	maxMsgLen uint16
	endian    binary.ByteOrder
}

// NewMsgParser create a new msg parser.
func NewMsgParser() *MsgParser {
	p := new(MsgParser)
	p.minMsgLen = 0
	p.maxMsgLen = RECOMMENDED_MAX_DATA_LEN
	p.endian = binary.LittleEndian
	return p
}

// SetMsgLen set min & max of msg total size.
func (p *MsgParser) SetMsgLen(minMsgLen uint16, maxMsgLen uint16) {
	var max uint16 = math.MaxUint16

	if minMsgLen != 0 {
		p.minMsgLen = minMsgLen
	}

	if p.minMsgLen > max {
		p.minMsgLen = max
	}

	if maxMsgLen != 0 {
		p.maxMsgLen = maxMsgLen
	}

	if p.maxMsgLen > max {
		p.maxMsgLen = max
	}
}

// SetByteOrder set the byte order of msg parser, used when read/write data from/to binary.
func (p *MsgParser) SetByteOrder(littleEndian bool) {
	if littleEndian {
		p.endian = binary.LittleEndian
	} else {
		p.endian = binary.BigEndian
	}
}

// ParseHead parse message header from binary.
func (p *MsgParser) ParseHead(data []byte) (*Head, error) {
	if len(data) != int(PACKET_HEAD_SIZE) {
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

// Read read one message(cmd, data) from reader.
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

	msgData := make([]byte, header.Size)
	if _, err := io.ReadFull(conn, msgData); err != nil {
		return 0, nil, err
	}

	return header.Cmd, msgData, nil
}

// Pack package cmd and data to binary.
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

// Write cmd and data to writer.
func (p *MsgParser) Write(conn io.Writer, cmd uint16, data []byte) error {
	msg, err := p.Pack(cmd, data)
	if err != nil {
		return err
	}
	conn.Write(msg)
	return nil
}
