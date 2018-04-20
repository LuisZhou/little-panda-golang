package db

import (
	"github.com/LuisZhou/lpge/db"
	"testing"
)

func TestDb(t *testing.T) {
	db.DsnConf["mysql"] = "gls:glsDbCenterBG@(192.168.163.133)/gls?charset=utf8"
	db.Create()

	c, err := db.GetDefaultConnection()
	t.Log(c, err)
}
