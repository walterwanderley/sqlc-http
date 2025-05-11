package frontend

import (
	"fmt"
	"path"
	"regexp"
	"slices"
	"strings"

	"github.com/walterwanderley/sqlc-grpc/converter"
	"github.com/walterwanderley/sqlc-grpc/metadata"
	httpmetadata "github.com/walterwanderley/sqlc-http/metadata"
)

type ServiceUI struct {
	*metadata.Service
	Package *metadata.Package
}

type DefinitionUI struct {
	*metadata.Definition
	UI bool
}

func (d *DefinitionUI) BreadCrumbs() []string {
	res := make([]string, 0)
	res = append(res, "switch c.Request.Pattern {")
	for _, pkg := range d.Packages {
		for _, svc := range pkg.Services {
			svcName := AddSpace(converter.ToPascalCase(svc.Name))
			for _, spec := range svc.HttpSpecs {
				if spec.Method != "" && spec.Path != "" {
					res = append(res, fmt.Sprintf(`case "%s %s":`, spec.Method, spec.Path))
					parents := parentServices(pkg, svc, &spec, false)
					var args strings.Builder
					for i := len(parents); i > 0; i-- {
						parent := parents[i-1]
						args.WriteString(fmt.Sprintf(`"%s", %s, `, parent.name, parent.path))
					}
					res = append(res, fmt.Sprintf(`return breadCrumbsFromStrings(%s"%s")`, args.String(), svcName))
				}
			}
		}
	}
	res = append(res, "default:")
	res = append(res, "switch c.Request.URL.Path {")
	for _, pkg := range d.Packages {
		for _, svc := range pkg.Services {
			svcName := AddSpace(converter.ToPascalCase(svc.Name))
			res = append(res, fmt.Sprintf(`case "/app/%s/%s":`, pkg.Package, converter.ToSnakeCase(svc.Name)))
			parents := parentServices(pkg, svc, nil, true)
			var args strings.Builder
			for i := len(parents); i > 0; i-- {
				parent := parents[i-1]
				args.WriteString(fmt.Sprintf(`"%s", %s, `, parent.name, parent.path))
			}
			res = append(res, fmt.Sprintf(`return breadCrumbsFromStrings(%s"%s")`, args.String(), svcName))
		}
	}
	res = append(res, "}")
	res = append(res, "}")
	return res
}

type servicePath struct {
	name string
	path string
}

func parentServices(pkg *metadata.Package, svc *metadata.Service, httpSpec *metadata.HttpSpec, ignorePathParams bool) []servicePath {
	pkgService := servicePath{
		name: AddSpace(converter.ToPascalCase(pkg.Package)),
		path: `"/"`,
	}
	if len(svc.HttpSpecs) == 0 && httpSpec == nil {
		return []servicePath{pkgService}
	}
	if httpSpec == nil {
		httpSpec = &svc.HttpSpecs[0]
	}
	list := make([]servicePath, 0)
	services := make([]*metadata.Service, 0)
	services = append(services, svc)
	currentPath := httpSpec.Path
	for {
		currentPath = path.Dir(currentPath)
		if currentPath == "/" || currentPath == "." {
			break
		}
		if ignorePathParams && len(specPathParams(currentPath)) > 0 {
			continue
		}
		service, servicePath := serviceByPath(pkg, services, currentPath, ignorePathParams)
		if servicePath != nil {
			list = append(list, *servicePath)
		}
		if service != nil {
			services = append(services, service)
		}
	}

	return append(list, pkgService)
}

func serviceByPath(pkg *metadata.Package, services []*metadata.Service, path string, ignorePathParams bool) (*metadata.Service, *servicePath) {
	for _, svc := range pkg.Services {
		if slices.Contains(services, svc) {
			continue
		}
		for _, spec := range svc.HttpSpecs {
			if strings.ToUpper(spec.Method) != "GET" {
				continue
			}
			if spec.Path == path {
				pathParams := specPathParams(path)
				if ignorePathParams && len(pathParams) > 0 {
					continue
				}
				resolvedPath := spec.Path
				for _, param := range pathParams {
					resolvedPath = strings.ReplaceAll(resolvedPath, fmt.Sprintf("{%s}", param), fmt.Sprintf(`" + c.Request.PathValue("%s") + "`, param))
				}
				resolvedPath = fmt.Sprintf(`"%s"`, resolvedPath)
				resolvedPath = strings.TrimSuffix(resolvedPath, ` + ""`)
				return svc, &servicePath{
					name: AddSpace(converter.ToPascalCase(svc.Name)),
					path: resolvedPath,
				}
			}
		}
	}
	return nil, nil
}

func specPathParams(path string) []string {
	re := regexp.MustCompile("{(.*?)}")
	params := re.FindAllString(path, 100)
	res := make([]string, 0)
	for _, p := range params {
		if len(p) <= 2 || p == "{$}" {
			continue
		}
		res = append(res, strings.TrimSuffix(strings.TrimSuffix(strings.TrimPrefix(p, "{"), "}"), "..."))
	}
	return res
}

type PackageUI struct {
	*metadata.Package
	UI bool
}

func (p *PackageUI) Name() string {
	return p.Package.Package
}

func (s *ServiceUI) ActionLabel() string {
	query := trimHeaderComments(strings.ReplaceAll(s.Sql, "`", ""))
	query = strings.ToUpper(query)
	switch {
	case strings.HasPrefix(query, "INSERT"):
		return "Insert"
	case strings.HasPrefix(query, "UPDATE"):
		return "Update"
	case strings.HasPrefix(query, "DELETE"):
		return "Delete"
	case strings.HasPrefix(query, "SELECT"):
		return "Search"
	default:
		return "Submit"
	}
}

func (s *ServiceUI) Title() string {
	return AddSpace(converter.ToPascalCase(s.Name))
}

func (s *ServiceUI) HtmxCall(nestedResponse bool) string {
	method := httpmetadata.HttpMethod(s.Service)
	path := httpmetadata.HttpPath(s.Service)
	path = strings.TrimSuffix(path, "{$}")
	if s.AutoSubmit() && s.HasArrayOutput() {
		path = path + "?limit=10"
		if nestedResponse {
			path = path + "&nested=true"
		}
	} else {
		if nestedResponse {
			path = path + "?nested=true"
		}
	}
	return fmt.Sprintf(`hx-%s="%s"`, strings.ToLower(method), strings.TrimPrefix(path, "/"))
}

func (s *ServiceUI) HasPathParam() bool {
	return hasPathParam(s.Service)
}

func (s *ServiceUI) AutoSubmit() bool {
	if httpmetadata.HttpMethod(s.Service) != "GET" {
		return false
	}
	if s.EmptyInput() {
		return true
	}
	if HasPagination(s.Service) {
		if s.HasCustomParams() {
			typ := s.InputTypes[0]
			m := s.Messages[converter.CanonicalName(typ)]
			for _, f := range m.Fields {
				name := converter.ToSnakeCase(f.Name)
				if name != "limit" && name != "offset" {
					return false
				}
			}
			return true
		}
		for _, inName := range s.InputNames {
			name := converter.ToSnakeCase(inName)
			if name != "limit" && name != "offset" {
				return false
			}
		}
		return true
	}

	return false
}

func (s *ServiceUI) HtmlInput() []string {
	res := make([]string, 0)
	if s.EmptyInput() {
		return res
	}
	hasPagination := HasPagination(s.Service)
	if s.HasCustomParams() {
		typ := s.InputTypes[0]
		m := s.Messages[converter.CanonicalName(typ)]
		for _, f := range m.Fields {
			name := converter.ToSnakeCase(f.Name)
			if hasPagination && name == "limit" {
				res = append(res, `<input type="hidden" name="limit" value="10"/>`)
				continue
			}
			if hasPagination && name == "offset" {
				res = append(res, `<input type="hidden" name="offset" value="0"/>`)
				continue
			}
			res = append(res, htmlInput(f.Name, f.Type, false)...)
		}
	} else {
		for i, inName := range s.InputNames {
			name := converter.ToSnakeCase(inName)
			if hasPagination && name == "limit" {
				res = append(res, `<input type="hidden" name="limit" value="10"/>`)
				continue
			}
			if hasPagination && name == "offset" {
				res = append(res, `<input type="hidden" name="offset" value="0"/>`)
				continue
			}
			res = append(res, htmlInput(inName, s.InputTypes[i], false)...)
		}
	}
	return res
}

func (s *ServiceUI) HasEditService() bool {
	return s.editService() != nil
}

func (s *ServiceUI) EditName() string {
	edit := s.editService()
	if edit == nil {
		return AddSpace(s.Name)
	}
	return AddSpace(edit.Name)
}

func (s *ServiceUI) EditHtmxCall() string {
	edit := s.editService()
	if edit == nil {
		return ""
	}
	return edit.HtmxCall(false)
}

func (s *ServiceUI) HtmlInputEdit() []string {
	res := make([]string, 0)
	serviceUI := s.editService()
	if serviceUI == nil {
		return res
	}
	pathParams := httpPathParams(serviceUI.Service)
	if serviceUI.HasCustomParams() {
		typ := serviceUI.InputTypes[0]
		m := serviceUI.Messages[converter.CanonicalName(typ)]
		for _, f := range m.Fields {
			param := converter.ToSnakeCase(f.Name)
			if slices.Contains(pathParams, param) {
				res = append(res, fmt.Sprintf(`<input type="hidden" name="%s" value="{{.Data.%s}}">`, param, f.Name))
				continue
			}
			res = append(res, htmlInput(f.Name, f.Type, true)...)
		}
	} else {
		for i, name := range serviceUI.InputNames {
			param := converter.ToSnakeCase(name)
			if slices.Contains(pathParams, param) {
				res = append(res, fmt.Sprintf(`<input type="hidden" name="%s" value="{{.Data.%s}}">`, param, name))
				continue
			}
			res = append(res, htmlInput(name, serviceUI.InputTypes[i], true)...)
		}
	}
	res = append(res, `<div class="row">`)
	res = append(res, fmt.Sprintf(`    <button class="btn waves-effect waves-light" type="submit">%s<i class="material-icons right">send</i></button>`, serviceUI.ActionLabel()))
	res = append(res, `    <button class="waves-effect waves-light btn" type="button"`)
	res = append(res, `        onclick="javascript:window.history.back()">Back</button>`)
	res = append(res, `</div>`)
	return res
}

func (s *ServiceUI) DeletePath() string {
	path := httpmetadata.HttpPath(s.Service)
	out, ok := s.Messages[converter.CanonicalName(s.Output)]
	if !ok {
		return ""
	}
	paths := make([]string, 0)
	for _, svc := range s.Package.Services {
		if httpmetadata.HttpMethod(svc) != "DELETE" {
			continue
		}
		p := httpmetadata.HttpPath(svc)
		if !strings.HasPrefix(p, path) {
			continue
		}
		params := httpPathParams(svc)
		queryParams := make([]string, 0)
		var incompatibleParams bool
		if svc.HasCustomParams() {
			m, ok := svc.Messages[converter.CanonicalName(svc.InputTypes[0])]
			if !ok {
				continue
			}
			for _, inField := range m.Fields {
				var found bool
				for _, outField := range out.Fields {
					if inField.Name == outField.Name {
						found = true
						break
					}
				}
				if !found {
					incompatibleParams = true
					break
				}

				var foundPathParam bool
				for _, param := range params {
					if param == converter.ToSnakeCase(inField.Name) {
						p = strings.ReplaceAll(p, fmt.Sprintf("{%s}", param), fmt.Sprintf(`{{$v.%s}}`, inField.Name))
						foundPathParam = true
						break
					}
				}
				if !foundPathParam {
					queryParam := converter.ToSnakeCase(inField.Name)
					queryParams = append(queryParams, fmt.Sprintf("%s={{$v.%s}}", queryParam, inField.Name))
				}

			}
		} else {
			for _, inField := range svc.InputNames {
				var found bool
				var outParam string
				for _, outField := range out.Fields {
					if converter.ToSnakeCase(inField) == converter.ToSnakeCase(outField.Name) {
						outParam = outField.Name
						found = true
						break
					}
				}
				if !found {
					incompatibleParams = true
					break
				}
				var foundPathParam bool
				for _, param := range params {
					if param == converter.ToSnakeCase(inField) {
						p = strings.ReplaceAll(p, fmt.Sprintf("{%s}", param), fmt.Sprintf(`{{$v.%s}}`, outParam))
						foundPathParam = true
						break
					}
				}
				if !foundPathParam {
					queryParam := converter.ToSnakeCase(inField)
					queryParams = append(queryParams, fmt.Sprintf("%s={{$v.%s}}", queryParam, outParam))
				}
			}
		}
		if incompatibleParams {
			continue
		}

		if len(queryParams) > 0 {
			p = p + "?" + strings.Join(queryParams, "&")
		}
		if p != "" {
			paths = append(paths, p)
		}
	}
	if len(paths) == 0 {
		return ""
	}
	slices.SortFunc(paths, func(a, b string) int {
		if len(a) < len(b) {
			return -1
		}
		if len(b) < len(a) {
			return 1
		}
		return 0
	})
	return paths[0]
}

func (s *ServiceUI) ViewPath() string {
	if !s.HasArrayOutput() {
		return ""
	}
	output := converter.CanonicalName(s.Output)
	out, ok := s.Messages[output]
	if !ok {
		return ""
	}
	paths := make([]string, 0)
	for _, svc := range s.Package.Services {
		if httpmetadata.HttpMethod(svc) != "GET" {
			continue
		}
		if svc.HasArrayOutput() {
			continue
		}
		if converter.CanonicalName(svc.Output) != output {
			continue
		}

		p := httpmetadata.HttpPath(svc)
		params := httpPathParams(svc)
		queryParams := make([]string, 0)
		var incompatibleParams bool
		if svc.HasCustomParams() {
			m, ok := svc.Messages[converter.CanonicalName(svc.InputTypes[0])]
			if !ok {
				continue
			}
			for _, inField := range m.Fields {
				var found bool
				for _, outField := range out.Fields {
					if inField.Name == outField.Name {
						found = true
						break
					}
				}
				if !found {
					incompatibleParams = true
					break
				}

				var foundPathParam bool
				for _, param := range params {
					if param == converter.ToSnakeCase(inField.Name) {
						p = strings.ReplaceAll(p, fmt.Sprintf("{%s}", param), fmt.Sprintf(`{{$v.%s}}`, inField.Name))
						foundPathParam = true
						break
					}
				}
				if !foundPathParam {
					queryParam := converter.ToSnakeCase(inField.Name)
					queryParams = append(queryParams, fmt.Sprintf("%s={{$v.%s}}", queryParam, inField.Name))
				}
			}
		} else {
			for _, inField := range svc.InputNames {
				var found bool
				var outParam string
				for _, outField := range out.Fields {
					if converter.ToSnakeCase(inField) == converter.ToSnakeCase(outField.Name) {
						outParam = outField.Name
						found = true
						break
					}
				}
				if !found {
					incompatibleParams = true
					break
				}
				var foundPathParam bool
				for _, param := range params {
					if param == converter.ToSnakeCase(inField) {
						p = strings.ReplaceAll(p, fmt.Sprintf("{%s}", param), fmt.Sprintf(`{{$v.%s}}`, outParam))
						foundPathParam = true
						break
					}
				}
				if !foundPathParam {
					queryParam := converter.ToSnakeCase(inField)
					queryParams = append(queryParams, fmt.Sprintf("%s={{$v.%s}}", queryParam, outParam))
				}
			}
		}
		if incompatibleParams {
			continue
		}

		if len(queryParams) > 0 {
			p = p + "?" + strings.Join(queryParams, "&")
		}
		if p != "" {
			paths = append(paths, p)
		}
	}
	if len(paths) == 0 {
		return ""
	}
	slices.SortFunc(paths, func(a, b string) int {
		if len(a) < len(b) {
			return -1
		}
		if len(b) < len(a) {
			return 1
		}
		return 0
	})
	return paths[0]
}

func (s *ServiceUI) AddPath() string {
	if httpmetadata.HttpMethod(s.Service) != "GET" {
		return ""
	}
	_, ok := s.Messages[converter.CanonicalName(s.Output)]
	if !s.HasArrayOutput() || !ok {
		return ""
	}

	path := httpmetadata.HttpPath(s.Service)
	for _, svc := range s.Package.Services {
		if httpmetadata.HttpMethod(svc) != "POST" {
			continue
		}
		if httpmetadata.HttpPath(svc) == path {
			return fmt.Sprintf("app/%s/%s", s.Package.Package, converter.ToSnakeCase(svc.Name))
		}
	}
	return ""
}

func (s *ServiceUI) EditPath() string {
	viewPathURL := s.ViewPath()
	var viewPath string
	if i := strings.Index(viewPath, "?"); i > 0 {
		viewPath = viewPath[0:i]
	} else {
		viewPath = viewPathURL
	}
	if viewPath == "" {
		return ""
	}
	out, ok := s.Messages[converter.CanonicalName(s.Output)]
	if !ok {
		return ""
	}
	for _, svc := range s.Package.Services {
		if httpmetadata.HttpMethod(svc) != "PUT" && httpmetadata.HttpMethod(svc) != "PATCH" {
			continue
		}
		p := httpmetadata.HttpPath(svc)
		params := httpPathParams(svc)
		for _, param := range params {
			var found bool
			for _, f := range out.Fields {
				if converter.ToSnakeCase(param) == converter.ToSnakeCase(f.Name) {
					found = true
					p = strings.ReplaceAll(p, fmt.Sprintf("{%s}", param), fmt.Sprintf(`{{$v.%s}}`, f.Name))
				}
			}
			if !found {
				p = ""
				break
			}
		}

		if p == viewPath {
			if strings.Contains(viewPathURL, "?") {
				return p + "&edit"
			}
			return p + "?edit"
		}
	}
	return ""
}

func (s *ServiceUI) HtmlOutput() []string {
	res := make([]string, 0)
	if s.EmptyOutput() || s.Output == "pgconn.CommandTag" || s.Output == "sql.Result" {
		return res
	}

	m, ok := s.Messages[converter.CanonicalName(s.Output)]
	if !ok {
		res = append(res, `<div class="col mb-5">`)
		if s.HasArrayOutput() {
			res = append(res, `    <ul>`)
			res = append(res, `        {{range .Data.list}}<li>`)
			res = append(res, `            {{.}}`)
			res = append(res, `        </li>`)
			res = append(res, `        {{end -}}`)
			res = append(res, `    </ul>`)
		} else {
			res = append(res, `    <div class="row">`)
			res = append(res, `        <div class="column">`)
			res = append(res, `            <p>{{.Data.value}}</p>`)
			res = append(res, `        </div>`)
			res = append(res, `    </div>`)
		}
		res = append(res, `</div>`)
		return res
	}

	if s.HasArrayOutput() {
		viewPath := s.ViewPath()
		deletePath := s.DeletePath()
		editPath := s.EditPath()
		addPath := s.AddPath()
		if addPath != "" {
			res = append(res, `<div class="row">`)
			res = append(res, fmt.Sprintf(`    <a class="btn-floating btn-large waves-effect waves-light red"><i class="material-icons" hx-get="%s" hx-push-url="true">add</i></a>`, strings.TrimPrefix(addPath, "/")))
			res = append(res, `</div>`)
		}
		res = append(res, `<div class="col s12">`)
		res = append(res, `<table>`)
		res = append(res, `    <thead><tr>`)
		for _, f := range m.Fields {
			attrName := AddSpace(converter.UpperFirstCharacter(f.Name))
			res = append(res, fmt.Sprintf(`        <th>%s</th>`, attrName))
		}
		if viewPath != "" || deletePath != "" || editPath != "" {
			res = append(res, `        <th></th> <!-- Ações -->`) // Ações
		}
		res = append(res, `    </tr></thead>`)
		res = append(res, `    <tbody>`)
		res = append(res, `        {{range $i, $v := .Data}}<tr id="row_{{$i}}">`)
		for _, f := range m.Fields {
			attrName := converter.UpperFirstCharacter(f.Name)
			if strings.HasSuffix(f.Type, "time.Time") || strings.HasSuffix(f.Type, "sql.NullTime") ||
				strings.HasSuffix(f.Type, "pgtype.Date") {
				res = append(res, fmt.Sprintf(`        <td>{{if $v.%s}}{{$v.%s.Format "02/01/2006"}}{{end}}</td>`, attrName, attrName))
				continue
			}
			res = append(res, fmt.Sprintf(`        <td>{{$v.%s}}</td>`, attrName))
		}

		if viewPath != "" || deletePath != "" || editPath != "" {
			res = append(res, `        <td>`) // Ações
			if viewPath != "" {
				res = append(res, fmt.Sprintf(`            <a class="btn-floating btn-small waves-effect waves-light red" 
  	href="javascript: void(0)" hx-push-url="true" hx-get="%s"><i class="material-icons">visibility</i></a>`, strings.TrimPrefix(viewPath, "/")))
			}
			if editPath != "" {
				res = append(res, fmt.Sprintf(`            <a class="btn-floating btn-small waves-effect waves-light red" 
  	href="javascript: void(0)" hx-push-url="true" hx-get="%s"><i class="material-icons">edit</i></a>`, strings.TrimPrefix(editPath, "/")))
			}
			if deletePath != "" {
				res = append(res, fmt.Sprintf(`            <a class="btn-floating btn-small waves-effect waves-light red" 
  	href="javascript: void(0)" hx-delete="%s" hx-swap="outerHTML"
	hx-target="#row_{{$i}}" hx-confirm="Are you shure?"><i class="material-icons">delete</i></a>`, strings.TrimPrefix(deletePath, "/")))
			}
			res = append(res, `        </td>`) // Ações
		}

		res = append(res, `        </tr>{{end}}`)
		res = append(res, `    </tbody>`)
		res = append(res, `</table>`)
		res = append(res, `</div>`)
		return res
	}

	res = append(res, `<div class="col mb-5">`)
	for _, f := range m.Fields {
		attrName := converter.UpperFirstCharacter(f.Name)
		label := AddSpace(attrName)
		res = append(res, `    <div class="row">`)
		res = append(res, `        <div class="column">`)
		res = append(res, fmt.Sprintf(`            <p><b>%s:</b> {{.Data.%s}}</p>`, label, attrName))
		res = append(res, `        </div>`)
		res = append(res, `    </div>`)
	}
	res = append(res, `</div>`)

	return res
}

func (s *ServiceUI) editService() *ServiceUI {
	if httpmetadata.HttpMethod(s.Service) != "GET" || s.HasArrayOutput() || s.EmptyOutput() {
		return nil
	}
	path := httpmetadata.HttpPath(s.Service)
	var service *metadata.Service
	for _, svc := range s.Package.Services {
		if httpmetadata.HttpPath(svc) != path {
			continue
		}
		if httpmetadata.HttpMethod(svc) == "PUT" || httpmetadata.HttpMethod(svc) == "PATCH" {
			if svc.EmptyInput() {
				continue
			}
			service = svc
			break
		}
	}
	if service == nil {
		return nil
	}
	return &ServiceUI{
		Service: service,
		Package: s.Package,
	}
}
