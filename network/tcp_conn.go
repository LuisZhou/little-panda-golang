package network

import (
	"github.com/LuisZhou/lpge/log"
	"net"
	"sync"
)

// TCPConn combile a tcp connection with a MsgParser, give the connection with proper capability of read/write msg.
// It also allocate some buffer channel to allow parallelly writing for others goroutines.
// The write goroutine of the connection starts when call newTCPConn(), and another goroutine of the connection starts when
// server(such as TCPServer) accept new connect, the server use Agent.Run() as the function for the goroutine, we normally
// use a for-loop to read msg for the agent in Agent.Run(), whe Agent.Run() exist, means the socket is close, and the
// read goroutine will do the clean for the agent.
type TCPConn struct {
	sync.Mutex
	conn      net.Conn
	writeChan chan []byte
	closeFlag bool
	msgParser *MsgParser
}

// newTCPConn do combile conn with msgParser, and allocate buffer channel of size 'pendingWriteNum' for write.
// This function also start a goroutine for read from write buffer, and write the data to conn.
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
		tcpConn.Destroy()
		log.Debug("exist")
	}()

	return tcpConn
}

// only called internal. The caller should get the lock.
func (tcpConn *TCPConn) doDestroy() {
	if !tcpConn.closeFlag {
		tcpConn.conn.(*net.TCPConn).SetLinger(0)
		tcpConn.conn.Close()

		// close the chan, let the goroutine to exist.
		close(tcpConn.writeChan)
		tcpConn.closeFlag = true
	}
}

// Destroy do destroy the connect, close the conn, and discards any unsent or unacknowledged data.
// This function normally be called by app.
func (tcpConn *TCPConn) Destroy() {
	tcpConn.Lock()
	defer tcpConn.Unlock()

	tcpConn.doDestroy()
}

func (tcpConn *TCPConn) doWrite(b []byte) {
	if len(tcpConn.writeChan) == cap(tcpConn.writeChan) {
		log.Debug("close conn: channel full")
		tcpConn.doDestroy()
		return
	}

	tcpConn.writeChan <- b
}

// Write do write data to write channel, write implements the io.Write interface.
func (tcpConn *TCPConn) Write(b []byte) (n int, err error) {
	tcpConn.Lock()
	defer tcpConn.Unlock()

	if tcpConn.closeFlag || b == nil {
		return
	}

	tcpConn.doWrite(b)

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
