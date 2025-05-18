package view

import (
	"fmt"
	"net/http"
	"strings"
)

const defaultLimit = 10

type Pagination struct {
	request *http.Request
	Limit   int64
	Offset  int64
}

func (p *Pagination) From() int64 {
	return p.Offset + 1
}

func (p *Pagination) To() int64 {
	return p.Offset + p.validLimit()
}

func (p *Pagination) Next() int64 {
	return p.Offset + p.validLimit()
}

func (p *Pagination) Prev() int64 {
	limit := p.validLimit()
	offset := p.Offset - limit
	if offset < 0 {
		offset = 0
	}
	return offset
}

func (p *Pagination) URL(limit, offset int64) string {
	if p == nil {
		return ""
	}
	if offset < 0 {
		offset = 0
	}
	if limit == 0 {
		limit = defaultLimit
	}
	var url strings.Builder
	url.WriteString(p.request.URL.Path)
	url.WriteString("?")
	for k := range p.request.URL.Query() {
		if k == "limit" || k == "offset" {
			continue
		}
		url.WriteString(fmt.Sprintf("%s=%s&", k, p.request.URL.Query().Get(k)))
	}
	if limit > 0 {
		url.WriteString(fmt.Sprintf("limit=%d&offset=%d", limit, offset))
	} else {
		url.WriteString(fmt.Sprintf("offset=%d", offset))
	}
	return url.String()
}

func (p *Pagination) validLimit() int64 {
	limit := p.Limit
	if limit == 0 {
		limit = 10
	}
	return limit
}
