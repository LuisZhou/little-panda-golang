package main

import (
	"fmt"
	"github.com/LuisZhou/lpge/conf"
	"github.com/LuisZhou/lpge/log"
	"github.com/LuisZhou/lpge/module"
	"os"
	"os/signal"
)

func Run(mods ...module.Module) {
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
	for i := 0; i < len(mods); i++ {
		module.Register(mods[i], "")
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
	Run()
}
