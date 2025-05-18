package view

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"sqlite-htmx/view/etag"
	watchersse "sqlite-htmx/view/watcher"
)

// Content-Security-Policy
const csp = "default-src 'self'; img-src 'self' data: ; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline';"

type key int

const (
	templatesContext key = iota
	messageContext
	paginationContext
)

var (
	ErrTemplateNotFound = errors.New("template not found")

	//go:embed static
	staticFS embed.FS

	//go:embed templates
	templatesFS embed.FS
	funcs       = template.FuncMap{}

	provider templatesProvider
)

func init() {
	context := os.Getenv("WEB_CONTEXT")
	if context == "" {
		context = "/"
	}
	funcs["WebContext"] = func() string {
		return context
	}
}

type templatesProvider interface {
	FullTemplate(string) (*template.Template, error)
	DynamicTemplate(string) (*template.Template, error)
	TemplatesFS() fs.FS
	DevMode() bool
}

func RegisterHandlers(mux *http.ServeMux, devMode bool) error {
	base := template.New("base.html").Funcs(funcs)
	content := template.New("content.html").Funcs(funcs)
	componentsTemplates := []string{
		"components/breadcrumbs.html",
		"components/hx-context.html",
		"components/messages-context.html",
		"components/message.html",
	}
	layoutTemplates := []string{
		"layout/base.html",
		"layout/header.html",
		"layout/footer.html",
	}

	if devMode {
		staticPath := filepath.Join("view", "static")
		mux.Handle("GET /static/", http.StripPrefix("/static", http.FileServer(http.FS(os.DirFS(staticPath)))))
		templatesPath := filepath.Join("view", "templates")
		provider = &templateDevRender{
			templatesFS:   os.DirFS(templatesPath),
			base:          base,
			content:       content,
			baseTemplates: append(componentsTemplates, layoutTemplates...),
			components:    componentsTemplates,
		}
		watcher, err := watchersse.New(staticPath, templatesPath)
		if err != nil {
			return err
		}
		watcher.Start(context.Background())
		mux.Handle("GET /reload", watcher)
	} else {
		mux.Handle("GET /static/", etag.Handler(staticFS, ""))
		subFS, err := fs.Sub(templatesFS, "templates")
		if err != nil {
			return err
		}
		tr := templateRender{
			templatesFS: subFS,
			content: template.Must(content.ParseFS(subFS,
				append(componentsTemplates, "layout/content.html")...)),
			base: template.Must(base.ParseFS(subFS,
				append(componentsTemplates, layoutTemplates...)...)),
		}
		if err := tr.Compile(); err != nil {
			return err
		}
		provider = &tr
	}
	mux.HandleFunc("GET /", templatesHandler)
	return nil
}

func templatesHandler(w http.ResponseWriter, r *http.Request) {
	if err := RenderHTML[any](w, r, nil); err != nil {
		slog.Error("render html", "error", err, "path", r.URL.Path)
	}
}

type templateOpts struct {
	Title   string
	Content any
	DevMode bool
}

func defaultTemplateOpts(content any) templateOpts {
	return templateOpts{
		Title:   "Sqlite-htmx",
		Content: content,
		DevMode: provider.DevMode(),
	}
}

func HXRequest(r *http.Request) bool {
	return r.Header.Get("hx-request") == "true"
}

func ContextWithPagination(ctx context.Context, pagination *Pagination) context.Context {
	return context.WithValue(ctx, paginationContext, pagination)
}

func ContextWithMessage(ctx context.Context, msg Message) context.Context {
	return context.WithValue(ctx, messageContext, msg)
}

func ContextWithTemplates(ctx context.Context, templates ...string) context.Context {
	return context.WithValue(ctx, templatesContext, templates)
}

func RenderHTML[T any](w http.ResponseWriter, r *http.Request, content T) (err error) {
	templates := contextTemplates(r)
	if len(templates) == 0 {
		if msg, ok := r.Context().Value(messageContext).(Message); ok {
			return msg.Render(w, r)
		}
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return nil
	}
	var (
		tmpl *template.Template
		obj  any
	)
	if HXRequest(r) {
		tmpl, err = provider.DynamicTemplate(templates[0])
		obj = Content[T]{
			Data:    content,
			Request: r,
		}
	} else {
		tmpl, err = provider.FullTemplate(templates[0])
		obj = defaultTemplateOpts(Content[T]{
			Data:    content,
			Request: r,
		})
	}
	if err != nil {
		switch {
		case errors.Is(err, ErrTemplateNotFound):
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return nil
		default:
			slog.ErrorContext(r.Context(), "render html", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if len(templates) > 1 {
		tmpl, err = tmpl.Clone()
		if err != nil {
			return err
		}
		tmpl, err = tmpl.ParseFS(provider.TemplatesFS(), templates[1:]...)
		if err != nil {
			return err
		}
	}

	w.Header().Set("Content-Type", "text/html;charset=utf-8")
	w.Header().Set("Content-Security-Policy", csp)
	w.WriteHeader(http.StatusOK)
	err = tmpl.Execute(w, obj)
	return err
}

type templateDevRender struct {
	templatesFS   fs.FS
	base          *template.Template
	content       *template.Template
	baseTemplates []string
	components    []string
}

func (t *templateDevRender) DevMode() bool {
	return true
}

func (t *templateDevRender) FullTemplate(path string) (*template.Template, error) {
	f, err := t.templatesFS.Open(path)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrTemplateNotFound, path)
	}
	f.Close()
	return template.Must(t.base.Clone()).ParseFS(t.templatesFS, append(t.baseTemplates, path)...)
}

func (t *templateDevRender) DynamicTemplate(path string) (*template.Template, error) {
	f, err := t.templatesFS.Open(path)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrTemplateNotFound, path)
	}
	f.Close()
	return template.Must(t.content.Clone()).ParseFS(t.templatesFS, append(t.components, "layout/content.html", path)...)
}

func (t *templateDevRender) TemplatesFS() fs.FS {
	return t.templatesFS
}

type templateRender struct {
	templatesFS      fs.FS
	fullTemplates    map[string]*template.Template
	dynamicTemplates map[string]*template.Template
	base             *template.Template
	content          *template.Template
}

func (t *templateRender) DevMode() bool {
	return false
}

func (t *templateRender) Compile() error {
	t.dynamicTemplates = make(map[string]*template.Template)
	t.fullTemplates = make(map[string]*template.Template)
	err := fs.WalkDir(t.templatesFS, ".", func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() || strings.HasPrefix(path, "components/") || strings.HasPrefix(path, "layout/") {
			return nil
		}

		if strings.HasSuffix(path, ".html") {
			dynamicClone, err := t.content.Clone()
			if err != nil {
				return err
			}
			tmpl, err := dynamicClone.ParseFS(t.templatesFS, path)
			if err != nil {
				return err
			}
			t.dynamicTemplates[path] = tmpl
			fullClone, err := t.base.Clone()
			if err != nil {
				return err
			}
			tmplFull, err := fullClone.ParseFS(t.templatesFS, path)
			if err != nil {
				return err
			}
			t.fullTemplates[path] = tmplFull
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("compiling templates: %w", err)
	}
	return nil
}

func (t *templateRender) FullTemplate(path string) (*template.Template, error) {
	tmpl, ok := t.fullTemplates[path]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrTemplateNotFound, path)
	}
	return tmpl, nil
}

func (t *templateRender) DynamicTemplate(path string) (*template.Template, error) {
	tmpl, ok := t.dynamicTemplates[path]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrTemplateNotFound, path)
	}
	return tmpl, nil
}

func (t *templateRender) TemplatesFS() fs.FS {
	return t.templatesFS
}

func contextTemplates(r *http.Request) []string {
	if v, ok := r.Context().Value(templatesContext).([]string); ok {
		return v
	}

	path := strings.TrimPrefix(strings.TrimPrefix(r.Pattern, r.Method+" "), "/")
	if path == "" {
		path = strings.TrimPrefix(r.URL.Path, "/")
	}
	if path == "" {
		path = "index"
	}
	path = path + ".html"
	return []string{path}
}
