package swgen_test

import (
	"fmt"

	"github.com/lazada/swgen"
)

func ExampleGenerator_GenDocument() {
	// PetsRequest defines all params for /pest request
	type PetsRequest struct {
		Tags  []string `schema:"tags" in:"query" required:"-" description:"tags to filter by"`
		Limit int32    `schema:"limit" in:"query" required:"-" description:"maximum number of results to return"`
	}

	// Pet contains information of a pet
	type Pet struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
		Tag  string `json:"tag"`
	}

	gen := swgen.NewGenerator()
	gen.SetHost("petstore.swagger.io").SetBasePath("/api")
	gen.SetInfo("Swagger Petstore (Simple)", "A sample API that uses a petstore as an example to demonstrate features in the swagger-2.0 specification", "http://helloreverb.com/terms/", "2.0")
	gen.SetLicense("MIT", "http://opensource.org/licenses/MIT")
	gen.SetContact("Swagger API team", "http://swagger.io", "foo@example.com")

	pathInf := swgen.PathItemInfo{
		Path:        "/pets",
		Method:      "GET",
		Title:       "findPets",
		Description: "Returns all pets from the system that the user has access to",
		Tag:         "v1",
		Deprecated:  false,
	}
	pathInf.AddExtendedField("x-example", "example")

	gen.SetPathItem(
		pathInf,
		PetsRequest{}, // request object
		nil,           // body data if any
		[]Pet{},       // response object
	)

	// extended field
	gen.SetUppercaseVersion(true)

	docData, _ := gen.GenDocument()
	fmt.Println(string(docData))

	// output:
	// {"swagger":"2.0","info":{"title":"Swagger Petstore (Simple)","description":"A sample API that uses a petstore as an example to demonstrate features in the swagger-2.0 specification","termsOfService":"http://helloreverb.com/terms/","contact":{"name":"Swagger API team","url":"http://swagger.io","email":"foo@example.com"},"license":{"name":"MIT","url":"http://opensource.org/licenses/MIT"},"version":"2.0"},"host":"petstore.swagger.io","basePath":"/api","schemes":["http","https"],"paths":{"/pets":{"get":{"tags":["v1"],"summary":"findPets","description":"Returns all pets from the system that the user has access to","parameters":[{"name":"tags","in":"query","type":"array","items":{"type":"string"},"collectionFormat":"multi","description":"tags to filter by","required":false},{"name":"limit","in":"query","type":"integer","format":"int32","description":"maximum number of results to return","required":false}],"responses":{"200":{"description":"request success","schema":{"type":"array","items":{"$ref":"#/definitions/Pet"}}}},"deprecated":false,"x-example":"example"}}},"definitions":{"Pet":{"type":"object","properties":{"id":{"type":"integer","format":"int64"},"name":{"type":"string"},"tag":{"type":"string"}}}},"x-uppercase-version":true}
}
