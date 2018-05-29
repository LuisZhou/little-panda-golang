package timer

import (
	"github.com/LuisZhou/lpge/util"
	"time"
)

// Timer compose sys timer and cb.
type Timer struct {
	t  *time.Timer
	cb func()
}

// Stop can stop sys timer anytime.
func (t *Timer) Stop() {
	t.t.Stop()
	t.cb = nil
}

// Cb execute cb of Timer.
func (t *Timer) Cb() {
	defer func() {
		t.cb = nil
		util.RecoverAndLog()
	}()

	if t.cb != nil {
		t.cb()
	}
}

// Dispatcher make control of the timer, make the callback of the timer serialize execute in the FIFO.
type Dispatcher struct {
	ChanTimer chan *Timer
}

// Create a new time dispatch.
func NewDispatcher(l int) *Dispatcher {
	disp := new(Dispatcher)
	disp.ChanTimer = make(chan *Timer, l)
	return disp
}

// AfterFunc push a Timer in channel after d.
func (disp *Dispatcher) AfterFunc(d time.Duration, cb func()) *Timer {
	t := new(Timer)
	t.cb = cb
	t.t = time.AfterFunc(d, func() {
		disp.ChanTimer <- t
	})
	return t
}

// Cron schedule AfterFunc-job.
type Cron struct {
	t *Timer
}

// Stop timer of Cron.
func (c *Cron) Stop() {
	if c.t != nil {
		c.t.Stop()
	}
}

// Start cron func.
func (disp *Dispatcher) CronFunc(cronExpr *CronExpr, _cb func()) *Cron {
	c := new(Cron)

	now := time.Now()
	nextTime := cronExpr.Next(now)
	if nextTime.IsZero() {
		return c
	}

	// callback
	var cb func()
	cb = func() {
		defer _cb()

		now := time.Now()
		nextTime := cronExpr.Next(now)
		if nextTime.IsZero() {
			return
		}
		c.t = disp.AfterFunc(nextTime.Sub(now), cb)
	}

	c.t = disp.AfterFunc(nextTime.Sub(now), cb)
	return c
}
