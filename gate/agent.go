package gate

import (
	"net"
)

type Agent interface {
	WriteMsg(cmd uint16, msg interface{})
	LocalAddr() net.Addr
	RemoteAddr() net.Addr
	Close()
	Destroy()
	UserData() interface{}
	SetUserData(data interface{})
	Handler(cmd uint16, msg interface{}) error
}
