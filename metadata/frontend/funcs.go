package frontend

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
	"time"
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
			res = append(res, `htmx.Success(w, r, http.StatusOK, fmt.Sprintf("Last insert ID: %d", lastInsertId))`)
		} else {
			res = append(res, "rowsAffected, _ := result.RowsAffected()")
			res = append(res, `htmx.Success(w, r, http.StatusOK, fmt.Sprintf("Rows affected: %d", rowsAffected))`)
		}
		return res
	}

	if s.Output == "pgconn.CommandTag" {
		res = append(res, `htmx.Success(w, r, http.StatusOK, fmt.Sprintf("Rows affected: %d", result.RowsAffected()))`)
		return res
	}

	res = append(res, "server.Encode(w, r, http.StatusOK, map[string]any{\"value\": result})")

	return res
}

func toHtmlType(typ string) string {
	if strings.HasPrefix(typ, "*") {
		return toHtmlType(typ[1:])
	}
	if strings.HasPrefix(typ, "[]") {
		return toHtmlType(typ[2:])
	}
	switch typ {
	case "bool", "sql.NullBool", "pgtype.Bool":
		return "checkbox"
	case "sql.NullInt32", "pgtype.Int4", "pgtype.Int2", "pgtype.Uint32", "int", "int64", "uint64", "int16", "int32", "uint16", "uint32":
		return "number"
	case "time.Time", "sql.NullTime", "pgtype.Date", "pgtype.Timestamp", "pgtype.Timestampz":
		return "date"
	default:
		return "text"
	}
}

func htmlInput(attr, typ string, fill bool) []string {
	attrName := converter.ToPascalCase(attr)
	required := !strings.Contains(typ, "*") && !strings.Contains(typ, "pgtype.") && !strings.Contains(typ, "sql.Null")
	label := AddSpace(attrName)
	if required {
		label = label + " *"
	}
	typ = toHtmlType(typ)
	res := make([]string, 0)
	attrFormName := converter.ToSnakeCase(attrName)
	var requiredAttr string
	if required {
		requiredAttr = "required"
	}
	switch typ {
	case "checkbox":
		res = append(res, `<div class="form-check">`)
		if fill {
			res = append(res, fmt.Sprintf(`    <input id="%s" name="%s" type="checkbox" class="form-check-input" value="{{.Data.%s}}" {{if .Data.%s}}checked{{end}}/>`, attrFormName, attrFormName, attr, attr))
		} else {
			res = append(res, fmt.Sprintf(`    <input id="%s" name="%s" type="checkbox" class="form-check-input" value=""/>`, attrFormName, attrFormName))
		}
		res = append(res, fmt.Sprintf(`    <label class="form-check-label" for="%s">%s</label>`, attrFormName, label))
		res = append(res, `</div>`)
	case "date":
		res = append(res, `<div class="mb-3">`)
		res = append(res, `    <div class="col-sm-4 col-md-2">`)

		if fill {
			res = append(res, fmt.Sprintf(`        <label for="%s" class="form-label">%s</label>`, attrFormName, label))
			res = append(res, fmt.Sprintf(`        <input id="%s" %s name="%s" type="date" class="form-control"{{if .Data.%s}} value="{{.Data.%s.Format "%s"}}"{{end}}/>`, attrFormName, requiredAttr, attrFormName, attr, attr, time.DateOnly))
		} else {
			res = append(res, fmt.Sprintf(`        <label for="%s" class="form-label">%s</label>`, attrFormName, label))
			res = append(res, fmt.Sprintf(`        <input id="%s" %s name="%s" type="date" class="form-control"/>`, attrFormName, requiredAttr, attrFormName))
		}
		res = append(res, `    </div>`)
		res = append(res, `</div>`)
	default:
		res = append(res, `<div class="mb-3">`)
		if fill {
			res = append(res, fmt.Sprintf(`    <label for="%s" class="form-label">%s</label>`, attrFormName, label))
			res = append(res, fmt.Sprintf(`    <input id="%s" %s name="%s" type="%s"{{if .Data.%s}} value="{{.Data.%s}}"{{end}} class="form-control"/>`, attrFormName, requiredAttr, attrFormName, typ, attr, attr))
		} else {
			res = append(res, fmt.Sprintf(`    <label for="%s" class="form-label">%s</label>`, attrFormName, label))
			res = append(res, fmt.Sprintf(`    <input id="%s" %s name="%s" type="%s" class="form-control"/>`, attrFormName, requiredAttr, attrFormName, typ))

		}
		res = append(res, `</div>`)
	}

	return res
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
