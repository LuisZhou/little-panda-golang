package model

import (
	"github.com/LuisZhou/lpge/model"
	"testing"
)

type ModelTest struct {
	model.Model
}

func TestModel(t *testing.T) {
	m := &ModelTest{}
	m.Save()
	t.Log(m.GetTableName())
}
