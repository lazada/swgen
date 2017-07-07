package swgen

import (
	"testing"
	"reflect"
)

type TestStruct1 struct {
	Id   uint
	Name string
}

type TestStruct2 struct {
	Id   uint
	Name string
}

func TestReflectTypeHash(t *testing.T) {
	var (
		ts1a, ts1b TestStruct1
		ts2        TestStruct2

		anon1a, anon1b struct {
			Id   uint
			Name string
		}

		anon2 = struct {
			Id   uint
			Name string
		}{}
	)

	if reflect.TypeOf(ts1a) != reflect.TypeOf(ts1b) {
		t.Error("Different reflect.Type on instances of the same named struct")
	}

	if ReflectTypeHash(reflect.TypeOf(ts1a)) == ReflectTypeHash(reflect.TypeOf(ts2)) {
		t.Error("Same reflect.Type on instances of different named structs:", ReflectTypeHash(reflect.TypeOf(ts1a)))
	}

	if reflect.TypeOf(anon1a) != reflect.TypeOf(anon1b) {
		t.Error("Different reflect.Type on instances of the same anonymous struct")
	}

	if ReflectTypeHash(reflect.TypeOf(anon1a)) != ReflectTypeHash(reflect.TypeOf(anon2)) {
		t.Error("Different reflect.Type on instances of the different anonymous structs with same fields")
	}
}
