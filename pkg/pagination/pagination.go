package pagination

import (
	"net/http"
	"strconv"
)

// ParseQueryParams extracts pagination parameters from query string.
// Default: limit=10, offset=0. Max limit: 100.
func ParseQueryParams(r *http.Request) (limit, offset int) {
	limit = 10
	offset = 0

	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
			if limit > 100 {
				limit = 100
			}
		}
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}
	return
}
