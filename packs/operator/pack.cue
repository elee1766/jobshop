package operator

// #Pack is the host's operator chat pool: sessions that talk to the
// operator and manage work across every pack instance (#248). Include it
// once, without a prefix.
#Pack: {
	let V = vars
	path:   *"packs/operator" | string
	prefix: *"" | string
	vars: {
		jobshop_repo:      string
		checkout:          string
		agent:             string
		operator_sessions: *0 | int & >=0
		// Local memory (#592), relative to the root dir.
		memory_dir: *"memory" | string
		memory_cap: *2200 | int & >=0
		...
	}
	pools: operator: {
		type:         "chat"
		attach:       "rw"
		agent:        V.agent
		repo:         V.checkout
		ref:          "origin/master"
		max_sessions: V.operator_sessions
		idle_timeout: "12h"
		turn_timeout: "1h"
		prompt_file:  "prompts/operator.md"
		// Cross-instance reads and actions; Desk's attention chats run here (#603).
		admin: ["config-reload", "config-show", "session-list", "pool-list", "pool-scale", "session-cordon", "session-config", "trigger-list", "attention", "attention-answer", "desk-command", "activity", "ticket-release", "ticket-undefer", "decision-resolve", "question-answer"]
	}
	prompt_rules: [{
		name: "operator-desk-links"
		at:   "turn"
		match: pools: ["operator"]
		text: """
			In Desk replies and escalation summaries, use Markdown links. Link Jobshop
			tickets and runs as [label](jobshop://ticket/<id>), and sessions as
			[label](jobshop://session/<id>), using real IDs in the current instance.
			Link Linear and GitHub references to their verified canonical HTTPS URLs.
			Do not invent destinations, wrap links in code, or assume plain references
			become links. Formatting changes must preserve blockers and decisions.

			"""
	}, {
		name: "operator-improve"
		file: "prompts/improve.md"
		at:   "after"
		match: {pools: ["operator"], keys: ["improve:*"]}
	}, {
		name: "operator-memory"
		at:   "system"
		match: pools: ["operator"]
		memory: {dir: V.memory_dir, cap: V.memory_cap}
	}]
	// Operator sessions never push to master (#249): they run
	// `jobshop guard push --target <branch>` first.
	hooks: [{
		on: "action"
		actions: ["push"]
		pools: ["operator"]
		handler: "guard.star:push"
		args: protected: ["master", "main"]
		timeout: "5s"
	}]
}
