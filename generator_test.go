package swgen

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/kr/pretty"
	"github.com/lazada/swgen/sample"
)

type TestSampleStruct struct {
	SimpleString string `json:"simple_string"`
	SimpleInt    int    `json:"simple_int"`
}

type testEmptyStruct struct{}

type testSimpleStruct struct {
	SimpleString  string  `json:"simple_string" schema:"simple_string"`
	SimpleInt     int     `json:"simple_int" schema:"simple_int"`
	SimpleInt32   int32   `json:"simple_int32" schema:"simple_int32"`
	SimpleInt64   int64   `json:"simple_int64" schema:"simple_int64"`
	SimpleUInt32  uint32  `json:"simple_uint32" schema:"simple_uint32"`
	SimpleUInt64  uint64  `json:"simple_uint64" schema:"simple_uint64"`
	SimpleFloat32 float32 `json:"simple_float32" schema:"simple_float32"`
	SimpleFloat64 float64 `json:"simple_float64" schema:"simple_float64"`
	SimpleBool    bool    `json:"simple_bool" schema:"simple_bool"`
	IgnoreField   string  `json:"-" schema:"-"`
}

type testSimpleSlices struct {
	ListString  []string  `json:"list_string"`
	ListInt     []int     `json:"list_int"`
	ListInt32   []int32   `json:"list_int32"`
	ListInt64   []int64   `json:"list_int64"`
	ListUInt32  []uint32  `json:"list_uint32"`
	ListUInt64  []uint64  `json:"list_uint64"`
	ListFloat32 []float32 `json:"list_float32"`
	ListFloat64 []float64 `json:"list_float64"`
	ListBool    []bool    `json:"list_bool"`
}

type testSimpleMaps struct {
	MapString  map[string]string  `json:"map_string"`
	MapInt     map[string]int     `json:"map_int"`
	MapInt32   map[string]int32   `json:"map_int32"`
	MapInt64   map[string]int64   `json:"map_int64"`
	MapUInt32  map[string]uint32  `json:"map_uint32"`
	MapUInt64  map[string]uint64  `json:"map_uint64"`
	MapFloat32 map[string]float32 `json:"map_float32"`
	MapFloat64 map[string]float64 `json:"map_float64"`
	MapBool    map[string]bool    `json:"map_bool"`
}

type testSimpleMapList struct {
	MapListString  []map[string]string  `json:"map_list_string"`
	MapListInt     []map[string]int     `json:"map_list_int"`
	MapListInt32   []map[string]int32   `json:"map_list_int32"`
	MapListInt64   []map[string]int64   `json:"map_list_int64"`
	MapListUInt32  []map[string]uint32  `json:"map_list_uint32"`
	MapListUInt64  []map[string]uint64  `json:"map_list_uint64"`
	MapListFloat32 []map[string]float32 `json:"map_list_float32"`
	MapListFloat64 []map[string]float64 `json:"map_list_float64"`
	MapListBool    []map[string]bool    `json:"map_list_bool"`
}

type testSubTypes struct {
	TestSimpleStruct  testSimpleStruct  `json:"test_simple_struct"`
	TestSimpleSlices  testSimpleSlices  `json:"test_simple_slices"`
	TestSimpleMaps    testSimpleMaps    `json:"test_simple_maps"`
	TestSimpleMapList testSimpleMapList `json:"test_simple_map_list"`
}

type testPathParam struct {
	ID  uint64 `json:"id" path:"id" required:"-"`
	Cat string `json:"category" path:"category"`
}

type simpleTestReplacement struct {
	ID  uint64 `json:"id"`
	Cat string `json:"category"`
}

type deepReplacementTag struct {
	TestField1 string `json:"test_field_1" swgen_type:"double"`
}

type testWrapParams struct {
	SimpleTestReplacement simpleTestReplacement `json:"simple_test_replacement"`
	ReplaceByTag          int                   `json:"should_be_sting" swgen_type:"string"`
	DeepReplacementTag    deepReplacementTag    `json:"deep_replacement"`
}

type simpleDateTime struct {
	Time time.Time `json:"time"`
}

type sliceDateTime struct {
	Items []simpleDateTime `json:"items"`
}

type mapDateTime struct {
	Items map[string]simpleDateTime `json:"items"`
}

type paramStructMap struct {
	Field1 int                   `schema:"field1"`
	Field2 string                `schema:"field2"`
	Field3 simpleTestReplacement `schema:"field3"`
}

type AnonymousField struct {
	AnonProp int `json:"anonProp"`
}

type mixedStruct struct {
	AnonymousField
	FieldQuery int `schema:"fieldQuery"`
	FieldBody  int `json:"fieldBody"`
}

type typeMapHolder struct {
	M typeMap `json:"m"`
}

type typeMap struct {
	R1 int `json:"1"`
	R2 int `json:"2"`
	R3 int `json:"3"`
	R4 int `json:"4"`
	R5 int `json:"5"`
}

type Gender int

func (Gender) GetEnumSlices() ([]interface{}, []string) {
	return []interface{}{
			PreferNotToDisclose,
			Male,
			Female,
			LGBT,
		}, []string{
			"PreferNotToDisclose",
			"Male",
			"Female",
			"LGBT",
		}
}

const (
	PreferNotToDisclose Gender = iota
	Male
	Female
	LGBT
)

type Flag string

func (Flag) GetEnumSlices() ([]interface{}, []string) {
	return []interface{}{Flag("Foo"), Flag("Bar")}, []string{"Foo", "Bar"}
}

type mixedStructWithEnumer struct {
	Gender Gender `schema:"gender"`
	Flag   Flag   `schema:"flag"`
}

type sliceType []testSimpleStruct

type NullFloat64 struct{}

func (NullFloat64) SwgenDefinition() (typeName string, typeDef SchemaObj, err error) {
	typeName = "NullFloat64"
	typeDef = SchemaFromCommonName(CommonNameFloat)
	return
}

type NullBool struct{}

func (NullBool) SwgenDefinition() (typeName string, typeDef SchemaObj, err error) {
	typeName = "NullBool"
	typeDef = SchemaFromCommonName(CommonNameBoolean)
	return
}

type NullString struct{}

func (NullString) SwgenDefinition() (typeName string, typeDef SchemaObj, err error) {
	typeName = "NullString"
	typeDef = SchemaFromCommonName(CommonNameString)
	return
}

type NullInt64 struct{}

func (NullInt64) SwgenDefinition() (typeName string, typeDef SchemaObj, err error) {
	typeName = "NullInt64"
	typeDef = SchemaFromCommonName(CommonNameLong)
	return
}

type NullDateTime struct{}

func (NullDateTime) SwgenDefinition() (typeName string, typeDef SchemaObj, err error) {
	typeName = "NullDateTime"
	typeDef = SchemaFromCommonName(CommonNameDateTime)
	return
}

type NullDate struct{}

func (NullDate) SwgenDefinition() (typeName string, typeDef SchemaObj, err error) {
	typeName = "NullDate"
	typeDef = SchemaFromCommonName(CommonNameDate)
	return
}

type NullTimestamp struct{}

func (NullTimestamp) SwgenDefinition() (typeName string, typeDef SchemaObj, err error) {
	typeName = "NullTimestamp"
	typeDef = SchemaFromCommonName(CommonNameLong)
	return
}

type testDefaults struct {
	Field1 int            `json:"field1" default:"25"`
	Field2 float64        `json:"field2" default:"25.5"`
	Field3 string         `json:"field3" default:"test"`
	Field4 bool           `json:"field4" default:"true"`
	Field5 []int          `json:"field5" default:"[1, 2, 3]"`
	Field6 map[string]int `json:"field6" default:"{\"test\": 1}"`
	Field7 *uint          `json:"field7" default:"25"`
}

type NullTypes struct {
	Float     NullFloat64   `json:"null_float"`
	Bool      NullBool      `json:"null_bool"`
	String    NullString    `json:"null_string"`
	Int       NullInt64     `json:"null_int"`
	DateTime  NullDateTime  `json:"null_date_time"`
	Date      NullDate      `json:"null_date"`
	Timestamp NullTimestamp `json:"null_timestamp"`
}

type Unknown struct {
	Anything interface{}      `json:"anything"`
	Whatever *json.RawMessage `json:"whatever"`
}

var _ IDefinition = definitionExample{}

type definitionExample struct{}

func (defEx definitionExample) SwgenDefinition() (typeName string, typeDef SchemaObj, err error) {
	return "", SchemaObj{
		Type:   "string",
		Format: "byte",
	}, nil
}

func createPathItemInfo(path, method, title, description, tag string, deprecated bool) PathItemInfo {
	return PathItemInfo{
		Path:        path,
		Method:      method,
		Title:       title,
		Description: description,
		Tag:         tag,
		Deprecated:  deprecated,
	}
}

func TestREST(t *testing.T) {
	gen := NewGenerator()
	gen.SetHost("localhost").
		SetBasePath("/").
		SetInfo("swgen title", "swgen description", "term", "2.0").
		SetLicense("BEER-WARE", "https://fedoraproject.org/wiki/Licensing/Beerware").
		SetContact("Dylan Noblitt", "http://example.com", "dylan.noblitt@example.com").
		AddExtendedField("x-service-type", ServiceTypeRest).
		ReflectGoTypes(true).
		IndentJSON(true)

	gen.AddTypeMap(simpleTestReplacement{}, "")
	gen.AddTypeMap(sliceType{}, float64(0))
	gen.AddTypeMap(typeMap{}, map[string]int{})

	var emptyInterface interface{}

	gen.SetPathItem(createPathItemInfo("/V1/test1", "GET", "test1 name", "test1 description", "v1", false), emptyInterface, emptyInterface, testSimpleStruct{})
	gen.SetPathItem(createPathItemInfo("/V1/test2", "GET", "test2 name", "test2 description", "v1", false), testSimpleStruct{}, emptyInterface, testSimpleSlices{})
	gen.SetPathItem(createPathItemInfo("/V1/test3", "PUT", "test3 name", "test3 description", "v1", false), emptyInterface, testSimpleSlices{}, testSimpleMaps{})
	gen.SetPathItem(createPathItemInfo("/V1/test4", "POST", "test4 name", "test4 description", "v1", false), emptyInterface, testSimpleMaps{}, testSimpleMapList{})
	gen.SetPathItem(createPathItemInfo("/V1/test5", "DELETE", "test5 name", "test5 description", "v1", false), emptyInterface, testSimpleMapList{}, testSubTypes{})
	gen.SetPathItem(createPathItemInfo("/V1/test6", "PATCH", "test6 name", "test6 description", "v1", false), emptyInterface, testSubTypes{}, testSimpleStruct{})
	gen.SetPathItem(createPathItemInfo("/V1/test7", "OPTIONS", "test7 name", "test7 description", "v1", false), emptyInterface, emptyInterface, testSimpleSlices{})
	gen.SetPathItem(createPathItemInfo("/V1/test8", "GET", "test8v1 name", "test8v1 description", "v1", false), paramStructMap{}, emptyInterface, map[string]testSimpleStruct{})
	gen.SetPathItem(createPathItemInfo("/V1/test9", "GET", "test9 name", "test9 description", "v1", false), mixedStruct{}, mixedStruct{}, map[string]testSimpleStruct{})

	gen.SetPathItem(createPathItemInfo("/V1/combine", "GET", "test1 name", "test1 description", "v1", true), emptyInterface, emptyInterface, testSimpleStruct{})
	gen.SetPathItem(createPathItemInfo("/V1/combine", "PUT", "test3 name", "test3 description", "v1", true), emptyInterface, testSimpleSlices{}, testSimpleMaps{})
	gen.SetPathItem(createPathItemInfo("/V1/combine", "POST", "test4 name", "test4 description", "v1", true), emptyInterface, testSimpleMaps{}, testSimpleMapList{})
	gen.SetPathItem(createPathItemInfo("/V1/combine", "DELETE", "test5 name", "test5 description", "v1", true), emptyInterface, testSimpleMapList{}, testSubTypes{})
	gen.SetPathItem(createPathItemInfo("/V1/combine", "PATCH", "test6 name", "test6 description", "v1", true), emptyInterface, testSubTypes{}, testSimpleStruct{})
	gen.SetPathItem(createPathItemInfo("/V1/combine", "OPTIONS", "test7 name", "test7 description", "v1", true), emptyInterface, testSubTypes{}, testSimpleStruct{})

	gen.SetPathItem(createPathItemInfo("/V1/pathParams/{category:[a-zA-Z]{32}}/{id:[0-9]+}", "GET", "test8 name", "test8 description", "V1", false), testPathParam{}, emptyInterface, testSimpleStruct{})

	//anonymous types:
	gen.SetPathItem(createPathItemInfo("/V1/anonymous1", "POST", "test10 name", "test10 description", "v1", false), emptyInterface, testSimpleStruct{}, map[string]int64{})
	gen.SetPathItem(createPathItemInfo("/V1/anonymous2", "POST", "test11 name", "test11 description", "v1", false), emptyInterface, testSimpleStruct{}, map[float64]string{})
	gen.SetPathItem(createPathItemInfo("/V1/anonymous3", "POST", "test12 name", "test12 description", "v1", false), emptyInterface, testSimpleStruct{}, []string{})
	gen.SetPathItem(createPathItemInfo("/V1/anonymous4", "POST", "test13 name", "test13 description", "v1", false), emptyInterface, testSimpleStruct{}, []int{})
	gen.SetPathItem(createPathItemInfo("/V1/anonymous5", "POST", "test14 name", "test14 description", "v1", false), emptyInterface, testSimpleStruct{}, "")
	gen.SetPathItem(createPathItemInfo("/V1/anonymous6", "POST", "test15 name", "test15 description", "v1", false), emptyInterface, testSimpleStruct{}, true)
	gen.SetPathItem(createPathItemInfo("/V1/anonymous7", "POST", "test16 name", "test16 description", "v1", false), emptyInterface, testSimpleStruct{}, map[string]testSimpleStruct{})

	gen.SetPathItem(createPathItemInfo("/V1/typeReplacement1", "POST", "test9 name", "test9 description", "v1", false), emptyInterface, testSubTypes{}, testWrapParams{})

	gen.SetPathItem(createPathItemInfo("/V1/date1", "POST", "test date 1 name", "test date 1 description", "v1", false), emptyInterface, testSimpleStruct{}, simpleDateTime{})
	gen.SetPathItem(createPathItemInfo("/V1/date2", "POST", "test date 2 name", "test date 2 description", "v1", false), emptyInterface, testSimpleStruct{}, sliceDateTime{})
	gen.SetPathItem(createPathItemInfo("/V1/date3", "POST", "test date 3 name", "test date 3 description", "v1", false), emptyInterface, testSimpleStruct{}, mapDateTime{})
	gen.SetPathItem(createPathItemInfo("/V1/date4", "POST", "test date 4 name", "test date 4 description", "v1", false), emptyInterface, testSimpleStruct{}, []mapDateTime{})

	gen.SetPathItem(createPathItemInfo("/V1/slice1", "POST", "test slice 1 name", "test slice 1 description", "v1", false), emptyInterface, testSimpleStruct{}, []mapDateTime{})
	gen.SetPathItem(createPathItemInfo("/V1/slice2", "POST", "test slice 2 name", "test slice 2 description", "v1", false), emptyInterface, testSimpleStruct{}, sliceType{})

	gen.SetPathItem(createPathItemInfo("/V1/IDefinition1", "POST", "test IDefinition1 name", "test IDefinition1 description", "v1", false), emptyInterface, definitionExample{}, definitionExample{})
	gen.SetPathItem(createPathItemInfo("/V1/nullTypes", "GET", "test nulltypes", "test nulltypes", "v1", false), emptyInterface, NullTypes{}, NullTypes{})

	gen.SetPathItem(createPathItemInfo("/V1/primitiveTypes1", "POST", "testPrimitives", "test Primitives", "v1", false), emptyInterface, "", 10)
	gen.SetPathItem(createPathItemInfo("/V1/primitiveTypes2", "POST", "testPrimitives", "test Primitives", "v1", false), emptyInterface, true, 1.1)
	gen.SetPathItem(createPathItemInfo("/V1/primitiveTypes3", "POST", "testPrimitives", "test Primitives", "v1", false), emptyInterface, int64(10), "")
	gen.SetPathItem(createPathItemInfo("/V1/primitiveTypes4", "POST", "testPrimitives", "test Primitives", "v1", false), emptyInterface, int64(10), "")

	gen.SetPathItem(createPathItemInfo("/V1/defaeults1", "GET", "default", "test defaults", "v1", false), emptyInterface, emptyInterface, testDefaults{})
	gen.SetPathItem(createPathItemInfo("/V1/unknown", "POST", "test unknown types", "test unknown types", "v1", false), emptyInterface, Unknown{}, Unknown{})

	gen.SetPathItem(createPathItemInfo("/V1/empty", "POST", "test empty struct", "test empty struct", "v1", false), testEmptyStruct{}, nil, testEmptyStruct{})

	gen.SetPathItem(createPathItemInfo("/V1/struct-collision", "POST", "test struct name collision", "test struct name collision", "v1", false), nil, TestSampleStruct{}, TestSampleStruct{})
	gen.SetPathItem(createPathItemInfo("/V2/struct-collision", "POST", "test struct name collision", "test struct name collision", "v2", false), nil, sample.TestSampleStruct{}, sample.TestSampleStruct{})

	gen.SetPathItem(createPathItemInfo("/V1/type-map", "POST", "test type mapping", "test type mapping", "v1", false), nil, nil, typeMapHolder{})

	bytes, err := gen.GenDocument()
	if err != nil {
		t.Fatalf("Failed to generate Swagger JSON document: %s", err.Error())
	}

	if err := writeLastRun("test_REST_last_run.json", bytes); err != nil {
		t.Fatalf("Failed write last run data to a file: %s", err.Error())
	}

	assertTrue(checkResult(bytes, "test_REST.json", t), t)
}

func TestJsonRpc(t *testing.T) {
	gen := NewGenerator()
	gen.SetHost("localhost")
	gen.SetInfo("swgen title", "swgen description", "term", "2.0")
	gen.SetLicense("BEER-WARE", "https://fedoraproject.org/wiki/Licensing/Beerware")
	gen.SetContact("Dylan Noblitt", "http://example.com", "dylan.noblitt@example.com")
	gen.AddExtendedField("x-service-type", ServiceTypeJSONRPC)
	gen.AddTypeMap(simpleTestReplacement{}, "")
	gen.AddTypeMap(sliceType{}, "")
	gen.IndentJSON(true)

	var emptyInterface interface{}

	gen.SetPathItem(createPathItemInfo("/V1/test1", "POST", "test1 name", "test1 description", "v1", true), emptyInterface, emptyInterface, testSimpleStruct{})
	gen.SetPathItem(createPathItemInfo("/V1/test2", "POST", "test2 name", "test2 description", "v1", true), testSimpleStruct{}, emptyInterface, testSimpleSlices{})
	gen.SetPathItem(createPathItemInfo("/V1/test3", "POST", "test3 name", "test3 description", "v1", true), emptyInterface, testSimpleSlices{}, testSimpleMaps{})
	gen.SetPathItem(createPathItemInfo("/V1/test4", "POST", "test4 name", "test4 description", "v1", true), emptyInterface, testSimpleMaps{}, testSimpleMapList{})
	gen.SetPathItem(createPathItemInfo("/V1/test5", "POST", "test5 name", "test5 description", "v1", true), emptyInterface, testSimpleMapList{}, testSubTypes{})
	gen.SetPathItem(createPathItemInfo("/V1/test6", "POST", "test6 name", "test6 description", "v1", true), emptyInterface, testSubTypes{}, testSimpleStruct{})
	gen.SetPathItem(createPathItemInfo("/V1/test7", "POST", "test7 name", "test7 description", "v1", true), emptyInterface, emptyInterface, testSimpleSlices{})
	gen.SetPathItem(createPathItemInfo("/V1/test8", "POST", "test8v1 name", "test8v1 description", "v1", true), emptyInterface, paramStructMap{}, map[string]testSimpleStruct{})
	gen.SetPathItem(createPathItemInfo("/V1/test9", "POST", "test9 name", "test9 description", "v1", true), mixedStruct{}, mixedStruct{}, map[string]testSimpleStruct{})
	gen.SetPathItem(createPathItemInfo("/V1/test10", "POST", "test10 name", "test10 description", "v1", true), mixedStructWithEnumer{}, mixedStruct{}, map[string]testSimpleStruct{})

	gen.SetPathItem(createPathItemInfo("/V1/typeReplacement1", "POST", "test9 name", "test9 description", "v1", false), emptyInterface, testSubTypes{}, testWrapParams{})

	//anonymous types:
	gen.SetPathItem(createPathItemInfo("/V1/anonymous1", "POST", "test10 name", "test10 description", "v1", false), emptyInterface, emptyInterface, map[string]int64{})
	gen.SetPathItem(createPathItemInfo("/V1/anonymous2", "POST", "test11 name", "test11 description", "v1", false), emptyInterface, emptyInterface, map[float64]string{})
	gen.SetPathItem(createPathItemInfo("/V1/anonymous3", "POST", "test12 name", "test12 description", "v1", false), emptyInterface, emptyInterface, []string{})
	gen.SetPathItem(createPathItemInfo("/V1/anonymous4", "POST", "test13 name", "test13 description", "v1", false), emptyInterface, emptyInterface, []int{})
	gen.SetPathItem(createPathItemInfo("/V1/anonymous5", "POST", "test14 name", "test14 description", "v1", false), emptyInterface, emptyInterface, "")
	gen.SetPathItem(createPathItemInfo("/V1/anonymous6", "POST", "test15 name", "test15 description", "v1", false), emptyInterface, emptyInterface, true)
	gen.SetPathItem(createPathItemInfo("/V1/anonymous7", "POST", "test16 name", "test16 description", "v1", false), emptyInterface, emptyInterface, map[string]testSimpleStruct{})

	gen.SetPathItem(createPathItemInfo("/V1/date1", "POST", "test date 1 name", "test date 1 description", "v1", false), emptyInterface, emptyInterface, simpleDateTime{})
	gen.SetPathItem(createPathItemInfo("/V1/date2", "POST", "test date 2 name", "test date 2 description", "v1", false), emptyInterface, emptyInterface, sliceDateTime{})
	gen.SetPathItem(createPathItemInfo("/V1/date3", "POST", "test date 3 name", "test date 3 description", "v1", false), emptyInterface, emptyInterface, mapDateTime{})
	gen.SetPathItem(createPathItemInfo("/V1/date4", "POST", "test date 4 name", "test date 4 description", "v1", false), emptyInterface, emptyInterface, []mapDateTime{})

	gen.SetPathItem(createPathItemInfo("/V1/slice1", "POST", "test slice 1 name", "test slice 1 description", "v1", false), emptyInterface, emptyInterface, []mapDateTime{})
	gen.SetPathItem(createPathItemInfo("/V1/slice2", "POST", "test slice 2 name", "test slice 2 description", "v1", false), emptyInterface, emptyInterface, sliceType{})

	gen.SetPathItem(createPathItemInfo("/V1/primitiveTypes1", "POST", "testPrimitives", "test Primitives", "v1", false), emptyInterface, "", 10)
	gen.SetPathItem(createPathItemInfo("/V1/primitiveTypes2", "POST", "testPrimitives", "test Primitives", "v1", false), emptyInterface, true, 1.1)
	gen.SetPathItem(createPathItemInfo("/V1/primitiveTypes3", "POST", "testPrimitives", "test Primitives", "v1", false), emptyInterface, int64(10), "")
	gen.SetPathItem(createPathItemInfo("/V1/primitiveTypes4", "POST", "testPrimitives", "test Primitives", "v1", false), emptyInterface, int64(10), "")

	gen.SetPathItem(createPathItemInfo("/V1/defaults1", "POST", "default", "test defaults", "v1", false), emptyInterface, emptyInterface, testDefaults{})

	bytes, err := gen.GenDocument()
	if err != nil {
		t.Fatalf("can not generate document: %s", err.Error())
	}

	if err := writeLastRun("test_JSON-RPC_last_run.json", bytes); err != nil {
		t.Fatalf("Failed write last run data to a file: %s", err.Error())
	}

	assertTrue(checkResult(bytes, "test_JSON-RPC.json", t), t)
}

func getTestDataDir(filename string) string {
	pwd, err := os.Getwd()
	if err != nil {
		return filename
	}

	return path.Join(pwd, "testdata", filename)
}

func writeLastRun(filename string, data []byte) error {
	return ioutil.WriteFile(getTestDataDir(filename), data, os.ModePerm)
}

func readTestFile(filename string) ([]byte, error) {
	bytes, readError := ioutil.ReadFile(getTestDataDir(filename))
	if readError != nil {
		return []byte{}, readError
	}

	return bytes, nil
}

func checkResult(generatedBytes []byte, expectedDataFileName string, t *testing.T) bool {
	expectedData := make(map[string]interface{})
	generatedData := make(map[string]interface{})

	expectedBytes, err := readTestFile(expectedDataFileName)
	if err != nil {
		t.Fatalf("can not read test file '%s': %s", expectedDataFileName, err.Error())
	}
	if err = json.Unmarshal(expectedBytes, &expectedData); err != nil {
		t.Fatalf("can not unmarshal '%s' data: %s", expectedDataFileName, err.Error())
	}
	if err = json.Unmarshal(generatedBytes, &generatedData); err != nil {
		t.Fatalf("can not unmarshal generated data: %s", err.Error())
	}

	for _, diff := range pretty.Diff(expectedData, generatedData) {
		pretty.Println(diff)
	}

	return reflect.DeepEqual(expectedData, generatedData)
}

func TestGenDocumentFunc(t *testing.T) {
	SetHost("localhost:1234")
	SetBasePath("/")
	SetContact("Test Name", "test@email.com", "http://test.com")
	SetLicense("MIT", "http://www.mit.com")
	SetInfo("Test API", "Generate api document", "term.com", "1.0.0")
	EnableCORS(false)

	info := PathItemInfo{
		Path:        "/v1/test/handler",
		Title:       "TestHandler",
		Description: "This is just a test handler with GET request",
		Method:      "GET",
	}

	if err := SetPathItem(info, nil, nil, nil); err != nil {
		t.Fatalf("error %v", err)
	}

	w := httptest.NewRecorder()
	r, err := http.NewRequest("GET", "http://localhost:1234/docs/swagger.json", nil)
	if err != nil {
		t.Fatalf("error when create request: %v", err)
	}

	ServeHTTP(w, r)

	responseDoc := Document{}
	if err := json.Unmarshal(w.Body.Bytes(), &responseDoc); err != nil {
		t.Fatalf("could not get response: %v", err)
	}

	if responseDoc.Host != "localhost:1234" {
		t.Fatalf("gen swagger json error: %v", responseDoc)
	}

	// generate again without base path
	data, _ := GenDocument()
	doc := Document{}
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatalf("could not get response: %v", err)
	}

	for path := range doc.Paths {
		if !strings.HasPrefix(path, "/v1") {
			t.Fatal("Path should be started by /v1")
		}
	}

	assertTrue(w.Header().Get("Access-Control-Allow-Origin") == "", t)
	assertTrue(w.Header().Get("Access-Control-Allow-Methods") == "", t)
	assertTrue(w.Header().Get("Access-Control-Allow-Headers") == "", t)
}

func TestCORSSupport(t *testing.T) {
	g := NewGenerator()
	g.EnableCORS(true, "X-ABC-Test").
		SetHost("localhost:1234")

	info := PathItemInfo{
		Path:        "/v1/test/handler",
		Title:       "TestHandler",
		Description: "This is just a test handler with GET request",
		Method:      "GET",
	}

	if err := g.SetPathItem(info, nil, nil, nil); err != nil {
		t.Fatalf("error %v", err)
	}

	w := httptest.NewRecorder()
	r, err := http.NewRequest("GET", "http://localhost:1234/docs/swagger.json", nil)
	if err != nil {
		t.Fatalf("error when create request: %v", err)
	}

	g.ServeHTTP(w, r)

	assertTrue(w.Header().Get("Access-Control-Allow-Origin") == "*", t)
	assertTrue(w.Header().Get("Access-Control-Allow-Methods") == "GET, POST, DELETE, PUT, PATCH, OPTIONS", t)
	assertTrue(w.Header().Get("Access-Control-Allow-Headers") == "Content-Type, api_key, Authorization, X-ABC-Test", t)
}
