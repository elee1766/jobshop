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
			Rule: "cue-only-in-packcue", From: "./pkg/hooks", To: "cuelang.org/go/...", Issue: "JOB-172",
			Why: "hooks/schema.go compiles a CUE schema for hook input and imports CUE directly. JOB-172 gives pack/cue the one evaluation API and hooks calls it.",
		},
	}
}
