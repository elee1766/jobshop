package desk

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/elee1766/jobshop/pkg/rpc"
)

// journalSearch proxies journal.search (#530). Query: q (LogsQL), since
// (Go duration), start and until (RFC3339), limit.
func (s *Server) journalSearch(r *http.Request) (any, error) {
	q := r.URL.Query()
	p := rpc.JournalSearchParams{Query: q.Get("q")}
	if v := q.Get("since"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			return nil, fmt.Errorf("%w: since: %w", errBadRequest, err)
		}
		p.Since = d
	}
	for _, f := range []struct {
		key string
		dst *time.Time
	}{{"start", &p.Start}, {"until", &p.Until}} {
		if v := q.Get(f.key); v != "" {
			t, err := time.Parse(time.RFC3339, v)
			if err != nil {
				return nil, fmt.Errorf("%w: %s: %w", errBadRequest, f.key, err)
			}
			*f.dst = t
		}
	}
	if v := q.Get("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("%w: limit: %w", errBadRequest, err)
		}
		p.Limit = n
	}
	var res rpc.JournalSearchResult
	if err := s.ctl.Call(r.Context(), rpc.MethodJournalSearch, p, &res); err != nil {
		return nil, err
	}
	return res, nil
}
