package network

import (
	"crypto/tls"
	"github.com/LuisZhou/lpge/log"
	"github.com/gorilla/websocket"
	"net"
	"net/http"
	"sync"
	"time"
)

// WebsocketConnSet is a map mapping WSConn to struct{}, used to manager all conn.
type WebsocketConnSet map[*WSConn]struct{}

// WSServer is a ws server.
type WSServer struct {
	Addr            string              // ws server address.
	MaxConnNum      int                 // max connection per server.
	PendingWriteNum int                 // write channel buffer number, per agent.
	MaxMsgLen       uint32              // max len of msg of ws.
	HTTPTimeout     time.Duration       // http timeout.
	CertFile        string              // cert file of http server.
	KeyFile         string              // key file of http server.
	NewAgent        func(*WSConn) Agent // new agent creator, called when clint come in.
	ln              net.Listener        // net listener.
	handler         *WSHandler          // ws handler.
}

// WSHandler is Handler use to handle ws request.
type WSHandler struct {
	maxConnNum      int                 // max connection per server, same as WSServer.
	pendingWriteNum int                 // write channel buffer number, per agent, same as WSServer.
	maxMsgLen       uint32              // max len of msg of ws.
	newAgent        func(*WSConn) Agent // new agent creator, called when clint come in.
	upgrader        websocket.Upgrader  // Upgrader used to get conn of ws.
	conns           WebsocketConnSet    // map of all WSConn of thie server.
	mutexConns      sync.Mutex          // Mutex protect conns.
	wg              sync.WaitGroup      // WaitGroup protect exist process of connects
}

// ServeHTTP is a handle function for one ws request.
func (handler *WSHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	conn, err := handler.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Debug("upgrade error: %v", err)
		return
	}
	conn.SetReadLimit(int64(handler.maxMsgLen))

	handler.wg.Add(1)
	defer handler.wg.Done()

	handler.mutexConns.Lock()
	if handler.conns == nil {
		handler.mutexConns.Unlock()
		conn.Close()
		return
	}
	if len(handler.conns) >= handler.maxConnNum {
		handler.mutexConns.Unlock()
		conn.Close()
		log.Debug("too many connections")
		return
	}
	wsConn := newWSConn(conn, handler.pendingWriteNum, handler.maxMsgLen)
	handler.conns[wsConn] = struct{}{}
	handler.mutexConns.Unlock()

	agent := handler.newAgent(wsConn)
	agent.Run()

	// cleanup
	wsConn.Close()
	handler.mutexConns.Lock()
	delete(handler.conns, wsConn)
	handler.mutexConns.Unlock()
	agent.OnClose()
}

func (server *WSServer) Start() {
	ln, err := net.Listen("tcp", server.Addr)
	if err != nil {
		log.Fatal("%v", err)
	}

	if server.MaxConnNum <= 0 {
		server.MaxConnNum = 100
		log.Release("invalid MaxConnNum, reset to %v", server.MaxConnNum)
	}
	if server.PendingWriteNum <= 0 {
		server.PendingWriteNum = 100
		log.Release("invalid PendingWriteNum, reset to %v", server.PendingWriteNum)
	}
	if server.MaxMsgLen <= 0 {
		server.MaxMsgLen = 4096
		log.Release("invalid MaxMsgLen, reset to %v", server.MaxMsgLen)
	}
	if server.HTTPTimeout <= 0 {
		server.HTTPTimeout = 10 * time.Second
		log.Release("invalid HTTPTimeout, reset to %v", server.HTTPTimeout)
	}
	if server.NewAgent == nil {
		log.Fatal("NewAgent must not be nil")
	}

	if server.CertFile != "" || server.KeyFile != "" {
		config := &tls.Config{}
		config.NextProtos = []string{"http/1.1"}

		var err error
		config.Certificates = make([]tls.Certificate, 1)
		config.Certificates[0], err = tls.LoadX509KeyPair(server.CertFile, server.KeyFile)
		if err != nil {
			log.Fatal("%v", err)
		}

		ln = tls.NewListener(ln, config)
	}

	server.ln = ln

	// create a handler.
	server.handler = &WSHandler{
		maxConnNum:      server.MaxConnNum,
		pendingWriteNum: server.PendingWriteNum,
		maxMsgLen:       server.MaxMsgLen,
		newAgent:        server.NewAgent,
		conns:           make(WebsocketConnSet),
		upgrader: websocket.Upgrader{
			HandshakeTimeout: server.HTTPTimeout,
			CheckOrigin:      func(_ *http.Request) bool { return true },
		},
	}

	// create a http server.
	httpServer := &http.Server{
		Addr:           server.Addr,
		Handler:        server.handler, // pass the handler to http server.
		ReadTimeout:    server.HTTPTimeout,
		WriteTimeout:   server.HTTPTimeout,
		MaxHeaderBytes: 1024,
	}

	go httpServer.Serve(ln)
}

func (server *WSServer) Close() {
	server.ln.Close()

	server.handler.mutexConns.Lock()
	for conn := range server.handler.conns {
		conn.Close()
	}
	server.handler.conns = nil
	server.handler.mutexConns.Unlock()

	server.handler.wg.Wait()
}
