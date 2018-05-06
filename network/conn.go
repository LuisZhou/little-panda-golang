package network

import (
	"net"
)

// Interface of Conn, present one session.
type Conn interface {
	// ReadMsg read msg from the conn.
	ReadMsg() (uint16, []byte, error)
	// WriteMsg write msg to conn.
	WriteMsg(cmd uint16, data []byte) error
	// LocalAddr get the server addr.
	LocalAddr() net.Addr
	// LocalAddr get the client addr.
	RemoteAddr() net.Addr
	// Close closes the conn and do clean up.
	Close()
}
