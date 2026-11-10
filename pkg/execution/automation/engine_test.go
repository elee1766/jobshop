package automation_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gfx-labs/jobshop/pkg/execution/automation"
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
        return result(actions=[start_run(workflow=args["workflow"], key=event.repository, inputs=event.payload, reason="enabled event policy")], output={"custom": event.payload, "metadata": "user data"})
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
		if result.Metadata.Entry != "main.star:handle" || result.Metadata.Digest != engine.Digest() || result.Metadata.Calls != 0 {
			t.Fatalf("incorrect host result metadata: %#v", result.Metadata)
		}
		if result.Output["custom"].(map[string]any)["count"] != int64(9007199254740993) {
			t.Fatalf("custom payload lost integer precision: %#v", result.Output)
		}
		if enabled {
			if len(result.Actions) != 1 || result.Actions[0].Reason != "enabled event policy" || result.Actions[0].Workflow != "review" || result.Actions[0].Inputs["count"] != int64(9007199254740993) || result.Actions[0].Inputs["unsigned"] != uint64(18446744073709551615) {
				t.Fatalf("wrong action: %#v", result)
			}
		}
	}
}

func TestCachedActionsAreIsolated(t *testing.T) {
	engine := hook(t, `action = result(actions=[start_run(workflow="review", key="cached", inputs={"nested": {"original": True}})], output={"nested": {"original": True}})
def handle():
    return action
`, automation.Options{})
	first, err := engine.Invoke(t.Context(), "main.star:handle")
	if err != nil {
		t.Fatal(err)
	}
	first.Actions[0].Inputs["nested"].(map[string]any)["original"] = false
	first.Output["nested"].(map[string]any)["original"] = false
	second, err := engine.Invoke(t.Context(), "main.star:handle")
	if err != nil {
		t.Fatal(err)
	}
	if second.Output["nested"].(map[string]any)["original"] != true {
		t.Fatal("caller mutation corrupted frozen module output")
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

func TestCommandStepAction(t *testing.T) {
	engine := hook(t, `def handle():
    return [
        start_step(id="gate",type="command",pool="gate",run=["make","","check"],timeout="10m",retry_after="1m",on_restart="fail",worktree=True,retries=3,needs_resources=["emu"]),
        start_step(id="shell",type="command",pool="gate",run="make check"),
        start_step(id="agent",type="agent",pool="dev",prompt="implement"),
    ]
`, automation.Options{})
	first, err := engine.Invoke(t.Context(), "main.star:handle")
	if err != nil {
		t.Fatal(err)
	}
	gate := first.Actions[0]
	argv, ok := gate.Run.([]string)
	if !ok || !slices.Equal(argv, []string{"make", "", "check"}) || gate.Type != "command" || gate.Timeout != "10m" || gate.RetryAfter != "1m" || gate.OnRestart != "fail" || !gate.Worktree || gate.Retries != 3 || !gate.RetriesSet || !slices.Equal(gate.NeedsResources, []string{"emu"}) {
		t.Fatalf("lost command step fields: %#v", gate)
	}
	if first.Actions[1].Run != "make check" || first.Actions[2].Type != "agent" || first.Actions[2].Run != nil {
		t.Fatalf("wrong shell or agent step: %#v", first.Actions[1:])
	}
	argv[0] = "mutated"
	second, err := engine.Invoke(t.Context(), "main.star:handle")
	if err != nil {
		t.Fatal(err)
	}
	if got := second.Actions[0].Run.([]string); got[0] != "make" {
		t.Fatalf("caller mutation corrupted frozen argv: %v", got)
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
	for _, failure := range []string{"", "started", "completed", "cancelled"} {
		t.Run(failure, func(t *testing.T) {
			var effects atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				effects.Add(1)
				_, _ = io.WriteString(w, r.Header.Get("Authorization"))
			}))
			defer server.Close()
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			var observations []automation.Observation
			engine := hook(t, "def handle(url):\n    return {\"status\": http.post(url, auth=\"service\", body=\"mutation\").status_code}\n", automation.Options{
				AllowHTTP: true, Auth: map[string]automation.Credential{"service": {Origins: []string{server.URL}, Headers: map[string]string{"Authorization": "Bearer private-token"}}},
				Observe: func(_ context.Context, o automation.Observation) error {
					if o.Phase == "started" && effects.Load() != 0 {
						t.Error("effect preceded started observation")
					}
					observations = append(observations, o)
					if failure == "cancelled" && o.Phase == "started" {
						cancel()
					}
					if o.Phase == failure {
						return errors.New("storage failed")
					}
					return nil
				},
			})
			_, err := engine.Invoke(ctx, "main.star:handle", server.URL)
			if failure == "" && err != nil {
				t.Fatal(err)
			}
			if failure != "" && err == nil {
				t.Fatal("observer failure was ignored")
			}
			expected := int32(1)
			if failure == "started" || failure == "cancelled" {
				expected = 0
			}
			if effects.Load() != expected {
				t.Fatalf("unexpected effect count %d", effects.Load())
			}
			if failure == "cancelled" && (len(observations) != 2 || observations[1].Phase != "failed") {
				t.Fatalf("cancellation before execution was not a known failure: %#v", observations)
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
	if err != nil || previous.Output["ok"] != true {
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
		{"scalar", "def handle():\n    return 1\n", "output must be an object", automation.Limits{}},
		{"result_output", "def handle():\n    return result(output=[])\n", "dict", automation.Limits{}},
		{"result_metadata", "def handle():\n    return result(metadata={\"entry\":\"forged\"})\n", "unexpected keyword", automation.Limits{}},
		{"result_actions", "def handle():\n    return result(actions=[{}])\n", "only action constructors", automation.Limits{}},
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
			result, err := engine.Invoke(t.Context(), "main.star:handle")
			if result.Metadata.Entry != "main.star:handle" || result.Metadata.Digest != engine.Digest() {
				t.Fatalf("error lost invocation metadata: %#v", result)
			}
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

// TestProcessHelper counts real external effects for cross-capability replay.
func TestProcessHelper(_ *testing.T) {
	if os.Getenv("AUTOMATION_TEST_PROCESS") != "count" {
		return
	}
	file, err := os.OpenFile(os.Getenv("AUTOMATION_TEST_FILE"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		os.Exit(2)
	}
	_, _ = file.WriteString("executed\n")
	_ = file.Close()
	fmt.Print("counted")
	os.Exit(0)
}

func TestStartStepAgentFields(t *testing.T) {
	engine := hook(t, `def handle():
    return [
        start_step(id="build",pool="dev",prompt="b"),
        start_step(id="review",pool="dev",prompt="r",needs=["build"],needs_resources=["pr:acme#1","emu"],
                   max_review_rounds=0,on_request_changes="build"),
        start_step(id="placed",prompt="p",placement={
            "required": ["pr:acme#1"],
            "prefer_for": "10m",
            "options": [{"pool":"flagship","resources":["gpu"],"weight":100}, {"pool":"dev"}],
        }),
    ]
`, automation.Options{})
	result, err := engine.Invoke(t.Context(), "main.star:handle")
	if err != nil {
		t.Fatal(err)
	}
	review := result.Actions[1]
	if strings.Join(review.NeedsResources, ",") != "pr:acme#1,emu" || review.MaxReviewRounds == nil || *review.MaxReviewRounds != 0 || review.OnRequestChanges != "build" {
		t.Fatalf("lost review step fields: %#v", review)
	}
	if result.Actions[0].MaxReviewRounds != nil {
		t.Fatal("unset max_review_rounds must stay nil, not 0")
	}
	p := result.Actions[2].Placement
	if p == nil || result.Actions[2].Pool != "" || p.PreferFor != "10m" || len(p.Options) != 2 || p.Options[0].Pool != "flagship" || p.Options[0].Weight != 100 || p.Options[0].Resources[0] != "gpu" || p.Required[0] != "pr:acme#1" {
		t.Fatalf("lost placement: %#v", p)
	}
	// A caller mutating the result must not reach the frozen module value.
	p.Options[0].Pool = "changed"
	again, err := engine.Invoke(t.Context(), "main.star:handle")
	if err != nil {
		t.Fatal(err)
	}
	if again.Actions[2].Placement.Options[0].Pool != "flagship" {
		t.Fatal("placement is shared between invocations")
	}
}

func TestStartStepScriptRetryAfter(t *testing.T) {
	root := t.TempDir()
	writeHook(t, root, "main.star", `def handle():
    return start_step(id="wait",pool="dev",script=poll,retry_after="30s")

def poll(run, step):
    return result(not_yet=True)

def bad(run, step):
    return result(not_yet=True, outcome="x")
`)
	engine, err := automation.Load(t.Context(), root, []string{"main.star:handle"}, automation.Options{})
	if err != nil {
		t.Fatal(err)
	}
	result, err := engine.Invoke(t.Context(), "main.star:handle")
	if err != nil {
		t.Fatal(err)
	}
	if result.Actions[0].RetryAfter != "30s" || result.Actions[0].Script != "main.star:poll" {
		t.Fatalf("lost retry_after: %#v", result.Actions[0])
	}
	step, err := engine.Step(t.Context(), "main.star:poll", time.Minute, nil, struct{}{}, struct{}{})
	if err != nil || !step.NotYet {
		t.Fatalf("not_yet lost: %#v %v", step, err)
	}
	if _, err := engine.Step(t.Context(), "main.star:bad", time.Minute, nil, struct{}{}, struct{}{}); err == nil || !strings.Contains(err.Error(), "not_yet excludes outcome") {
		t.Fatalf("not_yet with an outcome accepted: %v", err)
	}
}

func TestStartStepFieldValidation(t *testing.T) {
	for _, tc := range []struct{ name, call, want string }{
		{"resources not list", `start_step(id="a",pool="p",prompt="x",needs_resources="emu")`, "needs_resources"},
		{"resources blank", `start_step(id="a",pool="p",prompt="x",needs_resources=[""])`, "nonempty strings"},
		{"retry_after agent", `start_step(id="a",pool="p",prompt="x",retry_after="1m")`, "retry_after needs script"},
		{"command no run", `start_step(id="a",type="command",pool="p")`, "needs run"},
		{"command blank run", `start_step(id="a",type="command",pool="p",run=" ")`, "run must be nonempty"},
		{"command empty argv", `start_step(id="a",type="command",pool="p",run=[])`, "run must be nonempty"},
		{"command blank executable", `start_step(id="a",type="command",pool="p",run=["", "x"])`, "executable must be nonblank"},
		{"command non-string argv", `start_step(id="a",type="command",pool="p",run=["make", 1])`, "list of strings"},
		{"command prompt", `start_step(id="a",type="command",pool="p",run="true",prompt="x")`, "command excludes"},
		{"command output_schema", `start_step(id="a",type="command",pool="p",run="true",output_schema={})`, "command excludes"},
		{"command on_restart", `start_step(id="a",type="command",pool="p",run="true",on_restart="ignore")`, "on_restart"},
		{"command retry_after", `start_step(id="a",type="command",pool="p",run="true",retry_after="0s")`, "retry_after must be a positive duration"},
		{"command timeout", `start_step(id="a",type="command",pool="p",run="true",timeout="-1m")`, "timeout must be a positive duration"},
		{"command retry_pool", `start_step(id="a",type="command",pool="p",run="true",retry_pool="q")`, "excludes retry_pool"},
		{"command nul shell", `start_step(id="a",type="command",pool="p",run="true\x00")`, "NUL"},
		{"command nul argv", `start_step(id="a",type="command",pool="p",run=["make", "a\x00b"])`, "NUL"},
		{"command placement", `start_step(id="a",type="command",run="true",placement={"options":[{"pool":"p"}]})`, "placement is for agent steps"},
		{"command rounds", `start_step(id="a",type="command",pool="p",run="true",max_review_rounds=1)`, "max_review_rounds is for agent steps"},
		{"command request changes", `start_step(id="a",type="command",pool="p",run="true",needs=["b"],on_request_changes="b")`, "on_request_changes is for agent steps"},
		{"agent run", `start_step(id="a",pool="p",prompt="x",run="true")`, "need type"},
		{"agent on_restart", `start_step(id="a",type="agent",pool="p",prompt="x",on_restart="fail")`, "need type"},
		{"unknown type", `start_step(id="a",type="http",pool="p",prompt="x")`, "must be agent or command"},
		{"rounds negative", `start_step(id="a",pool="p",prompt="x",max_review_rounds=-1)`, "max_review_rounds"},
		{"rounds string", `start_step(id="a",pool="p",prompt="x",max_review_rounds="2")`, "max_review_rounds"},
		{"request changes not a need", `start_step(id="a",pool="p",prompt="x",needs=["b"],on_request_changes="c")`, "on_request_changes"},
		{"request changes no needs", `start_step(id="a",pool="p",prompt="x",on_request_changes="c")`, "on_request_changes"},
		{"pool missing", `start_step(id="a",prompt="x")`, "pool must be nonempty"},
		{"placement and pool", `start_step(id="a",pool="p",prompt="x",placement={"options":[{"pool":"p"}]})`, "set one or the other"},
		{"placement and resources", `start_step(id="a",prompt="x",needs_resources=["emu"],placement={"options":[{"pool":"p"}]})`, "set one or the other"},
		{"placement and retry_pool", `start_step(id="a",prompt="x",retry_pool="p",placement={"options":[{"pool":"p"}]})`, "retry_pool"},
		{"placement no options", `start_step(id="a",prompt="x",placement={"options":[]})`, "at least one option"},
		{"placement option no pool", `start_step(id="a",prompt="x",placement={"options":[{"weight":1}]})`, "pool is required"},
		{"placement bad prefer_for", `start_step(id="a",prompt="x",placement={"prefer_for":"soon","options":[{"pool":"p"}]})`, "prefer_for"},
		{"placement negative prefer_for", `start_step(id="a",prompt="x",placement={"prefer_for":"-1m","options":[{"pool":"p"}]})`, "prefer_for"},
		{"placement unknown key", `start_step(id="a",prompt="x",placement={"options":[{"pool":"p"}],"when":"x"})`, "unknown key"},
		{"placement option unknown key", `start_step(id="a",prompt="x",placement={"options":[{"pool":"p","when":"x"}]})`, "unknown key"},
		{"placement weight", `start_step(id="a",prompt="x",placement={"options":[{"pool":"p","weight":"1"}]})`, "weight"},
		{"placement resources", `start_step(id="a",prompt="x",placement={"required":"emu","options":[{"pool":"p"}]})`, "placement required"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			engine := hook(t, "def handle():\n    return "+tc.call+"\n", automation.Options{})
			_, err := engine.Invoke(t.Context(), "main.star:handle")
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("want error containing %q, got %v", tc.want, err)
			}
		})
	}
	for _, tc := range []struct{ name, call, want string }{
		{"retry_after", `start_step(id="a",pool="p",script=run_it,retry_after="0s")`, "retry_after must be a positive duration"},
		{"retry_after junk", `start_step(id="a",pool="p",script=run_it,retry_after="soon")`, "retry_after must be a positive duration"},
		{"script placement", `start_step(id="a",script=run_it,placement={"options":[{"pool":"p"}]})`, "placement is for agent steps"},
		{"script rounds", `start_step(id="a",pool="p",script=run_it,max_review_rounds=1)`, "max_review_rounds is for agent steps"},
		{"script request changes", `start_step(id="a",pool="p",script=run_it,needs=["b"],on_request_changes="b")`, "on_request_changes is for agent steps"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			engine := hook(t, "def run_it(run, step):\n    return result()\n\ndef handle():\n    return "+tc.call+"\n", automation.Options{})
			_, err := engine.Invoke(t.Context(), "main.star:handle")
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("want error containing %q, got %v", tc.want, err)
			}
		})
	}
}
