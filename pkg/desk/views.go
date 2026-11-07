package desk

import (
	"fmt"
	"net/http"
	"time"

	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

func (s *Server) attentionAnswer(r *http.Request) (any, error) {
	var p rpc.AttentionAnswerParams
	if err := s.decode(r, &p); err != nil {
		return nil, err
	}
	var out rpc.AttentionChatResult
	err := s.ctl.Call(r.Context(), rpc.MethodAttentionAnswer, p, &out)
	return out, err
}

func (s *Server) chatManage(r *http.Request) (any, error) {
	var p rpc.ChatManageParams
	if err := s.decode(r, &p); err != nil {
		return nil, err
	}
	p.Pool = s.cfg.ChatPool
	var out rpc.SessionInfo
	err := s.ctl.Call(r.Context(), rpc.MethodChatManage, p, &out)
	return out, err
}

// overview keeps run details complete within an explicit, reported bound.
func (s *Server) overview(r *http.Request) (any, error) {
	var runs []store.Ticket
	if err := s.ctl.Call(r.Context(), rpc.MethodTicketList, rpc.ListParams{Kind: store.KindRun, Status: store.StatusOpen, Limit: 101}, &runs); err != nil {
		return nil, err
	}
	out := struct {
		Runs      []rpc.RunResult `json:"runs"`
		Truncated bool            `json:"truncated"`
	}{Runs: []rpc.RunResult{}, Truncated: len(runs) > 100}
	for _, run := range runs[:min(len(runs), 100)] {
		var rr rpc.RunResult
		if err := s.ctl.Call(r.Context(), rpc.MethodRunShow, rpc.IDParams{ID: run.ID}, &rr); err != nil {
			return nil, err
		}
		out.Runs = append(out.Runs, rr)
	}
	return out, nil
}

func (s *Server) traceView(r *http.Request) (any, error) {
	var trace rpc.TraceNode
	if err := s.ctl.Call(r.Context(), rpc.MethodTicketTrace, rpc.TraceParams{ID: r.PathValue("ref"), Depth: 1}, &trace); err != nil {
		return nil, err
	}
	r.SetPathValue("id", trace.ID)
	return s.ticket(r)
}

func (s *Server) systemView(r *http.Request) (any, error) {
	var out struct {
		Status    rpc.SystemStatus       `json:"status"`
		Pools     rpc.PoolStatusResult   `json:"pools"`
		Resources rpc.ResourceListResult `json:"resources"`
		Sessions  []rpc.SessionInfo      `json:"sessions"`
	}
	calls := []struct {
		method         string
		params, result any
	}{
		{rpc.MethodSystemStatus, struct{}{}, &out.Status},
		{rpc.MethodPoolList, struct{}{}, &out.Pools},
		{rpc.MethodResourceList, struct{}{}, &out.Resources},
		{rpc.MethodSessionList, rpc.SessionListParams{}, &out.Sessions},
	}
	for _, call := range calls {
		if err := s.ctl.Call(r.Context(), call.method, call.params, call.result); err != nil {
			return nil, err
		}
	}
	return out, nil
}

func (s *Server) improvements(r *http.Request) (any, error) {
	var out struct {
		Learnings []store.Learning   `json:"learnings"`
		Edits     []store.MemoryEdit `json:"edits"`
	}
	if err := s.ctl.Call(r.Context(), rpc.MethodLearnList, rpc.LearnListParams{Limit: 500}, &out.Learnings); err != nil {
		return nil, err
	}
	if err := s.ctl.Call(r.Context(), rpc.MethodLearnEdits, rpc.LearnEditsParams{Limit: 500}, &out.Edits); err != nil {
		return nil, err
	}
	return out, nil
}

func focusUser(r *http.Request) string {
	if u := userFrom(r.Context()); u != "" {
		return u
	}
	return "admin"
}

func (s *Server) focusGet(r *http.Request) (any, error) {
	return map[string]time.Time{"last_focus": s.attn.load().Focus[focusUser(r)]}, nil
}

func (s *Server) focusSet(r *http.Request) (any, error) {
	var p struct {
		At time.Time `json:"at"`
	}
	if err := s.decode(r, &p); err != nil {
		return nil, err
	}
	// Use server time rather than trusting device clocks.
	at := time.Now().UTC()
	err := s.attn.update(func(st *attentionState) {
		if st.Focus == nil {
			st.Focus = map[string]time.Time{}
		}
		st.Focus[focusUser(r)] = at
	})
	if err != nil {
		return nil, fmt.Errorf("save focus: %w", err)
	}
	return map[string]time.Time{"last_focus": at}, nil
}
