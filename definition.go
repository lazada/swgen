package swgen

// Definition is a helper that implements interface IDefinition
type Definition struct {
	SchemaObj
	TypeName string
}

// SwgenDefinition return type name and definition that was set
func (s Definition) SwgenDefinition() (typeName string, typeDef SchemaObj, err error) {
	typeName = s.TypeName
	typeDef = s.SchemaObj
	return
}
