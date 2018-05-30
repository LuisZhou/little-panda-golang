package lpge

import (
	"github.com/LuisZhou/lpge/conf"
	"github.com/LuisZhou/lpge/console"
	"github.com/LuisZhou/lpge/log"
	"github.com/LuisZhou/lpge/module"
	"os"
	"os/signal"
)

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

	log.Release("LPGE %v starting up", 1.0)

	// module
	for k, v := range mods {
		module.Register(v, k)
	}

	// cluster
	//cluster.Init()

	// console
	console.Init()

	// close
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	sig := <-c
	log.Release("LPGE closing down (signal: %v)", sig)
	console.Destroy()
	//cluster.Destroy()
	module.Destroy()
}
