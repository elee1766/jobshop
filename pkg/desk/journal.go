package desk

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/elee1766/jobshop/pkg/rpc"
)

// sessionJournal proxies session.journal (#529). Query: since_seq,
// limit, kinds (comma-separated), from.
func (s *Server) sessionJournal(r *http.Request) (any, error) {
	q := r.URL.Query()
	p := rpc.SessionJournalParams{ID: r.PathValue("id"), From: q.Get("from")}
	for _, k := range strings.Split(q.Get("kinds"), ",") {
		if k = strings.TrimSpace(k); k != "" {
			p.Kinds = append(p.Kinds, k)
		}
	}
	if v := q.Get("since_seq"); v != "" {
		n, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("%w: since_seq: %w", errBadRequest, err)
		}
		p.SinceSeq = n
	}
	if v := q.Get("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("%w: limit: %w", errBadRequest, err)
		}
		p.Limit = n
	}
	var res rpc.SessionJournalResult
	if err := s.ctl.Call(r.Context(), rpc.MethodSessionJournal, p, &res); err != nil {
		return nil, err
	}
	return res, nil
}
