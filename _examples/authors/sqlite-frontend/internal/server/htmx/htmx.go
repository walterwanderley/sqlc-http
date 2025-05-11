package htmx

import (
	"net/http"
)

func HXRequest(r *http.Request) bool {
	return r.Header.Get("hx-request") == "true"
}
