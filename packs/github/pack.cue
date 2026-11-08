package github

// #Pack works one GitHub repo through issues and PRs. Workflows and
// triggers are Starlark; see README.md.
#Pack: {
	let V = vars
	let P = prefix
	path:   *"packs/github" | string
	prefix: *"" | string
	vars: {
		repo:                   string
		checkout:               string
		agent:                  string
		human:                  string
		pack_prefix:            P
		ci_workflow:            *"ci" | string
		ci_branch:              *"master" | string
		dev_sessions:           *2 | int & >=0
		refinery_sessions:      *1 | int & >=0
		refinery_model:         *"" | string
		refinery_effort:        *"" | string
		dev_pool:               *"\(P)dev" | string
		dev_model:              *"" | string
		dev_effort:             *"" | string
		dev_sonnet_sessions:    *1 | int & >=0
		dev_sonnet_model:       *"claude-sonnet-5-5" | string
		dev_sonnet_effort:      *"high" | string
		triage_model:           *"" | string
		triage_effort:          *"" | string
		deploy_sessions:        *3 | int & >=0
		merge_method:           *"squash" | string
		check_cmd:              *"true" | string
		required_approver:      *"" | string
		required_checks:        *"" | string
		guide:                  *"Read the repo's README and contributor docs (AGENTS.md, CONTRIBUTING.md) first." | string
		scan_after:             *"1h" | string
		scan_step_after:        *"3h" | string
		scan_human_after:       *"4h" | string
		authors:                *"" | string
		pr_review_skip_authors: *"" | string
		pr_review_skip_labels:  *"hold" | string
		release_pr_label:       *"autorelease: pending" | string
		deploy_check:           *"none (these agents do not deploy this repo); treat this check as passed." | string
		tracker:                *"github" | "linear"
		linear_workspace:       *"" | string
		linear_team:            *"" | string
		linear_target:          *"" | string
		linear_done_state:      *"Done" | string
		linear_skip_labels:     *"Blocked,target,proposed,tracking" | string
		linear_human:           *"" | string
		linear_intake_states:   *"Todo" | string
		linear_intake_labels:   *"" | string
		linear_new_state:       *"Todo" | string
		linear_intake_assignee: *"" | string
		linear_poll:            *"5m" | string
		docs_rule:              *"put a design note or runbook check in a new file of its own, named by the issue number. Never append a numbered section to a shared doc." | string
		escalation_chat_pool:   *"" | string
		propose_targets:        *true | bool
		asset_prefix:           *"" | string
		...
	}

	_pool: {
		agent:        V.agent
		repo:         V.checkout
		ref:          "origin/master"
		idle_timeout: "10m"
		turn_timeout: "2h"
		...
	}
	_admin: ["session-list", "session-restart", "trigger-list", "report-incidents", "attention", "activity", "handoff"]
	pools: {
		dev: _pool & {
			max_sessions: V.dev_sessions
			model:        V.dev_model
			effort:       V.dev_effort
			prompt_file:  "\(V.asset_prefix)prompts/\(V.tracker)/dev.md"
		}
		"dev-sonnet": _pool & {
			max_sessions: V.dev_sonnet_sessions
			model:        V.dev_sonnet_model
			effort:       V.dev_sonnet_effort
			prompt_file:  "\(V.asset_prefix)prompts/\(V.tracker)/dev.md"
		}
		plan: _pool & {
			max_sessions: 1
			prompt_file:  "\(V.asset_prefix)prompts/\(V.tracker)/plan.md"
			admin:        _admin
		}
		triage: _pool & {
			max_sessions: 1
			model:        V.triage_model
			effort:       V.triage_effort
			admin:        _admin
		}
		refinery: _pool & {
			max_sessions: V.refinery_sessions
			model:        V.refinery_model
			effort:       V.refinery_effort
			prompt_file:  "\(V.asset_prefix)prompts/refinery.md"
		}
		"deploy-check": _pool & {
			max_sessions: V.deploy_sessions
			prompt_file:  "\(V.asset_prefix)prompts/deploy-check.md"
		}
		...
	}

	// Handlers decide; the controller keeps per-subject history, cooldowns
	// and caps for the scan signals.
	_github: {repo: V.repo, ci_workflow: V.ci_workflow, ci_branch: V.ci_branch}
	triggers: {
		replan: {
			when: {pool: "dev", ready_below: 1}
			handler: "\(V.asset_prefix)triggers.star:replan"
			args: {dev_pool: "dev", plan_pool: "plan"}
			cooldown: "10m"
			max_open: 1
		}
		"master-red": {
			on: [{github: {event: "workflow_run", action: "completed", repo: V.repo, workflow: V.ci_workflow, branch: V.ci_branch, conclusion: "failure"}}]
			fallback: "5m"
			handler:  "\(V.asset_prefix)triggers.star:master_red"
			args:     _github
		}
		"release-pr": {
			on: [{github: {event: "check_suite", action: "completed", repo: V.repo}}]
			fallback: "5m"
			handler:  "\(V.asset_prefix)triggers.star:release_pr"
			args: {repo: V.repo, release_pr_label: V.release_pr_label}
		}
		scan: {
			when: scan: [
				{signal: "ready_unclaimed", after: V.scan_after},
				{signal: "pool_quarantined"},
				{signal: "step_idle", after: V.scan_step_after},
				{signal: "session_silent", after: V.scan_after},
				{signal: "needs_human", after: V.scan_human_after},
				{signal: "blocked", after: V.scan_after},
				{
					signal: "issue_no_run"
					repo:   V.repo
					script: "\(V.asset_prefix)triggers.star:issue_no_run"
					args: {repo: V.repo, human: V.human, authors: V.authors}
				},
				{
					signal:      "linear_issue_no_run"
					team:        V.linear_team
					states:      V.linear_intake_states
					labels:      V.linear_intake_labels
					skip_labels: V.linear_skip_labels
					poll:        V.linear_poll
				},
				{
					signal: "no_merge"
					script: "\(V.asset_prefix)triggers.star:no_merge"
					args:   _github
				},
			]
			run: {workflow: "triage", vars: {signal: "{{.signal}}", subject: "{{.subject}}", detail: "{{.detail}}"}}
			cooldown: "30m"
			max_runs: 2
		}
		"pr-unmerged": {
			when: scan: [{
				signal: "pr_unmerged"
				script: "\(V.asset_prefix)triggers.star:pr_unmerged"
				args: {repo: V.repo, human: V.human, skip_authors: V.pr_review_skip_authors, skip_labels: V.pr_review_skip_labels, release_pr_label: V.release_pr_label}
			}]
			run: {workflow: "pr-review", vars: {pr: "{{.subject}}"}, key: "pr-{{.subject}}"}
			cooldown: "30m"
			max_runs: 2
		}
		...
	}

	prompt_rules: [
		{name: "desk-links", at: "turn", file: "\(V.asset_prefix)prompts/rules/desk-links.md"},
		{name: "recall", at: "system", file: "\(V.asset_prefix)prompts/rules/recall.md"},
		{name: "escalate-policy", at: "system", file: "\(V.asset_prefix)prompts/rules/escalate-policy.md"},
		{name: "learn-nudge", at: "after", file: "\(V.asset_prefix)prompts/rules/learn-nudge.md"},
		{name: "dev-pool", at: "system", match: {pools: ["plan", "dev", "dev-sonnet"]}, file: "\(V.asset_prefix)prompts/rules/dev-pool.md"},
		...,
	]

	_workflow: {
		escalate_pool: "plan"
		escalate: {
			chat_pool:        V.escalation_chat_pool
			prompt_file:      "\(V.asset_prefix)prompts/escalation.md"
			pool_prompt_file: "\(V.asset_prefix)prompts/escalation-pool.md"
			hold_prompt_file: "\(V.asset_prefix)prompts/escalation-hold.md"
		}
		...
	}
	_star: "\(V.asset_prefix)workflows.star"
	workflows: {
		"pr-work": _workflow & {
			vars: ["issue", "dev_pool"]
			issue:   "issue"
			repo:    V.repo
			tracker: V.tracker
			if V.tracker == "github" {linear_workspace: ""}
			if V.tracker == "linear" {linear_workspace: V.linear_workspace}
			on_start:          "\(_star):pr_work_start"
			on_step_completed: "\(_star):pr_work_completed"
		}
		"pr-work-linear": _workflow & {
			vars: ["issue", "dev_pool"]
			issue:             "issue"
			repo:              V.repo
			tracker:           "linear"
			linear_workspace:  V.linear_workspace
			on_start:          "\(_star):pr_work_linear_start"
			on_step_completed: "\(_star):pr_work_linear_completed"
		}
		"pr-review": _workflow & {
			vars: ["pr"]
			pr:                "pr"
			repo:              V.repo
			on_start:          "\(_star):pr_review_start"
			on_step_completed: "\(_star):pr_review_completed"
		}
		triage: _workflow & {
			vars: ["signal", "subject", "detail"]
			on_start:          "\(_star):triage_start"
			on_step_completed: "\(_star):finish"
		}
		"ci-repair": _workflow & {
			vars: ["sha"]
			repo:              V.repo
			on_start:          "\(_star):ci_start"
			on_step_completed: "\(_star):ci_completed"
		}
		"release-merge": _workflow & {
			vars: ["pr"]
			pr:                "pr"
			repo:              V.repo
			on_start:          "\(_star):release_start"
			on_step_completed: "\(_star):finish"
		}
		drill: _workflow & {
			vars: ["issue"]
			issue:             "issue"
			repo:              V.repo
			on_start:          "\(_star):drill_start"
			on_step_completed: "\(_star):finish"
		}
		...
	}
	...
}
