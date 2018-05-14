package chanrpc_test

import (
	"fmt"
	"github.com/LuisZhou/lpge/chanrpc"
	"github.com/LuisZhou/lpge/log"
	"strings"
	"sync"
	"testing"
	"time"
)

// todo
// 1. we need bench mark test.

// further reading
// https://github.com/golang/go/wiki/TableDrivenTests

// Wait wait forever for receiving return from async channel using goroutine internal.
func Wait(c *chanrpc.Client, closeSig chan bool) {
	go func() {
		for {
			select {
			case <-closeSig:
				c.Close()
				return
			case ret := <-c.ChanAsynRet:
				c.Cb(ret)
			}
		}
	}()
}

// Start start execute coming call request.
func Start(s *chanrpc.Server, closeSig chan bool) {
	go func() {
		for {
			select {
			case <-closeSig:
				s.Close()
				return
			case ci := <-s.ChanCall:
				log.Debug("%v", ci)
				err := s.Exec(ci)
				if err != nil {
					log.Error("%v", err)
				}
			}
		}
	}()
}

func TestFloodServer(t *testing.T) {
	var wg sync.WaitGroup

	s := chanrpc.NewServer(10, 100)

	wg.Add(1)

	s.Register("add", func(args []interface{}) (ret interface{}, err error) {
		n1 := args[0].(int)
		n2 := args[1].(int)
		return n1 + n2, err
	})

	closesig := make(chan bool)

	Start(s, closesig)
	defer func() {
		closesig <- true
	}()

	c := chanrpc.NewClient(100, 0)
	defer func() {
		closesig <- true
	}()

	counter := 0

	Wait(c, closesig)

	for i := 0; i < 100; i++ {
		c.AsynCall(s, "add", 1, 2, func(ret interface{}, err error) {
			if err != nil {
				t.Log(err)
			} else {
				t.Log(ret)
			}
			counter++
			if counter+c.SkipCounter == 100 {
				wg.Done()
			}
		})
	}

	wg.Wait()
}

func TestFloodClient(t *testing.T) {
	closesig := make(chan bool)

	var wg sync.WaitGroup
	wg.Add(1)

	s := chanrpc.NewServer(1000, 10)
	s.Register("print", func(args []interface{}) (ret interface{}, err error) {
		n1 := args[0].(int)
		return n1, err
	})
	Start(s, closesig)
	defer func() {
		closesig <- true
	}()

	c := chanrpc.NewClient(1, 0)
	defer func() {
		c.Close()
	}()
	c.AllowOverFlood = true

	counter := 0
	go func() {
		for {
			c.Cb(<-c.ChanAsynRet)
			time.Sleep(time.Millisecond * 20)
		}
	}()

	for i := 0; i < 100; i++ {
		c.AsynCall(s, "print", i, func(ret interface{}, err error) {
			if err != nil {
				t.Log(err)
			} else {
				t.Log(ret)
			}
			counter++
			if counter+s.SkipCounter == 100 {
				wg.Done()
			}
		})
	}

	wg.Wait()
}

func TestError(t *testing.T) {
	closesig := make(chan bool)

	var wg sync.WaitGroup

	s := chanrpc.NewServer(10, 1000)
	s.Register("f0", func(args []interface{}) (ret interface{}, err error) {
		return nil, fmt.Errorf("%v", "err 1")
	})
	s.Register("f1", func(args []interface{}) (ret interface{}, err error) {
		panic("err 2")
		return nil, nil
	})
	Start(s, closesig)
	defer func() {
		closesig <- true
	}()

	_, err := chanrpc.SynCall(s, "f0", 123)
	if strings.Compare(err.(error).Error(), "err 1") != 0 {
		t.Error("err test fail")
	}

	c := chanrpc.NewClient(10, time.Millisecond*50) //s.Open(10, time.Millisecond*50)
	defer func() {
		closesig <- true
	}()

	wg.Add(1)

	Wait(c, closesig)

	c.AsynCall(s, "f1", 1, 2, func(ret interface{}, err error) {
		if strings.Compare(err.(error).Error(), "err 2") != 0 {
			t.Error("err test fail")
		}
		wg.Done()
	})

	wg.Wait()
}

func Example() {
	closesig := make(chan bool)

	var wg sync.WaitGroup

	s := chanrpc.NewServer(10, 1000)
	s.Register("f0", func(args []interface{}) (ret interface{}, err error) {
		fmt.Println("f0", len(args))
		return len(args), err
	})
	s.Register("add", func(args []interface{}) (ret interface{}, err error) {
		n1 := args[0].(int)
		n2 := args[1].(int)
		return n1 + n2, err
	})
	Start(s, closesig)
	defer func() {
		closesig <- true
	}()

	// 1. Example: sync call.

	l, _ := chanrpc.SynCall(s, "f0", 123)
	fmt.Println(l)

	// 2. Example: async call

	c := chanrpc.NewClient(10, time.Millisecond*50) //s.Open(10, time.Millisecond*50)
	defer func() {
		c.Close()
	}()

	wg.Add(1)

	// Wait wait for async callback.
	Wait(c, closesig)

	c.AsynCall(s, "add", 1, 2, func(ret interface{}, err error) {
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(ret)
		}
		wg.Done()
	})

	// 3. Example: async call, but do not wait for the callback.
	c.AsynCall(s, "add", 1, 2, func(ret interface{}, err error) {
		// leave empty
	})

	wg.Wait()

	// Output:
	// f0 1
	// 1
	// 3
}
