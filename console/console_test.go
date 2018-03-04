package console

import (
	"github.com/LuisZhou/lpge/conf"
	"github.com/LuisZhou/lpge/console"
	"testing"
)

func TestConsole(t *testing.T) {
	conf.ConsolePort = 6001
	console.Init()
}
