package chanrpc

import (
	"errors"
	"fmt"
	"github.com/LuisZhou/little-panda-golang/conf"
	"github.com/LuisZhou/little-panda-golang/log"
	"runtime"
)

type CallBack func(ret interface{}, err error)

// one server per goroutine (goroutine not safe)
// one client per goroutine (goroutine not safe)
type Server struct {
	// id -> function
	// function fomat:
	// func(args ... interface{}) interface{}
	functions    map[interface{}]interface{}
	ChanCall     chan *CallInfo
	closeSig     chan bool
	isUseRoutine bool
}

// wrapper of call info for server channel.
// do not use the cb here, what to callback is decided by the agent, or parent.
type CallInfo struct {
	id      interface{}   // id to map function
	args    []interface{} // args to call function
	chanRet chan *RetInfo // channel for return value for caller or agent of caller
	cb      CallBack
}

// wrapper of return info for caller or agent of caller.
type RetInfo struct {
	ret interface{}
	err error
	cb  CallBack
}

// Client struct.
type Client struct {
	s               *Server
	chanSyncRet     chan *RetInfo
	ChanAsynRet     chan *RetInfo // If the caller care the return of async call, just use a goroutine to wait the result.
	pendingAsynCall int
}

func NewServer(bufsize int) *Server {
	s := new(Server)
	s.functions = make(map[interface{}]interface{})
	s.ChanCall = make(chan *CallInfo, bufsize)
	s.closeSig = make(chan bool)
	return s
}

// you must call the function before calling Open and Go
func (s *Server) Register(id interface{}, f interface{}) {
	switch f.(type) {
	case func([]interface{}) interface{}:
	default:
		panic(fmt.Sprintf("function id %v: definition of function is invalid", id))
	}

	if _, ok := s.functions[id]; ok {
		panic(fmt.Sprintf("function id %v: already registered", id))
	}

	s.functions[id] = f
}

// can panic every where here.
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

	// todo: this is danger: if the client is block, for the reason of exhausted channel buffer.
	// The server will block too. Use select here.
	// selection:
	// 1. just drop the ret if no channel buffer.
	// 2. use large buffer.
	// 3. wait something, drop the ret if timeout.
	ci.chanRet <- ri
	return
}

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

	// todo:f return err ret or just use panic?
	ret := f.(func([]interface{}) interface{})(ci.args)
	return s.ret(ci, &RetInfo{ret: ret})

	// panic("bug")
}

// Why we need use a parameter for the CallInfo, not just use s.ChanCall.
// Because in the external world, we should use s.ChanCall in "select", if it is available, then Exec it.
// For in one Module, we may want there is single goroutine, for "thread-sare"

// I want to deprecate it.
func (s *Server) Exec(ci *CallInfo) {
	err := s.exec(ci)
	if err != nil {
		log.Error("%v", err)
	}
}

func (s *Server) Start() {
	// is this safe?
	s.isUseRoutine = true
	go func() {
		for {
			// add close signal channel.
			// select use the first or random.
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
		// log: server exit.
	}()
}

// let the server to do something, but no pending job for the ret.
// goroutine safe
func (s *Server) Go(id interface{}, args ...interface{}) {
	// check here is also correct.
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

// goroutine safe
func (s *Server) Call(id interface{}, args ...interface{}) (interface{}, error) {
	return s.Open(0).Call(id, args...)
}

func (s *Server) close() {
	// what's that side effect?
	close(s.ChanCall)

	// move this to Start() if goroutine in Start() get the close signal.
	for ci := range s.ChanCall {
		s.ret(ci, &RetInfo{
			err: errors.New("chanrpc server closed"),
		})
	}
}

// todo: is this ok?
func (s *Server) Close() {
	if s.isUseRoutine {
		s.closeSig <- true
	} else {
		s.close()
	}
}

// goroutine safe
func (s *Server) Open(l int) *Client {
	c := NewClient(l)
	c.Attach(s)
	return c
}

func NewClient(size int) *Client {
	c := new(Client)
	c.chanSyncRet = make(chan *RetInfo, 1)
	c.ChanAsynRet = make(chan *RetInfo, size)
	return c
}

func (c *Client) Attach(s *Server) {
	c.s = s
}

func (c *Client) call(ci *CallInfo, block bool) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()

	if block {
		c.s.ChanCall <- ci
	} else {
		// if full, just drop it.
		select {
		case c.s.ChanCall <- ci:
		default:
			err = errors.New("chanrpc channel full")
		}
	}
	return
}

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

func (c *Client) asynCall(id interface{}, args []interface{}, cb CallBack) {
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

func (c *Client) AsynCall(id interface{}, _args ...interface{}) {
	if len(_args) < 1 {
		panic("callback function not found")
	}

	args := _args[:len(_args)-1]
	cb := _args[len(_args)-1]

	var cb_cast CallBack

	if cb != nil {
		cb_tmp, ok := cb.(func(ret interface{}, err error)) // can not use Callback here!
		fmt.Println("cast", cb_tmp, ok)

		if ok == false {
			cb_cast = nil
		} else {
			cb_cast = cb_tmp
		}
	} else {
		cb_cast = nil
	}

	// too many calls
	if c.pendingAsynCall >= cap(c.ChanAsynRet) {
		execCb(&RetInfo{err: errors.New("too many calls"), cb: cb_cast})
		return
	}

	c.asynCall(id, args, cb_cast)
	c.pendingAsynCall++
}

func execCb(ri *RetInfo) {
	fmt.Println("execCb", ri.cb)

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

func (c *Client) Cb(ri *RetInfo) {
	c.pendingAsynCall--
	execCb(ri)
}

func (c *Client) Close() {
	for c.pendingAsynCall > 0 {
		c.Cb(<-c.ChanAsynRet)
	}
}

func (c *Client) Idle() bool {
	return c.pendingAsynCall == 0
}

func (c *Client) Long() {
	go func() {
		for {
			fmt.Println("Long wait")
			c.Cb(<-c.ChanAsynRet)
		}
	}()
}
