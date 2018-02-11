package chanrpc_test

import (
	"fmt"
	"github.com/LuisZhou/little-panda-golang/chanrpc"
	"strings"
	"sync"
	"testing"
	"time"
)

// todo
// 1. we need bench mark test.

// further reading
// https://github.com/golang/go/wiki/TableDrivenTests

func TestFloodServer(t *testing.T) {

	var wg sync.WaitGroup

	s := chanrpc.NewServer(10, 100)

	wg.Add(1)

	s.Register("add", func(args []interface{}) (ret interface{}, err error) {
		n1 := args[0].(int)
		n2 := args[1].(int)
		return n1 + n2, err
	})

	s.Start()

	c := s.Open(100, time.Millisecond*50)

	counter := 0

	c.Long()

	for i := 0; i < 100; i++ {
		c.AsynCall("add", 1, 2, func(ret interface{}, err error) {
			if err != nil {
				t.Log(err)
			} else {
				t.Log(ret)
			}

			counter++

			if counter == 100 {
				wg.Done()
			}
		})
	}

	wg.Wait()
}

func TestFloodClient(t *testing.T) {
	var wg sync.WaitGroup
	s := chanrpc.NewServer(1000, 50)

	wg.Add(1)

	s.Register("print", func(args []interface{}) (ret interface{}, err error) {
		n1 := args[0].(int)
		//n2 := args[1].(int)
		return n1, err
	})
	s.Start()

	c := s.Open(1, time.Millisecond*50)
	c.AllowOverFlood = true

	counter := 0

	go func() {
		for {
			c.Cb(<-c.ChanAsynRet)
			time.Sleep(time.Millisecond * 50)
		}
	}()

	for i := 0; i < 100; i++ {
		c.AsynCall("print", i, func(ret interface{}, err error) {
			if err != nil {
				t.Log(err)
			} else {
				t.Log(ret)
			}
			counter++
			if counter+s.SkipCounter() == 100 {
				wg.Done()
			}
		})
	}
	wg.Wait()

	s.Close()

	// Do not call c.Long() before, so do not need to Close().
	// c.Close()
}

func TestError(t *testing.T) {

	var wg sync.WaitGroup

	s := chanrpc.NewServer(10, 1000)
	// register handler
	s.Register("f0", func(args []interface{}) (ret interface{}, err error) {
		return nil, fmt.Errorf("%v", "err 1")
	})
	s.Register("f1", func(args []interface{}) (ret interface{}, err error) {
		panic("err 2")
		return nil, nil
	})

	s.Start()

	_, err := s.Call("f0", 123)
	if strings.Compare(err.(error).Error(), "err 1") != 0 {
		t.Error("err test fail")
	}

	//wg.Add(1)

	c := s.Open(10, time.Millisecond*50)

	wg.Add(1)

	// Long wait for async callback.
	c.Long()

	c.AsynCall("f1", 1, 2, func(ret interface{}, err error) {
		if strings.Compare(err.(error).Error(), "err 2") != 0 {
			t.Error("err test fail")
		}
		wg.Done()
	})

	wg.Wait()

	c.Close()
	s.Close()
}

func Example() {
	var wg sync.WaitGroup

	s := chanrpc.NewServer(10, 1000)
	// register handler
	s.Register("f0", func(args []interface{}) (ret interface{}, err error) {
		fmt.Println("f0", len(args))
		return len(args), err
	})
	// register handler
	s.Register("add", func(args []interface{}) (ret interface{}, err error) {
		n1 := args[0].(int)
		n2 := args[1].(int)
		return n1 + n2, err
	})

	s.Start()

	// 1. Example: sync call.

	l, _ := s.Call("f0", 123)
	fmt.Println(l)

	// 2. Example: async call

	c := s.Open(10, time.Millisecond*50)

	wg.Add(1)

	// Long wait for async callback.
	c.Long()

	c.AsynCall("add", 1, 2, func(ret interface{}, err error) {
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(ret)
		}
		wg.Done()
	})

	// 3. Example: async call, but do not wait for the callback.
	c.AsynCall("add", 1, 2, func(ret interface{}, err error) {
		// leave empty
	})

	wg.Wait()

	c.Close()
	s.Close()

	// Output:
	// f0 1
	// 1
	// 3
}
