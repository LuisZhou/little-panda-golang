package orm

import (
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"strings"
	// _ "github.com/jinzhu/gorm/dialects/postgres"
	// _ "github.com/jinzhu/gorm/dialects/sqlite"
	// _ "github.com/jinzhu/gorm/dialects/mssql"
	"github.com/jinzhu/gorm"
)

var (
	defaultConn *gorm.DB
	connections map[string]*gorm.DB
)

func Connect(info map[string]string, _default string) {
	connections = make(map[string]*gorm.DB)

	for k, v := range info {
		if v != "" {
			db, err := gorm.Open(k, v)
			if err != nil {
				panic("failed to connect database")
			}
			connections[k] = db
		}
	}

	defaultConn = connections[strings.ToLower(_default)]
}

func Default() *gorm.DB {
	return defaultConn
}

func Get(k string) *gorm.DB {
	return connections[k]
}

func Close() {
	if connections != nil {
		for _, v := range connections {
			v.Close()
		}
	}
}
