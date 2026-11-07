package automation_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/elee1766/jobshop/pkg/automation"
)

func hook(t *testing.T, source string, options automation.Options) *automation.Engine {
	t.Helper()
	root := t.TempDir()
	writeHook(t, root, "main.star", source)
	engine, err := automation.Load(t.Context(), root, []string{"main.star:handle"}, options)
	if err != nil {
		t.Fatal(err)
	}
	return engine
}

func writeHook(t *testing.T, root, name, source string) {
	t.Helper()
	path := filepath.Join(root, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestHookActionsAndCustomData(t *testing.T) {
	engine := hook(t, `def handle(event, args):
    if event.payload["enabled"]:
        return [start_run(workflow=args["workflow"], key=event.repository, inputs=event.payload, reason="enabled event policy")]
    return {"custom": event.payload, "nested": [None, True, 1.25]}
`, automation.Options{})
	type event struct {
		Repository string         `json:"repository"`
		Payload    map[string]any `json:"payload"`
	}
	for _, enabled := range []bool{true, false} {
		result, err := engine.Invoke(t.Context(), "main.star:handle", event{Repository: "org/repo", Payload: map[string]any{"enabled": enabled, "count": json.Number("9007199254740993"), "unsigned": json.Number("18446744073709551615")}}, map[string]any{"workflow": "review"})
		if err != nil {
			t.Fatal(err)
		}
		if enabled {
			if len(result.Actions) != 1 || result.Actions[0].Reason != "enabled event policy" || result.Actions[0].Workflow != "review" || result.Actions[0].Inputs["count"] != int64(9007199254740993) || result.Actions[0].Inputs["unsigned"] != uint64(18446744073709551615) {
				t.Fatalf("wrong action: %#v", result)
			}
		} else {
			data := result.Value.(map[string]any)
			if data["custom"].(map[string]any)["count"] != int64(9007199254740993) {
				t.Fatalf("custom payload lost integer precision: %#v", data)
			}
		}
	}
}

func TestCachedActionsAreIsolated(t *testing.T) {
	engine := hook(t, `action = start_run(workflow="review", key="cached", inputs={"nested": {"original": True}})
def handle():
    return [action]
`, automation.Options{})
	first, err := engine.Invoke(t.Context(), "main.star:handle")
	if err != nil {
		t.Fatal(err)
	}
	first.Actions[0].Inputs["nested"].(map[string]any)["original"] = false
	second, err := engine.Invoke(t.Context(), "main.star:handle")
	if err != nil {
		t.Fatal(err)
	}
	if second.Actions[0].Inputs["nested"].(map[string]any)["original"] != true {
		t.Fatal("caller mutation corrupted frozen module action")
	}
}

func TestStepOutputSchemaAndRetryIntent(t *testing.T) {
	engine := hook(t, `def handle():
    return [
        start_step(id="inherited",pool="dev",prompt="implement",output_schema={"type":"object","required":["summary"],"properties":{"summary":{"type":"string"}},"additionalProperties":False}),
        start_step(id="explicit",pool="dev",prompt="implement",retries=0,reason="single attempt policy"),
        finish_run(outcome="done",reason="all required work complete"),
    ]
`, automation.Options{})
	result, err := engine.Invoke(t.Context(), "main.star:handle")
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Actions) != 3 || result.Actions[0].RetriesSet || !result.Actions[1].RetriesSet || result.Actions[1].Retries != 0 {
		t.Fatalf("lost explicit versus inherited retry policy: %#v", result.Actions)
	}
	schema := result.Actions[0].OutputSchema
	if schema["type"] != "object" || schema["required"].([]any)[0] != "summary" || schema["additionalProperties"] != false {
		t.Fatalf("lost structured output schema: %#v", schema)
	}
	if result.Actions[1].Reason != "single attempt policy" || result.Actions[2].Reason != "all required work complete" {
		t.Fatalf("lost ordinary policy rationale: %#v", result.Actions)
	}
}

func TestCheckArityWithoutExecutingCallback(t *testing.T) {
	for _, tc := range []struct {
		name       string
		parameters string
		count      int
		valid      bool
	}{
		{"start", "run", 1, true},
		{"completed", "run, step", 2, true},
		{"missing", "run, step", 1, false},
		{"extra", "run", 2, false},
		{"default", "run, step=None", 1, true},
		{"variadic", "run, *args, **kwargs", 3, true},
		{"kwargs_are_not_positional", "run, **kwargs", 2, false},
		{"required_keyword", "run, *, step", 1, false},
		{"default_keyword", "run, *, step=None", 1, true},
		{"negative", "run", -1, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			engine := hook(t, "def handle("+tc.parameters+"):\n    fail(\"callback must not execute during signature checking\")\n", automation.Options{})
			if err := engine.CheckArity("main.star:handle", tc.count); (err == nil) != tc.valid {
				t.Fatalf("offline signature accepted=%v, want %v: %v", err == nil, tc.valid, err)
			}
			if err := engine.CheckArity("main.star:missing", tc.count); err == nil {
				t.Fatal("unloaded entry passed signature checking")
			}
		})
	}
}

func TestCapabilityObservationDurabilityBoundary(t *testing.T) {
	for _, failure := range []string{"", "started", "completed"} {
		t.Run(failure, func(t *testing.T) {
			var effects atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				effects.Add(1)
				_, _ = io.WriteString(w, r.Header.Get("Authorization"))
			}))
			defer server.Close()
			var observations []automation.Observation
			engine := hook(t, "def handle(url):\n    return http.post(url, auth=\"service\", body=\"mutation\").status_code\n", automation.Options{
				AllowHTTP: true, Auth: map[string]automation.Credential{"service": {Origins: []string{server.URL}, Headers: map[string]string{"Authorization": "Bearer private-token"}}},
				Observe: func(_ context.Context, o automation.Observation) error {
					if o.Phase == "started" && effects.Load() != 0 {
						t.Error("effect preceded started observation")
					}
					observations = append(observations, o)
					if o.Phase == failure {
						return errors.New("storage failed")
					}
					return nil
				},
			})
			_, err := engine.Invoke(t.Context(), "main.star:handle", server.URL)
			if failure == "" && err != nil {
				t.Fatal(err)
			}
			if failure != "" && err == nil {
				t.Fatal("observer failure was ignored")
			}
			expected := int32(1)
			if failure == "started" {
				expected = 0
			}
			if effects.Load() != expected {
				t.Fatalf("unexpected effect count %d", effects.Load())
			}
			data, marshalErr := json.Marshal(observations)
			if marshalErr != nil {
				t.Fatal(marshalErr)
			}
			if strings.Contains(string(data), "private-token") {
				t.Fatal("host credential leaked into observations")
			}
			if failure == "completed" && !strings.Contains(err.Error(), "ambiguous") {
				t.Fatalf("missing post-effect ambiguity: %v", err)
			}
		})
	}
}

func TestLoaderCapabilityAndFilesystemBoundaries(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { requests.Add(1); _, _ = io.WriteString(w, "{}") }))
	defer server.Close()
	root := t.TempDir()
	writeHook(t, root, "main.star", fmt.Sprintf("top = http.get(%q)\ndef handle():\n    return top\n", server.URL))
	if _, err := automation.Load(t.Context(), root, []string{"main.star:handle"}, automation.Options{AllowHTTP: true}); err == nil || !strings.Contains(err.Error(), "module loading") {
		t.Fatalf("load permitted capability call: %v", err)
	}
	if requests.Load() != 0 {
		t.Fatal("module loading performed a network call")
	}
	outside := t.TempDir()
	writeHook(t, outside, "outside.star", "def helper():\n    return 1\n")
	if err := os.Symlink(filepath.Join(outside, "outside.star"), filepath.Join(root, "escape.star")); err != nil {
		t.Fatal(err)
	}
	for _, module := range []string{"../outside.star", "escape.star"} {
		writeHook(t, root, "main.star", fmt.Sprintf("load(%q, \"helper\")\ndef handle():\n    return helper()\n", module))
		if _, err := automation.Load(t.Context(), root, []string{"main.star:handle"}, automation.Options{}); err == nil {
			t.Fatalf("loaded escaping module %s", module)
		}
	}
	writeHook(t, root, "lib/helper.star", "def helper():\n    return {\"ok\": True}\n")
	writeHook(t, root, "main.star", "load(\"lib/helper.star\", \"helper\")\ndef handle():\n    return helper()\n")
	engine, err := automation.Load(t.Context(), root, []string{"main.star:handle"}, automation.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if len(engine.Files()) != 2 {
		t.Fatalf("missing loaded watch files: %v", engine.Files())
	}
	for _, limits := range []automation.Limits{{Modules: 1}, {SourceBytes: 1}} {
		if _, err := automation.Load(t.Context(), root, []string{"main.star:handle"}, automation.Options{Limits: limits}); err == nil || !strings.Contains(err.Error(), "limit exceeded") {
			t.Fatalf("loader limit was not enforced: %v", err)
		}
	}
	if _, err := engine.Invoke(t.Context(), "main.star:handle"); err != nil {
		t.Fatal(err)
	}
	compiled := engine.Digest()
	writeHook(t, root, "lib/helper.star", "def helper():\n    return {\"ok\": False}\n")
	if compiled == "" || engine.Digest() != compiled {
		t.Fatal("compiled source identity followed mutable files")
	}
	previous, err := engine.Invoke(t.Context(), "main.star:handle")
	if err != nil || previous.Value.(map[string]any)["ok"] != true {
		t.Fatalf("loaded globals changed after disk write: %#v %v", previous, err)
	}
	reloaded, err := automation.Load(t.Context(), root, []string{"main.star:handle"}, automation.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if reloaded.Digest() == compiled {
		t.Fatal("changed imported source kept previous compiled identity")
	}
	if _, err := automation.Load(t.Context(), root, []string{"main.star:missing"}, automation.Options{}); err == nil {
		t.Fatal("accepted nonexistent callback")
	}
}

func TestHookLimitsAndDiagnostics(t *testing.T) {
	tests := []struct {
		name, source, want string
		limits             automation.Limits
	}{
		{"steps", "def handle():\n    total = 0\n    for n in range(10000):\n        total += n\n    return total\n", "too many steps", automation.Limits{Steps: 100}},
		{"print", "def handle():\n    print(\"123456789\")\n    return None\n", "output byte limit", automation.Limits{OutputBytes: 4}},
		{"result", "def handle():\n    return {\"payload\": \"123456789\"}\n", "data byte/depth limit", automation.Limits{DataBytes: 4}},
		{"action", "def handle():\n    return [start_step(id=\"a\",pool=\"p\",prompt=\"x\",retries=-1)]\n", "retries", automation.Limits{}},
		{"batch", "def handle():\n    return [finish_run(outcome=\"done\"), {}]\n", "only action constructors", automation.Limits{}},
		{"action_count", "def handle():\n    return [finish_run(outcome=\"done\"), finish_run(outcome=\"done\")]\n", "action count limit", automation.Limits{Actions: 1}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			engine := hook(t, test.source, automation.Options{Limits: test.limits})
			_, err := engine.Invoke(t.Context(), "main.star:handle")
			if err == nil || !strings.Contains(err.Error(), test.want) || !strings.Contains(err.Error(), "main.star") {
				t.Fatalf("missing contextual budget error: %v", err)
			}
		})
	}
	engine := hook(t, "def handle():\n    return 1\n", automation.Options{})
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if _, err := engine.Invoke(ctx, "main.star:handle"); err == nil {
		t.Fatal("cancelled invocation succeeded")
	}
}

func TestHTTPAuthJSONAndRedirects(t *testing.T) {
	var foreignRequests atomic.Int32
	foreign := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		foreignRequests.Add(1)
		if r.Header.Get("Authorization") != "" || r.Header.Get("X-Credential") != "" {
			t.Error("credential leaked across origin")
		}
	}))
	defer foreign.Close()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/public-redirect" {
			http.Redirect(w, r, foreign.URL, http.StatusFound)
			return
		}
		if r.Header.Get("Authorization") != "Bearer secret" {
			t.Error("missing named credential")
		}
		if r.URL.Path == "/redirect" {
			http.Redirect(w, r, foreign.URL, http.StatusFound)
			return
		}
		if r.Method != http.MethodPost || r.Header.Get("Content-Type") != "application/json" {
			t.Error("wrong JSON request contract")
		}
		data, err := io.ReadAll(r.Body)
		if err != nil {
			t.Error(err)
		}
		if string(data) != "{\"nested\":[1,true]}" {
			t.Errorf("wrong JSON body %s", data)
		}
		_, _ = io.WriteString(w, `{"value":9007199254740993}`)
	}))
	defer server.Close()
	options := automation.Options{AllowHTTP: true, Auth: map[string]automation.Credential{"service": {Origins: []string{server.URL}, Headers: map[string]string{"Authorization": "Bearer secret"}}}}
	engine := hook(t, `def handle(url):
    response = http.post(url, auth="service", json={"nested": [1, True]})
    return {"status": response.status_code, "value": response.json()["value"]}
`, options)
	result, err := engine.Invoke(t.Context(), "main.star:handle", server.URL)
	if err != nil {
		t.Fatal(err)
	}
	if result.Value.(map[string]any)["value"] != int64(9007199254740993) {
		t.Fatalf("JSON conversion lost data: %#v", result.Value)
	}
	for _, url := range []string{foreign.URL, server.URL + "/redirect"} {
		if _, err := engine.Invoke(t.Context(), "main.star:handle", url); err == nil {
			t.Fatal("allowed credential origin violation")
		}
	}
	if foreignRequests.Load() != 0 {
		t.Fatal("authenticated redirect reached another origin")
	}
	public := hook(t, "def handle(url):\n    return http.get(url, headers={\"Authorization\": \"caller-secret\", \"X-Credential\": \"custom-secret\"}).status_code\n", automation.Options{AllowHTTP: true})
	if _, err := public.Invoke(t.Context(), "main.star:handle", server.URL+"/public-redirect"); err != nil {
		t.Fatal(err)
	}
	if foreignRequests.Load() != 1 {
		t.Fatal("unauthenticated redirect did not reach destination")
	}
}

func TestHTTPResponseLimitAndCancellation(t *testing.T) {
	started := make(chan struct{})
	var smallRequests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/small" {
			smallRequests.Add(1)
			_, _ = io.WriteString(w, "{}")
			return
		}
		if r.URL.Path == "/large" {
			_, _ = io.WriteString(w, strings.Repeat("x", 100))
			return
		}
		close(started)
		<-r.Context().Done()
	}))
	defer server.Close()
	engine := hook(t, "def handle(url):\n    return http.get(url).text\n", automation.Options{AllowHTTP: true, Limits: automation.Limits{OutputBytes: 10}})
	if _, err := engine.Invoke(t.Context(), "main.star:handle", server.URL+"/large"); err == nil || !strings.Contains(err.Error(), "byte limit") {
		t.Fatalf("unbounded HTTP output: %v", err)
	}
	quota := hook(t, "def handle(url):\n    http.get(url)\n    return http.get(url).text\n", automation.Options{AllowHTTP: true, Limits: automation.Limits{Calls: 1}})
	if _, err := quota.Invoke(t.Context(), "main.star:handle", server.URL+"/small"); err == nil || !strings.Contains(err.Error(), "call limit") {
		t.Fatalf("unbounded capability calls: %v", err)
	}
	if smallRequests.Load() != 1 {
		t.Fatalf("call budget allowed %d requests", smallRequests.Load())
	}
	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan error, 1)
	go func() { _, err := engine.Invoke(ctx, "main.star:handle", server.URL); done <- err }()
	<-started
	cancel()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("HTTP cancellation succeeded")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("HTTP cancellation blocked")
	}
}

// TestProcessHelper is a real child process used to exercise shell boundaries.
func TestProcessHelper(_ *testing.T) {
	mode := os.Getenv("AUTOMATION_TEST_PROCESS")
	if mode == "" {
		return
	}
	switch mode {
	case "echo":
		cwd, err := os.Getwd()
		if err != nil {
			os.Exit(1)
		}
		_ = json.NewEncoder(os.Stdout).Encode(map[string]any{"cwd": cwd, "env": os.Getenv("HOOK_VALUE"), "args": os.Args[len(os.Args)-1]})
		os.Exit(7)
	case "large":
		fmt.Print(strings.Repeat("x", 100))
		os.Exit(0)
	case "count":
		file, err := os.OpenFile(os.Getenv("AUTOMATION_TEST_FILE"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
		if err != nil {
			os.Exit(2)
		}
		_, _ = file.WriteString("executed\n")
		_ = file.Close()
		fmt.Print("counted")
		os.Exit(0)
	case "wait":
		executable, err := os.Executable()
		if err != nil {
			os.Exit(2)
		}
		child := exec.Command(executable, "-test.run=^TestProcessHelper$")
		child.Env = append(os.Environ(), "AUTOMATION_TEST_PROCESS=child")
		if err := child.Run(); err != nil {
			os.Exit(2)
		}
		os.Exit(0)
	case "child":
		conn, err := net.Dial("tcp", os.Getenv("AUTOMATION_TEST_ADDRESS"))
		if err != nil {
			os.Exit(2)
		}
		_, _ = conn.Write([]byte("ready"))
		_, _ = io.Copy(io.Discard, conn)
		os.Exit(0)
	}
}

func TestShellArgvEnvironmentLimitsAndCancellation(t *testing.T) {
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	options := automation.Options{AllowShell: true, ShellDir: dir, ShellEnv: []string{"AUTOMATION_TEST_PROCESS=echo"}}
	engine := hook(t, `def handle(argv):
    result = shell(argv, env={"HOOK_VALUE": "literal"})
    return {"code": result.exit_code, "data": result.stdout}
`, options)
	argv := []string{executable, "-test.run=^TestProcessHelper$", "--", "$(false); untrusted"}
	result, err := engine.Invoke(t.Context(), "main.star:handle", argv)
	if err != nil {
		t.Fatal(err)
	}
	data := result.Value.(map[string]any)
	if data["code"] != int64(7) {
		t.Fatalf("wrong exit status %#v", data)
	}
	var output map[string]any
	if err := json.Unmarshal([]byte(data["data"].(string)), &output); err != nil {
		t.Fatal(err)
	}
	if output["args"] != "$(false); untrusted" || output["env"] != "literal" || output["cwd"] != dir {
		t.Fatalf("shell did not preserve explicit argv/env/cwd: %#v", output)
	}
	options.ShellEnv = []string{"AUTOMATION_TEST_PROCESS=large"}
	options.Limits.OutputBytes = 10
	large := hook(t, "def handle(argv):\n    return shell(argv).stdout\n", options)
	if _, err := large.Invoke(t.Context(), "main.star:handle", argv); err == nil || !strings.Contains(err.Error(), "output byte limit") {
		t.Fatalf("unbounded shell output: %v", err)
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = listener.Close() }()
	options.ShellEnv = []string{"AUTOMATION_TEST_PROCESS=wait", "AUTOMATION_TEST_ADDRESS=" + listener.Addr().String()}
	options.Limits.OutputBytes = 0
	waiting := hook(t, "def handle(argv):\n    return shell(argv).exit_code\n", options)
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	done := make(chan error, 1)
	go func() { _, err := waiting.Invoke(ctx, "main.star:handle", argv); done <- err }()
	connection, err := listener.Accept()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = connection.Close() }()
	ready := make([]byte, 5)
	if _, err := io.ReadFull(connection, ready); err != nil {
		t.Fatal(err)
	}
	cancel()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("shell cancellation succeeded")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("shell cancellation blocked")
	}
	if err := connection.SetReadDeadline(time.Now().Add(5 * time.Second)); err != nil {
		t.Fatal(err)
	}
	if _, err := connection.Read(make([]byte, 1)); !errors.Is(err, io.EOF) {
		t.Fatalf("shell descendant survived cancellation: %v", err)
	}
}
