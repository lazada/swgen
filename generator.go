package swgen

import (
	"encoding/json"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

// singleton package generator
var gen = NewGenerator()

// Generator create swagger document
type Generator struct {
	doc              Document
	host             string // address of api in host:port format
	corsEnabled      bool   // allow cross-origin HTTP request
	corsAllowHeaders []string

	definitions map[string]SchemaObj // list of all definition objects
	defMux      *sync.Mutex

	defQueue map[string]interface{}
	queueMux *sync.Mutex

	paths    map[string]PathItem // list all of paths object
	pathsMux *sync.RWMutex

	typesMap map[string]interface{}

	indentJSON     bool
	reflectGoTypes bool
}

// NewGenerator create a new Generator
func NewGenerator() *Generator {
	g := &Generator{}

	g.definitions = make(map[string]SchemaObj)
	g.defMux = &sync.Mutex{}

	g.defQueue = make(map[string]interface{})
	g.queueMux = &sync.Mutex{}

	g.paths = make(map[string]PathItem) // list all of paths object
	g.typesMap = make(map[string]interface{})
	g.pathsMux = &sync.RWMutex{}

	g.doc.Schemes = []string{"http", "https"}
	g.doc.Paths = make(map[string]PathItem)
	g.doc.Definitions = make(map[string]SchemaObj)
	g.doc.Version = "2.0"
	g.doc.BasePath = "/"

	// set default Access-Control-Allow-Headers of swagger.json
	g.corsAllowHeaders = []string{"Content-Type", "api_key", "Authorization"}

	return g
}

// IndentJSON controls JSON indentation
func (g *Generator) IndentJSON(enabled bool) *Generator {
	g.indentJSON = enabled
	return g
}

// ReflectGoTypes controls JSON indentation
func (g *Generator) ReflectGoTypes(enabled bool) *Generator {
	g.reflectGoTypes = enabled
	return g
}

// EnableCORS enable HTTP handler support CORS
func (g *Generator) EnableCORS(b bool, allowHeaders ...string) *Generator {
	g.corsEnabled = b
	if len(allowHeaders) != 0 {
		g.corsAllowHeaders = append(g.corsAllowHeaders, allowHeaders...)
	}

	return g
}

// EnableCORS enable HTTP handler support CORS
func EnableCORS(b bool, allowHeaders ...string) *Generator {
	return gen.EnableCORS(b, allowHeaders...)
}

// SetHost set host info for swagger specification
func (g *Generator) SetHost(host string) *Generator {
	g.host = host
	return g
}

// SetHost set host info for swagger specification
func SetHost(host string) *Generator {
	return gen.SetHost(host)
}

// SetBasePath set host info for swagger specification
func (g *Generator) SetBasePath(basePath string) *Generator {
	g.doc.BasePath = "/" + strings.Trim(basePath, "/")
	return g
}

// SetBasePath set host info for swagger specification
func SetBasePath(basePath string) *Generator {
	return gen.SetBasePath(basePath)
}

// SetContact set contact information for API
func (g *Generator) SetContact(name, url, email string) *Generator {
	ct := ContactObj{
		Name:  name,
		URL:   url,
		Email: email,
	}

	g.doc.Info.Contact = ct
	return g
}

// SetContact set contact information for API
func SetContact(name, url, email string) *Generator {
	return gen.SetContact(name, url, email)
}

// SetInfo set information about API
func (g *Generator) SetInfo(title, description, term, version string) *Generator {
	info := InfoObj{
		Title:          title,
		Description:    description,
		TermsOfService: term,
		Version:        version,
	}

	g.doc.Info = info
	return g
}

// SetInfo set information about API
func SetInfo(title, description, term, version string) *Generator {
	return gen.SetInfo(title, description, term, version)
}

// SetLicense set license information for API
func (g *Generator) SetLicense(name, url string) *Generator {
	ls := LicenseObj{
		Name: name,
		URL:  url,
	}

	g.doc.Info.License = ls
	return g
}

// SetLicense set license information for API
func SetLicense(name, url string) *Generator {
	return gen.SetLicense(name, url)
}

// SetType set service type
func AddExtendedField(name string, value interface{}) *Generator {
	return gen.AddExtendedField(name, value)
}

// AddExtendedField add field to additional data map
func (g *Generator) AddExtendedField(name string, value interface{}) *Generator {
	g.doc.AddExtendedField(name, value)
	return g
}

// AddTypeMap add rule to use dst interface instead of src
func (g *Generator) AddTypeMap(src interface{}, dst interface{}) *Generator {
	t := reflect.TypeOf(src)
	g.typesMap[t.String()] = dst
	return g
}

// AddTypeMap add rule to use dst interface instead of src
func AddTypeMap(src interface{}, dst interface{}) *Generator {
	return gen.AddTypeMap(src, dst)
}

// genDocument returns document specification in JSON string (in []byte)
func (g *Generator) genDocument(host string) ([]byte, error) {
	// ensure that all definition in queue is parsed before generating
	g.parseDefInQueue()
	g.doc.Definitions = g.definitions
	g.doc.Host = host
	g.doc.Paths = make(map[string]PathItem)

	for path, item := range g.paths {
		t, isServiceType := g.doc.data["x-service-type"].(ServiceType)
		if isServiceType && t == ServiceTypeJSONRPC {
			if !item.HasMethod("POST") {
				continue
			}

			item.Get = nil
			item.Put = nil
			item.Delete = nil
			item.Options = nil
			item.Head = nil
			item.Patch = nil
		}
		g.doc.Paths[path] = item
	}

	g.pathsMux.RLock()
	var (
		data []byte
		err  error
	)
	if g.indentJSON {
		data, err = json.MarshalIndent(g.doc, "", "  ")
	} else {
		data, err = json.Marshal(g.doc)
	}
	g.pathsMux.RUnlock()

	return data, err
}

// GenDocument returns document specification in JSON string (in []byte)
func (g *Generator) GenDocument() ([]byte, error) {
	return g.genDocument(g.host)
}

// GenDocument returns document specification in JSON string (in []byte)
func GenDocument() ([]byte, error) {
	return gen.GenDocument()
}

// ServeHTTP implements http.Handler to server swagger.json document
func (g *Generator) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	host := r.URL.Host
	if g.host != "" {
		host = g.host
	}
	data, err := g.genDocument(host)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))

	if g.corsEnabled {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, PUT, PATCH, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", strings.Join(g.corsAllowHeaders, ", "))
	}

	w.Write(data)
}

// ServeHTTP implements http.HandleFunc to server swagger.json document
func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	gen.ServeHTTP(w, r)
}

func (g *Generator) getTypeMapByString(src string) (interface{}, bool) {
	dstInterface, exists := g.typesMap[src]
	return dstInterface, exists
}
