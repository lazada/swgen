package swgen

import "testing"

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

func TestAdditionalDataJSONMarshal(t *testing.T) {
	// empty object
	obj := additionalData{}
	_, err := obj.marshalJSONWithStruct(nil)
	assertTrue(err == nil, t)

	obj.AddExtendedField("x-custom-field", 1)
	data, err := obj.marshalJSONWithStruct(struct{}{})
	assertTrue(err == nil, t)
	assertTrue(string(data) == `{"x-custom-field":1}`, t)
}

func assertTrue(v bool, t *testing.T) {
	if v != true {
		t.Fatal("value must return true")
	}
}

func assertFalse(v bool, t *testing.T) {
	if v != false {
		t.Fatal("value must return false")
	}
}
