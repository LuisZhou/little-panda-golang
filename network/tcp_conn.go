package network

import (
	"github.com/LuisZhou/lpge/log"
	"net"
	"sync"
)

// ConnSet is a map, mapping conn to struct{}, it is useful for maintain state of all connections.
type ConnSet map[net.Conn]struct{}

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
		//for b, ok := range tcpConn.writeChan {
		for {
			b, ok := <-tcpConn.writeChan

			if ok == false {
				break
			}

			if b == nil {
				break
			}

			_, err := conn.Write(b)
			if err != nil {
				break
			}
		}

		// clean up
		conn.Close()
		tcpConn.Lock()
		tcpConn.closeFlag = true
		tcpConn.Unlock()
	}()

	return tcpConn
}

func (tcpConn *TCPConn) doDestroy() {
	tcpConn.conn.(*net.TCPConn).SetLinger(0)
	// tcpConn.conn.Close()

	if !tcpConn.closeFlag {
		// this will change the tcpConn.closeFlag
		// this will do the conn.Close()
		close(tcpConn.writeChan)
		// tcpConn.closeFlag = true
	}
}

// Destroy do destroy the connect, close the conn, and discards any unsent or unacknowledged data.
// This function normally be called by app.
func (tcpConn *TCPConn) Destroy() {
	tcpConn.Lock()
	defer tcpConn.Unlock()

	tcpConn.doDestroy()
}

// Close do close the connect and clean up.
// This function normally when the socket is close by client or server.
func (tcpConn *TCPConn) Close() {
	tcpConn.Lock()
	defer tcpConn.Unlock()
	if tcpConn.closeFlag {
		return
	}

	// let the goroutine to break.
	tcpConn.doWrite(nil)
	tcpConn.closeFlag = true
}

func (tcpConn *TCPConn) doWrite(b []byte) {
	if len(tcpConn.writeChan) == cap(tcpConn.writeChan) {
		log.Debug("close conn: channel full")
		tcpConn.doDestroy()
		return
	}

	tcpConn.writeChan <- b
}

// Write is the api for writing raw data to the connection. So it implement the io.writer interface.
func (tcpConn *TCPConn) Write(b []byte) (n int, err error) {

	// I don't think we need a lock, write to a closed socket is just fine, except will get an error.
	// Use a lock will make serialization.

	// tcpConn.Lock()
	// defer tcpConn.Unlock()

	if tcpConn.closeFlag || b == nil {
		return
	}

	tcpConn.doWrite(b)

	// compatible with io.writer
	return len(b), nil
}

// Write is the api for reading raw data from the connection. So it implement the io.reader interface.
func (tcpConn *TCPConn) Read(b []byte) (int, error) {
	return tcpConn.conn.Read(b)
}

func (tcpConn *TCPConn) LocalAddr() net.Addr {
	return tcpConn.conn.LocalAddr()
}

func (tcpConn *TCPConn) RemoteAddr() net.Addr {
	return tcpConn.conn.RemoteAddr()
}

// Write is the api for reading data from the connection, and parse the data, return cmd, data, err.
func (tcpConn *TCPConn) ReadMsg() (uint16, []byte, error) {
	return tcpConn.msgParser.Read(tcpConn)
}

// Write is the api for write msg to the connection, msg is composed by cmd and data.
func (tcpConn *TCPConn) WriteMsg(cmd uint16, data []byte) error {
	return tcpConn.msgParser.Write(tcpConn, cmd, data)
}
