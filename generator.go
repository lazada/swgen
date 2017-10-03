package swgen

import (
	"encoding/json"
	"net/http"
	"reflect"
	"strings"
	"sync"
)

// Generator create swagger document
type Generator struct {
	doc  Document
	host string // address of api in host:port format

	corsMu           sync.RWMutex // mutex for CORS public API
	corsEnabled      bool         // allow cross-origin HTTP request
	corsAllowHeaders []string

	definitionAdded map[string]bool           // index of TypeNames
	definitions     defMap                    // list of all definition objects
	defQueue        map[reflect.Type]struct{} // queue of reflect.Type objects waiting for analysis
	paths           map[string]PathItem       // list all of paths object
	typesMap        map[reflect.Type]interface{}

	indentJSON     bool
	reflectGoTypes bool

	mu sync.Mutex // mutex for Generator's public API
}

type defMap map[reflect.Type]SchemaObj

func (m *defMap) GenDefinitions() (result map[string]SchemaObj) {
	if m == nil {
		return nil
	}

	result = make(map[string]SchemaObj)
	for t, typeDef := range *m {
		typeDef.Ref = "" // first (top) level Swagger definitions are never references
		if _, ok := result[typeDef.TypeName]; ok {
			typeName := goType(t)
			result[typeName] = typeDef
		} else {
			result[typeDef.TypeName] = typeDef
		}
	}
	return
}

// NewGenerator create a new Generator
func NewGenerator() *Generator {
	g := &Generator{}

	g.definitions = make(map[reflect.Type]SchemaObj)
	g.definitionAdded = make(map[string]bool)

	g.defQueue = make(map[reflect.Type]struct{})
	g.paths = make(map[string]PathItem) // list all of paths object
	g.typesMap = make(map[reflect.Type]interface{})

	g.doc.Schemes = []string{"http", "https"}
	g.doc.Paths = make(map[string]PathItem)
	g.doc.Definitions = make(map[string]SchemaObj)
	g.doc.SecurityDefinitions = make(map[string]SecurityDef)
	g.doc.Version = "2.0"
	g.doc.BasePath = "/"

	// set default Access-Control-Allow-Headers of swagger.json
	g.corsAllowHeaders = []string{"Content-Type", "api_key", "Authorization"}

	return g
}

// IndentJSON controls JSON indentation
func (g *Generator) IndentJSON(enabled bool) *Generator {
	g.mu.Lock()
	g.indentJSON = enabled
	g.mu.Unlock()
	return g
}

// ReflectGoTypes controls JSON indentation
func (g *Generator) ReflectGoTypes(enabled bool) *Generator {
	g.mu.Lock()
	g.reflectGoTypes = enabled
	g.mu.Unlock()
	return g
}

// EnableCORS enable HTTP handler support CORS
func (g *Generator) EnableCORS(b bool, allowHeaders ...string) *Generator {
	g.corsMu.Lock()
	g.corsEnabled = b
	if len(allowHeaders) != 0 {
		g.corsAllowHeaders = append(g.corsAllowHeaders, allowHeaders...)
	}
	g.corsMu.Unlock()
	return g
}

func (g *Generator) writeCORSHeaders(w http.ResponseWriter) {
	g.corsMu.RLock()
	defer g.corsMu.RUnlock()

	if !g.corsEnabled {
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, PUT, PATCH, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", strings.Join(g.corsAllowHeaders, ", "))
}

// SetHost set host info for swagger specification
func (g *Generator) SetHost(host string) *Generator {
	g.mu.Lock()
	g.host = host
	g.mu.Unlock()
	return g
}

// SetBasePath set host info for swagger specification
func (g *Generator) SetBasePath(basePath string) *Generator {
	basePath = "/" + strings.Trim(basePath, "/")
	g.mu.Lock()
	g.doc.BasePath = basePath
	g.mu.Unlock()
	return g
}

// SetContact set contact information for API
func (g *Generator) SetContact(name, url, email string) *Generator {
	ct := ContactObj{
		Name:  name,
		URL:   url,
		Email: email,
	}

	g.mu.Lock()
	g.doc.Info.Contact = ct
	g.mu.Unlock()
	return g
}

// SetInfo set information about API
func (g *Generator) SetInfo(title, description, term, version string) *Generator {
	info := InfoObj{
		Title:          title,
		Description:    description,
		TermsOfService: term,
		Version:        version,
	}

	g.mu.Lock()
	g.doc.Info = info
	g.mu.Unlock()
	return g
}

// SetLicense set license information for API
func (g *Generator) SetLicense(name, url string) *Generator {
	ls := LicenseObj{
		Name: name,
		URL:  url,
	}

	g.mu.Lock()
	g.doc.Info.License = ls
	g.mu.Unlock()
	return g
}

// AddExtendedField add vendor extension field to document
func (g *Generator) AddExtendedField(name string, value interface{}) *Generator {
	g.mu.Lock()
	g.doc.AddExtendedField(name, value)
	g.mu.Unlock()
	return g
}

// AddSecurityDefinition adds shared security definition to document
func (g *Generator) AddSecurityDefinition(name string, def SecurityDef) *Generator {
	g.mu.Lock()
	g.doc.SecurityDefinitions[name] = def
	g.mu.Unlock()
	return g
}

// AddTypeMap add rule to use dst interface instead of src
func (g *Generator) AddTypeMap(src interface{}, dst interface{}) *Generator {
	g.mu.Lock()
	g.typesMap[reflect.TypeOf(src)] = dst
	g.mu.Unlock()
	return g
}

func (g *Generator) getMappedType(t reflect.Type) (dst interface{}, found bool) {
	dst, found = g.typesMap[t]
	return
}

func (g *Generator) isJSONRPC() bool {
	serviceType, found := g.doc.data["x-service-type"]
	if !found {
		return false
	}

	t, isServiceType := serviceType.(ServiceType)
	return isServiceType && t == ServiceTypeJSONRPC
}

// genDocument returns document specification in JSON string (in []byte)
func (g *Generator) genDocument(host *string) ([]byte, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	// ensure that all definition in queue is parsed before generating
	g.parseDefInQueue()
	g.doc.Definitions = g.definitions.GenDefinitions()
	if g.host != "" || host == nil {
		g.doc.Host = g.host
	} else {
		g.doc.Host = *host
	}
	g.doc.Paths = make(map[string]PathItem)

	isJSONRPC := g.isJSONRPC()
	for path, item := range g.paths {
		if isJSONRPC {
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

	var (
		data []byte
		err  error
	)
	if g.indentJSON {
		data, err = json.MarshalIndent(g.doc, "", "  ")
	} else {
		data, err = json.Marshal(g.doc)
	}

	return data, err
}

// GenDocument returns document specification in JSON string (in []byte)
func (g *Generator) GenDocument() ([]byte, error) {
	// pass nil here to set host as g.host
	return g.genDocument(nil)
}

// ServeHTTP implements http.Handler to server swagger.json document
func (g *Generator) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	data, err := g.genDocument(&r.URL.Host)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")

	g.writeCORSHeaders(w)

	w.Write(data)
}
