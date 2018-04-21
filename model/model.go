package model

import (
	"fmt"
	"reflect"
)

// model interface

type Model struct {
	__table string
}

// model api

func (m *Model) Save() {
	fmt.Println("Save")
}

func (m *Model) GetTableName() string {

	fmt.Println(reflect.TypeOf(m))

	msgType := reflect.TypeOf(m)
	var name string
	if msgType.Kind() != reflect.Ptr {
		name = msgType.Kind().String()
	} else {
		name = msgType.Elem().String()
	}

	return name
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
