package swgen

type commonName string

const (
	// CommonNameInteger data type is integer, format int32 (signed 32 bits)
	CommonNameInteger commonName = "integer"
	// CommonNameLong data type is integer, format int64 (signed 64 bits)
	CommonNameLong commonName = "long"
	// CommonNameFloat data type is number, format float
	CommonNameFloat commonName = "float"
	// CommonNameDouble data type is number, format double
	CommonNameDouble commonName = "double"
	// CommonNameString data type is string
	CommonNameString commonName = "string"
	// CommonNameByte data type is string, format byte (base64 encoded characters)
	CommonNameByte commonName = "byte"
	// CommonNameBinary data type is string, format binary (any sequence of octets)
	CommonNameBinary commonName = "binary"
	// CommonNameBoolean data type is boolean
	CommonNameBoolean commonName = "boolean"
	// CommonNameDate data type is string, format date (As defined by full-date - RFC3339)
	CommonNameDate commonName = "date"
	// CommonNameDateTime data type is string, format date-time (As defined by date-time - RFC3339)
	CommonNameDateTime commonName = "dateTime"
	// CommonNamePassword data type is string, format password
	CommonNamePassword commonName = "password"
)

type typeFormat struct {
	Type   string
	Format string
}

var commonNamesMap = map[commonName]typeFormat{
	CommonNameInteger:  {"integer", "int32"},
	CommonNameLong:     {"integer", "int64"},
	CommonNameFloat:    {"number", "float"},
	CommonNameDouble:   {"number", "double"},
	CommonNameString:   {"string", ""},
	CommonNameByte:     {"string", "byte"},
	CommonNameBinary:   {"string", "binary"},
	CommonNameBoolean:  {"boolean", ""},
	CommonNameDate:     {"string", "date"},
	CommonNameDateTime: {"string", "date-time"},
	CommonNamePassword: {"string", "password"},
}

func isCommonName(typeName string) (ok bool) {
	_, ok = commonNamesMap[commonName(typeName)]
	return
}

// SchemaFromCommonName create SchemaObj from common name of data types
// supported types: https://github.com/OAI/OpenAPI-Specification/blob/master/versions/2.0.md#data-types
func SchemaFromCommonName(name commonName) SchemaObj {
	data, ok := commonNamesMap[name]
	if ok {
		return SchemaObj{
			Type:   data.Type,
			Format: data.Format,
		}
	}

	return SchemaObj{
		Type: string(name),
	}
}
