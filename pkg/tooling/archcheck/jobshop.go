package archcheck

// Jobshop is the ownership table and boundary rules for this repository
// (JOB-165, parent JOB-164). Paths are today's paths. A child that moves a
// package updates its row, and adds the rules the move makes possible, in the
// same PR. docs/notes/JOB-165-package-boundaries.md explains the decisions.
func Jobshop() Config {
	return Config{
		Families: families(),
		Owners:   owners(),
		Rules:    rules(),
		Only:     onlyRules(),
		Known:    known(),
	}
}

// families lists the layers. A family's May list is every other family its
// production code may import, so a new edge between families fails until
// someone decides it is wanted. The family graph must have no cycle.
func families() []Family {
	return []Family{
		{Name: "support"},
		{Name: "ids"},
		// domain is the shared business types and invariants (JOB-166). It
		// imports nothing of the module and no third-party package, so any
		// family can take a type from it without taking an implementation.
		{Name: "domain", StdOnly: true},
		// pack is a leaf of the module apart from itself: it assembles files and
		// returns data, and config decodes it. It may not import config, store,
		// automation or any provider (rule pack-is-independent).
		{Name: "pack"},
		{Name: "store", May: []string{"support", "ids", "domain"}},
		{Name: "knowledge", May: []string{"store", "support", "domain"}},
		// config reads store types it has not moved yet and compiles hook
		// programs to reject a bad script at load (config/hookcheck.go).
		{Name: "config", May: []string{"support", "ids", "pack", "store", "automation", "domain"}},
		{Name: "integrations", May: []string{"support", "config", "domain"}},
		{Name: "automation", May: []string{"integrations", "domain"}},
		{Name: "wire", May: []string{"support", "config", "store", "pack", "domain"}},
		// sandbox is process isolation and resource mechanics: it takes a
		// config and a process spec and knows nothing of sessions, so a
		// command path can use it (JOB-167).
		{Name: "sandbox", May: []string{"support", "config", "domain"}},
		{Name: "session", May: []string{"support", "config", "sandbox", "wire", "domain"}},
		{Name: "journal", May: []string{"support", "store", "wire", "config", "domain"}},
		{Name: "conversation", May: []string{"support", "config", "store", "wire", "domain"}},
		{Name: "orchestration", May: []string{
			"support", "ids", "store", "knowledge", "config", "integrations",
			"automation", "wire", "sandbox", "session", "journal", "conversation", "pack", "domain",
		}},
		{Name: "transport", May: []string{
			"support", "ids", "store", "knowledge", "config", "integrations",
			"wire", "session", "journal", "conversation", "orchestration", "tooling", "pack", "domain",
		}},
		{Name: "tooling", May: []string{"wire"}},
		{Name: "content", May: []string{"config", "store", "wire", "orchestration", "domain"}},
	}
}

func owners() []Owner {
	return []Owner{
		// support: separate cross-cutting primitives, no umbrella package.
		{Pkg: "./pkg/support/clock", Target: "stay", Family: "support", Child: "-"},
		{Pkg: "./pkg/support/redact", Target: "stay", Family: "support", Child: "-"},
		{Pkg: "./pkg/support/logctx", Target: "stay", Family: "support", Child: "-"},
		{Pkg: "./pkg/support/telemetry", Target: "stay", Family: "support", Child: "-"},

		// ids: the generator and its id shape checks stay a leaf. Only the two
		// store backends and a store test use it besides config and control, and
		// none needs it for a type, so domain holds no identifier kinds (JOB-166).
		{Pkg: "./pkg/ids", Target: "stay (leaf generator)", Family: "ids", Child: "-"},

		// domain: shared business types and invariants (JOB-166).
		{Pkg: "./pkg/domain", Target: "stay", Family: "domain", Child: "-"},

		// pack: assembly, provenance and watch paths, with the standalone CUE
		// evaluator and the pack template language beneath it (JOB-172). pack
		// takes the few facts it needs as data, so it never imports config.
		{Pkg: "./pkg/pack", Target: "stay", Family: "pack", Child: "-"},
		{Pkg: "./pkg/pack/cue", Target: "stay", Family: "pack", Child: "-"},
		{Pkg: "./pkg/pack/tmpl", Target: "stay", Family: "pack", Child: "-"},

		// storage
		{Pkg: "./pkg/storage/store", Target: "storage/store", Family: "store", Child: "JOB-173"},
		{Pkg: "./pkg/storage/store/sqlite", Target: "storage/store/sqlite", Family: "store", Child: "JOB-173"},
		{Pkg: "./pkg/storage/store/sqlite/sqldb", Target: "storage/store/sqlite/sqldb", Family: "store", Child: "JOB-173"},
		{Pkg: "./pkg/storage/store/postgres", Target: "storage/store/postgres", Family: "store", Child: "JOB-173"},
		{Pkg: "./pkg/storage/store/postgres/pgdb", Target: "storage/store/postgres/pgdb", Family: "store", Child: "JOB-173"},
		{Pkg: "./pkg/storage/store/storerules", Target: "storage/store/storerules (I/O-free rules both backends share)", Family: "store", Child: "JOB-173"},
		{Pkg: "./pkg/storage/store/storemigrate", Target: "storage/store/storemigrate", Family: "store", Child: "JOB-173"},
		{Pkg: "./pkg/storage/store/storetest", Target: "storage/store/storetest", Family: "store", Child: "JOB-173", Harness: true},
		{Pkg: "./pkg/storage/store/writekey", Target: "storage/store/writekey (a store collection; control is the only caller)", Family: "store", Child: "JOB-173"},

		{Pkg: "./pkg/storage/journal", Target: "storage/journal (frames, VictoriaLogs client, reads, search, recall index)", Family: "journal", Child: "JOB-173"},
		{Pkg: "./pkg/storage/journal/vltest", Target: "storage/journal/vltest", Family: "journal", Child: "JOB-173", Harness: true},
		{Pkg: "./pkg/storage/journal/ship", Target: "storage/journal/ship", Family: "journal", Child: "JOB-173"},
		{Pkg: "./pkg/storage/journal/compress", Target: "storage/journal/compress", Family: "journal", Child: "JOB-173"},
		{Pkg: "./pkg/storage/journal/journalvl", Target: "storage/journal (tests only)", Family: "journal", Child: "JOB-173", Harness: true},

		// knowledge: memory stays usable without the database-backed records.
		// learnings, memory edits, the curator gate's judgement and the
		// curation mechanics (JOB-170). memory is a leaf under it.
		{Pkg: "./pkg/knowledge", Target: "stay", Family: "knowledge", Child: "-"},
		{Pkg: "./pkg/knowledge/memory", Target: "stay (leaf: files and skills, no storage)", Family: "knowledge", Child: "-"},

		// config
		{Pkg: "./pkg/config", Target: "stay (service configuration loading and composition)", Family: "config", Child: "-"},
		{Pkg: "./pkg/config/cron", Target: "stay (parsed by config at load, evaluated by triggers)", Family: "config", Child: "-"},

		// integrations (JOB-169): one client per provider, no shared tracker
		// interface. apicall stays public because automation and cli sit outside
		// the family and need the token source and the observer.
		{Pkg: "./pkg/integrations/apicall", Target: "stay", Family: "integrations", Child: "-"},
		{Pkg: "./pkg/integrations/github", Target: "stay", Family: "integrations", Child: "-"},
		{Pkg: "./pkg/integrations/linear", Target: "stay", Family: "integrations", Child: "-"},
		{Pkg: "./pkg/integrations/identity", Target: "stay", Family: "integrations", Child: "-"},
		{Pkg: "./pkg/integrations/internal/transport", Target: "stay (token renewal and body limit both clients share)", Family: "integrations", Child: "-"},

		// execution
		// JOB-167: Starlark runtime and its capabilities, process isolation
		// and resource mechanics, and the session host mechanics, each in
		// its own package under pkg/execution.
		{Pkg: "./pkg/execution/automation/...", Target: "stay", Family: "automation", Child: "-"},
		{Pkg: "./pkg/execution/automation", Target: "stay", Family: "automation", Child: "-"},
		{Pkg: "./pkg/execution/sandbox", Target: "stay (isolation, cgroups, ACLs, mounts, isolated command; no session lifecycle)", Family: "sandbox", Child: "-"},
		{Pkg: "./pkg/execution/session", Target: "stay (host start, adopt, stop, liveness, slots, homes, ACP client, spec.json)", Family: "session", Child: "-"},

		// wire
		{Pkg: "./pkg/transport/rpc", Target: "stay", Family: "wire", Child: "-"},
		{Pkg: "./pkg/transport/rpc/rpcschema", Target: "stay", Family: "wire", Child: "-"},

		// conversation
		{Pkg: "./pkg/conversation", Target: "stay (managed chat rules, escalation chat policy, message envelope, attachments)", Family: "conversation", Child: "JOB-171"},
		{Pkg: "./pkg/conversation/attention", Target: "stay (attention projection and item choices, a leaf of conversation)", Family: "conversation", Child: "JOB-171"},
		{Pkg: "./pkg/conversation/question", Target: "stay (question ticket content, a leaf of conversation)", Family: "conversation", Child: "JOB-171"},

		// orchestration
		{Pkg: "./pkg/orchestration/control", Target: "stay", Family: "orchestration", Child: "-"},
		{Pkg: "./pkg/orchestration/control/controltest", Target: "stay", Family: "orchestration", Child: "-", Harness: true},
		{Pkg: "./pkg/orchestration/activity", Target: "stay", Family: "orchestration", Child: "-"},
		{Pkg: "./pkg/orchestration/browse", Target: "stay", Family: "orchestration", Child: "-"},
		{Pkg: "./pkg/orchestration/ingest/queue", Target: "stay", Family: "orchestration", Child: "-"},
		{Pkg: "./pkg/orchestration/ingest", Target: "stay", Family: "orchestration", Child: "-"},
		{Pkg: "./pkg/orchestration/linearsvc", Target: "stay", Family: "orchestration", Child: "-"},
		{Pkg: "./pkg/orchestration/scaleplan", Target: "stay", Family: "orchestration", Child: "-"},
		{Pkg: "./pkg/orchestration/slotprovider", Target: "stay", Family: "orchestration", Child: "-"},
		{Pkg: "./pkg/orchestration/stepexec", Target: "stay", Family: "orchestration", Child: "-"},
		{Pkg: "./pkg/orchestration/threads", Target: "stay", Family: "orchestration", Child: "-"},
		{Pkg: "./pkg/orchestration/triggers", Target: "stay", Family: "orchestration", Child: "-"},
		{Pkg: "./pkg/orchestration/workflow", Target: "stay", Family: "orchestration", Child: "-"},
		{Pkg: "./pkg/orchestration/hooks", Target: "stay", Family: "orchestration", Child: "-"},

		// transport
		{Pkg: "./cmd/jobshop", Target: "stay", Family: "transport", Child: "-"},
		{Pkg: "./pkg/transport/cli", Target: "stay", Family: "transport", Child: "-"},
		{Pkg: "./pkg/transport/desk", Target: "stay", Family: "transport", Child: "-"},
		{Pkg: "./pkg/transport/gateway", Target: "stay", Family: "transport", Child: "-"},

		// tooling
		{Pkg: "./pkg/tooling/runbook", Target: "stay", Family: "tooling", Child: "-"},
		{Pkg: "./pkg/tooling/flaky", Target: "stay", Family: "tooling", Child: "-"},
		{Pkg: "./pkg/tooling/flaky/cmd/ci-flaky", Target: "stay", Family: "tooling", Child: "-"},
		{Pkg: "./pkg/transport/rpc/rpcschema/cmd/deskgen", Target: "stay", Family: "tooling", Child: "-"},
		{Pkg: "./pkg/tooling/archcheck", Target: "stay", Family: "tooling", Child: "-"},
		{Pkg: "./pkg/tooling/archcheck/cmd/archcheck", Target: "stay", Family: "tooling", Child: "-"},

		// content with tests only
		{Pkg: "./deploy", Target: "stay (docs and dashboards tests)", Family: "content", Child: "-"},
		{Pkg: "./packs", Target: "stay (bundled pack content tests)", Family: "content", Child: "-"},
	}
}
