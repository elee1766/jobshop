package github

// #Pack keeps the GitHub pack's names and policy. Select callbacks only after
// draining legacy runs. Controller-owned stall discovery stays declarative.
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
		stall_model:            *"" | string
		stall_effort:           *"" | string
		deploy_sessions:        *3 | int & >=0
		merge_method:           *"squash" | string
		check_cmd:              *"true" | string
		required_approver:      *"" | string
		required_checks:        *"" | string
		guide:                  *"Read the repo's README and contributor docs (AGENTS.md, CONTRIBUTING.md) first." | string
		stall_after:            *"1h" | string
		stall_step_after:       *"3h" | string
		stall_human_after:      *"4h" | string
		authors:                *"" | string
		stall_pr_skip_authors:  *"" | string
		stall_pr_skip_labels:   *"hold" | string
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
		starlark_workflows:     *false | bool
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
		stall: _pool & {
			max_sessions: 1
			model:        V.stall_model
			effort:       V.stall_effort
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

	// The runtime owns stall subjects, durable history caps, and closed-ticket cooldowns.
	triggers: {
		replan: {
			when: {pool: "dev", ready_below: 1}
			ticket: {pool: "plan", title: "Plan next step", priority: 1, labels: ["plan"]}
			cooldown: "10m"
		}
		"master-red": {
			on: [{github: {event: "workflow_run", action: "completed", repo: V.repo, workflow: V.ci_workflow, branch: V.ci_branch, conclusion: "failure"}}]
			fallback: "5m"
			when: {
				check:         "gh run list --repo \(V.repo) --branch \(V.ci_branch) --workflow \(V.ci_workflow) --status completed --limit 1 --json conclusion,headSha --jq '.[] | select(.conclusion == \"failure\") | .headSha' | grep ."
				check_timeout: "30s"
			}
			run: {workflow: "ci-repair", vars: {sha: "{{.check_output}}"}, key: "{{.check_output}}"}
		}
		"release-pr": {
			on: [{github: {event: "check_suite", action: "completed", repo: V.repo}}]
			fallback: "5m"
			when: {
				check:         "test -n \"\(V.release_pr_label)\" && gh pr list --repo \(V.repo) --state open --label \"\(V.release_pr_label)\" --json number,isDraft,statusCheckRollup --jq ' [.[] | select(.isDraft | not) | select((.statusCheckRollup | length) > 0 and all(.statusCheckRollup[]; .conclusion == \"SUCCESS\" or .conclusion == \"SKIPPED\" or .conclusion == \"NEUTRAL\")) | .number] | sort | .[0] // empty' | grep ."
				check_timeout: "30s"
			}
			run: {workflow: "release-merge", vars: {pr: "{{.check_output}}"}, key: "release-pr-{{.check_output}}"}
		}
		stall: {
			when: {
				check_timeout: "1m"
				stall: [
					{signal: "ready_unclaimed", after: V.stall_after},
					{signal: "pool_quarantined"},
					{signal: "step_stalled", after: V.stall_step_after},
					{signal: "session_silent", after: V.stall_after},
					{signal: "needs_human", after: V.stall_human_after},
					{signal: "blocked", after: V.stall_after},
					{
						signal: "issue_no_run"
						repo:   V.repo
						check:  "gh issue list --repo \(V.repo) --state open --limit 200 --json number,labels,title,createdAt,author --jq ' (\"\(V.human),\(V.authors)\" | ascii_downcase | split(\",\") | map(select(. != \"\"))) as $ok | .[] | select(.author.is_bot or ([.author.login | ascii_downcase] | inside($ok)) or ([.labels[].name] | any(. == \"accepted\"))) | select([.labels[].name] | any(. == \"target\" or . == \"needs-decision\" or . == \"needs-human\" or . == \"tracking\" or . == \"proposed\") | not) | select(.title | test(\"^drill\"; \"i\") | not) | select((now - (.createdAt | fromdateiso8601)) > 1800) | \"\\(.number) open, no run\"'"
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
						check:  "test \"$(gh pr list --repo \(V.repo) --state open --json number --jq length)\" -gt 0 && gh api repos/\(V.repo)/commits/\(V.ci_branch) --jq 'select((now - (.commit.committer.date | fromdateiso8601)) > 21600) | \"\\(.sha[0:12]) last commit \\(.commit.committer.date)\"'"
					},
				]
			}
			run: {workflow: "stall-check", vars: {signal: "{{.stall_signal}}", subject: "{{.stall_subject}}", detail: "{{.stall_detail}}"}}
			cooldown: "30m"
			max_runs: 2
		}
		"pr-unmerged": {
			when: {
				check_timeout: "1m"
				stall: [{
					signal: "pr_unmerged"
					check:  "gh pr list --repo \(V.repo) --state open --json number,isDraft,updatedAt,statusCheckRollup,author,labels --jq ' (\"\(V.human),\(V.stall_pr_skip_authors)\" | ascii_downcase | split(\",\") | map(select(. != \"\"))) as $skipa | (\"\(V.stall_pr_skip_labels)\" | ascii_downcase | split(\",\") | map(select(. != \"\"))) as $skipl | .[] | select(.isDraft | not) | select(.author.login | ascii_downcase | IN($skipa[]) | not) | select(any(.labels[].name | ascii_downcase; IN($skipl[])) | not) | select((\"\(V.release_pr_label)\" | ascii_downcase) as $rl | $rl == \"\" or (any(.labels[].name | ascii_downcase; . == $rl) | not)) | select((.statusCheckRollup | length) > 0 and all(.statusCheckRollup[]; .conclusion == \"SUCCESS\" or .conclusion == \"SKIPPED\" or .conclusion == \"NEUTRAL\")) | select((now - (.updatedAt | fromdateiso8601)) > 7200) | \"\\(.number) green, unmerged since \\(.updatedAt)\"'"
				}]
			}
			run: {workflow: "pr-review", vars: {pr: "{{.stall_subject}}"}, key: "pr-{{.stall_subject}}"}
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
	workflows: {
		"pr-work": _workflow & {
			vars: ["issue", "dev_pool"]
			issue:   "issue"
			repo:    V.repo
			tracker: V.tracker
			if V.tracker == "github" {linear_workspace: ""}
			if V.tracker == "linear" {linear_workspace: V.linear_workspace}
			if V.starlark_workflows {
				on_start:          "\(V.asset_prefix)workflows.star:pr_work_start"
				on_step_completed: "\(V.asset_prefix)workflows.star:pr_work_completed"
			}
			if !V.starlark_workflows {
				steps: [
					{id: "implement", pool: "{{.dev_pool}}", retry_pool: "dev", retries: 3, title: "Implement #{{.issue}}", prompt_file: "\(V.asset_prefix)prompts/\(V.tracker)/implement.md"},
					{id: "review-merge", pool: "refinery", retries: 2, when: {step: "implement", outcome: "pr"}, title: "Review and merge PR for #{{.issue}}", prompt_file: "\(V.asset_prefix)prompts/\(V.tracker)/merge.md"},
					{id: "post-merge", pool: "deploy-check", retries: 2, when: {step: "implement", outcome: "pr"}, title: "Check master CI and deploy for #{{.issue}}", prompt_file: "\(V.asset_prefix)prompts/\(V.tracker)/post-merge.md"},
				]
			}
		}
		"pr-work-linear": _workflow & {
			vars: ["issue", "dev_pool"]
			issue:            "issue"
			repo:             V.repo
			tracker:          "linear"
			linear_workspace: V.linear_workspace
			if V.starlark_workflows {
				on_start:          "\(V.asset_prefix)workflows.star:pr_work_linear_start"
				on_step_completed: "\(V.asset_prefix)workflows.star:pr_work_linear_completed"
			}
			if !V.starlark_workflows {
				steps: [
					{id: "implement", pool: "{{.dev_pool}}", retry_pool: "dev", retries: 3, title: "Implement {{.issue}}", prompt_file: "\(V.asset_prefix)prompts/linear/implement.md"},
					{id: "review-merge", pool: "refinery", retries: 2, when: {step: "implement", outcome: "pr"}, title: "Review and merge PR for {{.issue}}", prompt_file: "\(V.asset_prefix)prompts/linear/merge.md"},
					{id: "post-merge", pool: "deploy-check", retries: 2, when: {step: "implement", outcome: "pr"}, title: "Check master CI and deploy for {{.issue}}", prompt_file: "\(V.asset_prefix)prompts/linear/post-merge.md"},
				]
			}
		}
		"pr-review": _workflow & {
			vars: ["pr"]
			pr:   "pr"
			repo: V.repo
			if V.starlark_workflows {
				on_start:          "\(V.asset_prefix)workflows.star:pr_review_start"
				on_step_completed: "\(V.asset_prefix)workflows.star:pr_review_completed"
			}
			if !V.starlark_workflows {
				steps: [
					{id: "review-merge", pool: "refinery", retries: 2, title: "Review and merge PR #{{.pr}}", prompt_file: "\(V.asset_prefix)prompts/pr-review/merge.md"},
					{id: "post-merge", pool: "deploy-check", retries: 2, when: {step: "review-merge", outcome: "merged"}, title: "Check master CI and deploy for PR #{{.pr}}", prompt_file: "\(V.asset_prefix)prompts/pr-review/post-merge.md"},
				]
			}
		}
		"stall-check": _workflow & {
			vars: ["signal", "subject", "detail"]
			if V.starlark_workflows {
				on_start:          "\(V.asset_prefix)workflows.star:stall_start"
				on_step_completed: "\(V.asset_prefix)workflows.star:finish"
			}
			if !V.starlark_workflows {
				steps: [{id: "diagnose", pool: "stall", retries: 1, title: "Stall {{.signal}}: {{.subject}}", prompt_file: "\(V.asset_prefix)prompts/stall.md"}]
			}
		}
		"ci-repair": _workflow & {
			vars: ["sha"]
			repo: V.repo
			if V.starlark_workflows {
				on_start:          "\(V.asset_prefix)workflows.star:ci_start"
				on_step_completed: "\(V.asset_prefix)workflows.star:ci_completed"
			}
			if !V.starlark_workflows {
				steps: [
					{id: "fix", pool: "dev", retries: 2, title: "Fix red master CI at {{.sha}}", prompt_file: "\(V.asset_prefix)prompts/ci-fix.md"},
					{id: "merge", pool: "refinery", retries: 2, title: "Merge CI fix for {{.sha}}", prompt_file: "\(V.asset_prefix)prompts/ci-merge.md"},
				]
			}
		}
		"release-merge": _workflow & {
			vars: ["pr"]
			pr:   "pr"
			repo: V.repo
			if V.starlark_workflows {
				on_start:          "\(V.asset_prefix)workflows.star:release_start"
				on_step_completed: "\(V.asset_prefix)workflows.star:finish"
			}
			if !V.starlark_workflows {
				steps: [{id: "merge", pool: "refinery", retries: 2, title: "Merge release PR #{{.pr}}", prompt_file: "\(V.asset_prefix)prompts/release-merge.md"}]
			}
		}
		drill: _workflow & {
			vars: ["issue"]
			issue: "issue"
			repo:  V.repo
			if V.starlark_workflows {
				on_start:          "\(V.asset_prefix)workflows.star:drill_start"
				on_step_completed: "\(V.asset_prefix)workflows.star:finish"
			}
			if !V.starlark_workflows {
				steps: [{id: "fail", pool: "dev", retries: 1, title: "Escalation drill for #{{.issue}}", prompt_file: "\(V.asset_prefix)prompts/drill.md"}]
			}
		}
		...
	}
	...
}
