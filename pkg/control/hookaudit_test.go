package control_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/control"
	"github.com/elee1766/jobshop/pkg/control/controltest"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
	"github.com/elee1766/jobshop/pkg/store/sqlite"
)

func TestHookAuditPublicProjectionSurvivesCredentialRotation(t *testing.T) {
	for _, recoveredState := range []string{"done", "failed"} {
		t.Run(recoveredState, func(t *testing.T) { testHookAuditCredentialRotation(t, recoveredState) })
	}
}

func testHookAuditCredentialRotation(t *testing.T, recoveredState string) {
	t.Helper()
	const hostSecret = "host-credential-for-public-audit-test"
	const callerPassword = "caller-password-must-stay-private"
	const responsePassword = "response-password-must-stay-private"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"padding": strings.Repeat("a", 12<<10), "echo": hostSecret, "nested": map[string]any{"password": responsePassword}})
	}))
	defer server.Close()
	root := hookTestRoot(t, `def begin(run):
    response = http.get(run.inputs["url"])
    print("diagnostic " + run.inputs["host"])
    return [start_step(id="work",pool="dev",prompt="work " + run.inputs["host"],inputs={"normal":run.inputs["host"],"nested":{"api_key":run.inputs["password"]},"body":response.json()})]
def completed(run, step):
    return []
def handle(event, args):
    return []
`)
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, "credential"), []byte(hostSecret), 0o600))
	data, err := os.ReadFile(root.Config())
	require.NoError(t, err)
	grants := fmt.Sprintf("automation:\n  http: true\n  credentials:\n    api:\n      origins: [%q]\n      header: X-API-Key\n      secret: credential\n", server.URL)
	require.NoError(t, os.WriteFile(root.Config(), append(data, []byte(grants)...), 0o600))
	c := controltest.Start(t, root, control.Options{})
	var run rpc.RunResult
	c.Call(rpc.MethodRunStart, rpc.RunStartParams{Workflow: "dynamic", Inputs: map[string]any{"url": server.URL, "host": hostSecret, "password": callerPassword}}, &run)
	var records []store.Record
	c.PassUntil(func() bool {
		c.Call(rpc.MethodAutomationList, rpc.AutomationListParams{Run: run.Run.ID}, &records)
		return len(records) == 1 && records[0].State == "done"
	}, "hook did not finish")
	assert.Empty(t, records[0].Holder)
	assert.True(t, records[0].LeaseUntil.IsZero())
	assert.NotContains(t, string(records[0].Value), `"actions"`)
	attemptCount := 1
	check := func() {
		t.Helper()
		var audit rpc.AutomationResult
		c.Call(rpc.MethodAutomationShow, rpc.IDParams{ID: records[0].Key}, &audit)
		encoded, err := json.Marshal(audit)
		require.NoError(t, err)
		for _, private := range []string{hostSecret, callerPassword, responsePassword, `"plan"`, `"read_set"`, `"run_input"`, `"step_input"`, `"input_digest"`, `"input"`} {
			assert.NotContains(t, string(encoded), private)
		}
		require.Len(t, audit.Attempts, attemptCount)
		require.Len(t, audit.Calls, 1)
		assert.Less(t, len(encoded), 32<<10)
		var value any
		require.NoError(t, json.Unmarshal(encoded, &value))
		artifacts := map[string]bool{}
		var walk func(any)
		walk = func(value any) {
			switch v := value.(type) {
			case map[string]any:
				if path, ok := v["artifact"].(string); ok {
					require.True(t, filepath.IsLocal(path))
					artifacts[path] = true
				}
				for _, item := range v {
					walk(item)
				}
			case []any:
				for _, item := range v {
					walk(item)
				}
			}
		}
		walk(value)
		require.NotEmpty(t, artifacts)
		for path := range artifacts {
			artifact, err := os.ReadFile(filepath.Join(root.Dir, path))
			require.NoError(t, err)
			for _, private := range []string{hostSecret, callerPassword, responsePassword} {
				assert.NotContains(t, string(artifact), private)
			}
			assert.Contains(t, string(artifact), "[REDACTED]")
			info, err := os.Stat(filepath.Join(root.Dir, path))
			require.NoError(t, err)
			assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())
		}
	}
	check()
	c.Stop()
	// Restore the actual prepared journal to model a crash before its terminal acknowledgment.
	db, err := sqlite.Open(t.Context(), root.DB(), sqlite.Options{})
	require.NoError(t, err)
	invocations, err := db.Collection(t.Context(), store.CollectionSpec{Name: "hook_invocations", Index: []store.Field{{Name: "run", Path: "run"}, {Name: "trigger", Path: "trigger"}, {Name: "event", Path: "event.delivery"}}})
	require.NoError(t, err)
	prepared, err := invocations.Get(t.Context(), records[0].Key)
	require.NoError(t, err)
	prepared.State, prepared.Holder, prepared.LeaseUntil = "planned", "", time.Time{}
	_, err = invocations.Update(t.Context(), prepared)
	require.NoError(t, err)
	if recoveredState == "failed" {
		// Without an applied marker, the rotated manifest must reject this prepared graph.
		ticket, err := db.Get(t.Context(), run.Run.ID)
		require.NoError(t, err)
		fields, err := store.DecodeFields[store.RunFields](ticket)
		require.NoError(t, err)
		fields.Automation.Applied = nil
		encoded, err := json.Marshal(fields)
		require.NoError(t, err)
		_, err = db.Update(t.Context(), "test", ticket.ID, store.Update{Fields: encoded, IfRevision: ticket.Revision})
		require.NoError(t, err)
	}
	require.NoError(t, db.Close())
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, "credential"), []byte("rotated-credential"), 0o600))
	c = controltest.Start(t, root, control.Options{})
	c.PassUntil(func() bool {
		c.Call(rpc.MethodAutomationList, rpc.AutomationListParams{Run: run.Run.ID}, &records)
		if len(records) != 1 || records[0].State != recoveredState {
			return false
		}
		var audit rpc.AutomationResult
		c.Call(rpc.MethodAutomationShow, rpc.IDParams{ID: records[0].Key}, &audit)
		return len(audit.Attempts) == 2 && audit.Attempts[1].State == recoveredState
	}, "prepared invocation did not recover")
	attemptCount = 2
	check()
	assert.Len(t, hookRun(t, c, run.Run.ID).Steps, 1)
}

func TestHookColdFallbackVerifiesAcceptedManifest(t *testing.T) {
	const script = `def begin(run):
    return []
def completed(run, step):
    return []
def handle(event, args):
    return []
`
	root := hookTestRoot(t, script)
	c := controltest.Start(t, root, control.Options{})
	c.Stop()
	packPath := filepath.Join(root.Dir, "pack.yaml")
	pack, err := os.ReadFile(packPath)
	require.NoError(t, err)
	broken := strings.Replace(string(pack), root.Dir, "{{.future_binary_var}}", 1)
	require.NoError(t, os.WriteFile(packPath, []byte(broken), 0o600))
	c = controltest.Start(t, root, control.Options{})
	var attention rpc.AttentionResult
	c.Call(rpc.MethodAttention, struct{}{}, &attention)
	require.NotEmpty(t, attention.Items)
	c.Stop()
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, "hooks.star"), []byte(script+"\n# source drift\n"), 0o600))
	ctx, cancel := context.WithTimeout(t.Context(), controltest.DialWait)
	defer cancel()
	err = control.Run(ctx, root, control.Options{})
	require.ErrorContains(t, err, "accepted hook manifest changed")
}

func TestHookPublicInventoryIsBoundedBeyondOnePage(t *testing.T) {
	root := hookTestRoot(t, `def begin(run):
    return []
def completed(run, step):
    return []
def handle(event, args):
    return []
`)
	c := controltest.Start(t, root, control.Options{})
	var last rpc.RunResult
	for i := range 105 {
		c.Call(rpc.MethodRunStart, rpc.RunStartParams{Workflow: "dynamic", Key: fmt.Sprint(i)}, &last)
	}
	c.PassUntil(func() bool {
		fields, err := store.DecodeFields[store.RunFields](hookRun(t, c, last.Run.ID).Run)
		require.NoError(t, err)
		return fields.Automation.Version == 1
	}, "workflow callbacks did not finish")
	var records []store.Record
	c.Call(rpc.MethodAutomationList, rpc.AutomationListParams{}, &records)
	assert.Len(t, records, 50)
	c.Call(rpc.MethodAutomationList, rpc.AutomationListParams{Limit: 100000}, &records)
	assert.Len(t, records, 100)
}
