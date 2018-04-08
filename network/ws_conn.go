package network

import (
	"errors"
	"github.com/LuisZhou/lpge/log"
	"github.com/gorilla/websocket"
	"net"
	_ "strconv"
	"sync"
)

type WebsocketConnSet map[*websocket.Conn]struct{}

type WSConn struct {
	sync.Mutex
	conn      *websocket.Conn
	writeChan chan []byte
	maxMsgLen uint32
	closeFlag bool
}

func newWSConn(conn *websocket.Conn, pendingWriteNum int, maxMsgLen uint32) *WSConn {
	wsConn := new(WSConn)
	wsConn.conn = conn
	wsConn.writeChan = make(chan []byte, pendingWriteNum)
	wsConn.maxMsgLen = maxMsgLen

	go func() {
		for b := range wsConn.writeChan {
			if b == nil {
				break
			}

			err := conn.WriteMessage(websocket.BinaryMessage, b)
			if err != nil {
				break
			}
		}

		conn.Close()
		wsConn.Lock()
		wsConn.closeFlag = true
		wsConn.Unlock()
	}()

	return wsConn
}

func (wsConn *WSConn) doDestroy() {
	wsConn.conn.UnderlyingConn().(*net.TCPConn).SetLinger(0)
	wsConn.conn.Close()

	if !wsConn.closeFlag {
		close(wsConn.writeChan)
		wsConn.closeFlag = true
	}
}

func (wsConn *WSConn) Destroy() {
	wsConn.Lock()
	defer wsConn.Unlock()

	wsConn.doDestroy()
}

func (wsConn *WSConn) Close() {
	wsConn.Lock()
	defer wsConn.Unlock()
	if wsConn.closeFlag {
		return
	}

	wsConn.doWrite(nil)
	wsConn.closeFlag = true
}

func (wsConn *WSConn) doWrite(b []byte) {
	if len(wsConn.writeChan) == cap(wsConn.writeChan) {
		log.Debug("close conn: channel full")
		wsConn.doDestroy()
		return
	}

	wsConn.writeChan <- b
}

func (wsConn *WSConn) LocalAddr() net.Addr {
	return wsConn.conn.LocalAddr()
}

func (wsConn *WSConn) RemoteAddr() net.Addr {
	return wsConn.conn.RemoteAddr()
}

// goroutine not safe
func (wsConn *WSConn) ReadMsg() (uint16, []byte, error) {
	_, b, err := wsConn.conn.ReadMessage()
	if err != nil {
		return 0, b, err
	}

	var cmd uint16 = uint16((uint16(b[0]) & 0xff) | (uint16(b[1]) << 8 & 0xff00))
	return cmd, b[2:], err
}

// args must not be modified by the others goroutines
func (wsConn *WSConn) WriteMsg(cmd uint16, data []byte) error {
	wsConn.Lock()
	defer wsConn.Unlock()
	if wsConn.closeFlag {
		return errors.New("ws connect has closed")
	}

	var msgLen uint32 = uint32(len(data)) + uint32(2)
	if msgLen > wsConn.maxMsgLen {
		return errors.New("message too long")
	} else if msgLen < 2 {
		return errors.New("message too short")
	}

	wsConn.doWrite(append([]byte{byte(cmd & 0xff), byte(cmd >> 8 & 0xff)}, data[:]...))
	return nil
}
