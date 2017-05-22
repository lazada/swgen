package swgen

import (
	"testing"
)

func TestDefinition(t *testing.T) {
	var obj IDefinition
	obj = Definition{
		TypeName:  "MyName",
		SchemaObj: SchemaObj{Type: "integer", Format: "int64"},
	}

	typeName, _, err := obj.SwgenDefinition()
	assertTrue(typeName == "MyName", t)
	assertTrue(err == nil, t)
}
