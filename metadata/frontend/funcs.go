package frontend

import (
	"regexp"
	"slices"
	"strings"
	"unicode"

	"github.com/walterwanderley/sqlc-grpc/converter"
	"github.com/walterwanderley/sqlc-grpc/metadata"

	httpmetadata "github.com/walterwanderley/sqlc-http/metadata"
)

func AddSpace(s string) string {
	var result []rune
	for i, r := range s {
		if i > 0 && unicode.IsUpper(r) && unicode.IsLower(rune(s[i-1])) {
			result = append(result, ' ')
		}
		result = append(result, r)
	}
	return string(result)
}

func HasPagination(s *metadata.Service) bool {
	if _, ok := s.Messages[converter.CanonicalName(s.Output)]; !ok {
		return false
	}
	if !s.HasArrayOutput() {
		return false
	}
	if httpmetadata.HttpMethod(s) != "GET" {
		return false
	}
	pathParams := httpPathParams(s)
	if slices.Contains(pathParams, "limit") || slices.Contains(pathParams, "offset") {
		return false
	}
	var hasLimit, hasOffset bool
	if s.HasCustomParams() {
		typ := s.InputTypes[0]
		m := s.Messages[converter.CanonicalName(typ)]
		for _, f := range m.Fields {
			name := converter.ToSnakeCase(f.Name)
			if name == "limit" {
				hasLimit = true
				continue
			}
			if name == "offset" {
				hasOffset = true
			}
		}
	} else {
		for _, input := range s.InputNames {
			name := converter.ToSnakeCase(input)
			if name == "limit" {
				hasLimit = true
				continue
			}
			if name == "offset" {
				hasOffset = true
			}
		}
	}

	return hasLimit && hasOffset
}

func OutputUI(s *metadata.Service) []string {
	res := make([]string, 0)
	if s.EmptyOutput() {
		res = append(res, `server.Success(w, r, http.StatusOK, "Success")`)
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
			res = append(res, httpmetadata.BindToSerializable("r", "item", attrName, f.Type)...)
		}
		res = append(res, "res = append(res, item)")
		res = append(res, "}")
		res = append(res, "server.Encode(w, r, http.StatusOK, res)")
		return res
	}

	if m != nil {
		res = append(res, "var res response")
		for _, f := range m.Fields {
			res = append(res, httpmetadata.BindToSerializable("result", "res", converter.UpperFirstCharacter(f.Name), f.Type)...)
		}
		res = append(res, "server.Encode(w, r, http.StatusOK, res)")
		return res
	}

	if s.Output == "sql.Result" {
		if strings.Contains(strings.ToUpper(s.Sql), "INSERT ") {
			res = append(res, "lastInsertId, _ := result.LastInsertId()")
			res = append(res, `server.Success(w, r, http.StatusOK, fmt.Sprintf("Last insert ID: %d", lastInsertId))`)
		} else {
			res = append(res, "rowsAffected, _ := result.RowsAffected()")
			res = append(res, "if rowsAffected < 1 {")
			res = append(res, `    server.Warning(w, r, http.StatusOK, fmt.Sprintf("Rows affected: %d", rowsAffected))`)
			res = append(res, "} else {")
			res = append(res, `    server.Success(w, r, http.StatusOK, fmt.Sprintf("Rows affected: %d", rowsAffected))`)
			res = append(res, "}")
		}
		return res
	}

	if s.Output == "pgconn.CommandTag" {
		res = append(res, "if rowsAffected := result.RowsAffected(); rowsAffected < 1 {")
		res = append(res, `    server.Warning(w, r, http.StatusOK, fmt.Sprintf("Rows affected: %d", rowsAffected))`)
		res = append(res, "} else {")
		res = append(res, `    server.Success(w, r, http.StatusOK, fmt.Sprintf("Rows affected: %d", rowsAffected))`)
		res = append(res, "}")
		return res
	}

	res = append(res, "server.Encode(w, r, http.StatusOK, map[string]any{\"value\": result})")

	return res
}

func ToServiceUI(pkg *metadata.Package, s *metadata.Service) *ServiceUI {
	return &ServiceUI{Service: s, Package: pkg}
}

func hasPathParam(s *metadata.Service) bool {
	path := httpmetadata.HttpPath(s)
	path = strings.TrimSuffix(path, "{$}")
	return strings.Contains(path, "{")
}

func trimHeaderComments(s string) string {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "--") && !strings.HasPrefix(s, "/*") {
		return s
	}
	i := strings.Index(s, "\n")
	if i != -1 {
		return trimHeaderComments(s[i+1:])
	}
	return strings.TrimSpace(s)
}

func httpPathParams(s *metadata.Service) []string {
	re := regexp.MustCompile("{(.*?)}")
	params := re.FindAllString(httpmetadata.HttpPath(s), 100)
	res := make([]string, 0)
	for _, p := range params {
		if len(p) <= 2 || p == "{$}" {
			continue
		}
		res = append(res, strings.TrimSuffix(strings.TrimSuffix(strings.TrimPrefix(p, "{"), "}"), "..."))
	}
	return res
}
