package network

import (
	"errors"
	"github.com/LuisZhou/lpge/log"
	"github.com/gorilla/websocket"
	"net"
	_ "strconv"
	"sync"
)

const WS_HEADER_SIZE uint32 = 2

// WSConn repsent one session of WS, giving the ablility of RW (msg) for agent.
type WSConn struct {
	sync.Mutex
	conn      *websocket.Conn
	writeChan chan []byte
	maxMsgLen uint32
	closeFlag bool
}

// newWSConn create a new WSConn.
func newWSConn(conn *websocket.Conn, pendingWriteNum int, maxMsgLen uint32) *WSConn {
	wsConn := new(WSConn)
	wsConn.conn = conn
	wsConn.writeChan = make(chan []byte, pendingWriteNum)
	wsConn.maxMsgLen = maxMsgLen

	go func() {
		for {
			b, ok := <-wsConn.writeChan

			if ok == false || b == nil {
				break
			}

			err := conn.WriteMessage(websocket.BinaryMessage, b)
			if err != nil {
				break
			}
		}

		wsConn.Close()
		log.Debug("wsConn write routine exist")
	}()

	return wsConn
}

// doClose do the clean, and only called by internal. The caller should first get the lock.
func (wsConn *WSConn) doClose() {
	if !wsConn.closeFlag {
		wsConn.conn.UnderlyingConn().(*net.TCPConn).SetLinger(0)
		wsConn.conn.Close()

		log.Debug("doClose()")
		close(wsConn.writeChan)
		wsConn.closeFlag = true
	}
}

// Close do destroy the connect.
func (wsConn *WSConn) Close() {
	wsConn.Lock()
	defer wsConn.Unlock()

	wsConn.doClose()
}

// doWrite do write data to write channel, write implements the io.Write interface.
func (wsConn *WSConn) doWrite(b []byte) {
	if len(wsConn.writeChan) == cap(wsConn.writeChan) {
		log.Debug("close conn: channel full")
		wsConn.doClose()
		return
	}

	wsConn.writeChan <- b
}

// LocalAddr returns the local network address.
func (wsConn *WSConn) LocalAddr() net.Addr {
	return wsConn.conn.LocalAddr()
}

// RemoteAddr returns the remote network address.
func (wsConn *WSConn) RemoteAddr() net.Addr {
	return wsConn.conn.RemoteAddr()
}

// ReadMsg is the api for reading msg from the connection.
func (wsConn *WSConn) ReadMsg() (uint16, []byte, error) {
	_, b, err := wsConn.conn.ReadMessage()
	if err != nil {
		return 0, b, err
	}

	var cmd uint16 = uint16((uint16(b[0]) & 0xff) | (uint16(b[1]) << 8 & 0xff00))
	return cmd, b[WS_HEADER_SIZE:], err
}

// Write is the api for writing msg to the connection.
func (wsConn *WSConn) WriteMsg(cmd uint16, data []byte) error {
	wsConn.Lock()
	defer wsConn.Unlock()
	if wsConn.closeFlag {
		return errors.New("ws connect has closed")
	}

	var msgLen uint32 = uint32(len(data)) + WS_HEADER_SIZE
	if msgLen > wsConn.maxMsgLen {
		return errors.New("message too long")
	} else if msgLen < WS_HEADER_SIZE {
		return errors.New("message too short")
	}

	// use little endian. !!double_check!!
	wsConn.doWrite(append([]byte{byte(cmd & 0xff), byte(cmd >> 8 & 0xff)}, data[:]...))
	return nil
}
