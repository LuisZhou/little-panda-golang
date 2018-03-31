package main

import (
	"fmt"
	"github.com/LuisZhou/lpge/conf"
	"github.com/LuisZhou/lpge/gate"
	"github.com/LuisZhou/lpge/log"
	"github.com/LuisZhou/lpge/module"
	"os"
	"os/signal"
	"time"
)

//func Run(mods ...module.Module) {
func Run(mods map[string]module.Module) {
	// logger
	if conf.LogLevel != "" {
		logger, err := log.New(conf.LogLevel, conf.LogPath, conf.LogFlag)
		if err != nil {
			panic(err)
		}
		log.Export(logger)
		defer logger.Close()
	}

	log.Release("Leaf %v starting up", 1.0)

	// module
	for k, v := range mods {
		module.Register(v, k)
	}
	//module.Init()

	// cluster
	//cluster.Init()

	// console
	//console.Init()

	// close
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	sig := <-c
	log.Release("Leaf closing down (signal: %v)", sig)
	//console.Destroy()
	//cluster.Destroy()
	module.Destroy()
}

func main() {
	fmt.Println(conf.LenStackBuf)
	_gate := &gate.Gate{
		MaxConnNum:      2,
		PendingWriteNum: 20,
		MaxMsgLen:       4096,
		HTTPTimeout:     10 * time.Second,
		TCPAddr:         "127.0.0.1:3563",
		LittleEndian:    true,
	}

	Run(map[string]module.Module{
		"gate": _gate,
	})
}
