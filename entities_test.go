package swgen

import (
	"testing"
)

func TestPathItemHasMethod(t *testing.T) {
	item := PathItem{}
	item.Get = &OperationObj{}

	assertTrue(item.HasMethod("GET"), t)
	assertFalse(item.HasMethod("POST"), t)
	assertFalse(item.HasMethod("PUT"), t)
	assertFalse(item.HasMethod("HEAD"), t)
	assertFalse(item.HasMethod("DELETE"), t)
	assertFalse(item.HasMethod("OPTIONS"), t)
	assertFalse(item.HasMethod("PATCH"), t)
	assertFalse(item.HasMethod(""), t)
}

func assertTrue(v bool, t *testing.T) {
	if v != true {
		t.Fatalf("value must return true")
	}
}

func assertFalse(v bool, t *testing.T) {
	if v != false {
		t.Fatalf("value must return false")
	}
}
