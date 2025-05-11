package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/walterwanderley/sqlc-grpc/converter"
	"github.com/walterwanderley/sqlc-grpc/metadata"
	"golang.org/x/tools/imports"

	httpmetadata "github.com/walterwanderley/sqlc-http/metadata"
	"github.com/walterwanderley/sqlc-http/metadata/frontend"
	"github.com/walterwanderley/sqlc-http/templates"
)

func process(def *metadata.Definition, appendMode bool, generateFrontend bool) error {
	return fs.WalkDir(templates.Files, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			log.Println("ERROR ", err.Error())
			return err
		}

		newPath := strings.TrimSuffix(path, ".tmpl")

		if d.IsDir() {
			if (strings.HasSuffix(newPath, "templates") ||
				strings.HasSuffix(newPath, "etag") ||
				strings.HasSuffix(newPath, "htmx") ||
				strings.HasSuffix(newPath, "watcher") ||
				strings.HasSuffix(newPath, "web") ||
				strings.HasSuffix(newPath, "swagger") ||
				strings.HasSuffix(newPath, "css") ||
				strings.HasSuffix(newPath, "js") ||
				strings.HasSuffix(newPath, "app") ||
				strings.HasSuffix(newPath, "layout")) && !generateFrontend {
				return nil
			}

			if strings.HasSuffix(newPath, "instrumentation") && (!def.DistributedTracing && !def.Metric) {
				return nil
			}
			if strings.HasSuffix(newPath, "trace") && !def.DistributedTracing {
				return nil
			}
			if strings.HasSuffix(newPath, "metric") && !def.Metric {
				return nil
			}
			if strings.HasSuffix(newPath, "litestream") && !(def.Database() == "sqlite" && def.Litestream) {
				return nil
			}

			if strings.HasSuffix(newPath, "litefs") && !(def.Database() == "sqlite" && def.LiteFS) {
				return nil
			}
			if _, err := os.Stat(newPath); os.IsNotExist(err) {
				err := os.MkdirAll(newPath, 0750)
				if err != nil {
					return err
				}
			}
			return nil
		}

		if newPath == "templates.go" {
			return nil
		}

		if (strings.HasSuffix(newPath, ".html") || strings.HasSuffix(newPath, ".css") || strings.HasSuffix(newPath, ".svg") ||
			strings.HasSuffix(newPath, ".js") || strings.HasSuffix(newPath, "templates.go")) && !generateFrontend {
			return nil
		}

		log.Println(path, "...")

		in, err := templates.Files.Open(path)
		if err != nil {
			return err
		}
		defer in.Close()

		if strings.HasSuffix(newPath, "service.go") {
			tpl, err := io.ReadAll(in)
			if err != nil {
				return err
			}
			for _, pkg := range def.Packages {
				err = genFromTemplate(path, string(tpl), pkg, true, generateFrontend, filepath.Join(pkg.SrcPath, "service.go"))
				if err != nil {
					return err
				}
			}
			return nil
		}

		if strings.HasSuffix(newPath, "service_factory.go") {
			tpl, err := io.ReadAll(in)
			if err != nil {
				return err
			}
			for _, pkg := range def.Packages {
				newPath := filepath.Join(pkg.SrcPath, "service_factory.go")
				if appendMode && fileExists(newPath) {
					continue
				}
				err = genFromTemplate(path, string(tpl), pkg, true, generateFrontend, newPath)
				if err != nil {
					return err
				}
			}
			return nil
		}

		if strings.HasSuffix(newPath, "routes.go") {
			tpl, err := io.ReadAll(in)
			if err != nil {
				return err
			}
			for _, pkg := range def.Packages {
				newPath := filepath.Join(pkg.SrcPath, "routes.go")
				err = genFromTemplate(path, string(tpl), pkg, true, generateFrontend, newPath)
				if err != nil {
					return err
				}
			}
			return nil
		}

		if strings.HasSuffix(newPath, "request.html") {
			if !generateFrontend {
				return nil
			}
			tpl, err := io.ReadAll(in)
			if err != nil {
				return err
			}
			dir := strings.TrimSuffix(newPath, "request.html")
			for _, pkg := range def.Packages {
				dest := filepath.Join(dir, pkg.Package)
				if _, err := os.Stat(dest); os.IsNotExist(err) {
					err := os.MkdirAll(dest, 0750)
					if err != nil {
						return err
					}
				}
				for _, service := range pkg.Services {
					destFile := filepath.Join(dest, (converter.ToSnakeCase(service.Name) + ".html"))
					if appendMode && fileExists(destFile) {
						return nil
					}
					err = genFromTemplate(path, string(tpl), &frontend.ServiceUI{Service: service, Package: pkg}, false, generateFrontend, destFile)
					if err != nil {
						return err
					}
				}
			}
			return nil
		}

		if strings.HasSuffix(newPath, "response.html") {
			if !generateFrontend {
				return nil
			}
			tpl, err := io.ReadAll(in)
			if err != nil {
				return err
			}

			for _, pkg := range def.Packages {
				for _, service := range pkg.Services {
					if service.EmptyOutput() || service.Output == "sql.Result" || service.Output == "pgconn.CommandTag" {
						continue
					}
					path := httpmetadata.HttpPath(service)
					path = strings.TrimSuffix(path, "/")
					destFile := filepath.Join("templates", path+".html")

					if appendMode && fileExists(destFile) {
						return nil
					}
					destDir := filepath.Dir(destFile)
					if _, err := os.Stat(destDir); os.IsNotExist(err) {
						err := os.MkdirAll(destDir, 0750)
						if err != nil {
							return err
						}
					}
					err = genFromTemplate(path, string(tpl), &frontend.ServiceUI{Service: service, Package: pkg}, false, generateFrontend, destFile)
					if err != nil {
						return err
					}
				}
			}
			return nil
		}

		if strings.HasSuffix(newPath, "openapi.yml") {
			tpl, err := io.ReadAll(in)
			if err != nil {
				return err
			}
			openapiDef, err := httpmetadata.LoadOpenApi(newPath, appendMode, def)
			if err != nil {
				return err
			}
			return genFromTemplate(path, string(tpl), openapiDef, false, generateFrontend, newPath)
		}

		if strings.HasSuffix(newPath, "tracing.go") && !def.DistributedTracing {
			return nil
		}

		if strings.HasSuffix(newPath, "metric.go") && !def.Metric {
			return nil
		}

		if strings.HasSuffix(newPath, "migration.go") && def.MigrationPath == "" {
			return nil
		}

		if strings.HasSuffix(newPath, "litestream.go") && !(def.Database() == "sqlite" && def.Litestream) {
			return nil
		}

		if (strings.HasSuffix(newPath, "litefs.go") || strings.HasSuffix(newPath, "forward.go")) && !(def.Database() == "sqlite" && def.LiteFS) {
			return nil
		}

		if strings.HasSuffix(newPath, "templates/templates.go") {
			if !generateFrontend {
				return nil
			}
			tpl, err := io.ReadAll(in)
			if err != nil {
				return err
			}
			return genFromTemplate(path, string(tpl), &frontend.DefinitionUI{Definition: def, UI: generateFrontend}, true, generateFrontend, newPath)
		}

		if strings.HasSuffix(path, ".tmpl") {
			tpl, err := io.ReadAll(in)
			if err != nil {
				return err
			}
			goCode := strings.HasSuffix(newPath, ".go")
			if goCode && appendMode && fileExists(newPath) && !strings.HasSuffix(newPath, "registry.go") {
				return nil
			}
			return genFromTemplate(path, string(tpl), def, goCode, generateFrontend, newPath)
		}

		if appendMode && fileExists(newPath) {
			return nil
		}

		out, err := os.Create(newPath)
		if err != nil {
			return err
		}
		defer out.Close()

		_, err = io.Copy(out, in)
		return err
	})
}

func genFromTemplate(name, tmp string, data interface{}, goSource, generateFrontend bool, outPath string) error {
	w, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer w.Close()

	var b bytes.Buffer

	templates.Funcs["UI"] = func() bool {
		return generateFrontend
	}

	t, err := template.New(name).Funcs(templates.Funcs).Parse(tmp)
	if err != nil {
		return err
	}
	err = t.Execute(&b, data)
	if err != nil {
		return fmt.Errorf("execute template error: %w", err)
	}

	var src []byte
	if goSource {
		src, err = imports.Process("", b.Bytes(), nil)
		if err != nil {
			fmt.Println(b.String())
			return fmt.Errorf("organize imports error: %w", err)
		}
	} else {
		src = b.Bytes()
	}

	fmt.Fprintf(w, "%s", string(src))
	return nil

}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !errors.Is(err, os.ErrNotExist)
}
