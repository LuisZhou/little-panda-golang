package console

import (
	"github.com/LuisZhou/lpge/conf"
	"github.com/LuisZhou/lpge/console"
	"sync"
	"testing"
)

func TestConsole(t *testing.T) {
	var wg sync.WaitGroup

	wg.Add(1)

	conf.ConsolePort = 6001
	console.Init()

	wg.Wait()
}
