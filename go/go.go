/* Package g provide linear execute or callback of operation. */
package g

import (
	"container/list"
	"fmt"
	"github.com/LuisZhou/lpge/conf"
	"github.com/LuisZhou/lpge/log"
	"runtime"
	"sync"
)

// Go is a type that provides linear execute call of function.
type Go struct {
	ChanCb    chan func()
	pendingGo int
}

// New create a go instance.
func New(l int) *Go {
	g := new(Go)
	g.ChanCb = make(chan func(), l)
	return g
}

// Go to execute f, and linear call cb. No routine-protect for execute of f. each f execute in seperate goroutine.
func (g *Go) Go(f func(), cb func()) {
	g.pendingGo++

	go func() {
		defer func() {
			g.ChanCb <- cb
			if r := recover(); r != nil {
				var err error
				if conf.LenStackBuf > 0 {
					buf := make([]byte, conf.LenStackBuf)
					l := runtime.Stack(buf, false)
					err = fmt.Errorf("%v: %s", r, buf[:l])
				} else {
					err = fmt.Errorf("%v", r)
				}
				log.Error(err.Error())
			}
		}()

		f()
	}()
}

// Cb execute cb.
func (g *Go) Cb(cb func()) {
	defer func() {
		g.pendingGo--
		if r := recover(); r != nil {
			var err error
			if conf.LenStackBuf > 0 {
				buf := make([]byte, conf.LenStackBuf)
				l := runtime.Stack(buf, false)
				err = fmt.Errorf("%v: %s", r, buf[:l])
			} else {
				err = fmt.Errorf("%v", r)
			}
			log.Error(err.Error())
		}
	}()

	if cb != nil {
		cb()
	}
}

// Close execute all pending cb in g.
func (g *Go) Close() {
	for g.pendingGo > 0 {
		g.Cb(<-g.ChanCb)
	}
}

// Idle return if g is idle.
func (g *Go) Idle() bool {
	return g.pendingGo == 0
}

// LinearGo is struct composes f and cb.
type LinearGo struct {
	f  func()
	cb func()
}

// LinearContext is a type that provides linear execute linearGo.
type LinearContext struct {
	g              *Go
	linearGo       *list.List
	mutexLinearGo  sync.Mutex
	mutexExecution sync.Mutex
}

// NewLinearContext create a new LinearContext.
func (g *Go) NewLinearContext() *LinearContext {
	c := new(LinearContext)
	c.g = g
	c.linearGo = list.New()
	return c
}

// Go start execute linearGo in list of LinearContext.
func (c *LinearContext) Go(f func(), cb func()) {
	c.g.pendingGo++

	c.mutexLinearGo.Lock()
	c.linearGo.PushBack(&LinearGo{f: f, cb: cb})
	c.mutexLinearGo.Unlock()

	go func() {
		c.mutexExecution.Lock()
		defer c.mutexExecution.Unlock()

		c.mutexLinearGo.Lock()
		e := c.linearGo.Remove(c.linearGo.Front()).(*LinearGo)
		c.mutexLinearGo.Unlock()

		defer func() {
			c.g.ChanCb <- e.cb
			if r := recover(); r != nil {
				var err error
				if conf.LenStackBuf > 0 {
					buf := make([]byte, conf.LenStackBuf)
					l := runtime.Stack(buf, false)
					err = fmt.Errorf("%v: %s", r, buf[:l])
				} else {
					err = fmt.Errorf("%v", r)
				}
				log.Error(err.Error())
			}
		}()

		e.f()
	}()
}
