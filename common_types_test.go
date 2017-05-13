package swgen

import "testing"

func TestSchemaFromCommonName(t *testing.T) {
	so := SchemaFromCommonName(CommonNameInteger)
	assertTrue(so.Type == "integer", t)
	assertTrue(so.Format == "int32", t)

	so = SchemaFromCommonName("file")
	assertTrue(so.Type == "file", t)
	assertTrue(so.Format == "", t)
}
