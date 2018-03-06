// Pacakge chanrpc provides base communication between service/module in the same process.
package chanrpc

import (
	"errors"
	"fmt"
	"github.com/LuisZhou/lpge/conf"
	"github.com/LuisZhou/lpge/log"
	"runtime"
	"time"
)

// RpcCommon is struct contain some field shared by server and client.
type RpcCommon struct {
	isStarted   bool
	timeout     time.Duration
	skipCounter int
	closeSig    chan bool
}

// CallInfo is struct contain informaction of call to server.
type CallInfo struct {
	id      interface{}
	args    []interface{}
	chanRet chan *RetInfo
	cb      func(interface{}, error)
}

// RetInfo is struct contain informaction of return of call to server.
type RetInfo struct {
	ret interface{}
	err error
	cb  func(interface{}, error)
}

// Server is struct defines a chanrpc server.
type Server struct {
	functions map[interface{}]interface{}
	ChanCall  chan *CallInfo
	RpcCommon
}

// Client is struct defines chanrpc client.
type Client struct {
	s               *Server
	chanSyncRet     chan *RetInfo
	ChanAsynRet     chan *RetInfo
	pendingAsynCall int
	AllowOverFlood  bool
	RpcCommon
}

// SkipCounter get the total number of timeout of write channel.
func (com *RpcCommon) SkipCounter() int {
	return com.skipCounter
}

// NewServer new a server. bufsize define the buffer size of call channel,
// timeout define most wait time for available of channel of clent when return the result to.
func NewServer(bufsize int, timeout time.Duration) *Server {
	s := new(Server)
	s.functions = make(map[interface{}]interface{})
	s.ChanCall = make(chan *CallInfo, bufsize)
	s.closeSig = make(chan bool)
	s.timeout = timeout
	return s
}

// Register register handler f for id. The handler accept any size of any type parameter, return ret as result,
// and err if any error happen.
func (s *Server) Register(id interface{}, f func([]interface{}) (ret interface{}, err error)) {
	if _, ok := s.functions[id]; ok {
		panic(fmt.Sprintf("function id %v: already registered", id))
	}
	s.functions[id] = f
}

// ret write result of ci to channel provided by ci.
func (s *Server) ret(ci *CallInfo, ri *RetInfo) (err error) {
	if ci.chanRet == nil {
		return
	}

	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()

	ri.cb = ci.cb

	select {
	case ci.chanRet <- ri:
	case <-time.After(time.Millisecond * s.timeout):
		s.skipCounter++
	}
	return
}

// exec execute call request internally, server should recover from any panic and error.
func (s *Server) exec(ci *CallInfo) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if conf.LenStackBuf > 0 {
				buf := make([]byte, conf.LenStackBuf)
				l := runtime.Stack(buf, false)
				err = fmt.Errorf("%v: %s", r, buf[:l])
			} else {
				err = fmt.Errorf("%v", r)
			}

			s.ret(ci, &RetInfo{err: fmt.Errorf("%v", r)})
		}
	}()

	f := s.functions[ci.id]
	if f == nil {
		panic(fmt.Sprintf("no function for %s", ci.id))
	}

	ret, err := f.(func([]interface{}) (ret interface{}, err error))(ci.args)
	return s.ret(ci, &RetInfo{ret: ret, err: err})
}

// Exec execute call request from ci. Do not go to the buf channel, just call the handle for ci.
func (s *Server) Exec(ci *CallInfo) {
	err := s.exec(ci)
	if err != nil {
		log.Error("%v", err)
	}
}

// Start start execute coming call request. Its internal goroutine loops forever, execute any coming call request from
// channel, if no request, just block.
func (s *Server) Start() {
	s.isStarted = true
	go func() {
		for {
			select {
			case <-s.closeSig:
				s.close()
				return
			case ci := <-s.ChanCall:
				err := s.exec(ci)
				if err != nil {
					log.Error("%v", err)
				}
			}
		}
	}()
}

// Go async call server with id and args, and no any callback is needed of this sync call.
func (s *Server) Go(id interface{}, args ...interface{}) {
	f := s.functions[id]
	if f == nil {
		return
	}

	defer func() {
		recover()
	}()

	s.ChanCall <- &CallInfo{
		id:   id,
		args: args,
	}
}

// Call do a sync call request to server.
func (s *Server) Call(id interface{}, args ...interface{}) (interface{}, error) {
	return s.Open(0, 0).Call(id, args...)
}

// close shutdown the call channel and return closed-msg to pending call requested before close.
func (s *Server) close() {
	close(s.ChanCall)

	for ci := range s.ChanCall {
		s.ret(ci, &RetInfo{
			err: errors.New("chanrpc server closed"),
		})
	}
}

// Close do shutdown goroutine if has called s.Start() before, and close the call channel and return closed-msg to
// pending call requested before close.
func (s *Server) Close() {
	if s.isStarted {
		s.closeSig <- true
		s.isStarted = false
	} else {
		s.close()
	}
}

// Open open a server to return a client of this server, provide bufsize define the async buffer size of aysnc callback
// channel, and timeout define max wait time when send async call request.
func (s *Server) Open(bufsize int, timeout time.Duration) *Client {
	c := NewClient(bufsize, timeout)
	c.Attach(s)
	return c
}

// NewClient create a client, but not specify which server to attach, provide bufsize define the async buffer size of
// aysnc callback channel, and timeout define max wait time when send async call request.
func NewClient(size int, timeout time.Duration) *Client {
	c := new(Client)
	c.chanSyncRet = make(chan *RetInfo, 1)
	c.ChanAsynRet = make(chan *RetInfo, size)
	c.timeout = timeout
	c.closeSig = make(chan bool)
	return c
}

// Attach attach a client to a server.
func (c *Client) Attach(s *Server) {
	c.s = s
}

// call send the request call ci to server channel, if block is true, the call is sync call, or is a async call.
// If it is a sync call, it allow timeout if server is too busy.
func (c *Client) call(ci *CallInfo, block bool) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()

	// todo: should I first test if the server support the ci support the request.
	// should I wrapper getSupportRequest().

	if block {
		c.s.ChanCall <- ci
	} else {
		select {
		case c.s.ChanCall <- ci:
		case <-time.After(c.timeout):
			c.skipCounter++
			c.pendingAsynCall--
			err = errors.New("server timeout")
		}
	}
	return
}

// validate validate the call id can map to handler of the server.
func (c *Client) validate(id interface{}) (f interface{}, err error) {
	if c.s == nil {
		err = errors.New("server not attached")
		return
	}

	f = c.s.functions[id]
	if f == nil {
		err = fmt.Errorf("function id %v: function not registered", id)
		return
	}

	return
}

// Call do a sync call to attatched server.
func (c *Client) Call(id interface{}, args ...interface{}) (interface{}, error) {
	_, err := c.validate(id)
	if err != nil {
		return nil, err
	}

	err = c.call(&CallInfo{
		id:      id,
		args:    args,
		chanRet: c.chanSyncRet,
	}, true)

	if err != nil {
		return nil, err
	}

	ri := <-c.chanSyncRet
	return ri.ret, ri.err
}

// asyncCall do a async call to server. It is a private method, only called by client internal.
func (c *Client) asynCall(id interface{}, args []interface{}, cb func(interface{}, error)) {
	_, err := c.validate(id)
	if err != nil {
		c.ChanAsynRet <- &RetInfo{err: err, cb: cb}
		return
	}

	err = c.call(&CallInfo{
		id:      id,
		args:    args,
		chanRet: c.ChanAsynRet,
		cb:      cb,
	}, false)

	if err != nil {
		c.ChanAsynRet <- &RetInfo{err: err, cb: cb}
		return
	}
}

// AsyncCall do a async call to server.
func (c *Client) AsynCall(id interface{}, _args ...interface{}) {
	if len(_args) < 1 {
		panic("callback function not found")
	}

	args := _args[:len(_args)-1]
	cb := _args[len(_args)-1]

	switch cb.(type) {
	case func(ret interface{}, err error):
	default:
		panic("definition of callback function is invalid")
	}

	if c.AllowOverFlood == false && c.pendingAsynCall >= cap(c.ChanAsynRet) {
		execCb(&RetInfo{err: errors.New("too many calls"), cb: cb.(func(interface{}, error))})
		return
	}

	c.asynCall(id, args, cb.(func(interface{}, error)))
	c.pendingAsynCall++
}

// exeCb do exec the callback of ri. It is a private method, only called by client internal.
func execCb(ri *RetInfo) {
	defer func() {
		if r := recover(); r != nil {
			if conf.LenStackBuf > 0 {
				buf := make([]byte, conf.LenStackBuf)
				l := runtime.Stack(buf, false)
				log.Error("%v: %s", r, buf[:l])
			} else {
				log.Error("%v", r)
			}
		}
	}()

	if ri.cb != nil {
		ri.cb(ri.ret, ri.err)
	}
	return
}

// Cb do exec the callback of ri when the async call is finish handled by server.
func (c *Client) Cb(ri *RetInfo) {
	c.pendingAsynCall--
	execCb(ri)
}

// Wait wait forever for receiving return from async channel using goroutine internal.
func (c *Client) Wait() {
	c.isStarted = true
	go func() {
		for {
			select {
			case <-c.closeSig:
				c.close()
				return
			case ret := <-c.ChanAsynRet:
				c.Cb(ret)
			}
		}
	}()
}

// close do the clean.
func (c *Client) close() {
	//The rule of thumb here is that only writers should close channels.
}

// Close close goroutine internal if it exist, and do other cleanup job.
func (c *Client) Close() {
	if c.isStarted {
		c.closeSig <- true
		c.isStarted = false
	} else {
		c.close()
	}
}
