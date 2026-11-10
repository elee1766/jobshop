package stepexec

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gfx-labs/acphost/pkg/proc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gfx-labs/jobshop/pkg/config"
	"github.com/gfx-labs/jobshop/pkg/domain"
)

// fakeHost records the command it was asked to run.
type fakeHost struct {
	cmd    Cmd
	exit   proc.Exit
	err    error
	stdout string
	stderr string
}

func (h *fakeHost) Run(_ context.Context, c Cmd) (proc.Exit, error) {
	h.cmd = c
	_, _ = c.Stdout.Write([]byte(h.stdout))
	_, _ = c.Stderr.Write([]byte(h.stderr))
	return h.exit, h.err
}

func (h *fakeHost) Worktree(string, string, string, string, string) error { return nil }

func (h *fakeHost) SlotEnv(context.Context, domain.Ticket) []string { return []string{"SLOT=1"} }

func (h *fakeHost) HostEnv() []string { return []string{"JOBSHOP_BIN=/bin/jobshop"} }

func newExec(h Host) Executor {
	return Executor{Root: "/root", Host: h, HTTP: http.DefaultClient, Now: time.Now}
}

func TestCommandResults(t *testing.T) {
	cmdStep := config.Step{ID: "s", Type: config.StepCommand, Run: []string{"echo", "{{.name}}"}}
	tpl := Templates{Vars: map[string]any{"name": "bob"}}
	for _, c := range []struct {
		name string
		host fakeHost
		want string
		err  string
	}{
		{"ok", fakeHost{stdout: "hi\noutcome: green\n"}, ResultOK, ""},
		{"not yet", fakeHost{exit: proc.Exit{ExitCode: ExitNotYet}}, ResultNotYet, ""},
		{"failed", fakeHost{exit: proc.Exit{ExitCode: 2}, stderr: "a\nb\n"}, ResultFailed, "exit 2: a | b"},
		{"signal", fakeHost{exit: proc.Exit{Signal: "KILL"}}, ResultFailed, "exit signal KILL"},
	} {
		t.Run(c.name, func(t *testing.T) {
			h := c.host
			res := newExec(&h).command(t.Context(), &config.Config{}, domain.Ticket{ID: "t1"}, cmdStep, domain.StepFields{}, tpl, "run/s/1")
			assert.Equal(t, c.want, res.Result)
			assert.Equal(t, c.err, res.Err)
			assert.Equal(t, "echo bob", res.Command)
			assert.Equal(t, []string{"echo", "bob"}, h.cmd.Argv)
			assert.Equal(t, "/root", h.cmd.Dir)
			assert.Contains(t, h.cmd.Env, "JOBSHOP_VAR_NAME=bob")
			assert.Contains(t, h.cmd.Env, "JOBSHOP_STEP_KEY=run/s/1")
			assert.Contains(t, h.cmd.Env, "JOBSHOP_TICKET=t1")
			assert.Contains(t, h.cmd.Env, "SLOT=1")
			assert.Contains(t, h.cmd.Env, "JOBSHOP_BIN=/bin/jobshop")
			if c.name == "ok" {
				assert.Equal(t, "green", res.Outcome)
			}
		})
	}
}

func TestCommandBadTemplate(t *testing.T) {
	h := &fakeHost{}
	s := config.Step{ID: "s", Type: config.StepCommand, Run: "echo {{.missing"}
	res := newExec(h).command(t.Context(), &config.Config{}, domain.Ticket{}, s, domain.StepFields{}, Templates{Vars: map[string]any{}}, "k")
	assert.Equal(t, ResultFailed, res.Result)
	assert.NotEmpty(t, res.Err)
	assert.Empty(t, h.cmd.Argv, "the command must not run")
}

func TestHTTPRetryAfter(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "run/s/1", r.Header.Get("Idempotency-Key"))
		w.Header().Set("Retry-After", "30")
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()
	s := config.Step{ID: "s", Type: config.StepHTTP, Method: "POST", URL: srv.URL}
	res := newExec(&fakeHost{}).httpStep(t.Context(), &config.Config{}, s, Templates{Vars: map[string]any{}}, "run/s/1")
	assert.Equal(t, ResultNotYet, res.Result)
	assert.Equal(t, 30*time.Second, res.RetryAfter)
	assert.Equal(t, "POST "+srv.URL, res.Command)
}

func TestHTTPFailureStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer srv.Close()
	s := config.Step{ID: "s", Type: config.StepHTTP, Method: "GET", URL: srv.URL}
	res := newExec(&fakeHost{}).httpStep(t.Context(), &config.Config{}, s, Templates{Vars: map[string]any{}}, "k")
	assert.Equal(t, ResultFailed, res.Result)
	assert.Equal(t, "400 Bad Request", res.Err)
}

func TestOutcomeOf(t *testing.T) {
	assert.Equal(t, "queued", OutcomeOf(`{"outcome":"queued"}`))
	assert.Empty(t, OutcomeOf(`{"x":1}`))
	assert.Equal(t, "green", OutcomeOf("log\noutcome: green\n"))
	assert.Empty(t, OutcomeOf("log\n"))
}

func TestLimitedBuffer(t *testing.T) {
	var b LimitedBuffer
	_, err := b.Write([]byte(strings.Repeat("x", OutputMax+10)))
	require.NoError(t, err)
	assert.Len(t, b.String(), OutputMax+len("\n[truncated]"))
}

func TestPlaceNoPool(t *testing.T) {
	pl, err := newExec(&fakeHost{}).Place(&config.Config{}, config.Step{}, domain.StepFields{Run: "r"}, Templates{})
	require.NoError(t, err)
	assert.Equal(t, Place{Dir: "/root"}, pl)
	assert.Equal(t, "/root/runs/r", RunDir("/root", "r"))
}
