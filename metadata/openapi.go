package metadata

import (
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/walterwanderley/sqlc-grpc/converter"
	"github.com/walterwanderley/sqlc-grpc/metadata"
)

func GroupByPath(pkg *metadata.Package) map[string][]*metadata.Service {
	paths := make(map[string][]*metadata.Service)
	for _, s := range pkg.Services {
		path := HttpPath(s)
		services, ok := paths[path]
		if !ok {
			services := []*metadata.Service{s}
			paths[path] = services
			continue
		}
		paths[path] = append(services, s)

	}
	return paths
}

func ApiParameters(s *metadata.Service) []string {
	res := make([]string, 0)
	if s.EmptyInput() {
		return res
	}

	method := HttpMethod(s)
	pathParams := httpPathParams(s)

	var hasParametersAttribute bool
	for i, typ := range s.InputTypes {
		m, ok := s.Messages[converter.CanonicalName(typ)]
		if ok {
			for _, f := range m.Fields {
				name := converter.ToSnakeCase(converter.CanonicalName(f.Name))
				inType := "query"
				if slices.Contains[[]string](pathParams, name) {
					inType = "path"
				}
				if inType == "query" && (method != "GET" && method != "DELETE") {
					continue
				}
				if !hasParametersAttribute {
					hasParametersAttribute = true
					res = append(res, "parameters:")
				}
				res = append(res, fmt.Sprintf("  - name: %s", name))
				res = append(res, fmt.Sprintf("    in: %s", inType))
				res = append(res, "    schema:")
				oasType, oasFormat := toOpenAPITypeAndFormat(f.Type)
				res = append(res, fmt.Sprintf("      type: %s", oasType))
				if oasFormat != "" {
					res = append(res, fmt.Sprintf("      format: %s", oasFormat))
				}
			}
			continue
		}
		name := converter.ToSnakeCase(converter.CanonicalName(s.InputNames[i]))
		inType := "query"
		if slices.Contains[[]string](pathParams, name) {
			inType = "path"
		}
		if inType == "query" && (method != "GET" && method != "DELETE") {
			continue
		}
		if !hasParametersAttribute {
			hasParametersAttribute = true
			res = append(res, "parameters:")
		}
		res = append(res, fmt.Sprintf("  - name: %s", name))
		res = append(res, fmt.Sprintf("    in: %s", inType))
		res = append(res, "    schema:")
		oasType, oasFormat := toOpenAPITypeAndFormat(typ)
		res = append(res, fmt.Sprintf("      type: %s", oasType))
		if oasFormat != "" {
			res = append(res, fmt.Sprintf("      format: %s", oasFormat))
		}

	}

	if s.HasArrayParams() {
		res = append(res, "requestBody:")
		res = append(res, "  content:")
		res = append(res, "    application/json:")

		schemaRes := make([]string, 0)

		schemaRes = append(schemaRes, "      schema:")
		schemaRes = append(schemaRes, "        type: array")
		schemaRes = append(schemaRes, "        items:")
		typ := converter.CanonicalName(s.InputTypes[0])
		m := s.Messages[typ]
		if m == nil {
			oasType, oasFormat := toOpenAPITypeAndFormat(typ)
			schemaRes = append(schemaRes, fmt.Sprintf("          type: %s", oasType))
			schemaRes = append(schemaRes, fmt.Sprintf("          format: %s", oasFormat))
			res = append(res, schemaRes...)
			res = append(res, "    application/x-www-form-urlencoded:")
			res = append(res, schemaRes...)
			return res
		}
		schemaRes = append(schemaRes, "          type: object")
		schemaRes = append(schemaRes, "          properties:")
		for _, f := range m.Fields {
			name := converter.ToSnakeCase(converter.CanonicalName(f.Name))
			schemaRes = append(schemaRes, fmt.Sprintf("            %s:", name))
			oasType, oasFormat := toOpenAPITypeAndFormat(f.Type)
			schemaRes = append(schemaRes, fmt.Sprintf("              type: %s", oasType))
			if oasFormat != "" {
				schemaRes = append(schemaRes, fmt.Sprintf("              format: %s", oasFormat))
			}
		}
		res = append(res, schemaRes...)
		res = append(res, "    application/x-www-form-urlencoded:")
		res = append(res, schemaRes...)
		return res
	}

	if method == "GET" || method == "DELETE" {
		return res
	}

	res = append(res, "requestBody:")
	res = append(res, "  content:")
	res = append(res, "    application/json:")

	schemaRes := make([]string, 0)
	schemaRes = append(schemaRes, "      schema:")
	schemaRes = append(schemaRes, "        type: object")
	schemaRes = append(schemaRes, "        properties:")
	for i, typ := range s.InputTypes {
		m := s.Messages[converter.CanonicalName(typ)]
		if m == nil {
			name := converter.ToSnakeCase(converter.CanonicalName(s.InputNames[i]))
			if slices.Contains[[]string](pathParams, name) {
				continue
			}
			schemaRes = append(schemaRes, fmt.Sprintf("          %s:", name))
			oasType, oasFormat := toOpenAPITypeAndFormat(typ)
			schemaRes = append(schemaRes, fmt.Sprintf("            type: %s", oasType))
			if oasFormat != "" {
				schemaRes = append(schemaRes, fmt.Sprintf("            format: %s", oasFormat))
			}
			continue
		}
		for _, f := range m.Fields {
			name := converter.ToSnakeCase(converter.CanonicalName(f.Name))
			if slices.Contains[[]string](pathParams, name) {
				continue
			}
			schemaRes = append(schemaRes, fmt.Sprintf("          %s:", name))
			oasType, oasFormat := toOpenAPITypeAndFormat(f.Type)
			schemaRes = append(schemaRes, fmt.Sprintf("            type: %s", oasType))
			if oasFormat != "" {
				schemaRes = append(schemaRes, fmt.Sprintf("            format: %s", oasFormat))
			}
		}
	}
	res = append(res, schemaRes...)
	res = append(res, "    application/x-www-form-urlencoded:")
	res = append(res, schemaRes...)

	return res
}

func ApiResponse(s *metadata.Service) []string {
	res := make([]string, 0)
	if s.EmptyOutput() {
		return res
	}
	res = append(res, "content:")
	res = append(res, "  application/json:")
	res = append(res, "    schema:")
	m := s.Messages[converter.CanonicalName(s.Output)]
	if s.HasArrayOutput() {
		res = append(res, "      type: array")
		res = append(res, "      items:")
		if m == nil {
			oasType, oasFormat := toOpenAPITypeAndFormat(converter.CanonicalName(s.Output))
			res = append(res, fmt.Sprintf("        type: %s", oasType))
			res = append(res, fmt.Sprintf("        format: %s", oasFormat))
			return res
		}

		res = append(res, fmt.Sprintf("        $ref: \"#/components/schemas/%s%s\"", m.PackageName, m.Name))
		return res
	}

	if m != nil {
		res = append(res, fmt.Sprintf("      $ref: \"#/components/schemas/%s%s\"", m.PackageName, m.Name))
		return res
	}

	if s.Output == "sql.Result" {
		res = append(res, "      type: object")
		res = append(res, "      properties:")
		res = append(res, "        last_insert_id:")
		res = append(res, "          type: integer")
		res = append(res, "          format: int64")
		res = append(res, "        rows_affected:")
		res = append(res, "          type: integer")
		res = append(res, "          format: int64")
		return res
	}

	if s.Output == "pgconn.CommandTag" {
		res = append(res, "      type: object")
		res = append(res, "      properties:")
		res = append(res, "        rows_affected:")
		res = append(res, "          type: integer")
		res = append(res, "          format: int64")
		return res
	}

	return res
}

func ApiComponentSchemas(pkg *metadata.Package) []string {
	res := make([]string, 0)
	messages := make([]*metadata.Message, 0)
	for _, s := range pkg.Services {
		if s.EmptyOutput() {
			continue
		}
		m, ok := s.Messages[converter.CanonicalName(s.Output)]
		if !ok {
			continue
		}
		if !slices.Contains[[]*metadata.Message](messages, m) {
			messages = append(messages, m)
		}
	}
	slices.SortFunc[[]*metadata.Message](messages, func(a, b *metadata.Message) int {
		return strings.Compare(a.Name, b.Name)
	})
	for _, m := range messages {
		res = append(res, fmt.Sprintf("%s%s:", m.PackageName, m.Name))
		for _, line := range m.CustomProtoOptions {
			res = append(res, fmt.Sprintf("  %s", line))
		}
		res = append(res, "  type: object")
		res = append(res, "  properties:")
		for _, f := range m.Fields {
			name := converter.ToSnakeCase(converter.CanonicalName(f.Name))
			res = append(res, fmt.Sprintf("    %s:", name))
			oasType, oasFormat := toOpenAPITypeAndFormat(f.Type)
			res = append(res, fmt.Sprintf("      type: %s", oasType))
			if oasFormat != "" {
				res = append(res, fmt.Sprintf("      format: %s", oasFormat))
			}
		}
	}
	return res
}

type EditableOpenApi struct {
	*metadata.Definition
	Info                  []string
	Tags                  []string
	UserDefinedPaths      []string
	UserDefinedSchemas    []string
	UserDefinedComponents []string
	ExtraDefinitions      []string
}

func LoadOpenApi(path string, append bool, def *metadata.Definition) (*EditableOpenApi, error) {
	var editable EditableOpenApi
	editable.Definition = def
	_, err := os.Stat(path)
	fileExists := !errors.Is(err, os.ErrNotExist)
	if !append || !fileExists {
		return &editable, nil
	}

	openapiFile, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var openApiMap map[string]any
	err = yaml.Unmarshal(openapiFile, &openApiMap)
	if err != nil {
		return nil, err
	}

	info, ok := openApiMap["info"]
	if ok {
		out, err := yaml.Marshal(info)
		if err != nil {
			return nil, err
		}
		editable.Info = strings.Split(string(out), "\n")
	}
	tags, ok := openApiMap["tags"]
	if ok {
		out, err := yaml.Marshal(tags)
		if err != nil {
			return nil, err
		}
		editable.Tags = strings.Split(string(out), "\n")
	}
	paths, ok := openApiMap["paths"]
	if ok {
		pathsMap, ok := paths.(map[string]any)
		if ok {
			for path, operations := range pathsMap {
				if containsPath(def, path) {
					operationsMap, ok := operations.(map[string]any)
					if ok {
						for method, operation := range operationsMap {
							svc, ok := lookupService(def, method, path)
							if !ok {
								continue
							}
							operationMap, ok := operation.(map[string]any)
							if !ok {
								continue
							}
							delete(operationMap, "parameters")
							delete(operationMap, "requestBody")
							delete(operationMap, "responses")
							if len(operationMap) > 0 {
								out, err := yaml.Marshal(operationMap)
								if err != nil {
									return nil, err
								}
								svc.CustomProtoOptions = strings.Split(string(out), "\n")
							}
						}
					}
					delete(pathsMap, path)
				}
			}
			if len(pathsMap) > 0 {
				out, err := yaml.Marshal(pathsMap)
				if err != nil {
					return nil, err
				}
				editable.UserDefinedPaths = strings.Split(string(out), "\n")
			}
		}
	}
	components, ok := openApiMap["components"]
	if ok {
		componentsMap, ok := components.(map[string]any)
		if ok {
			schemas, ok := componentsMap["schemas"]
			if ok {
				schemasMap, ok := schemas.(map[string]any)
				if ok {
					for packageStruct, schema := range schemasMap {
						if msg, ok := lookupMessage(def, packageStruct); ok {
							schemaMap, ok := schema.(map[string]any)
							if ok {
								delete(schemaMap, "type")
								delete(schemaMap, "properties")
								if len(schemaMap) > 0 {
									out, err := yaml.Marshal(schemaMap)
									if err != nil {
										return nil, err
									}
									msg.CustomProtoOptions = strings.Split(string(out), "\n")
								}

							}
							delete(schemasMap, packageStruct)
						}
					}
					if len(schemasMap) > 0 {
						out, err := yaml.Marshal(schemasMap)
						if err != nil {
							return nil, err
						}
						editable.UserDefinedSchemas = strings.Split(string(out), "\n")
					}
				}

			}
			delete(componentsMap, "schemas")
			if len(componentsMap) > 0 {
				out, err := yaml.Marshal(componentsMap)
				if err != nil {
					return nil, err
				}
				editable.UserDefinedComponents = strings.Split(string(out), "\n")
			}
		}
	}
	delete(openApiMap, "openapi")
	delete(openApiMap, "info")
	delete(openApiMap, "tags")
	delete(openApiMap, "paths")
	delete(openApiMap, "components")
	if len(openApiMap) > 0 {
		out, err := yaml.Marshal(openApiMap)
		if err != nil {
			return nil, err
		}
		editable.ExtraDefinitions = strings.Split(string(out), "\n")
	}

	return &editable, nil
}

func toOpenAPITypeAndFormat(typ string) (oasType string, oasFormat string) {
	if strings.HasPrefix(typ, "*") {
		return toOpenAPITypeAndFormat(typ[1:])
	}
	if strings.HasPrefix(typ, "[]") {
		return toOpenAPITypeAndFormat(typ[2:])
	}
	switch typ {
	case "json.RawMessage":
		return "string", "byte"
	case "byte":
		return "string", "binary"
	case "bool", "sql.NullBool", "pgtype.Bool":
		return "boolean", ""
	case "sql.NullInt32", "pgtype.Int4", "pgtype.Int2", "pgtype.Uint32", "int16", "int32", "uint16", "uint32":
		return "integer", "int32"
	case "int", "int64", "uint64", "sql.NullInt64", "pgtype.Int8":
		return "integer", "int64"
	case "float32", "pgtype.Float4":
		return "number", "float"
	case "float64", "sql.NullFloat64", "pgtype.Float8":
		return "number", "double"
	case "time.Time", "sql.NullTime", "pgtype.Timestamp", "pgtype.Timestampz":
		return "string", "date-time"
	case "pgtype.Date":
		return "string", "date"
	case "pgtype.UUID", "uuid.UUID":
		return "string", "uuid"
	default:
		return "string", ""
	}
}

func containsPath(def *metadata.Definition, path string) bool {
	for _, pkg := range def.Packages {
		for _, s := range pkg.Services {
			if path == HttpPath(s) {
				return true
			}
		}
	}
	return false
}

func lookupMessage(def *metadata.Definition, packageStruct string) (*metadata.Message, bool) {
	for _, pkg := range def.Packages {
		for _, msg := range pkg.Messages {
			if msg.PackageName+msg.Name == packageStruct {
				return msg, true
			}
		}
	}
	return nil, false
}

func lookupService(def *metadata.Definition, method, path string) (*metadata.Service, bool) {
	method = strings.ToUpper(method)
	for _, pkg := range def.Packages {
		for _, s := range pkg.Services {
			if path == HttpPath(s) && method == HttpMethod(s) {
				return s, true
			}
		}
	}
	return nil, false
}
