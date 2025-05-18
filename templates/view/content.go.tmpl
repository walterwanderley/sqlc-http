package view

import (
	"fmt"
	"net/http"
	"os"
	"strings"
)

type Content[T any] struct {
	Data    T
	Request *http.Request
}

func (c Content[T]) HxRequest() bool {
	return HXRequest(c.Request)
}

func (c Content[T]) BreadCrumbsFromRequest() []breadCrumb {
	return breadCrumbsFromRequest(c.Request)
}

func (c Content[T]) Pagination() *Pagination {
	pagination, _ := c.Request.Context().Value(paginationContext).(*Pagination)
	if pagination != nil {
		pagination.request = c.Request
	}
	return pagination
}

func (c Content[T]) BaseHref() string {
	schema := "http"
	if forwardedProto := c.Request.Header.Get("X-Forwarded-Proto"); forwardedProto != "" {
		schema = forwardedProto
	} else if c.Request.TLS != nil {
		schema = "https"
	}
	context := os.Getenv("WEB_CONTEXT")
	return fmt.Sprintf("%s://%s%s/", schema, c.Request.Host, strings.TrimSuffix(context, "/"))
}

func (c Content[T]) HasQuery(key string) bool {
	return c.Request.URL.Query().Has(key)
}

func (c Content[T]) Query(key string) string {
	return c.Request.URL.Query().Get(key)
}

func (c Content[T]) MessageContext() *Message {
	if msg, ok := c.Request.Context().Value(messageContext).(Message); ok {
		return &msg
	}
	return nil
}
