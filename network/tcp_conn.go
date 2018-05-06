package network

import (
	"errors"
	"github.com/LuisZhou/lpge/log"
	"net"
	"sync"
)

// TCPConn repsent one session of TCP, giving the ablility of RW (msg) for agent.
type TCPConn struct {
	sync.Mutex
	conn      net.Conn
	writeChan chan []byte
	closeFlag bool
	msgParser *MsgParser
}

// newTCPConn create a new TCPConn.
func newTCPConn(conn net.Conn, pendingWriteNum int, msgParser *MsgParser) *TCPConn {
	tcpConn := new(TCPConn)
	tcpConn.conn = conn
	tcpConn.writeChan = make(chan []byte, pendingWriteNum)
	tcpConn.msgParser = msgParser

	go func() {
		for {
			b, ok := <-tcpConn.writeChan
			if ok == false || b == nil {
				break
			}

			_, err := conn.Write(b)
			if err != nil {
				break
			}
		}
		tcpConn.Close()
		log.Debug("tcpConn write routine exist")
	}()

	return tcpConn
}

// doDestroy do the clean, and only called by internal. The caller should first get the lock.
func (tcpConn *TCPConn) doClose() {
	if !tcpConn.closeFlag {
		tcpConn.conn.(*net.TCPConn).SetLinger(0)
		tcpConn.conn.Close()

		// close the chan, let the goroutine to exist.
		close(tcpConn.writeChan)
		tcpConn.closeFlag = true
	}
}

// Destroy do destroy the connect.
func (tcpConn *TCPConn) Close() {
	tcpConn.Lock()
	defer tcpConn.Unlock()
	tcpConn.doClose()
}

// Write do write data to write channel, write implements the io.Write interface.
func (tcpConn *TCPConn) Write(b []byte) (n int, err error) {
	tcpConn.Lock()
	defer tcpConn.Unlock()

	if tcpConn.closeFlag || b == nil {
		return 0, errors.New("conn closed")
	}

	if len(tcpConn.writeChan) == cap(tcpConn.writeChan) {
		log.Debug("close conn: channel full")
		tcpConn.doClose()
		return 0, errors.New("channel full")
	}

	tcpConn.writeChan <- b

	return len(b), nil
}

// Read implements the io.Reader interface.
func (tcpConn *TCPConn) Read(b []byte) (int, error) {
	return tcpConn.conn.Read(b)
}

// LocalAddr returns the local network address.
func (tcpConn *TCPConn) LocalAddr() net.Addr {
	return tcpConn.conn.LocalAddr()
}

// RemoteAddr returns the remote network address.
func (tcpConn *TCPConn) RemoteAddr() net.Addr {
	return tcpConn.conn.RemoteAddr()
}

// ReadMsg is the api for reading msg from the connection.
func (tcpConn *TCPConn) ReadMsg() (uint16, []byte, error) {
	// msgParser read from io.Reader
	return tcpConn.msgParser.Read(tcpConn)
}

// Write is the api for writing msg to the connection.
func (tcpConn *TCPConn) WriteMsg(cmd uint16, data []byte) error {
	// msgParse Write cmd, data to io.Writer
	return tcpConn.msgParser.Write(tcpConn, cmd, data)
}
