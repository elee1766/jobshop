package archcheck

// known lists the violations that exist today, each with the child that
// removes it. It is the measured baseline and only shrinks: an entry whose
// edge is gone fails the check as stale, and a new violation is not covered by
// an old entry because each entry names one source and one target. A package
// pattern is for a subtree that moves together, never for "everything".
func known() []Known {
	return []Known{
		{
			Rule: "control-is-top", From: "./pkg/cli", To: "./pkg/control", Issue: "JOB-174",
			Why: "cli.Main hosts the controller (`jobshop controller` calls control.Run with control.Options). JOB-174 narrows this to one entry point so the CLI's client commands cannot reach controller internals.",
		},
		{
			Rule: "control-is-top", From: "./pkg/desk", To: "./pkg/control", Issue: "JOB-174",
			Why: "Two desk tests (push_test, systemautomation_test) build control.Options directly. They move to controltest's option helpers when Desk moves under transport.",
		},
		{
			Rule: "command-paths-no-session", From: "./pkg/control/stepexec", To: "./pkg/control/sessionrt", Issue: "JOB-167",
			Why: "stepexec takes sessionrt.GitCommonDir to place a step's run dir. JOB-167 moves that path helper to the process and isolation mechanics a command may use, then the session lifecycle is out of the command path.",
		},
		{
			Rule: "command-paths-no-acp", From: "./pkg/control/stepexec", To: "github.com/coder/acp-go-sdk", Issue: "JOB-167",
			Why: "Reached through sessionrt (see the command-paths-no-session entry). It goes with that edge.",
		},
		{
			Rule: "cue-only-in-packcue", From: "./pkg/hooks", To: "cuelang.org/go/...", Issue: "JOB-172",
			Why: "hooks/schema.go compiles a CUE schema for hook input and imports CUE directly. JOB-172 gives pack/cue the one evaluation API and hooks calls it.",
		},
	}
}
