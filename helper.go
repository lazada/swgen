package swgen

import (
	"reflect"
	"fmt"
)

func ReflectTypeHash(t reflect.Type) uint32 {
	return uint32(reflect.Indirect(reflect.ValueOf(t)).FieldByName("hash").Uint())
}

func ReflectTypeReliableName(t reflect.Type) string {
	if t.Name() != "" {
		return t.Name()
	}
	return fmt.Sprintf("anon_%08x", ReflectTypeHash(t))
}
