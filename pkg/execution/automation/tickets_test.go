package automation_test

import (
	"context"
	"strings"
	"testing"

	"github.com/gfx-labs/jobshop/pkg/execution/automation"
	"github.com/gfx-labs/jobshop/pkg/execution/automation/api"
)

// commentReader answers like the controller: a "comments" list only when
// the query asks for it.
func commentReader(seen *[]api.TicketQuery) api.TicketReader {
	return func(_ context.Context, q api.TicketQuery) ([]any, error) {
		*seen = append(*seen, q)
		out := []any{
			map[string]any{"id": "js-a"},
			map[string]any{"id": "js-b"},
		}
		if q.Comments {
			out[0].(map[string]any)["comments"] = []any{
				map[string]any{"id": int64(1), "author": "u", "body": "first", "created_at": "2026-11-10T00:00:00Z"},
				map[string]any{"id": int64(2), "author": "v", "body": "PR https://example.test/1", "created_at": "2026-11-10T00:01:00Z"},
			}
			out[1].(map[string]any)["comments"] = []any{}
		}
		return out, nil
	}
}

func TestStoreReadComments(t *testing.T) {
	var seen []api.TicketQuery
	engine := hook(t, `def handle(event, args):
    out = {}
    for name, rows in [("tickets", tickets(comments=True)), ("runs", runs(comments=True))]:
        out[name] = [[c["id"], c["author"], c["body"], c["created_at"]] for c in rows[0]["comments"]]
        out[name + "_empty"] = len(rows[1]["comments"])
    out["off"] = [("comments" in t) for t in tickets() + runs()]
    return out
`, automation.Options{Tickets: commentReader(&seen)})
	result, err := engine.Invoke(t.Context(), "main.star:handle", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"tickets", "runs"} {
		rows := result.Output[name].([]any)
		if len(rows) != 2 {
			t.Fatalf("%s comments %#v, want two in order", name, rows)
		}
		first := rows[0].([]any)
		if first[0] != int64(1) || first[1] != "u" || first[2] != "first" || first[3] != "2026-11-10T00:00:00Z" {
			t.Fatalf("%s first comment %#v", name, first)
		}
		if rows[1].([]any)[2] != "PR https://example.test/1" {
			t.Fatalf("%s second comment %#v", name, rows[1])
		}
		if result.Output[name+"_empty"] != int64(0) {
			t.Fatalf("%s: a ticket without comments must give an empty list", name)
		}
	}
	for _, has := range result.Output["off"].([]any) {
		if has != false {
			t.Fatalf("comments key present without comments=True: %#v", result.Output["off"])
		}
	}
	if len(seen) != 4 || !seen[0].Comments || !seen[1].Comments || seen[2].Comments || seen[3].Comments || seen[1].Kind != "run" {
		t.Fatalf("queries %#v", seen)
	}
}

func TestStoreReadCommentsOverDataLimit(t *testing.T) {
	big := strings.Repeat("x", 4096)
	reader := func(_ context.Context, q api.TicketQuery) ([]any, error) {
		row := map[string]any{"id": "js-a"}
		if q.Comments {
			row["comments"] = []any{map[string]any{"id": int64(1), "author": "u", "body": big, "created_at": "2026-11-10T00:00:00Z"}}
		}
		return []any{row}, nil
	}
	engine := hook(t, `def handle(event, args):
    tickets()
    return runs(comments=True)
`, automation.Options{Tickets: reader, Limits: automation.Limits{DataBytes: 1024}})
	_, err := engine.Invoke(t.Context(), "main.star:handle", nil, nil)
	if err == nil || !strings.Contains(err.Error(), "runs: structured data byte/depth limit exceeded") {
		t.Fatalf("got %v, want the limit error naming runs", err)
	}
}
