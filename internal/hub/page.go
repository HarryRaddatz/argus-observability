package hub

import (
	"net/http"
	"strconv"
)

const (
	pageDefault = 50
	pageMax     = 200
	offsetMax   = 100_000
)

// pageWindow reads limit and offset. A missing limit uses def. Values above max are clamped.
func pageWindow(r *http.Request, def int) (limit, offset int) {
	limit = def
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 {
			limit = n
		}
	}
	if limit > pageMax {
		limit = pageMax
	}
	if raw := r.URL.Query().Get("offset"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 {
			offset = n
		}
	}
	if offset > offsetMax {
		offset = offsetMax
	}
	return limit, offset
}
