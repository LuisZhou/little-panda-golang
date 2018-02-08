package chanrpc_test

import (
	"fmt"
	"github.com/LuisZhou/little-panda-golang/chanrpc"
	"sync"
	"testing"
	//"sync"
)

// func TestPrintSomething(t *testing.T) {
// 	fmt.Println("Say hi")
// 	t.Log("Say bye")
// }

//func Example() {
func TestPrintSomething(t *testing.T) {
	fmt.Println("start test")

	var wg sync.WaitGroup

	s := chanrpc.NewServer(10)
	s.Register("f0", func(args []interface{}) interface{} {
		fmt.Println("f0", len(args))
		return len(args)
	})

	s.Register("add", func(args []interface{}) interface{} {
		n1 := args[0].(int)
		n2 := args[1].(int)
		fmt.Println(n1, n2)
		return n1 + n2
	})

	s.Start()

	l, _ := s.Call("f0", 123)
	fmt.Println(l)

	c := s.Open(10)

	wg.Add(1)

	c.AsynCall("add", 1, 2, func(ret interface{}, err error) {
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(ret)
		}
		wg.Done()
	})

	c.Long()

	wg.Wait()

	// Output:
	// f0 1
	// 1
}
