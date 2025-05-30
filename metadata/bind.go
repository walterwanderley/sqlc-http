package metadata

import (
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/walterwanderley/sqlc-grpc/converter"
	"github.com/walterwanderley/sqlc-grpc/metadata"
)

func HandlerTypes(s *metadata.Service, ui bool) []string {
	res := make([]string, 0)

	requestAttributes := RequestTypeAttributes(s)
	if len(requestAttributes) > 0 {
		res = append(res, "type request struct {")
		res = append(res, requestAttributes...)
		res = append(res, "}")
	}
	// UI generate a message
	if ui && (s.Output == "sql.Result" || s.Output == "pgconn.CommandTag") {
		return res
	}
	responseAttributes := ResponseTypeAttributes(s)
	if len(responseAttributes) > 0 {
		res = append(res, "type response struct {")
		res = append(res, responseAttributes...)
		res = append(res, "}")
	}

	return res
}

func RequestTypeAttributes(s *metadata.Service) []string {
	res := make([]string, 0)
	if !s.EmptyInput() {
		if s.HasCustomParams() {
			typ := s.InputTypes[0]
			m := s.Messages[converter.CanonicalName(typ)]
			for _, f := range m.Fields {
				attrName := converter.UpperFirstCharacter(f.Name)
				res = append(res, fmt.Sprintf("%s %s `json:\"%s\"`", attrName, toSerializableType(f.Type), converter.ToSnakeCase(attrName)))
			}
		} else {
			for i, name := range s.InputNames {
				attrName := converter.UpperFirstCharacter(name)
				typ := s.InputTypes[i]
				res = append(res, fmt.Sprintf("%s %s `json:\"%s\"`", attrName, toSerializableType(typ), converter.ToSnakeCase(attrName)))
			}
		}
	}
	return res
}

func ResponseTypeAttributes(s *metadata.Service) []string {
	res := make([]string, 0)
	if !s.EmptyOutput() {
		if s.Output == "sql.Result" {
			res = append(res, "LastInsertId int64 `json:\"last_insert_id\"`")
			res = append(res, "RowsAffected int64 `json:\"rows_affected\"`")
		}
		if s.Output == "pgconn.CommandTag" {
			res = append(res, "RowsAffected int64 `json:\"rows_affected\"`")
		}

		m, ok := s.Messages[converter.CanonicalName(s.Output)]
		if !ok {
			return res
		}
		for _, f := range m.Fields {
			attrName := converter.UpperFirstCharacter(f.Name)
			res = append(res, fmt.Sprintf("%s %s `json:\"%s,omitempty\"`", attrName, toSerializableType(f.Type), converter.ToSnakeCase(attrName)))
		}
	}
	return res
}

func InputHttp(s *metadata.Service) []string {
	res := make([]string, 0)
	if s.EmptyInput() {
		return res
	}

	method := HttpMethod(s)
	pathParams := httpPathParams(s)

	if method == "GET" || method == "DELETE" {
		res = append(res, "var req request")
		for i, typ := range s.InputTypes {
			m, ok := s.Messages[converter.CanonicalName(typ)]
			if ok {
				for _, f := range m.Fields {
					if slices.Contains(pathParams, converter.ToSnakeCase(f.Name)) {
						res = append(res, BindStringToSerializable("r.PathValue", "req", converter.UpperFirstCharacter(f.Name), f.Type)...)
						continue
					}
					res = append(res, BindStringToSerializable("r.URL.Query().Get", "req", converter.UpperFirstCharacter(f.Name), f.Type)...)
				}
				continue
			}
			if slices.Contains(pathParams, converter.ToSnakeCase(s.InputNames[i])) {
				res = append(res, BindStringToSerializable("r.PathValue", "req", converter.UpperFirstCharacter(s.InputNames[i]), typ)...)
				continue
			}
			res = append(res, BindStringToSerializable("r.URL.Query().Get", "req", converter.UpperFirstCharacter(s.InputNames[i]), typ)...)

		}
	} else {
		res = append(res, "req, err := server.Decode[request](r)")
		res = append(res, "if err != nil { http.Error(w, err.Error(), http.StatusUnprocessableEntity)")
		res = append(res, "return }")
		for i, typ := range s.InputTypes {
			m, ok := s.Messages[converter.CanonicalName(typ)]
			if ok {
				for _, f := range m.Fields {
					if slices.Contains(pathParams, converter.ToSnakeCase(f.Name)) {
						res = append(res, BindStringToSerializable("r.PathValue", "req", converter.UpperFirstCharacter(f.Name), f.Type)...)
						continue
					}
				}
				continue
			}
			if slices.Contains(pathParams, converter.ToSnakeCase(s.InputNames[i])) {
				res = append(res, BindStringToSerializable("r.PathValue", "req", converter.UpperFirstCharacter(s.InputNames[i]), typ)...)
				continue
			}
		}
	}

	if s.HasCustomParams() {
		typ := s.InputTypes[0]
		in := s.InputNames[0]
		if strings.HasPrefix(typ, "*") {
			res = append(res, fmt.Sprintf("%s := new(%s)", in, typ[1:]))
		} else {
			res = append(res, fmt.Sprintf("var %s %s", in, typ))
		}
		m := s.Messages[converter.CanonicalName(typ)]
		for _, f := range m.Fields {
			attrName := converter.UpperFirstCharacter(f.Name)
			res = append(res, bindToGo("req", fmt.Sprintf("%s.%s", in, attrName), attrName, f.Type, false)...)
		}
	} else {
		for i, n := range s.InputNames {
			res = append(res, bindToGo("req", n, converter.UpperFirstCharacter(n), s.InputTypes[i], true)...)
		}
	}

	return res
}

func OutputHttp(s *metadata.Service) []string {
	res := make([]string, 0)
	if s.EmptyOutput() {
		return res
	}
	m := s.Messages[converter.CanonicalName(s.Output)]
	if s.HasArrayOutput() {
		if m == nil {
			res = append(res, "server.Encode(w, r, http.StatusOK, map[string]any{\"list\": result})")
			return res
		}
		res = append(res, "res := make([]response, 0)")
		res = append(res, "for _, r := range result {")
		res = append(res, "var item response")
		for _, f := range m.Fields {
			attrName := converter.UpperFirstCharacter(f.Name)
			res = append(res, BindToSerializable("r", "item", attrName, f.Type)...)
		}
		res = append(res, "res = append(res, item)")
		res = append(res, "}")
		res = append(res, "server.Encode(w, r, http.StatusOK, res)")
		return res
	}

	if m != nil {
		res = append(res, "var res response")
		for _, f := range m.Fields {
			res = append(res, BindToSerializable("result", "res", converter.UpperFirstCharacter(f.Name), f.Type)...)
		}
		res = append(res, "server.Encode(w, r, http.StatusOK, res)")
		return res
	}

	if s.Output == "sql.Result" {
		res = append(res, "lastInsertId, _ := result.LastInsertId()")
		res = append(res, "rowsAffected, _ := result.RowsAffected()")
		res = append(res, "server.Encode(w, r, http.StatusOK, response{")
		res = append(res, "LastInsertId: lastInsertId,")
		res = append(res, "RowsAffected: rowsAffected,")
		res = append(res, "})")
		return res
	}

	if s.Output == "pgconn.CommandTag" {
		res = append(res, "server.Encode(w, r, http.StatusOK, response{")
		res = append(res, "RowsAffected: result.RowsAffected(),")
		res = append(res, "})")
		return res
	}

	res = append(res, "server.Encode(w, r, http.StatusOK, map[string]any{\"value\": result})")

	return res
}

func HttpMethod(s *metadata.Service) string {
	if len(s.HttpSpecs) > 0 {
		return strings.ToUpper(s.HttpSpecs[0].Method)
	}
	return strings.ToUpper(s.HttpMethod())
}

func HttpPath(s *metadata.Service) string {
	path := s.HttpPath()
	if len(s.HttpSpecs) > 0 {
		path = s.HttpSpecs[0].Path
	}
	path = strings.TrimPrefix(path, "/")
	return "/" + path
}

func httpPathParams(s *metadata.Service) []string {
	re := regexp.MustCompile("{(.*?)}")
	params := re.FindAllString(HttpPath(s), 100)
	res := make([]string, 0)
	for _, p := range params {
		if len(p) <= 2 || p == "{$}" {
			continue
		}
		res = append(res, strings.TrimSuffix(strings.TrimSuffix(strings.TrimPrefix(p, "{"), "}"), "..."))
	}
	return res
}
