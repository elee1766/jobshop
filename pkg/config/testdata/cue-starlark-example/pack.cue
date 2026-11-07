package city

import demopack "jobshop.example/automation/packs/demo"

default_project: "js"
projects: js: {}
agents: demo: {
	command: "jcode"
	args: ["acp"]
}

packs: demo: demopack.#Pack & {
	vars: {
		repo: "acme/widget"
		enabled: true
	}
}
