package metadata

import (
	"fmt"
	"strings"

	"github.com/walterwanderley/sqlc-grpc/converter"
)

func toSerializableType(typ string) string {
	if strings.HasPrefix(typ, "*") {
		return toSerializableType(typ[1:])
	}
	if strings.HasPrefix(typ, "[]") {
		return "[]" + toSerializableType(typ[2:])
	}
	switch typ {
	case "json.RawMessage":
		return "json.RawMessage"
	case "byte":
		return "byte"
	case "bool":
		return "bool"
	case "sql.NullBool", "pgtype.Bool":
		return "*bool"
	case "sql.NullInt32", "pgtype.Int4", "pgtype.Int2":
		return "*int32"
	case "pgtype.Uint32":
		return "*uint32"
	case "int":
		return "int"
	case "int64":
		return "int64"
	case "uint64":
		return "uint64"
	case "int16":
		return "int16"
	case "int32":
		return "int32"
	case "uint16":
		return "uint32"
	case "uint32":
		return "uint32"
	case "sql.NullInt64", "pgtype.Int8":
		return "*int64"
	case "float32":
		return "float32"
	case "float64":
		return "float64"
	case "pgtype.Float4":
		return "*float32"
	case "sql.NullFloat64", "pgtype.Float8":
		return "*float64"
	case "sql.NullString", "pgtype.Text", "pgtype.UUID":
		return "*string"
	case "time.Time":
		return "time.Time"
	case "sql.NullTime", "pgtype.Date", "pgtype.Timestamp", "pgtype.Timestampz":
		return "*time.Time"
	case "string", "uuid.UUID", "net.HardwareAddr", "net.IP":
		return "string"
	default:
		if _, elementType := originalAndElementType(typ); elementType != "" {
			return elementType
		}
		return converter.UpperFirstCharacter(typ)
	}
}

func originalAndElementType(typ string) (original, element string) {
	typ = strings.TrimPrefix(typ, "[]")
	t := strings.Split(typ, ".")
	return t[0], strings.Join(t[1:], ".")
}

func BindToSerializable(src, dst, attrName, attrType string) []string {
	res := make([]string, 0)
	switch attrType {
	case "sql.NullBool", "pgtype.Bool":
		res = append(res, fmt.Sprintf("if %s.%s.Valid {", src, attrName))
		res = append(res, fmt.Sprintf("%s.%s = &%s.%s.Bool }", dst, attrName, src, attrName))
	case "pgtype.Int2":
		res = append(res, fmt.Sprintf("if %s.%s.Valid {", src, attrName))
		res = append(res, fmt.Sprintf("%s.%s = &%s.%s.Int16 }", dst, attrName, src, attrName))
	case "pgtype.Uint32":
		res = append(res, fmt.Sprintf("if %s.%s.Valid {", src, attrName))
		res = append(res, fmt.Sprintf("%s.%s = &%s.%s.Uint32 }", dst, attrName, src, attrName))
	case "sql.NullInt32", "pgtype.Int4":
		res = append(res, fmt.Sprintf("if %s.%s.Valid {", src, attrName))
		res = append(res, fmt.Sprintf("%s.%s = &%s.%s.Int32 }", dst, attrName, src, attrName))
	case "sql.NullInt64", "pgtype.Int8":
		res = append(res, fmt.Sprintf("if %s.%s.Valid {", src, attrName))
		res = append(res, fmt.Sprintf("%s.%s = &%s.%s.Int64 }", dst, attrName, src, attrName))
	case "pgtype.Float4":
		res = append(res, fmt.Sprintf("if %s.%s.Valid {", src, attrName))
		res = append(res, fmt.Sprintf("%s.%s = &%s.%s.Float32 }", dst, attrName, src, attrName))
	case "sql.NullFloat64", "pgtype.Float8":
		res = append(res, fmt.Sprintf("if %s.%s.Valid {", src, attrName))
		res = append(res, fmt.Sprintf("%s.%s = &%s.%s.Float64 }", dst, attrName, src, attrName))
	case "sql.NullString", "pgtype.Text":
		res = append(res, fmt.Sprintf("if %s.%s.Valid {", src, attrName))
		res = append(res, fmt.Sprintf("%s.%s = &%s.%s.String }", dst, attrName, src, attrName))
	case "sql.NullTime", "pgtype.Date", "pgtype.Timestamp", "pgtype.Timestampz":
		res = append(res, fmt.Sprintf("if %s.%s.Valid {", src, attrName))
		res = append(res, fmt.Sprintf("%s.%s = &%s.%s.Time }", dst, attrName, src, attrName))
	case "time.Time":
		res = append(res, fmt.Sprintf("%s.%s = %s.%s", dst, attrName, src, attrName))
	case "uuid.UUID", "net.HardwareAddr", "net.IP":
		res = append(res, fmt.Sprintf("%s.%s = %s.%s.String()", dst, attrName, src, attrName))
	case "pgtype.UUID":
		res = append(res, fmt.Sprintf("if v, err := json.Marshal(%s.%s); err == nil {", src, attrName))
		res = append(res, "str := string(v)")
		res = append(res, fmt.Sprintf("%s.%s = &str", dst, attrName))
		res = append(res, "}")
	case "int16":
		res = append(res, fmt.Sprintf("%s.%s = %s.%s", dst, attrName, src, attrName))
	default:
		_, elementType := originalAndElementType(attrType)
		if elementType != "" {
			res = append(res, fmt.Sprintf("%s.%s = %s(%s.%s)", dst, attrName, elementType, src, attrName))
		} else {
			res = append(res, fmt.Sprintf("%s.%s = %s.%s", dst, attrName, src, attrName))
		}
	}
	return res
}

func bindToGo(src, dst, attrName, attrType string, newVar bool) []string {
	res := make([]string, 0)
	switch attrType {
	case "sql.NullBool", "pgtype.Bool":
		if newVar {
			res = append(res, fmt.Sprintf("var %s %s", dst, attrType))
		}
		res = append(res, fmt.Sprintf("if %s.%s != nil {", src, attrName))
		res = append(res, fmt.Sprintf("%s = %s{Valid: true, Bool: *%s.%s}", dst, attrType, src, attrName))
		res = append(res, "}")
	case "pgtype.Int2":
		if newVar {
			res = append(res, fmt.Sprintf("var %s %s", dst, attrType))
		}
		res = append(res, fmt.Sprintf("if %s.%s != nil {", src, attrName))
		res = append(res, fmt.Sprintf("%s = %s{Valid: true, Int16: *%s.%s}", dst, attrType, src, attrName))
		res = append(res, "}")
	case "pgtype.Uint32":
		if newVar {
			res = append(res, fmt.Sprintf("var %s %s", dst, attrType))
		}
		res = append(res, fmt.Sprintf("if %s.%s != nil {", src, attrName))
		res = append(res, fmt.Sprintf("%s = %s{Valid: true, Uint32: *%s.%s}", dst, attrType, src, attrName))
		res = append(res, "}")
	case "sql.NullInt32", "pgtype.Int4":
		if newVar {
			res = append(res, fmt.Sprintf("var %s %s", dst, attrType))
		}
		res = append(res, fmt.Sprintf("if %s.%s != nil {", src, attrName))
		res = append(res, fmt.Sprintf("%s = %s{Valid: true, Int32: *%s.%s}", dst, attrType, src, attrName))
		res = append(res, "}")
	case "sql.NullInt64", "pgtype.Int8":
		if newVar {
			res = append(res, fmt.Sprintf("var %s %s", dst, attrType))
		}
		res = append(res, fmt.Sprintf("if %s.%s != nil {", src, attrName))
		res = append(res, fmt.Sprintf("%s = %s{Valid: true, Int64: *%s.%s}", dst, attrType, src, attrName))
		res = append(res, "}")
	case "pgtype.Float4":
		if newVar {
			res = append(res, fmt.Sprintf("var %s %s", dst, attrType))
		}
		res = append(res, fmt.Sprintf("if %s.%s != nil {", src, attrName))
		res = append(res, fmt.Sprintf("%s = %s{Valid: true, Float32: *%s.%s}", dst, attrType, src, attrName))
		res = append(res, "}")
	case "sql.NullFloat64", "pgtype.Float8":
		if newVar {
			res = append(res, fmt.Sprintf("var %s %s", dst, attrType))
		}
		res = append(res, fmt.Sprintf("if %s.%s != nil {", src, attrName))
		res = append(res, fmt.Sprintf("%s = %s{Valid: true, Float64: *%s.%s}", dst, attrType, src, attrName))
		res = append(res, "}")
	case "sql.NullString", "pgtype.Text":
		if newVar {
			res = append(res, fmt.Sprintf("var %s %s", dst, attrType))
		}
		res = append(res, fmt.Sprintf("if %s.%s != nil {", src, attrName))
		res = append(res, fmt.Sprintf("%s = %s{Valid: true, String: *%s.%s}", dst, attrType, src, attrName))
		res = append(res, "}")
	case "sql.NullTime", "pgtype.Date", "pgtype.Timestamp", "pgtype.Timestampz":
		if newVar {
			res = append(res, fmt.Sprintf("var %s %s", dst, attrType))
		}
		res = append(res, fmt.Sprintf("if %s.%s != nil {", src, attrName))
		res = append(res, fmt.Sprintf("%s.Valid = true", dst))
		res = append(res, fmt.Sprintf("%s.Time = *%s.%s }", dst, src, attrName))
	case "time.Time":
		if newVar {
			res = append(res, fmt.Sprintf("var %s %s", dst, attrType))
		}
		res = append(res, fmt.Sprintf("%s = %s.%s", dst, src, attrName))
	case "uuid.UUID":
		if newVar {
			res = append(res, fmt.Sprintf("var %s %s", dst, attrType))
		}
		res = append(res, fmt.Sprintf("if v, err := uuid.Parse(%s.%s); err != nil {", src, attrName))
		res = append(res, "http.Error(w, err.Error(), http.StatusUnprocessableEntity)")
		res = append(res, fmt.Sprintf("return } else { %s = v }", dst))
	case "pgtype.UUID":
		if newVar {
			res = append(res, fmt.Sprintf("var %s %s", dst, attrType))
		}
		res = append(res, fmt.Sprintf("if v := %s.%s; v != nil {", src, attrName))
		res = append(res, fmt.Sprintf("if err := json.Unmarshal([]byte(v), &%s); err != nil {", dst))
		res = append(res, "http.Error(w, err.Error(), http.StatusUnprocessableEntity)")
		res = append(res, "return nil, err }")
		res = append(res, "}")
	case "net.HardwareAddr":
		if newVar {
			res = append(res, fmt.Sprintf("var %s %s", dst, attrType))
		}
		res = append(res, fmt.Sprintf("if v, err = net.ParseMAC(%s.%s); err != nil {", src, attrName))
		res = append(res, "http.Error(w, err.Error(), http.StatusUnprocessableEntity)")
		res = append(res, fmt.Sprintf("return nil, err } else { %s = v }", dst))
	case "net.IP":
		if newVar {
			res = append(res, fmt.Sprintf("%s := net.ParseIP(%s.%s)", dst, src, attrName))
		} else {
			res = append(res, fmt.Sprintf("%s = net.ParseIP(%s.%s)", dst, src, attrName))
		}
	case "int16":
		if newVar {
			res = append(res, fmt.Sprintf("%s := int16(%s.%s)", dst, src, attrName))
		} else {
			res = append(res, fmt.Sprintf("%s = int16(%s.%s)", dst, src, attrName))
		}
	case "int":
		if newVar {
			res = append(res, fmt.Sprintf("%s := int(%s.%s)", dst, src, attrName))
		} else {
			res = append(res, fmt.Sprintf("%s = int(%s.%s)", dst, src, attrName))
		}
	case "uint16":
		if newVar {
			res = append(res, fmt.Sprintf("%s := uint16(%s.%s)", dst, src, attrName))
		} else {
			res = append(res, fmt.Sprintf("%s = uint16(%s.%s)", dst, src, attrName))
		}
	default:
		originalType, elementType := originalAndElementType(attrType)
		if newVar {
			if elementType != "" {
				res = append(res, fmt.Sprintf("%s := %s(%s.%s)", dst, originalType, src, attrName))
			} else {
				res = append(res, fmt.Sprintf("%s := %s.%s", dst, src, attrName))
			}
		} else {
			if elementType != "" {
				res = append(res, fmt.Sprintf("%s = %s(%s.%s)", dst, originalType, src, attrName))
			} else {
				res = append(res, fmt.Sprintf("%s = %s.%s", dst, src, attrName))
			}
		}
	}
	return res
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
