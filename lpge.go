package lpge

import (
	"github.com/LuisZhou/lpge/conf"
	"github.com/LuisZhou/lpge/console"
	"github.com/LuisZhou/lpge/log"
	"github.com/LuisZhou/lpge/module"
	"github.com/LuisZhou/lpge/orm"
	"github.com/jinzhu/gorm"
	"os"
	"os/signal"
)

func Migrate(f func(*gorm.DB)) {
	if conf.DBConfig.Default != "" {
		conns := make(map[string]string)
		conns["mysql"] = conf.DBConfig.Mysql
		conns["postgres"] = conf.DBConfig.Postgres
		conns["sqlite"] = conf.DBConfig.Sqlite
		conns["mssql"] = conf.DBConfig.Mssql
		orm.Connect(conns, conf.DBConfig.Default)

		f(orm.Default())
	}
}

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
