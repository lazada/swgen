package swgen

import "net/http"

// singleton package generator
var gen = NewGenerator()

// EnableCORS enable HTTP handler support CORS
func EnableCORS(b bool, allowHeaders ...string) *Generator {
	return gen.EnableCORS(b, allowHeaders...)
}

// SetHost set host info for swagger specification
func SetHost(host string) *Generator {
	return gen.SetHost(host)
}

// SetBasePath set host info for swagger specification
func SetBasePath(basePath string) *Generator {
	return gen.SetBasePath(basePath)
}

// SetContact set contact information for API
func SetContact(name, url, email string) *Generator {
	return gen.SetContact(name, url, email)
}

// SetInfo set information about API
func SetInfo(title, description, term, version string) *Generator {
	return gen.SetInfo(title, description, term, version)
}

// SetLicense set license information for API
func SetLicense(name, url string) *Generator {
	return gen.SetLicense(name, url)
}

// AddExtendedField add vendor extension field to document
func AddExtendedField(name string, value interface{}) *Generator {
	return gen.AddExtendedField(name, value)
}

// AddTypeMap add rule to use dst interface instead of src
func AddTypeMap(src interface{}, dst interface{}) *Generator {
	return gen.AddTypeMap(src, dst)
}

// GenDocument returns document specification in JSON string (in []byte)
func GenDocument() ([]byte, error) {
	return gen.GenDocument()
}

// ServeHTTP implements http.HandleFunc to server swagger.json document
func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	gen.ServeHTTP(w, r)
}
