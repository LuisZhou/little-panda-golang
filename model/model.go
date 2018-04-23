package model

import (
	"fmt"
	"github.com/LuisZhou/lpge/util"
	"github.com/jinzhu/inflection"
	"reflect"
	"strings"
)

// model interface

type Model struct {
	__table string
}

// model api

type ModelInterface interface {
}

func Save(m interface{}) {
	var v reflect.Value

	t := reflect.TypeOf(m)
	if t.Kind() != reflect.Ptr {
		v = reflect.ValueOf(m)
	} else {
		v = reflect.ValueOf(m).Elem()
		fmt.Println(v, v.NumField())
	}

	for i := 0; i < v.NumField(); i++ {
		value := v.Field(i).Interface()
		key := v.Type().Field(i).Name
		fmt.Println(key, ':', value)
	}
}

func SetTableName(m *Model, table_name string) {
	m.__table = table_name
}

func GetTableName(m interface{}) string {
	msgType := reflect.TypeOf(m)
	var name string
	if msgType.Kind() != reflect.Ptr {
		name = msgType.String()
	} else {
		name = msgType.Elem().String()
	}
	arr := strings.Split(name, ".")
	name = arr[len(arr)-1]
	return inflection.Plural(util.CamelCaseToUnderscore(name))
}

// common api
// where

type whereS struct {
	left  interface{}
	op    interface{}
	right interface{}
}

type query struct {
	where whereS
}

func newQuery() {

}
