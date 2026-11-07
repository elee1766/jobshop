package demo

#Pack: {
	let V = vars
	path: "packs/demo"
	vars: {
		repo: string & =~"^[^/]+/[^/]+$"
		branch: *"main" | string
		enabled: *false | bool
	}
	pools: dev: {
		agent: "demo"
		dir: "."
		max_sessions: 0
		prompt_file: "prompt.md"
	}
	workflows: change: {
		enabled: V.enabled
		vars: ["task"]
		steps: [{
			id: "implement"
			pool: "dev"
			title: "Implement: {{.task}}"
			prompt_file: "implement.md"
			retries: 1
		}]
	}
	workflows: hooked: {
		enabled: V.enabled
		project: "js"
		vars: ["task", "pool"]
		on_start: "workflow.star:begin"
		on_step_completed: "workflow.star:completed"
	}
	triggers: "demo-event": {
		enabled: false
		on: [{source: "generic", event: "tasks"}]
		handler: "events.star:handle"
		args: {workflow: "demo-hooked", pool: "demo-dev"}
		cooldown: "1m"
		max_open: 1
	}
}
