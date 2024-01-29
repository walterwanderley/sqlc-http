package templates

import (
	"embed"
	"html/template"
	"strings"

	"github.com/walterwanderley/sqlc-grpc/converter"

	"github.com/walterwanderley/sqlc-http/metadata"
)

//go:embed *
var Files embed.FS

var Funcs = template.FuncMap{
	"UpperFirstCharacter": converter.UpperFirstCharacter,
	"HandlerTypes":        metadata.HandlerTypes,
	"Input":               metadata.InputHttp,
	"Output":              metadata.OutputHttp,
	"UpperCase":           strings.ToUpper,
	"LowerCase":           strings.ToLower,
	"GroupByPath":         metadata.GroupByPath,
	"ApiParameters":       metadata.ApiParameters,
	"ApiResponse":         metadata.ApiResponse,
	"HttpMethod":          metadata.HttpMethod,
	"HttpPath":            metadata.HttpPath,
}
