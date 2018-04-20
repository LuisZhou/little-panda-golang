package db

import (
	"fmt"
	"reflect"
	"testing"
)

func TestModel(t *testing.T) {
	x := struct {
		Foo string
		Bar int
	}{"foo", 2}

	v := reflect.ValueOf(x)

	values := make([]interface{}, v.NumField())

	for i := 0; i < v.NumField(); i++ {
		//values[i] = v.Field(i).Interface()
		values[i] = v.Type().Field(i).Name
	}

	fmt.Println(values)
}
