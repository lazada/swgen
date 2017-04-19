package swgen

import (
	"encoding/json"
	"fmt"
	"sync"
	"testing"
)

type Person struct {
	Name       *PersonName `json:"name"`
	SecondName PersonName  `json:"second_name"`
	Age        uint        `json:"age"`
	Children   []Person    `json:"children"`
	Tags       []string    `json:"tags"`
	Weight     float64     `json:"weight"`
	Active     bool        `json:"active"`
	Balance    float32     `json:"balance"`
}

type PersonName struct {
	First    string `json:"first_name"`
	Middle   string `json:"middle_name"`
	Last     string `json:"last_name"`
	Nickname string `schema:"-"`
	_        string
}

type Employee struct {
	Person
	Salary float64 `json:"salary"`
}

// PreferredWarehouseRequest is request object of get preferred warehouse handler
type PreferredWarehouseRequest struct {
	Items              []string `schema:"items" description:"List of simple sku"`
	IDCustomerLocation uint64   `schema:"id_customer_location" description:"-"`
}

func TestResetDefinitions(t *testing.T) {
	ts := &Person{}
	if _, err := ParseDefinition(ts); err != nil {
		t.Fatalf("%v", err)
	}

	if len(gen.definitions) == 0 {
		t.Fatalf("len of gen.definitions must be greater than 0")
	}

	ResetDefinitions()
	if len(gen.definitions) != 0 {
		t.Fatalf("len of gen.definitions must be equal to 0")
	}
}

func TestParseDefinition(t *testing.T) {
	ts := &Person{}
	if _, err := ParseDefinition(ts); err != nil {
		t.Fatalf("%v", err)
	}
}

func TestParseDefinitionEmptyInterface(t *testing.T) {
	var ts interface{}
	if _, err := ParseDefinition(&ts); err != nil {
		t.Fatalf("%v", err)
	}
}

func TestParseDefinitionNonEmptyInterface(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Panic expected for non-empty interface")
		}
	}()

	var ts interface {
		Test()
	}

	if _, err := ParseDefinition(&ts); err != nil {
		t.Fatalf("%v", err)
	}
}

func TestParseDefinitionWithEmbeddedStruct(t *testing.T) {
	ts := &Employee{}
	typeDef, err := ParseDefinition(ts)
	if err != nil {
		t.Fatalf("%v", err)
	}
	name := typeDef.TypeName
	propertiesCount := len(gen.definitions[name].Properties)
	expectedPropertiesCount := 9
	if propertiesCount != expectedPropertiesCount {
		t.Fatalf("Expected %d properties, got %d : %#v", expectedPropertiesCount, propertiesCount, gen.definitions[name].Properties)
	}
}

func TestParseDefinitionString(t *testing.T) {
	typeDef, err := ParseDefinition("string")
	name := typeDef.TypeName
	if err != nil {
		t.Fatalf("Error parsing string: %+v", err)
	}
	if name != "string" {
		t.Fatalf("Wrong type name. Expect %q, got %q", "string", name)
	}
}

func TestParseDefinitionArray(t *testing.T) {
	type Names []string
	typeDef, err := ParseDefinition(Names{})
	if err != nil {
		t.Fatalf("Error while parsing array of string: %v", err)
	}

	if typeDef.TypeName != "Names" {
		t.Fatalf("Wrong type name. Expected: Names, Obtained: %v", typeDef.TypeName)
	}

	// re-parse with pointer input
	// should get from definition list
	_, err = ParseDefinition(&Names{})
	if err != nil {
		t.Fatalf("Error while parsing array of string: %v", err)
	}

	// try to parse a named map
	type MapList map[string]string
	_, err = ParseDefinition(&MapList{})
	if err != nil {
		t.Fatalf("Error while parsing map string to string: %v", err)
	}

	// named array of object
	type Person struct{}
	type Persons []*Person
	_, err = ParseDefinition(&Persons{})
	if err != nil {
		t.Fatalf("Error while parsing array of object: %v", err)
	}
}

func TestParseParameter(t *testing.T) {
	p := &PreferredWarehouseRequest{}
	name, params, err := ParseParameter(p)

	if err != nil {
		t.Fatalf("error %v", err)
	}

	if name != "PreferredWarehouseRequest" {
		t.Fatalf("name of parameter is %s, expected is PreferredWarehouseRequest", name)
	}

	if len(params) != 2 {
		t.Fatalf("number of parameter should be 2")
	}
}

func TestParseParameterError(t *testing.T) {
	_, _, err := ParseParameter(true)
	if err == nil {
		t.Fatalf("it should return error")
	}
}

//
// test and data for TestSetPathItem
//

func TestSetPathItem(t *testing.T) {
	h := &testHandler{}

	methods := []string{"GET", "POST", "HEAD", "PUT", "OPTIONS", "DELETE", "PATCH"}

	for _, method := range methods {
		info := PathItemInfo{
			Path:        "/v1/test/handler",
			Title:       "TestHandler",
			Description: fmt.Sprintf("This is just a test handler with %s request", method),
			Method:      method,
		}
		err := SetPathItem(info, h.GetRequestBuffer(method), h.GetBodyBuffer(), h.GetResponseBuffer(method))
		if err != nil {
			t.Fatalf("error %v", err)
		}
	}
}

func TestResetPaths(t *testing.T) {
	TestSetPathItem(t)

	if len(gen.paths) == 0 {
		t.Fatalf("len of gen.paths must be greater than 0")
	}

	ResetPaths()
	if len(gen.paths) != 0 {
		t.Fatalf("len of gen.paths must be equal to 0")
	}
}

//
// benchmark and parallel testing
//

func BenchmarkParseDefinitionsParallel(b *testing.B) {
	var (
		mu sync.Mutex
		i  int
	)

	data := []interface{}{
		&Person{},
		&PersonName{},
		&PreferredWarehouseRequest{},
	}

	b.SetParallelism(10)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			mu.Lock()
			i++
			input := data[i%3]
			mu.Unlock()

			if _, err := ParseDefinition(input); err != nil {
				b.Fatalf("%v", err)
			}
		}
	})
}

func BenchmarkSetPathItem(b *testing.B) {
	h := &testHandler{}

	infos := []PathItemInfo{
		{
			Path:        "/v1/test/handler",
			Title:       "TestHandler",
			Description: "This is just a test handler with GET request",
			Method:      "GET",
		},
		{
			Path:        "/v1/test/handler",
			Title:       "TestHandler",
			Description: "This is just a test handler with POST reqest",
			Method:      "POST",
		},
	}

	var (
		mu sync.Mutex
		i  int
	)

	b.SetParallelism(10)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			mu.Lock()
			i++
			info := infos[i%2]
			mu.Unlock()

			err := SetPathItem(
				info,
				h.GetRequestBuffer(info.Method),
				h.GetBodyBuffer(),
				h.GetResponseBuffer(info.Method),
			)

			if err != nil {
				b.Fatalf("error %v", err)
			}
		}
	})
}

// testHandler can handle POST and GET request
type testHandler struct{}

func (th *testHandler) GetName() string {
	return "TestHandle"
}

func (th *testHandler) GetDescription() string {
	return "This handler for test ParsePathItem"
}

func (th *testHandler) GetVersion() string {
	return "v1"
}

func (th *testHandler) GetRoute() string {
	return "/test/handler"
}

func (th *testHandler) GetRequestBuffer(_ string) interface{} {
	return &PersonName{}
}

func (th *testHandler) GetResponseBuffer(method string) interface{} {
	if method == "GET" {
		return nil
	}

	return &PreferredWarehouseRequest{}
}

func (th *testHandler) GetBodyBuffer() interface{} {
	return &Person{}
}

func (th *testHandler) HandlePost(_ interface{}, _ interface{}) (response interface{}, err error) {
	// yes, I can handle a POST request
	return
}

func (th *testHandler) HandleGet(_ interface{}) (response interface{}, err error) {
	// yes, I can handle a GET request
	return
}

//
// Test helper
//

func getReadableJSON(i interface{}, t *testing.T) []byte {
	data, err := json.MarshalIndent(i, "", "  ")
	if err != nil {
		t.Fatalf("error while parsing struct to JSON string: %v", err)
	}

	return data
}
