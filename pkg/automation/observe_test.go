package automation_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/elee1766/jobshop/pkg/automation"
	"go.starlark.net/starlark"
)

func TestRecordedCapabilitiesReuseRealResults(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		if r.URL.Path == "/decide" {
			_, _ = io.WriteString(w, `{"answers":{"decision":{"type":"choice","choice":"b","confidence":0.7}}}`)
			return
		}
		_, _ = io.WriteString(w, `{"value":{"observed":true}}`)
	}))
	defer server.Close()
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	countFile := filepath.Join(t.TempDir(), "shell-count")
	records := map[int]automation.Observation{}
	options := automation.Options{
		AllowHTTP: true, AllowShell: true, ShellEnv: []string{"AUTOMATION_TEST_PROCESS=count", "AUTOMATION_TEST_FILE=" + countFile},
		Jev: automation.JevOptions{Enabled: true, Endpoint: server.URL + "/decide", APIKey: "private-reuse-key"},
		Observe: func(_ context.Context, o automation.Observation) error {
			if o.Phase == "completed" {
				if o.Redacted {
					return errors.New("cannot retain changed result")
				}
				data, err := json.Marshal(o)
				if err != nil {
					return err
				}
				var recorded automation.Observation
				if err := json.Unmarshal(data, &recorded); err != nil {
					return err
				}
				records[o.Index] = recorded
			}
			return nil
		},
		Replay: func(_ context.Context, o automation.Observation) (any, bool, error) {
			recorded, ok := records[o.Index]
			if !ok {
				return nil, false, nil
			}
			if recorded.InputDigest != o.InputDigest || recorded.Capability != o.Capability {
				return nil, false, errors.New("recorded input identity mismatch")
			}
			return recorded.Result, true, nil
		},
	}
	engine := hook(t, `def handle(url, argv):
    response = http.post(url, json={"request": "observe"})
    command = shell(argv)
    decision = jev(instructions="Pick one", options={"a": "first", "b": "second"})
    return [start_run(workflow=decision.choice, key="recorded", inputs={"value":response.json(),"command":command.stdout})]
`, options)
	argv := []string{executable, "-test.run=^TestProcessHelper$"}
	first, err := engine.Invoke(t.Context(), "main.star:handle", server.URL, argv)
	if err != nil {
		t.Fatal(err)
	}
	second, err := engine.Invoke(t.Context(), "main.star:handle", server.URL, argv)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("restored capability results changed callback behavior: %#v %#v", first, second)
	}
	if requests.Load() != 2 || len(records) != 3 {
		t.Fatalf("reused calls executed again: %d requests, %d records", requests.Load(), len(records))
	}
	count, err := os.ReadFile(countFile)
	if err != nil {
		t.Fatal(err)
	}
	if string(count) != "executed\n" {
		t.Fatalf("shell effect repeated on recorded-result reuse: %q", count)
	}
	if _, err := engine.Invoke(t.Context(), "main.star:handle", server.URL+"/changed", argv); err == nil || !strings.Contains(err.Error(), "lookup failed") {
		t.Fatalf("changed input reused a recorded result: %v", err)
	}
	if requests.Load() != 2 {
		t.Fatal("input mismatch repeated an external effect")
	}
}

func TestObservedLargeHTTPBodyCompletesAndRestores(t *testing.T) {
	payload := strings.Repeat("<", 600<<10)
	body := `{"payload":"` + payload + `"}`
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
		_, _ = io.WriteString(w, body)
	}))
	defer server.Close()
	var completed automation.Observation
	engine := hook(t, `def handle(url):
    response = http.get(url)
    return {"bytes":len(response.body), "aliases":response.text == response.body, "payload":len(response.json()["payload"])}
`, automation.Options{
		AllowHTTP: true,
		Observe: func(_ context.Context, o automation.Observation) error {
			if o.Phase == "completed" {
				encoded, err := json.Marshal(o)
				if err != nil {
					return err
				}
				return json.Unmarshal(encoded, &completed)
			}
			return nil
		},
		Replay: func(_ context.Context, o automation.Observation) (any, bool, error) {
			if completed.Phase == "" {
				return nil, false, nil
			}
			if completed.InputDigest != o.InputDigest || completed.Capability != o.Capability || completed.Index != o.Index || completed.Redacted {
				return nil, false, errors.New("recorded observation is not reusable")
			}
			return completed.Result, true, nil
		},
	})
	first, err := engine.Invoke(t.Context(), "main.star:handle", server.URL)
	if err != nil {
		t.Fatal(err)
	}
	if completed.Phase != "completed" {
		t.Fatalf("large response did not complete observation: %#v", completed)
	}
	canonical := completed.Result.(map[string]any)
	if _, duplicate := canonical["text"]; duplicate || canonical["body"] != body {
		t.Fatal("HTTP observation did not retain one canonical body")
	}
	second, err := engine.Invoke(t.Context(), "main.star:handle", server.URL)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(first, second) || requests.Load() != 1 {
		t.Fatalf("large response restore changed behavior or repeated HTTP: %#v %#v, requests=%d", first, second, requests.Load())
	}
	value := first.Output
	if value["bytes"] != int64(len(body)) || value["aliases"] != true || value["payload"] != int64(len(payload)) {
		t.Fatalf("restored response aliases/JSON differ: %#v", value)
	}
}

func TestInterruptedMutationIsUnknownAndNotRetried(t *testing.T) {
	mutated := make(chan struct{})
	var effects atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		effects.Add(1)
		close(mutated)
		<-r.Context().Done()
	}))
	defer server.Close()
	var records []automation.Observation
	options := automation.Options{AllowHTTP: true, Observe: func(_ context.Context, o automation.Observation) error { records = append(records, o); return nil }}
	source := `def handle(url):
    return http.post(url, {"Authorization":"Token caller-token","Cookie":"session=caller-cookie"}, body="mutate").text
`
	engine := hook(t, source, options)
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	done := make(chan error, 1)
	go func() { _, err := engine.Invoke(ctx, "main.star:handle", server.URL); done <- err }()
	<-mutated
	cancel()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("interrupted mutation succeeded")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("mutation cancellation blocked")
	}
	if len(records) != 2 || records[0].Phase != "started" || records[1].Phase != "unknown" || records[0].Index != 1 || records[1].Index != 1 {
		t.Fatalf("lost ordered unknown effect evidence: %#v", records)
	}
	terminal := records[1]
	if terminal.StartedAt.IsZero() || terminal.FinishedAt.IsZero() || terminal.FinishedAt.Before(terminal.StartedAt) || terminal.Duration < 0 || len(terminal.InputDigest) != 64 {
		t.Fatalf("missing bounded timing/input identity: %#v", terminal)
	}
	data, err := json.Marshal(records)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "caller-token") || strings.Contains(string(data), "caller-cookie") {
		t.Fatal("caller authorization/cookie leaked into call history")
	}
	options.Replay = func(_ context.Context, o automation.Observation) (any, bool, error) {
		if o.InputDigest != terminal.InputDigest {
			t.Error("same inputs changed digest")
		}
		return nil, false, errors.New("UNKNOWN external outcome requires reconciliation")
	}
	retry := hook(t, source, options)
	if _, err := retry.Invoke(t.Context(), "main.star:handle", server.URL); err == nil || !strings.Contains(err.Error(), "reconcile") {
		t.Fatalf("unknown effect was retryable: %v", err)
	}
	if effects.Load() != 1 {
		t.Fatalf("unknown mutation was repeated: %d", effects.Load())
	}
}

func TestKnownSecretsAreAbsentFromHookDiagnostics(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, strings.TrimPrefix(r.Header.Get("X-Private-Key"), "prefix="))
	}))
	defer server.Close()
	var recorded automation.Observation
	options := automation.Options{
		AllowHTTP: true, Auth: map[string]automation.Credential{"service": {Origins: []string{server.URL}, Headers: map[string]string{"X-Private-Key": "prefix=raw-host-key"}, Secrets: []string{"raw-host-key"}}},
		Observe: func(_ context.Context, o automation.Observation) error {
			if o.Phase == "completed" {
				recorded = o
			}
			return nil
		},
	}
	engine := hook(t, `def handle(url):
    response = http.get(url, auth="service")
    print(response.text)
    fail(response.text)
`, options)
	result, err := engine.Invoke(t.Context(), "main.star:handle", server.URL)
	if err == nil || strings.Contains(err.Error(), "raw-host-key") || strings.Contains(result.Logs, "raw-host-key") {
		t.Fatalf("known secret escaped diagnostic boundary: %#v %v", result, err)
	}
	var evaluation *starlark.EvalError
	if !errors.As(err, &evaluation) || strings.Contains(evaluation.Backtrace(), "raw-host-key") {
		t.Fatal("safe EvalError traceback was not preserved")
	}
	if !strings.Contains(result.Logs, "[redacted]") || !recorded.Redacted {
		t.Fatalf("secret echo was not marked as changed: %#v %#v", result, recorded)
	}
	data, marshalErr := json.Marshal(recorded)
	if marshalErr != nil {
		t.Fatal(marshalErr)
	}
	if strings.Contains(string(data), "raw-host-key") {
		t.Fatal("known secret escaped observation boundary")
	}
}
