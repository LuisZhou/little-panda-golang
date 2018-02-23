package network

import (
	"net"
)

type Conn interface {
	ReadMsg() (uint16, []byte, error)
	WriteMsg(cmd uint16, data []byte) error
	LocalAddr() net.Addr
	RemoteAddr() net.Addr
	Close()
	Destroy()
}
