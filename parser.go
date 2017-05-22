package swgen

import (
	"encoding"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	refDefinitionPrefix = "#/definitions/"
)

var (
	typeOfJSONRawMsg      = reflect.TypeOf((*json.RawMessage)(nil)).Elem()
	typeOfTime            = reflect.TypeOf((*time.Time)(nil)).Elem()
	typeOfTextUnmarshaler = reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()
)

// IParameter allows to return custom parameters
type IParameter interface {
	SwgenParameter() (name string, params []ParamObj, err error)
}

// IDefinition allows to return custom definitions
type IDefinition interface {
	SwgenDefinition() (typeName string, typeDef SchemaObj, err error)
}

func (g *Generator) addDefinition(name string, def SchemaObj) {
	g.defMux.Lock()
	defer g.defMux.Unlock()
	g.definitions[name] = def
}

func (g *Generator) defExists(name string) (b bool) {
	g.defMux.Lock()
	defer g.defMux.Unlock()

	_, b = g.definitions[name]
	return b
}

func (g *Generator) addToDefQueue(name string, i interface{}) {
	g.queueMux.Lock()
	defer g.queueMux.Unlock()

	g.defQueue[name] = i
}

func (g *Generator) defInQueue(name string) (b bool) {
	g.queueMux.Lock()
	defer g.queueMux.Unlock()

	_, b = g.defQueue[name]
	return b
}

func (g *Generator) getDefinition(name string) (SchemaObj, bool) {
	def, ok := g.definitions[name]

	return def, ok
}

func (g *Generator) deleteDefinition(name string) {
	g.queueMux.Lock()
	defer g.queueMux.Unlock()

	delete(g.definitions, name)
}

//
// Parse swagger schema object
// see http://swagger.io/specification/#schemaObject
//

// ResetDefinitions will remove all exists definitions and init again
func (g *Generator) ResetDefinitions() {
	g.defMux.Lock()
	g.definitions = make(map[string]SchemaObj)
	g.defMux.Unlock()

	g.queueMux.Lock()
	g.defQueue = make(map[string]interface{})
	g.queueMux.Unlock()
}

// ResetDefinitions will remove all exists definitions and init again
func ResetDefinitions() {
	gen.ResetDefinitions()
}

func (g *Generator) parse(i interface{}, t reflect.Type) (schema SchemaObj, err error) {
	var (
		typeName string
		typeDef  SchemaObj
	)

	if definition, ok := i.(IDefinition); ok {
		typeName, typeDef, err = definition.SwgenDefinition()
		if err != nil {
			return typeDef, err
		}
		if typeName == "" {
			typeName = t.Name()
		}
		typeDef.TypeName = typeName
		if g.defExists(typeName) {
			return SchemaObj{Ref: refDefinitionPrefix + typeName, TypeName: typeName}, nil
		}
		defer g.parseDefInQueue()
		if g.reflectGoTypes {
			typeDef.GoType = goType(t)
		}
		g.addDefinition(typeName, typeDef)

		return SchemaObj{Ref: refDefinitionPrefix + typeName, TypeName: typeName}, nil
	}

	if newInterface, ok := g.getTypeMapByString(t.String()); ok {
		return g.ParseDefinition(newInterface)
	}

	// if pointer get the underlying element
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Name() == "" {
		typeDef = g.genSchemaForType(t)
		return typeDef, nil
	}

	switch t.Kind() {
	case reflect.Struct:
		typeName = t.Name()

		if g.defExists(typeName) {
			return SchemaObj{Ref: refDefinitionPrefix + typeName, TypeName: typeName}, nil
		}

		typeDef = SchemaObj{Type: "object"}
		typeDef.Properties = g.parseDefinitionProperties(t, &typeDef)
	case reflect.Slice, reflect.Array:
		elemType := t.Elem()
		if elemType.Kind() == reflect.Ptr {
			elemType = elemType.Elem()
		}

		typeName = t.Name()

		if g.defExists(typeName) {
			return SchemaObj{Ref: refDefinitionPrefix + typeName, TypeName: typeName}, nil
		}

		var itemSchema SchemaObj
		if elemType.Kind() != reflect.Struct || (elemType.Kind() == reflect.Struct && elemType.Name() != "") {
			itemSchema = g.genSchemaForType(elemType)
		} else {
			itemSchema = SchemaObj{
				Type: "object",
			}
			itemSchema.Properties = g.parseDefinitionProperties(elemType, &itemSchema)
		}

		typeDef = SchemaObj{Type: "array"}
		typeDef.Items = &itemSchema
	case reflect.Map:
		typeName = t.Name()
		typeDef = SchemaObj{Type: "object"}

		elemType := t.Elem()
		if elemType.Kind() == reflect.Ptr {
			elemType = elemType.Elem()
		}

		if g.defExists(typeName) {
			return SchemaObj{Ref: refDefinitionPrefix + typeName, TypeName: typeName}, nil
		}

		itemDef := g.genSchemaForType(elemType)
		typeDef.AdditionalProperties = &itemDef
	default:
		typeDef = g.genSchemaForType(t)
		typeDef.TypeName = typeDef.Type
		return typeDef, nil
	}

	defer g.parseDefInQueue()

	if g.reflectGoTypes {
		typeDef.GoType = goType(t)
	}
	g.addDefinition(typeName, typeDef)
	return SchemaObj{Ref: refDefinitionPrefix + typeName, TypeName: typeName}, nil
}

func goType(t reflect.Type) (s string) {
	s = t.Name()
	pkgPath := t.PkgPath()
	if pkgPath != "" {
		pos := strings.Index(pkgPath, "/vendor/")
		if pos != -1 {
			pkgPath = pkgPath[pos+8:]
		}
		s = pkgPath + "." + s
	}

	ts := t.String()
	typeRef := s

	pos := strings.LastIndex(typeRef, "/")
	if pos != -1 {
		typeRef = typeRef[pos+1:]
	}

	if typeRef != ts {
		s = s + "::" + t.String()
	}

	switch t.Kind() {
	case reflect.Slice:
		return "[]" + goType(t.Elem())
	case reflect.Ptr:
		return "*" + goType(t.Elem())
	case reflect.Map:
		return "map[" + goType(t.Key()) + "]" + goType(t.Elem())
	}

	return
}

// ParseDefinition create a DefObj from input object, it should be a non-nil pointer to anything
// it reuse schema/json tag for property name.
func (g *Generator) ParseDefinition(i interface{}) (schema SchemaObj, err error) {
	return g.parse(i, reflect.TypeOf(i))
}

func (g *Generator) parseDefinitionProperties(t reflect.Type, parent *SchemaObj) map[string]SchemaObj {
	properties := make(map[string]SchemaObj, t.NumField())
	if g.reflectGoTypes {
		parent.GoPropertyNames = make(map[string]string, t.NumField())
		parent.GoPropertyTypes = make(map[string]string, t.NumField())
	}

	for i := 0; i < t.NumField(); i = i + 1 {
		field := t.Field(i)

		// we can't access the value of un-exportable field
		if field.PkgPath != "" {
			continue
		}

		if field.Anonymous {
			fieldProperties := g.parseDefinitionProperties(field.Type, parent)
			for propertyName, property := range fieldProperties {
				properties[propertyName] = property
			}
			continue
		}

		// don't check if it's omitted
		var tag string
		if tag = field.Tag.Get("json"); tag == "-" || tag == "" {
			continue
		}

		propName := strings.Split(tag, ",")[0]
		var (
			obj SchemaObj
		)

		if dataType := field.Tag.Get("swgen_type"); dataType != "" {
			obj = SchemaFromCommonName(commonName(dataType))
		} else {
			obj = g.genSchemaForType(field.Type)
		}

		if defaultTag := field.Tag.Get("default"); defaultTag != "" {
			if defaultValue, err := g.caseDefaultValue(field.Type, defaultTag); err == nil {
				obj.Default = defaultValue
			}
		}
		if g.reflectGoTypes {
			if obj.Ref == "" {
				obj.GoType = goType(field.Type)
			}
			parent.GoPropertyNames[propName] = field.Name
			parent.GoPropertyTypes[propName] = goType(field.Type)
		}

		properties[propName] = obj
	}

	return properties
}

func (g *Generator) caseDefaultValue(t reflect.Type, defaultValue string) (interface{}, error) {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	kind := t.Kind()

	switch kind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.ParseInt(defaultValue, 10, 64)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.ParseUint(defaultValue, 10, 64)
	case reflect.Float32, reflect.Float64:
		return strconv.ParseFloat(defaultValue, 64)
	case reflect.String:
		return defaultValue, nil
	case reflect.Bool:
		return strconv.ParseBool(defaultValue)
	default:
		instance := reflect.New(t).Interface()
		if err := json.Unmarshal([]byte(defaultValue), instance); err != nil {
			return nil, err
		}
		return reflect.Indirect(reflect.ValueOf(instance)).Interface(), nil
	}
}

// ParseDefinition create a DefObj from input object, it should be a pointer to a struct,
// it reuse schema/json tag for property name.
func ParseDefinition(i interface{}) (typeDef SchemaObj, err error) {
	return gen.ParseDefinition(i)
}

func (g *Generator) parseDefInQueue() {
	g.queueMux.Lock()
	length := len(g.defQueue)

	if length == 0 {
		g.queueMux.Unlock()
		return
	}

	done := make(chan string, length)
	for _, i := range g.defQueue {
		go func(i interface{}) {
			typeDef, _ := g.ParseDefinition(i)
			done <- typeDef.TypeName
		}(i)
	}

	g.queueMux.Unlock()

	for i := 0; i < length; i = i + 1 {
		name := <-done
		g.queueMux.Lock()
		delete(g.defQueue, name)
		g.queueMux.Unlock()
	}

	close(done)
}

func (g *Generator) genSchemaForType(fType reflect.Type) SchemaObj {
	for fType.Kind() == reflect.Ptr {
		fType = fType.Elem()
	}

	smObj := SchemaObj{}
	switch fType.Kind() {
	case reflect.Bool:
		smObj = SchemaFromCommonName(CommonNameBoolean)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Uint, reflect.Uint8, reflect.Uint16:
		smObj = SchemaFromCommonName(CommonNameInteger)
	case reflect.Int64, reflect.Uint32, reflect.Uint64:
		smObj = SchemaFromCommonName(CommonNameLong)
	case reflect.Float32:
		smObj = SchemaFromCommonName(CommonNameFloat)
	case reflect.Float64:
		smObj = SchemaFromCommonName(CommonNameDouble)
	case reflect.String:
		smObj = SchemaFromCommonName(CommonNameString)
	case reflect.Array, reflect.Slice:
		if fType != typeOfJSONRawMsg {
			smObj.Type = "array"
			itemSchema := g.genSchemaForType(fType.Elem())
			smObj.Items = &itemSchema
		}
	case reflect.Map:
		smObj.Type = "object"
		itemSchema := g.genSchemaForType(fType.Elem())
		smObj.AdditionalProperties = &itemSchema
	case reflect.Struct:
		switch {
		case fType == typeOfTime:
			smObj = SchemaFromCommonName(CommonNameDateTime)
		case reflect.PtrTo(fType).Implements(typeOfTextUnmarshaler):
			smObj.Type = "string"
		default:
			name := fType.Name()
			smObj.Ref = refDefinitionPrefix + name
			if !g.defExists(name) || !g.defInQueue(name) {
				stcInterface := reflect.Zero(fType).Interface()
				g.addToDefQueue(name, stcInterface)
			}
		}
	case reflect.Interface:
		if fType.NumMethod() > 0 {
			panic("Non-empty interface is not supported: " + fType.String())
		}
	default:
		panic(fmt.Sprintf("type %s is not supported: %s", fType.Kind(), fType.String()))
	}

	if g.reflectGoTypes && smObj.Ref == "" {
		smObj.GoType = goType(fType)
	}

	return smObj
}

//
// Parse struct to swagger parameter object of operation object
// see http://swagger.io/specification/#parameterObject
//

// ParseParameter parse input struct to swagger parameter object
func (g *Generator) ParseParameter(i interface{}) (name string, params []ParamObj, err error) {
	if param, ok := i.(IParameter); ok {
		return param.SwgenParameter()
	}

	v := reflect.ValueOf(i)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		err = errors.New("input must be a struct")
		return
	}

	if newInterface, ok := g.getTypeMapByString(v.Type().String()); ok {
		return g.ParseParameter(newInterface)
	}

	sType := v.Type()

	name = sType.Name()
	params = []ParamObj{}

	for i := 0; i < sType.NumField(); i = i + 1 {
		field := sType.Field(i)
		// we can't access the value of un-exportable or anonymous fields
		if field.PkgPath != "" || field.Anonymous {
			continue
		}

		// don't check if it's omitted
		var nameTag string

		var inPath bool
		if nameTag = field.Tag.Get("schema"); nameTag == "-" || nameTag == "" {
			inPath = true
			if nameTag = field.Tag.Get("path"); nameTag == "-" || nameTag == "" {
				continue
			}
		}

		paramName := strings.Split(nameTag, ",")[0]
		param := ParamObj{}
		if g.reflectGoTypes {
			param.AddExtendedField("x-go-name", field.Name)
			param.AddExtendedField("x-go-type", goType(field.Type))
		}

		param.Name = paramName

		if e, isEnumer := reflect.Zero(field.Type).Interface().(enumer); isEnumer {
			param.Enum.Enum, param.Enum.EnumNames = e.GetEnumSlices()
		}

		if descTag := field.Tag.Get("description"); descTag != "-" && descTag != "" {
			param.Description = descTag
		}

		if reqTag := field.Tag.Get("required"); reqTag == "-" || reqTag == "false" {
			param.Required = false
		} else {
			param.Required = true
		}

		if inTag := field.Tag.Get("in"); inTag != "-" && inTag != "" {
			param.In = inTag // todo: validate IN value
		} else if inPath {
			param.In = "path"
		} else {
			param.In = "query"
		}

		var schema SchemaObj
		if swGenType := field.Tag.Get("swgen_type"); swGenType != "" {
			schema = SchemaFromCommonName(commonName(swGenType))
		} else {
			if newInterface, ok := g.getTypeMapByString(field.Type.String()); ok {
				schema = g.genSchemaForType(reflect.TypeOf(newInterface))
			} else {
				schema = g.genSchemaForType(field.Type)
			}
		}

		if schema.Type == "" {
			panic("dont support struct " + v.Type().Name() + " in property " + field.Name + " of parameter struct")
		}

		param.Type = schema.Type
		param.Format = schema.Format

		if schema.Type == "array" && schema.Items != nil {
			if schema.Items.Ref != "" || schema.Items.Type == "array" {
				panic("dont support array of struct or nested array in parameter")
			}

			param.Items = &ParamItemObj{
				Type:   schema.Items.Type,
				Format: schema.Items.Format,
			}
			param.CollectionFormat = "multi" // default for now
		}

		params = append(params, param)
	}

	return
}

// ParseParameter parse input struct to swagger parameter object
func ParseParameter(i interface{}) (name string, params []ParamObj, err error) {
	return gen.ParseParameter(i)
}

// ResetPaths remove all current paths
func (g *Generator) ResetPaths() {
	g.pathsMux.Lock()
	g.paths = make(map[string]PathItem)
	g.pathsMux.Unlock()
}

// ResetPaths remove all current paths
func ResetPaths() {
	gen.ResetPaths()
}

var regexFindPathParameter = regexp.MustCompile(`\{([^}:]+)(:[^\/]+)?(?:\})`)

// SetPathItem register path item with some information and input, output
func (g *Generator) SetPathItem(info PathItemInfo, params interface{}, body interface{}, response interface{}) error {
	var (
		item  PathItem
		found bool
	)

	pathParametersSubmatches := regexFindPathParameter.FindAllStringSubmatch(info.Path, -1)
	if len(pathParametersSubmatches) > 0 {
		for _, submatch := range pathParametersSubmatches {
			if submatch[2] != "" { // Remove gorilla.Mux-style regexp in path
				info.Path = strings.Replace(info.Path, submatch[0], "{"+submatch[1]+"}", 1)
			}
		}
	}

	g.pathsMux.RLock()
	item, found = g.paths[info.Path]
	g.pathsMux.RUnlock()

	if found && item.HasMethod(info.Method) {
		return nil
	}

	if !found {
		item = PathItem{}
	}

	operationObj := &OperationObj{}
	operationObj.Summary = info.Title
	operationObj.Description = info.Description
	operationObj.Deprecated = info.Deprecated
	operationObj.additionalData = info.additionalData
	if info.Tag != "" {
		operationObj.Tags = []string{info.Tag}
	}

	if params != nil {
		if g.reflectGoTypes {
			operationObj.AddExtendedField("x-request-go-type", goType(reflect.TypeOf(params)))
		}

		if _, params, err := g.ParseParameter(params); err == nil {
			operationObj.Parameters = params
		} else {
			return err
		}
	}

	operationObj.Responses = g.parseResponseObject(response)

	if body != nil {
		if g.reflectGoTypes {
			operationObj.AddExtendedField("x-request-go-type", goType(reflect.TypeOf(body)))
		}

		typeDef, err := g.ParseDefinition(body)
		if err != nil {
			return err
		}

		if !g.isSchemaObjectEmpty(typeDef) {
			param := ParamObj{
				Name:     "body",
				In:       "body",
				Required: true,
				Schema:   &typeDef,
			}

			if operationObj.Parameters == nil {
				operationObj.Parameters = make([]ParamObj, 0, 1)
			}

			operationObj.Parameters = append(operationObj.Parameters, param)
		} else {
			g.deleteDefinition(typeDef.TypeName)
		}
	}

	switch strings.ToUpper(info.Method) {
	case "GET":
		item.Get = operationObj
	case "POST":
		item.Post = operationObj
	case "PUT":
		item.Put = operationObj
	case "DELETE":
		item.Delete = operationObj
	case "OPTIONS":
		item.Options = operationObj
	case "HEAD":
		item.Head = operationObj
	case "PATCH":
		item.Patch = operationObj
	}

	g.pathsMux.Lock()
	g.paths[info.Path] = item
	g.pathsMux.Unlock()

	return nil
}

func (g *Generator) isSchemaObjectEmpty(obj SchemaObj) bool {
	if obj.Ref != "" {
		if def, ok := g.getDefinition(obj.TypeName); ok {
			return g.isSchemaObjectEmpty(def)
		} else if !g.defInQueue(obj.TypeName) {
			return true
		}
	}

	if _, ok := commonNamesMap[commonName(obj.TypeName)]; ok {
		return false
	}

	switch obj.Type {
	case "object":
		return len(obj.Properties) == 0
	case "array":
		return obj.Items == nil
	default:
		return len(obj.Properties) == 0 && obj.AdditionalProperties == nil && obj.Format == ""
	}
}

// SetPathItem register path item with some information and input, output
func SetPathItem(info PathItemInfo, params interface{}, body interface{}, response interface{}) error {
	return gen.SetPathItem(info, params, body, response)
}

func (g *Generator) parseResponseObject(responseObj interface{}) (res Responses) {
	res = make(Responses)

	if responseObj != nil {
		schema, err := g.ParseDefinition(responseObj)
		if err != nil {
			panic(fmt.Sprintf("could not create schema object for response %v", responseObj))
		}
		// since we only response json object
		// so, type of response object is always object
		resObj := ResponseObj{Description: "request success", Schema: &schema}
		res["200"] = resObj
	} else {
		res["200"] = ResponseObj{
			Description: "request success",
			Schema:      &SchemaObj{Type: "null"},
		}
	}

	return res
}
