package chanrpc_test

import (
	"fmt"
	"github.com/LuisZhou/little-panda-golang/chanrpc"
	"sync"
	"testing"
	"time"
)

// todo
// 1. we need bench mark test.

// further reading
// https://github.com/golang/go/wiki/TableDrivenTests

// func TestFloodServer(t *testing.T) {

// 	var wg sync.WaitGroup

// 	s := chanrpc.NewServer(10, 100)

// 	wg.Add(1)

// 	s.Register("add", func(args []interface{}) (ret interface{}, err error) {
// 		n1 := args[0].(int)
// 		n2 := args[1].(int)
// 		return n1 + n2, err
// 	})

// 	s.Start()

// 	c := s.Open(100)

// 	counter := 0

// 	c.Long()

// 	for i := 0; i < 100; i++ {
// 		c.AsynCall("add", 1, 2, func(ret interface{}, err error) {
// 			if err != nil {
// 				t.Log(err)
// 			} else {
// 				t.Log(ret)
// 			}

// 			counter++

// 			if counter == 100 {
// 				wg.Done()
// 			}
// 		})
// 	}

// 	wg.Wait()
// }

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

	c := s.Open(1)

	counter := 0

	go func() {
		for {
			c.Cb(<-c.ChanAsynRet)
			time.Sleep(time.Millisecond * 50)
		}
	}()

	// var closeSig chan bool = make(chan bool)

	for i := 0; i < 100; i++ {
		c.AsynCall("print", i, func(ret interface{}, err error) {
			if err != nil {
				t.Log(err)
			} else {
				t.Log(ret)
			}

			counter++

			fmt.Println("what?", counter)
			if counter > 50 {
				wg.Done()
			}
		})
	}

	//c.Long()

	// go func() {
	// 	for {
	// 		a := <-closeSig
	// 		_ = a
	// 		break
	// 	}
	// }()

	wg.Wait()
}

// func Example() {
// 	var wg sync.WaitGroup

// 	s := chanrpc.NewServer(10, 1000)
// 	// register handler
// 	s.Register("f0", func(args []interface{}) (ret interface{}, err error) {
// 		fmt.Println("f0", len(args))
// 		return len(args), err
// 	})
// 	// register handler
// 	s.Register("add", func(args []interface{}) (ret interface{}, err error) {
// 		n1 := args[0].(int)
// 		n2 := args[1].(int)
// 		return n1 + n2, err
// 	})

// 	s.Start()

// 	// Example: sync call.

// 	l, _ := s.Call("f0", 123)
// 	fmt.Println(l)

// 	// Example: async call

// 	c := s.Open(10)

// 	wg.Add(1)

// 	// Long wait for async callback.
// 	c.Long()

// 	c.AsynCall("add", 1, 2, func(ret interface{}, err error) {
// 		if err != nil {
// 			fmt.Println(err)
// 		} else {
// 			fmt.Println(ret)
// 		}
// 		wg.Done()
// 	})

// 	wg.Wait()

// 	c.Close()
// 	s.Close()

// 	// Output:
// 	// f0 1
// 	// 1
// 	// 3
// }
