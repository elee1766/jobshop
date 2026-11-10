package desk

import (
	"fmt"
	"net/http"
	"slices"
	"time"

	"github.com/gfx-labs/jobshop/pkg/domain"
	"github.com/gfx-labs/jobshop/pkg/transport/rpc"
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
	var p struct {
		rpc.ChatManageParams
		Refs []string `json:"refs,omitempty"`
	}
	if err := decodeLimit(r, &p, rpc.MaxAttachmentBytes*4/3+maxChatText+16384); err != nil {
		return nil, err
	}
	context, err := s.referenceContext(r, p.Refs)
	if err != nil {
		return nil, err
	}
	p.Context += context
	p.Pool = s.cfg.ChatPool
	var out rpc.SessionInfo
	err = s.ctl.Call(r.Context(), rpc.MethodChatManage, p.ChatManageParams, &out)
	return out, err
}

// overview keeps run details complete within an explicit, reported bound.
func (s *Server) overview(r *http.Request) (any, error) {
	var runs []domain.Ticket
	if err := s.ctl.Call(r.Context(), rpc.MethodTicketList, rpc.ListParams{Kind: domain.KindRun, Status: domain.StatusOpen, Limit: 101}, &runs); err != nil {
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
	var graph rpc.GraphResult
	if err := s.ctl.Call(r.Context(), rpc.MethodTicketGraph, rpc.GraphParams{ID: r.PathValue("ref"), Limit: 2000}, &graph); err != nil {
		return nil, err
	}
	r.SetPathValue("id", graph.Root)
	// The timeline replays every node's history: only when Desk shows it
	// (?timeline=1, #1103). It comes first so the holders of its open
	// working bars join the one session.list call (JOB-184).
	var tl *rpc.TimelineResult
	var holders []string
	if r.URL.Query().Get("timeline") != "" {
		tl = new(rpc.TimelineResult)
		if err := s.ctl.Call(r.Context(), rpc.MethodTicketTimeline, rpc.GraphParams{ID: graph.Root, Limit: 2000}, tl); err != nil {
			return nil, err
		}
		holders = openWorkers(tl)
	}
	value, err := s.ticketView(r, holders)
	if err != nil {
		return nil, err
	}
	view := value.(TicketView)
	view.Graph = &graph
	view.Timeline = tl
	return view, nil
}

// openWorkers are the sessions holding a ticket in an open working bar.
func openWorkers(tl *rpc.TimelineResult) []string {
	var out []string
	for _, row := range tl.Rows {
		for _, seg := range row.Segments {
			if seg.State == rpc.TimelineWorking && seg.End.IsZero() && seg.Actor != "" && !slices.Contains(out, seg.Actor) {
				out = append(out, seg.Actor)
			}
		}
	}
	return out
}

func (s *Server) systemView(r *http.Request) (any, error) {
	var out struct {
		Status     rpc.SystemStatus        `json:"status"`
		Pools      rpc.PoolStatusResult    `json:"pools"`
		Resources  rpc.ResourceListResult  `json:"resources"`
		Sessions   []rpc.SessionInfo       `json:"sessions"`
		Ingest     rpc.IngestWorkersResult `json:"ingest"`
		Automation AutomationList          `json:"automation"`
	}
	calls := []struct {
		method         string
		params, result any
	}{
		{rpc.MethodSystemStatus, struct{}{}, &out.Status},
		{rpc.MethodPoolList, struct{}{}, &out.Pools},
		{rpc.MethodResourceList, struct{}{}, &out.Resources},
		{rpc.MethodSessionList, rpc.SessionListParams{}, &out.Sessions},
		{rpc.MethodIngestWorkers, struct{}{}, &out.Ingest},
	}
	for _, call := range calls {
		if err := s.ctl.Call(r.Context(), call.method, call.params, call.result); err != nil {
			return nil, err
		}
	}
	var err error
	out.Automation, err = s.automationList(r)
	return out, err
}

func (s *Server) improvements(r *http.Request) (any, error) {
	var out struct {
		Learnings []domain.Learning   `json:"learnings"`
		Edits     []domain.MemoryEdit `json:"edits"`
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
