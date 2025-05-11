package templates

import (
	"embed"
	"html/template"
	"strings"

	"github.com/walterwanderley/sqlc-grpc/converter"

	"github.com/walterwanderley/sqlc-http/metadata"
	"github.com/walterwanderley/sqlc-http/metadata/frontend"
)

//go:embed *
var Files embed.FS

var Funcs = template.FuncMap{
	"AddSpace":      frontend.AddSpace,
	"HasPagination": frontend.HasPagination,
	"OutputUI":      frontend.OutputUI,

	"UpperCase": strings.ToUpper,
	"LowerCase": strings.ToLower,

	"UpperFirstCharacter": converter.UpperFirstCharacter,
	"SnakeCase":           converter.ToSnakeCase,
	"KebabCase":           converter.ToKebabCase,
	"PascalCase":          converter.ToPascalCase,

	"HandlerTypes":        metadata.HandlerTypes,
	"Input":               metadata.InputHttp,
	"Output":              metadata.OutputHttp,
	"GroupByPath":         metadata.GroupByPath,
	"ApiParameters":       metadata.ApiParameters,
	"ApiResponse":         metadata.ApiResponse,
	"ApiComponentSchemas": metadata.ApiComponentSchemas,
	"HttpMethod":          metadata.HttpMethod,
	"HttpPath":            metadata.HttpPath,
}
