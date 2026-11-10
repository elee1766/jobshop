package desk

import (
	"cmp"
	"encoding/json"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/gfx-labs/jobshop/pkg/storage/store"
	"github.com/gfx-labs/jobshop/pkg/transport/rpc"
)

const automationPageSize = 50

// AutomationSummary excludes snapshots, plans and capability bodies from live lists.
type AutomationSummary struct {
	Key       string    `json:"key"`
	State     string    `json:"state"`
	Revision  int64     `json:"revision"`
	Attempts  int64     `json:"attempts"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Trigger   string    `json:"trigger,omitempty"`
	Run       string    `json:"run,omitempty"`
	Step      string    `json:"step,omitempty"`
	Entry     string    `json:"entry"`
	Event     *struct {
		Source   string `json:"source"`
		Event    string `json:"event"`
		Delivery string `json:"delivery"`
	} `json:"event,omitempty"`
	Error string `json:"error,omitempty"`
}

// AutomationRecord keeps the collection envelope independent of its JSON payload.
type AutomationRecord struct {
	Key       string          `json:"key"`
	State     string          `json:"state"`
	Revision  int64           `json:"revision"`
	Attempts  int64           `json:"attempts"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
	Value     json.RawMessage `json:"value"`
}

// AutomationList is the bounded invocation window in system live state.
type AutomationList struct {
	Items     []AutomationSummary `json:"items"`
	Truncated bool                `json:"truncated"`
}

// AutomationAudit is the selected invocation and its ordered execution trace.
type AutomationAudit struct {
	ID         string             `json:"id"`
	Invocation AutomationRecord   `json:"invocation"`
	Attempts   []AutomationRecord `json:"attempts"`
	Calls      []AutomationRecord `json:"calls"`
	Truncated  bool               `json:"truncated"`
}

func (s *Server) automationList(r *http.Request) (AutomationList, error) {
	var records []store.Record
	err := s.ctl.Call(r.Context(), rpc.MethodAutomationList, rpc.AutomationListParams{Limit: automationPageSize + 1}, &records)
	out := AutomationList{Items: []AutomationSummary{}, Truncated: len(records) > automationPageSize}
	if err != nil {
		return out, err
	}
	for _, rec := range records[:min(len(records), automationPageSize)] {
		var item AutomationSummary
		if err := json.Unmarshal(rec.Value, &item); err != nil {
			return out, err
		}
		item.Key, item.State, item.Revision, item.Attempts = rec.Key, rec.State, rec.Revision, rec.Attempts
		item.CreatedAt, item.UpdatedAt = rec.CreatedAt, rec.UpdatedAt
		texts := []*string{&item.Trigger, &item.Run, &item.Step, &item.Entry, &item.Error}
		if item.Event != nil {
			texts = append(texts, &item.Event.Source, &item.Event.Event, &item.Event.Delivery)
		}
		for _, text := range texts {
			if len(*text) > 1024 {
				*text = strings.ToValidUTF8((*text)[:1024], "") + "…"
			}
		}
		out.Items = append(out.Items, item)
	}
	return out, nil
}

func (s *Server) automationView(r *http.Request) (any, error) {
	var result rpc.AutomationResult
	id := r.URL.Query().Get("audit")
	if err := s.ctl.Call(r.Context(), rpc.MethodAutomationShow, rpc.IDParams{ID: id}, &result); err != nil {
		return map[string]string{"id": id, "error": err.Error()}, nil
	}
	out := AutomationAudit{ID: id, Invocation: automationRecord(result.Invocation), Attempts: []AutomationRecord{}, Calls: []AutomationRecord{}}
	// Collection keys are strings, but capability indices and attempts are numbers.
	order := func(records []store.Record, call bool) {
		slices.SortStableFunc(records, func(a, b store.Record) int {
			index := func(rec store.Record) (int64, int64) {
				var value struct {
					Attempt     int64 `json:"attempt"`
					Observation struct {
						Index int64 `json:"index"`
					} `json:"observation"`
				}
				_ = json.Unmarshal(rec.Value, &value)
				if call {
					return value.Attempt, value.Observation.Index
				}
				return value.Attempt, 0
			}
			attemptA, indexA := index(a)
			attemptB, indexB := index(b)
			if comparison := cmp.Compare(attemptA, attemptB); comparison != 0 {
				return comparison
			}
			if comparison := cmp.Compare(indexA, indexB); comparison != 0 {
				return comparison
			}
			return a.CreatedAt.Compare(b.CreatedAt)
		})
	}
	order(result.Attempts, false)
	order(result.Calls, true)
	out.Truncated = result.Truncated || len(result.Attempts) > automationPageSize || len(result.Calls) > 200
	for _, rec := range result.Attempts[:min(len(result.Attempts), automationPageSize)] {
		out.Attempts = append(out.Attempts, automationRecord(rec))
	}
	for _, rec := range result.Calls[:min(len(result.Calls), 200)] {
		out.Calls = append(out.Calls, automationRecord(rec))
	}
	return out, nil
}

func automationRecord(rec store.Record) AutomationRecord {
	return AutomationRecord{Key: rec.Key, State: rec.State, Revision: rec.Revision, Attempts: rec.Attempts, CreatedAt: rec.CreatedAt, UpdatedAt: rec.UpdatedAt, Value: rec.Value}
}
