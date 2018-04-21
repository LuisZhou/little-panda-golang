package util

import (
	"github.com/LuisZhou/lpge/util"
	"github.com/jinzhu/inflection"
	"testing"
)

func TestCamelCaseToUnderscore(t *testing.T) {
	r1 := util.CamelCaseToUnderscore("MyName")
	t.Log(r1)
	r1 = inflection.Plural(r1)
	t.Log(r1)

	r1 = util.CamelCaseToUnderscore("TestMy")
	t.Log(r1)
	r1 = inflection.Plural(r1)
	t.Log(r1)
}
