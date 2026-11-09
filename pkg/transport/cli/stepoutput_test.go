package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/control/controltest"
	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/transport/rpc"
)

func TestCloseStructuredOutput(t *testing.T) {
	root := controltest.NewRoot(t, `projects: {js: {}}
agents: {worker: {command: echo}}
pools: {dev: {agent: worker, max_sessions: 0, dir: $ROOT}}
workflows:
  example:
    steps:
      - {id: inspect, pool: dev, prompt: Inspect, title: Inspect}
`)
	controller := controltest.StartWith(t, root)
	var run rpc.RunResult
	controller.Call(rpc.MethodRunStart, rpc.RunStartParams{Workflow: "example"}, &run)
	require.Len(t, run.Steps, 1)
	id := run.Steps[0].ID
	path := filepath.Join(root.Dir, "result.json")
	show := func() domain.Ticket {
		var detail rpc.ShowResult
		controller.Call(rpc.MethodTicketShow, rpc.IDParams{ID: id}, &detail)
		return detail.Ticket
	}
	for _, body := range []string{"null", "[]", "{} {}", `{"bad":`, strings.Repeat(" ", maxStepOutputBytes+1)} {
		require.NoError(t, os.WriteFile(path, []byte(body), 0o600))
		err := (closeCmd{ID: id, Resolution: "done", OutputFile: path}).Run(&Globals{Socket: root.AdminSocket()})
		require.Error(t, err)
		require.Equal(t, domain.StatusOpen, show().Status, "invalid output must not close the step")
	}
	body := `{"count":9007199254740993,"routing":{"selected":"review"},"ready":true}`
	require.NoError(t, os.WriteFile(path, []byte(body), 0o600))
	previous := stdout
	t.Cleanup(func() { stdout = previous })
	var response bytes.Buffer
	stdout = &response
	require.Equal(t, 0, Main([]string{"--socket", root.AdminSocket(), "close", id, "--outcome", "approve", "--output-file", path}, nil))
	closed := show()
	require.Equal(t, domain.StatusClosed, closed.Status)
	require.Contains(t, string(closed.Fields), `9007199254740993`, "structured integer must survive CLI, RPC and store without float coercion")
	fields, err := domain.DecodeFields[domain.StepFields](closed)
	require.NoError(t, err)
	require.Equal(t, "approve", fields.Outcome)
	require.Equal(t, true, fields.StructuredOutput["ready"])
	response.Reset()
	require.Equal(t, 0, Main([]string{"--socket", root.AdminSocket(), "ticket", "show", id}, nil))
	require.Contains(t, response.String(), "9007199254740993", "operator inspection must preserve structured integers too")
}
