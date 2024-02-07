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
		res = append(res, "      schema:")
		res = append(res, "        type: array")
		res = append(res, "        items:")
		typ := converter.CanonicalName(s.InputTypes[0])
		m := s.Messages[typ]
		if m == nil {
			oasType, oasFormat := toOpenAPITypeAndFormat(typ)
			res = append(res, fmt.Sprintf("          type: %s", oasType))
			res = append(res, fmt.Sprintf("          format: %s", oasFormat))
			return res
		}
		res = append(res, "          type: object")
		res = append(res, "          properties:")
		for _, f := range m.Fields {
			name := converter.ToSnakeCase(converter.CanonicalName(f.Name))
			res = append(res, fmt.Sprintf("            %s:", name))
			oasType, oasFormat := toOpenAPITypeAndFormat(f.Type)
			res = append(res, fmt.Sprintf("              type: %s", oasType))
			if oasFormat != "" {
				res = append(res, fmt.Sprintf("              format: %s", oasFormat))
			}
		}
		return res
	}

	if method == "GET" || method == "DELETE" {
		return res
	}

	res = append(res, "requestBody:")
	res = append(res, "  content:")
	res = append(res, "    application/json:")
	res = append(res, "      schema:")
	res = append(res, "        type: object")
	res = append(res, "        properties:")
	for i, typ := range s.InputTypes {
		m := s.Messages[converter.CanonicalName(typ)]
		if m == nil {
			name := converter.ToSnakeCase(converter.CanonicalName(s.InputNames[i]))
			if slices.Contains[[]string](pathParams, name) {
				continue
			}
			res = append(res, fmt.Sprintf("          %s:", name))
			oasType, oasFormat := toOpenAPITypeAndFormat(typ)
			res = append(res, fmt.Sprintf("            type: %s", oasType))
			if oasFormat != "" {
				res = append(res, fmt.Sprintf("            format: %s", oasFormat))
			}
			continue
		}
		for _, f := range m.Fields {
			name := converter.ToSnakeCase(converter.CanonicalName(f.Name))
			if slices.Contains[[]string](pathParams, name) {
				continue
			}
			res = append(res, fmt.Sprintf("          %s:", name))
			oasType, oasFormat := toOpenAPITypeAndFormat(f.Type)
			res = append(res, fmt.Sprintf("            type: %s", oasType))
			if oasFormat != "" {
				res = append(res, fmt.Sprintf("            format: %s", oasFormat))
			}
		}
	}

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
	Info             []string
	Tags             []string
	UserDefinedPaths []string
	ExtraDefinitions []string
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
	case "bool":
		return "boolean", ""
	case "sql.NullBool", "pgtype.Bool":
		return "*boolean", ""
	case "sql.NullInt32", "pgtype.Int4", "pgtype.Int2":
		return "integer", "int32"
	case "pgtype.Uint32":
		return "integer", "int32"
	case "int":
		return "integer", "int64"
	case "int64":
		return "integer", "int64"
	case "uint64":
		return "integer", "int64"
	case "int16":
		return "integer", "int32"
	case "int32":
		return "integer", "int32"
	case "uint16":
		return "integer", "int32"
	case "uint32":
		return "integer", "int32"
	case "sql.NullInt64", "pgtype.Int8":
		return "*integer", "int64"
	case "float32":
		return "number", "float"
	case "float64":
		return "number", "double"
	case "pgtype.Float4":
		return "number", "float"
	case "sql.NullFloat64", "pgtype.Float8":
		return "number", "double"
	case "sql.NullString", "pgtype.Text":
		return "string", ""
	case "time.Time":
		return "string", "date-time"
	case "sql.NullTime", "pgtype.Timestamp", "pgtype.Timestampz":
		return "string", "date-time"
	case "pgtype.Date":
		return "string", "date"
	case "string", "net.HardwareAddr", "net.IP":
		return "string", ""
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
