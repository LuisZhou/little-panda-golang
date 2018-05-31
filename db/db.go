// Package db create connection of different database.
package db

import (
	"database/sql"
	"errors"
	_ "github.com/go-sql-driver/mysql"
)

type Connection struct {
	driver         string
	dataSourceName string
	conn           *sql.DB
}

var (
	Connections map[string]Connection
	Default     string
	DsnConf     map[string]string
)

func init() {
	Default = "mysql"
	DsnConf = make(map[string]string)
}

func Create() {
	Connections = map[string]Connection{
		"mysql": Connection{
			driver:         "mysql",
			dataSourceName: DsnConf["mysql"],
		},
	}

	for k, v := range Connections {
		if v.dataSourceName {
			var err error
			v.conn, err = sql.Open(v.driver, v.dataSourceName)
			if err != nil {
				panic(err.Error())
			}

			if err = v.conn.Ping(); err != nil {
				panic(err.Error())
			}

			Connections[k] = v
		}
	}
}

func GetDefaultConnection() (db *sql.DB, err error) {
	c, ok := Connections[Default]
	if !ok {
		return nil, errors.New("connection for default is empty")
	}
	return c.conn, nil
}
