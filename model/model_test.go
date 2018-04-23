package model

import (
	"github.com/LuisZhou/lpge/model"
	"testing"
)

type ModelTest struct {
	model.Model
	a string
	b int
}

func TestModel(t *testing.T) {
	m := &ModelTest{}
	model.Save(m)
	t.Log(model.GetTableName(m))
	t.Log(model.GetTableName(ModelTest{}))
}
