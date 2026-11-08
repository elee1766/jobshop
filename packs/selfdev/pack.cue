package selfdev

import (
	github "github.com/elee1766/jobshop/packs/github"
	"list"
)

#Settings: {
	repo:                 string
	agent:                string
	dev_sessions:         *2 | int & >=0
	refinery_sessions:    *1 | int & >=0
	refinery_model:       *"" | string
	refinery_effort:      *"" | string
	dev_model:            *"" | string
	dev_effort:           *"" | string
	dev_sonnet_sessions:  *1 | int & >=0
	dev_sonnet_model:     *"claude-sonnet-5-5" | string
	dev_sonnet_effort:    *"high" | string
	triage_model:         *"" | string
	triage_effort:        *"" | string
	memory_dir:           *"memory" | string
	memory_cap:           *2200 | int & >=0
	curate_min_learnings: *10 | int & >=0
	curate_every:         *"24h" | string
	curate_scope: *["repo:elee1766/jobshop", "pool:dev", "pool:dev-sonnet", "pool:plan", "pool:refinery", "pool:triage", "pool:deploy-check", "pool:learn"] | [...string]
	evaluate_runs:        *20 | int & >=0
	evaluate_age:         *"72h" | string
	learn_sessions:       *1 | int & >=0
	escalation_chat_pool: *"" | string
	linear_workspace:     *"" | string
	linear_team:          *"" | string
	linear_intake_labels: *"" | string
	linear_human:         *"" | string
}

#Pack: {
	let V = #Config
	let P = prefix
	#Config: #Settings
	path:    *"packs/selfdev" | string
	prefix:  *"" | string

	_github: github.#Pack & {
		prefix: P
		vars: {
			asset_prefix:         "../github/"
			repo:                 "elee1766/jobshop"
			checkout:             V.repo
			agent:                V.agent
			dev_sessions:         V.dev_sessions
			refinery_sessions:    V.refinery_sessions
			refinery_model:       V.refinery_model
			refinery_effort:      V.refinery_effort
			dev_model:            V.dev_model
			dev_effort:           V.dev_effort
			dev_sonnet_sessions:  V.dev_sonnet_sessions
			dev_sonnet_model:     V.dev_sonnet_model
			dev_sonnet_effort:    V.dev_sonnet_effort
			triage_model:         V.triage_model
			triage_effort:        V.triage_effort
			escalation_chat_pool: V.escalation_chat_pool
			linear_workspace:     V.linear_workspace
			linear_team:          V.linear_team
			linear_intake_labels: V.linear_intake_labels
			linear_human:         V.linear_human
			human:                "elee1766"
			authors:              "zexoverz,hhenry88,forgotten-memory"
			ci_workflow:          "ci"
			check_cmd:            "go build ./... && go vet ./... && go test ./... && golangci-lint run ./... && nix fmt -- --check $(git ls-files '*.nix')"
			docs_rule:            "behavior as built goes in a new file `docs/notes/<issue>-<slug>.md` (short, dated), a runbook check in a new file `docs/verify/<slug>.md`. Do not edit docs/plan.md (frozen 2026-11-05); a design decision is one dated comment on the decision log issue #569. Never append a numbered section to docs/gasserver-verify.md."
			guide:                "Jobshop is Go, module github.com/elee1766/jobshop. Read AGENTS.md and docs/ first (docs/plan.md is the design archive; newer decisions are on issue #569)."
			deploy_check:         "`jobshop version --server` prints the running controller's `jobshop <version> (<commit>)`. Plain `jobshop version` is this session's binary, fixed when the session started, so it is stale after a self-deploy (#180). The commit must be the merge commit (or the needed SHA) or a descendant: `git fetch origin && git merge-base --is-ancestor <sha> <commit>`. Self-deploy rebuilds with the newest green master commit every 10 minutes."
		}
	}

	// Runtime templates use GitHub's repo name, not the checkout input.
	vars: {
		for name, value in _github.vars {"\(name)": value}
		for name, value in V if name != "repo" {"\(name)": value}
	}
	pools: {
		for name, pool in _github.pools {"\(name)": pool}
		learn: {
			agent:        V.agent
			repo:         V.repo
			ref:          "origin/master"
			max_sessions: V.learn_sessions
			idle_timeout: "10m"
			turn_timeout: "2h"
			prompt_file:  "prompts/learn.md"
			admin: ["memory-write"]
		}
	}
	workflows: {
		for name, workflow in _github.workflows {"\(name)": workflow}
		curate: {
			escalate_pool:     "plan"
			escalate:          _escalate
			on_start:          "workflow.star:curate"
			on_step_completed: "workflow.star:completed"
		}
		evaluate: {
			escalate_pool:     "plan"
			escalate:          _escalate
			on_start:          "workflow.star:evaluate"
			on_step_completed: "workflow.star:completed"
		}
		"live-check": {
			vars: ["runbook", "issue"]
			issue:             "issue"
			repo:              "elee1766/jobshop"
			escalate_pool:     "plan"
			on_start:          "workflow.star:live_check"
			on_step_completed: "workflow.star:completed"
		}
	}
	_escalate: {
		chat_pool:        V.escalation_chat_pool
		prompt_file:      "../github/prompts/escalation.md"
		pool_prompt_file: "../github/prompts/escalation-pool.md"
		hold_prompt_file: "../github/prompts/escalation-hold.md"
	}
	triggers: {
		for name, trigger in _github.triggers {"\(name)": trigger}
		evaluate: {
			when: edits: {runs: V.evaluate_runs, age: V.evaluate_age}
			handler: "triggers.star:start"
			args: {workflow: "evaluate", key: "evaluate", title: "Evaluate recent runs and edits"}
			max_open: 1
		}
		curate: {
			when: learnings: {min: V.curate_min_learnings, every: V.curate_every, scope: V.curate_scope}
			handler: "triggers.star:start"
			args: {workflow: "curate", key: "curate", title: "Curate learnings"}
			max_open: 1
		}
	}
	prompt_rules: list.Concat([_github.prompt_rules, [{
		name: "selfdev-memory"
		at:   "system"
		match: chat: false
		memory: {dir: V.memory_dir, cap: V.memory_cap}
	}, {
		name: "selfdev-guide"
		at:   "system"
		match: {chat: false, repos: ["elee1766/jobshop"]}
		file: "prompts/guide.md"
	}, {
		name: "selfdev-live-check"
		at:   "system"
		match: {chat: false, pools: ["plan"], repos: ["elee1766/jobshop"]}
		file: "prompts/live-check.md"
	}]])
}
