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

const expectedStructName string = "swgen.TestStruct1"

func testReflectRealTypeName(t *testing.T) {
	var (
		t1 TestStruct1
		t2 *TestStruct1
		t3 interface{} = t1
		t4 interface{} = t2
		t5             = &t3
		t6             = &t4
	)

	test := func(i interface{}, testName string) {
		var typeName string
		defer func() {
			if r := recover(); r != nil {
				t.Log("Panic on", testName, r)
			}
		}()
		if typeName = ReflectRealTypeName(i); typeName != expectedStructName {
			t.Errorf("Test failed on %s: expected '%s', got '%s'.", testName, expectedStructName, typeName)
		} else {
			t.Logf("Test %s ok: returned '%s', Type.Name() = '%s', Type.String() = '%s'",
				testName, typeName, reflect.TypeOf(i).Name(), reflect.TypeOf(i).String())
		}
	}

	test(t1, "t1")
	test(t2, "t2")
	test(t3, "t3")
	test(t4, "t4")
	test(t5, "t5")
	test(t6, "t6")
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
