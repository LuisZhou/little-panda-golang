package network

import (
	"github.com/LuisZhou/lpge/log"
	"net"
	"sync"
	"time"
)

// ConnSet is map of all TCPConn.
type ConnSet map[*TCPConn]struct{}

// TCPServer is a tcp server.
type TCPServer struct {
	Addr            string               // tcp server address
	MaxConnNum      int                  // max connection per server.
	PendingWriteNum int                  // write channel buffer number, per agent.
	NewAgent        func(*TCPConn) Agent // new agent creator, called when clint come in.
	ln              net.Listener         // tcp listener.
	conns           ConnSet              // map of all TCPConns of thie server.
	mutexConns      sync.Mutex           // Mutex protect conns.
	wgLn            sync.WaitGroup       // WaitGroup protects start process to finish of server.
	wgConns         sync.WaitGroup       // WaitGroup protect exist process of connects
	MinMsgLen       uint16               // min Msg Len of MsgParser of server.
	MaxMsgLen       uint16               // max Msg Len of MsgParser of server.
	LittleEndian    bool                 // define endia of MsgParser.
	msgParser       *MsgParser           // point ot MsgParser.
}

// Start do start running of server, server runs in a one goroutine.
func (server *TCPServer) Start() {
	server.init()
	go server.run()
}

// init do initition according to the parameter of server.
func (server *TCPServer) init() {
	ln, err := net.Listen("tcp", server.Addr)
	if err != nil {
		log.Fatal("%v", err)
	}
	server.ln = ln

	if server.MaxConnNum <= 0 {
		server.MaxConnNum = 100
		log.Release("invalid MaxConnNum, reset to %v", server.MaxConnNum)
	}
	if server.PendingWriteNum <= 0 {
		server.PendingWriteNum = 100
		log.Release("invalid PendingWriteNum, reset to %v", server.PendingWriteNum)
	}
	if server.NewAgent == nil {
		log.Fatal("NewAgent must not be nil")
	}

	server.conns = make(ConnSet)

	msgParser := NewMsgParser()
	msgParser.SetMsgLen(server.MinMsgLen, server.MaxMsgLen)
	msgParser.SetByteOrder(server.LittleEndian)
	server.msgParser = msgParser
}

// run starts to do the job of server.
func (server *TCPServer) run() {
	server.wgLn.Add(1)
	defer server.wgLn.Done()

	var tempDelay time.Duration
	for {
		conn, err := server.ln.Accept()

		// when error happen.
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				log.Release("accept error: %v; retrying in %v", err, tempDelay)
				time.Sleep(tempDelay)
				continue
			}
			return
		}
		tempDelay = 0

		// create new TCPConn.
		server.mutexConns.Lock()
		if len(server.conns) >= server.MaxConnNum {
			server.mutexConns.Unlock()
			conn.Close()
			log.Debug("too many connections")
			continue
		}
		tcpConn := newTCPConn(conn, server.PendingWriteNum, server.msgParser)
		server.conns[tcpConn] = struct{}{}
		server.mutexConns.Unlock()

		// add one wait for the connecion
		server.wgConns.Add(1)

		// create new agent for the new TCPConn.
		agent := server.NewAgent(tcpConn)

		// create a new goroutine for running of the agent.
		go func() {
			// once the connection encounter error from Run, the Run() should exist.
			agent.Run()
			// cleanup
			tcpConn.Close()
			server.mutexConns.Lock()
			delete(server.conns, tcpConn)
			server.mutexConns.Unlock()
			agent.OnClose()
			// exist process finish.
			server.wgConns.Done()
		}()
	}
}

// Close shut down the listener of tcp and clean up all TCPConn.
func (server *TCPServer) Close() {
	// shut down all the close conn, cause the goroutine of server to exist.
	server.ln.Close()
	// wait run goroutine of the server to exist.
	server.wgLn.Wait()
	// protect the conns
	server.mutexConns.Lock()
	for conn := range server.conns {
		conn.Close()
	}
	server.conns = nil
	server.mutexConns.Unlock()
	// wait all conn exit.
	server.wgConns.Wait()
}
