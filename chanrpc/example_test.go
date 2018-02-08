package chanrpc_test

import (
	"fmt"
	"github.com/LuisZhou/little-panda-golang/chanrpc"
	//"sync"
)

func Example() {
	s := chanrpc.NewServer(10)
	s.Register("f0", func(args ...interface{}) interface{} {
		fmt.Println("f0", len(args))
		return len(args)
	})

	s.Start()

	l, _ := s.Call("f0", 123)
	fmt.Println(l)

	// Output:
	// f0 1
	// 1
}
