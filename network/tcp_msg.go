package network

import (
	"encoding/binary"
	"errors"
	"io"
	"math"
)

// ---------------------
// | Size | Cmd | Data |
// ---------------------

const PACKET_HEAD_SIZE int = 4

type Head struct {
	Size uint16
	Cmd  uint16
}

type MsgParser struct {
	lenMsgLen    int
	minMsgLen    uint32
	maxMsgLen    uint32
	littleEndian bool
}

func NewMsgParser() *MsgParser {
	p := new(MsgParser)
	p.lenMsgLen = 2
	p.minMsgLen = 1
	p.maxMsgLen = 4096
	p.littleEndian = false

	return p
}

// It's dangerous to call the method on reading or writing
func (p *MsgParser) SetMsgLen(lenMsgLen int, minMsgLen uint32, maxMsgLen uint32) {
	// fix to short now.
	p.lenMsgLen = 2

	if minMsgLen != 0 {
		p.minMsgLen = minMsgLen
	}
	if maxMsgLen != 0 {
		p.maxMsgLen = maxMsgLen
	}

	var max uint32
	switch p.lenMsgLen {
	case 1:
		max = math.MaxUint8
	case 2:
		max = math.MaxUint16
	case 4:
		max = math.MaxUint32
	}
	if p.minMsgLen > max {
		p.minMsgLen = max
	}
	if p.maxMsgLen > max {
		p.maxMsgLen = max
	}
}

// It's dangerous to call the method on reading or writing
func (p *MsgParser) SetByteOrder(littleEndian bool) {
	p.littleEndian = littleEndian
}

// ParseHead
func (p *MsgParser) ParseHead(data []byte) (*Head, error) {
	if len(data) != 4 {
		return nil, fmt.Errorf("HEAD ERROR")
	}
	var head Head
	buf_read := bytes.NewBuffer(data[:PACKET_HEAD_SIZE])

	var byteorder ByteOrder
	if p.littleEndian {
		byteorder = binary.LittleEndian
	} else {
		byteorder = binary.BigEndian
	}

	err := binary.Read(buf_read, byteorder, &head)
	if err != nil {
		return nil, err
	}
	return &head, nil
}

// goroutine safe
func (p *MsgParser) Read(r io.Reader) ([]byte, error) {
	bufMsgLen := make([]byte, PACKET_HEAD_SIZE)

	// read len
	if _, err := io.ReadFull(conn, bufMsgLen); err != nil {
		return nil, err
	}

	header, parse_err := ParseHead(bufMsgLen)
	if parse_err != nil {
		return nil, err
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

	msg := make([]byte, PACKET_HEAD_SIZE+head.Size)

	var byteorder ByteOrder
	if p.littleEndian {
		byteorder = binary.LittleEndian
	} else {
		byteorder = binary.BigEndian
	}

	binary.Write(msg, byteorder, head)

	// byte order is not care when writing byte.
	msg.Write(data)

	return msg, nil
}

// goroutine safe
func (p *MsgParser) Write(w io.Writer, cmd uint16, data []byte) error {
	msg, err := Pack(cmd, data)
	if err != nil {
		return err
	}
	conn.Write(msg)
	return nil
}
