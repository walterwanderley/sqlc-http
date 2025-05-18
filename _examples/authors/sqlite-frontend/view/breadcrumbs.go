package view

import "net/http"

type breadCrumb struct {
	Name string
	Href string
}

func breadCrumbsFromRequest(r *http.Request) []breadCrumb {
	switch r.Pattern {
	case "POST /authors":
		return breadCrumbsFromStrings("Authors", "/", "Create Author")
	case "DELETE /authors/{id}":
		return breadCrumbsFromStrings("Authors", "/", "List Authors", "/authors", "Delete Author")
	case "GET /authors/{id}":
		serviceName := "Get Author"
		if r.URL.Query().Has("edit") {
			serviceName = "Update Author"
		}
		return breadCrumbsFromStrings("Authors", "/", "List Authors", "/authors", serviceName)
	case "GET /authors":
		return breadCrumbsFromStrings("Authors", "/", "List Authors")
	case "PUT /authors/{id}":
		return breadCrumbsFromStrings("Authors", "/", "List Authors", "/authors", "Update Author")
	case "PATCH /authors/{id}/bio":
		return breadCrumbsFromStrings("Authors", "/", "List Authors", "/authors", "Get Author", "/authors/"+r.PathValue("id"), "Update Author Bio")
	default:
		switch r.URL.Path {
		case "/app/authors/create_author":
			return breadCrumbsFromStrings("Authors", "/", "Create Author")
		case "/app/authors/delete_author":
			return breadCrumbsFromStrings("Authors", "/", "List Authors", "/authors", "Delete Author")
		case "/app/authors/get_author":
			return breadCrumbsFromStrings("Authors", "/", "List Authors", "/authors", "Get Author")
		case "/app/authors/list_authors":
			return breadCrumbsFromStrings("Authors", "/", "List Authors")
		case "/app/authors/update_author":
			return breadCrumbsFromStrings("Authors", "/", "List Authors", "/authors", "Update Author")
		case "/app/authors/update_author_bio":
			return breadCrumbsFromStrings("Authors", "/", "List Authors", "/authors", "Update Author Bio")
		}
	}

	return nil
}

func breadCrumbsFromStrings(items ...string) []breadCrumb {
	breadcrumbs := make([]breadCrumb, 0)
	for i := 0; i < len(items); i = i + 2 {
		var bc breadCrumb
		bc.Name = items[i]
		j := i + 1
		if j < len(items) {
			bc.Href = items[j]
		}
		breadcrumbs = append(breadcrumbs, bc)
	}
	return breadcrumbs
}
