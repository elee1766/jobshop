package sessionrt

import (
	"slices"
	"strings"

	host "github.com/elee1766/acphost/pkg/acphost"
	"github.com/elee1766/acphost/pkg/trace"
)

// DefaultShellProbe is what an agent's tool shell needs to work a ticket
// (JOB-67): the jobshop CLI, gh and git, a shell, and the core utils agent
// commands use (`$(dirname $(which go))` emptied PATH in #780).
var DefaultShellProbe = []string{"jobshop", "gh", "git", "sh", "env", "dirname", "which", "awk"}

// ProbeShell checks the env an agent's tool shell inherits (the
// controller's, overridden by the spec env) before a host is spawned:
// JOBSHOP_SESSION set and every command in want on its PATH. A miss is a
// permanent spawn error naming what is missing, so the pool is
// quarantined instead of sessions starting into "command not found" and
// burning step attempts (#780). Sandboxed agents resolve commands inside
// the sandbox, so they are not probed here.
func ProbeShell(spec host.Spec, inherited, want []string) error {
	if len(want) == 0 || spec.Sandbox != nil {
		return nil
	}
	env := map[string]string{}
	for _, kv := range slices.Concat(inherited, spec.Env) {
		k, v, _ := strings.Cut(kv, "=")
		env[k] = v
	}
	var missing []string
	if env["JOBSHOP_SESSION"] == "" {
		missing = append(missing, "$JOBSHOP_SESSION")
	}
	path := []string{"PATH=" + env["PATH"]}
	for _, c := range want {
		if trace.LookCommand(c, path) != nil {
			missing = append(missing, c)
		}
	}
	if len(missing) == 0 {
		return nil
	}
	return host.Permanentf("agent shell probe: missing %s on the agent PATH", strings.Join(missing, ", "))
}
