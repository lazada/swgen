package swgen

import (
	"fmt"
	"reflect"
)

// ReflectTypeHash returns private (unexported) `hash` field of the Golang internal reflect.rtype struct for a given reflect.Type
// This hash is used to (quasi-)uniquely identify a reflect.Type value
func ReflectTypeHash(t reflect.Type) uint32 {
	return uint32(reflect.Indirect(reflect.ValueOf(t)).FieldByName("hash").Uint())
}

// ReflectTypeReliableName returns real name of given reflect.Type, if it is non-empty, or auto-generates "anon_*"]
// name for anonymous structs
func ReflectTypeReliableName(t reflect.Type) string {
	if t.Name() != "" {
		return t.Name()
	}
	return fmt.Sprintf("anon_%08x", ReflectTypeHash(t))
}
