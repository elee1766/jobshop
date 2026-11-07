package automation_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/elee1766/jobshop/pkg/automation"
)

func TestDocumentedHookExamples(t *testing.T) {
	root := "../config/testdata/cue-starlark-example/packs/demo"
	entries := []string{"workflow.star:begin", "workflow.star:completed", "events.star:handle"}
	engine, err := automation.Load(t.Context(), root, entries, automation.Options{})
	if err != nil {
		t.Fatal(err)
	}
	run := struct {
		Inputs map[string]any `json:"inputs"`
	}{Inputs: map[string]any{"task": "fix documented example", "pool": "demo-dev"}}
	begin, err := engine.Invoke(t.Context(), entries[0], run)
	if err != nil {
		t.Fatal(err)
	}
	if len(begin.Actions) != 1 || begin.Actions[0].Kind != "start_step" || begin.Actions[0].ID != "implement" || begin.Actions[0].Pool != "demo-dev" || begin.Actions[0].PromptFile != "implement.md" || begin.Actions[0].Inputs["task"] != "fix documented example" {
		t.Fatalf("wrong documented begin action: %#v", begin)
	}
	step := struct {
		ID         string         `json:"id"`
		Outcome    string         `json:"outcome"`
		Resolution string         `json:"resolution"`
		Output     map[string]any `json:"output"`
	}{ID: "implement", Outcome: "done", Output: map[string]any{"summary": "implemented"}}
	complete, err := engine.Invoke(t.Context(), entries[1], run, step)
	if err != nil {
		t.Fatal(err)
	}
	if len(complete.Actions) != 1 || complete.Actions[0].Kind != "finish_run" || complete.Actions[0].Outcome != "done" || complete.Actions[0].Output["summary"] != "implemented" {
		t.Fatalf("wrong documented completion action: %#v", complete)
	}
	event := struct {
		Source  string         `json:"source"`
		Payload map[string]any `json:"payload"`
	}{Source: "example/tasks", Payload: map[string]any{"tasks": []any{map[string]any{"key": "TASK-1", "task": "fix documented example"}}}}
	trigger, err := engine.Invoke(t.Context(), entries[2], event, map[string]any{"workflow": "demo-implement", "pool": "demo-dev"})
	if err != nil {
		t.Fatal(err)
	}
	if trigger.Output["selected"] != int64(1) || len(trigger.Actions) != 1 || trigger.Actions[0].Kind != "start_run" || trigger.Actions[0].Key != "demo:TASK-1" || trigger.Actions[0].Workflow != "demo-implement" || trigger.Actions[0].Inputs["pool"] != "demo-dev" {
		t.Fatalf("wrong documented trigger action: %#v", trigger)
	}
}

func TestDocumentedCapabilityExamples(t *testing.T) {
	root := "../config/testdata/cue-starlark-example/packs/demo"
	entries := []string{"capabilities.star:http_probe", "capabilities.star:shell_probe", "capabilities.star:decide"}
	engine, err := automation.Load(t.Context(), root, entries, automation.Options{})
	if err != nil {
		t.Fatal(err)
	}
	calls := []struct {
		entry string
		args  []any
	}{
		{entries[0], []any{"http://127.0.0.1"}},
		{entries[1], nil},
		{entries[2], []any{map[string]any{"issue": "example"}}},
	}
	for _, call := range calls {
		if _, err := engine.Invoke(t.Context(), call.entry, call.args...); err == nil || !strings.Contains(err.Error(), "disabled") {
			t.Fatalf("documented capability not denied by default: %s %v", call.entry, err)
		}
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			_, _ = io.WriteString(w, `{"custom":{"ok":true}}`)
			return
		}
		if r.Header.Get("Authorization") != "Bearer example-key" {
			t.Error("Jev fixture omitted host key")
		}
		_, _ = io.WriteString(w, `{"answers":{"decision":{"type":"choice","choice":"scoped"}}}`)
	}))
	defer server.Close()
	engine, err = automation.Load(t.Context(), root, entries, automation.Options{AllowHTTP: true, AllowShell: true, Jev: automation.JevOptions{Enabled: true, Endpoint: server.URL, APIKey: "example-key"}})
	if err != nil {
		t.Fatal(err)
	}
	result, err := engine.Invoke(t.Context(), entries[0], server.URL)
	if err != nil || result.Output["custom"].(map[string]any)["ok"] != true {
		t.Fatalf("HTTP fixture result: %#v %v", result, err)
	}
	t.Run("shell", func(t *testing.T) {
		if _, err := os.Stat("/bin/printf"); err != nil {
			t.Skip("documented Unix /bin/printf executable is unavailable")
		}
		result, err := engine.Invoke(t.Context(), entries[1])
		if err != nil || result.Output["stdout"] != "ok" {
			t.Fatalf("shell fixture result: %#v %v", result, err)
		}
	})
	result, err = engine.Invoke(t.Context(), entries[2], map[string]any{"issue": "example"})
	if err != nil || result.Output["choice"] != "scoped" || len(result.Decisions) != 1 {
		t.Fatalf("Jev fixture result: %#v %v", result, err)
	}
}
