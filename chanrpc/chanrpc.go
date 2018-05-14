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

// RpcCommon is common field shared by rpc server and client.
type RpcCommon struct {
	timeout     time.Duration // timeout of write to channel.
	SkipCounter int           // counter count the number of skip return when timeout.
}

// CallInfo is info of call to server.
type CallInfo struct {
	id      interface{}              // call id.
	args    []interface{}            // call args.
	chanRet chan *RetInfo            // return channel, such as ci.chanRet <- ri.
	cb      func(interface{}, error) // callback which will pass to RetInfo.
}

// RetInfo is info of return from server.
type RetInfo struct {
	ret interface{}              // value of return.
	err error                    // err of return.
	cb  func(interface{}, error) // callback of return.
}

// Server is a chanrpc server.
type Server struct {
	functions map[interface{}]interface{} // map id(cmd) to handler.
	ChanCall  chan *CallInfo              // channel of CallInfo.
	RpcCommon                             // common variate.
}

// Client is a chanrpc client.
type Client struct {
	chanSyncRet     chan *RetInfo // channel of RetInfo for sync call.
	ChanAsynRet     chan *RetInfo // channel of RetInfo for asyn call.
	pendingAsynCall int           // number of not returned asyn call.
	RpcCommon                     // common variate.
}

// NewServer new a server.
//
// bufsize define buffer size of called channel
// timeout define timeout for writing to channel of clent, prevent from infinite blocking.
func NewServer(bufsize int, timeout time.Duration) *Server {
	s := new(Server)
	s.functions = make(map[interface{}]interface{})
	s.ChanCall = make(chan *CallInfo, bufsize)
	s.timeout = timeout
	return s
}

// Register register handler for id.
func (s *Server) Register(id interface{}, f func([]interface{}) (ret interface{}, err error)) {
	if _, ok := s.functions[id]; ok {
		panic(fmt.Sprintf("function id %v: already registered", id))
	}
	s.functions[id] = f
}

// ret write result to channel provided by ci.
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
		s.SkipCounter++
	}
	return
}

// Exec execute call request from channel.
func (s *Server) Exec(ci *CallInfo) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if conf.LenStackBuf > 0 {
				buf := make([]byte, conf.LenStackBuf)
				l := runtime.Stack(buf, false)
				err = fmt.Errorf("%v: %s", r, buf[:l])
			} else {
				err = fmt.Errorf("%v", r)
			}
			log.Debug("%v", r)
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

// Close do the shutdown of the server.
func (s *Server) Close() {
	close(s.ChanCall)
	for ci := range s.ChanCall {
		err := s.Exec(ci)
		if err != nil {
			log.Error("%v", err)
		}
	}
}

// NewClient create a rpc client.
//
// bufsize define the async buffer size of aysnc callback channel
// timeout define max wait time when send async call request.
func NewClient(size int, timeout time.Duration) *Client {
	c := new(Client)
	c.chanSyncRet = make(chan *RetInfo, 1)
	c.ChanAsynRet = make(chan *RetInfo, size)
	c.timeout = timeout
	return c
}

// call send the request call ci to server channel. If block is true, this will wait forever if channel of server is busy.
func (c *Client) call(s *Server, ci *CallInfo, block bool) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()

	if block {
		s.ChanCall <- ci
	} else {
		select {
		case s.ChanCall <- ci:
		case <-time.After(c.timeout):
			c.SkipCounter++
			c.pendingAsynCall--
			err = errors.New("server is too busy.")
		}
	}
	return
}

// Call do a sync call to attatched server.
func (c *Client) SynCall(s *Server, id interface{}, args ...interface{}) (interface{}, error) {
	_, err := validate(s, id)
	if err != nil {
		return nil, err
	}

	err = c.call(s, &CallInfo{
		id:      id,
		args:    args,
		chanRet: c.chanSyncRet,
	}, true)

	if err != nil {
		return nil, err
	}

	// sync.
	ri := <-c.chanSyncRet
	return ri.ret, ri.err
}

// AsyncCall do a async call to server.
func (c *Client) AsynCall(s *Server, id interface{}, _args ...interface{}) error {
	args := _args[:len(_args)-1]
	cb := _args[len(_args)-1]

	switch cb.(type) {
	case func(ret interface{}, err error):
	default:
		args = _args
		cb = nil
	}

	var _cb func(interface{}, error)
	if cb != nil {
		_cb = cb.(func(interface{}, error))
	}

	c.pendingAsynCall++

	_, err := validate(s, id)
	if err != nil {
		c.ChanAsynRet <- &RetInfo{err: err, cb: _cb}
		return fmt.Errorf("no matching function for asynCall call: %v", id)
	}

	if c.pendingAsynCall > cap(c.ChanAsynRet) {
		c.ChanAsynRet <- &RetInfo{err: errors.New("too many calls of client"), cb: _cb}
		return nil
	}

	err = c.call(s, &CallInfo{
		id:      id,
		args:    args,
		chanRet: c.ChanAsynRet,
		cb:      _cb,
	}, false)

	if err != nil {
		c.ChanAsynRet <- &RetInfo{err: err, cb: _cb}
		return err
	}

	return nil
}

// Cb do exec the callback of ri when the async call is finish handled by server.
func (c *Client) Cb(ri *RetInfo) *RetInfo {
	c.pendingAsynCall--

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
	return ri
}

// Close do close goroutine internal if it exist, and do other cleanup job.
func (c *Client) Close() {
	//The rule of thumb here is that only writers should close channels.
}

// Helper function.

// SynCall do a sync call request to server.
func SynCall(s *Server, id interface{}, args ...interface{}) (interface{}, error) {
	// !!double_check!!
	c := NewClient(0, 1*time.Second)
	return c.SynCall(s, id, args...)
}

// validate if the server can handle the id.
func validate(s *Server, id interface{}) (f interface{}, err error) {
	f = s.functions[id]
	if f == nil {
		err = fmt.Errorf("function id %v: function not registered", id)
		return
	}
	return
}
