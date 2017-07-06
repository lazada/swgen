package swgen

import (
	"reflect"
	"fmt"
)

func ReflectRealTypeName(i interface{}) string {
	v := reflect.ValueOf(i)
	for v.Type().Kind() == reflect.Ptr || v.Type().Kind() == reflect.Interface {
		if v.IsNil() {
			v = reflect.New(v.Type().Elem()).Elem()
		} else {
			v = v.Elem()
		}
	}
	return v.Type().String()
}

func ReflectTypeHash(t reflect.Type) uint32 {
	return uint32(reflect.Indirect(reflect.ValueOf(t)).FieldByName("hash").Uint())
}

func ReflectTypeReliableName(t reflect.Type) string {
	if t.Name() != "" {
		return t.Name()
	}
	return fmt.Sprintf("anon_%08x", ReflectTypeHash(t))
}
