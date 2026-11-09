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
		{Name: "pack"},
		{Name: "store", May: []string{"support", "ids", "domain"}},
		{Name: "knowledge", May: []string{"store", "support", "domain"}},
		// config reads store types it has not moved yet and compiles hook
		// programs to reject a bad script at load (config/hookcheck.go).
		{Name: "config", May: []string{"support", "ids", "pack", "store", "automation", "domain"}},
		{Name: "integrations", May: []string{"support", "config", "domain"}},
		{Name: "automation", May: []string{"integrations", "domain"}},
		{Name: "wire", May: []string{"support", "config", "store", "domain"}},
		{Name: "session", May: []string{"support", "config", "wire", "domain"}},
		{Name: "journal", May: []string{"support", "store", "wire", "config", "domain"}},
		{Name: "conversation", May: []string{"support", "config", "store", "wire", "domain"}},
		{Name: "orchestration", May: []string{
			"support", "ids", "store", "knowledge", "config", "integrations",
			"automation", "wire", "session", "journal", "conversation", "pack", "domain",
		}},
		{Name: "transport", May: []string{
			"support", "ids", "store", "knowledge", "config", "integrations",
			"wire", "session", "journal", "conversation", "orchestration", "tooling", "domain",
		}},
		{Name: "tooling", May: []string{"wire"}},
		{Name: "content", May: []string{"config", "store", "wire", "orchestration", "domain"}},
	}
}

func owners() []Owner {
	return []Owner{
		// support: separate cross-cutting primitives, no umbrella package.
		{Pkg: "./pkg/clock", Target: "support/clock", Family: "support", Child: "JOB-174"},
		{Pkg: "./pkg/redact", Target: "support/redact", Family: "support", Child: "JOB-174"},
		{Pkg: "./pkg/logctx", Target: "support/logctx", Family: "support", Child: "JOB-174"},
		{Pkg: "./pkg/telemetry", Target: "support/telemetry", Family: "support", Child: "JOB-174"},

		// ids: the generator and its id shape checks stay a leaf. Only the two
		// store backends and a store test use it besides config and control, and
		// none needs it for a type, so domain holds no identifier kinds (JOB-166).
		{Pkg: "./pkg/ids", Target: "stay (leaf generator)", Family: "ids", Child: "-"},

		// domain: shared business types and invariants (JOB-166).
		{Pkg: "./pkg/domain", Target: "stay", Family: "domain", Child: "-"},

		{Pkg: "./pkg/packcue", Target: "pack/cue", Family: "pack", Child: "JOB-172"},

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
		{Pkg: "./pkg/learning", Target: "knowledge", Family: "knowledge", Child: "JOB-170"},
		{Pkg: "./pkg/memedit", Target: "knowledge (merged with learning's coordination)", Family: "knowledge", Child: "JOB-170"},
		{Pkg: "./pkg/memory", Target: "knowledge/memory", Family: "knowledge", Child: "JOB-170"},

		// config
		{Pkg: "./pkg/config", Target: "config", Family: "config", Child: "JOB-172"},
		{Pkg: "./pkg/config/tmpl", Target: "config/tmpl", Family: "config", Child: "-"},
		{Pkg: "./pkg/cron", Target: "config/cron (parsed by config at load, evaluated by triggers)", Family: "config", Child: "JOB-172"},

		// integrations (JOB-169): one client per provider, no shared tracker
		// interface. apicall stays public because automation and cli sit outside
		// the family and need the token source and the observer.
		{Pkg: "./pkg/integrations/apicall", Target: "stay", Family: "integrations", Child: "-"},
		{Pkg: "./pkg/integrations/github", Target: "stay", Family: "integrations", Child: "-"},
		{Pkg: "./pkg/integrations/linear", Target: "stay", Family: "integrations", Child: "-"},
		{Pkg: "./pkg/integrations/identity", Target: "stay", Family: "integrations", Child: "-"},
		{Pkg: "./pkg/integrations/internal/transport", Target: "stay (token renewal and body limit both clients share)", Family: "integrations", Child: "-"},

		// execution
		{Pkg: "./pkg/automation/...", Target: "execution/automation/...", Family: "automation", Child: "JOB-167"},
		{Pkg: "./pkg/automation", Target: "execution/automation", Family: "automation", Child: "JOB-167"},
		{Pkg: "./pkg/control/sessionrt", Target: "execution/session + execution/sandbox (split by file)", Family: "session", Child: "JOB-167"},
		{Pkg: "./pkg/hostspec", Target: "execution/session", Family: "session", Child: "JOB-167"},

		// wire
		{Pkg: "./pkg/rpc", Target: "transport/rpc", Family: "wire", Child: "JOB-174"},
		{Pkg: "./pkg/rpc/rpcschema", Target: "transport/rpc/rpcschema", Family: "wire", Child: "JOB-174"},

		// conversation
		{Pkg: "./pkg/conversation", Target: "stay (managed chat rules, escalation chat policy, message envelope, attachments)", Family: "conversation", Child: "JOB-171"},
		{Pkg: "./pkg/conversation/attention", Target: "stay (attention projection and item choices, a leaf of conversation)", Family: "conversation", Child: "JOB-171"},
		{Pkg: "./pkg/conversation/question", Target: "stay (question ticket content, a leaf of conversation)", Family: "conversation", Child: "JOB-171"},

		// orchestration
		{Pkg: "./pkg/control", Target: "orchestration/control", Family: "orchestration", Child: "JOB-168"},
		{Pkg: "./pkg/control/controltest", Target: "orchestration/control/controltest", Family: "orchestration", Child: "JOB-168", Harness: true},
		{Pkg: "./pkg/control/activity", Target: "orchestration/activity", Family: "orchestration", Child: "JOB-168"},
		{Pkg: "./pkg/browse", Target: "orchestration/browse (reads Linear and GitHub for Desk)", Family: "orchestration", Child: "JOB-168"},
		{Pkg: "./pkg/control/ingestqueue", Target: "orchestration/ingest", Family: "orchestration", Child: "JOB-168"},
		{Pkg: "./pkg/ingest", Target: "orchestration/ingest", Family: "orchestration", Child: "JOB-168"},
		{Pkg: "./pkg/control/linearsvc", Target: "orchestration/linearsvc", Family: "orchestration", Child: "JOB-168"},
		{Pkg: "./pkg/control/scaleplan", Target: "orchestration/scaleplan", Family: "orchestration", Child: "JOB-168"},
		{Pkg: "./pkg/control/slotprovider", Target: "orchestration/slotprovider", Family: "orchestration", Child: "JOB-168"},
		{Pkg: "./pkg/control/stepexec", Target: "orchestration/stepexec", Family: "orchestration", Child: "JOB-168"},
		{Pkg: "./pkg/control/threads", Target: "orchestration/threads", Family: "orchestration", Child: "JOB-168"},
		{Pkg: "./pkg/control/triggers", Target: "orchestration/triggers", Family: "orchestration", Child: "JOB-168"},
		{Pkg: "./pkg/control/waitengine", Target: "orchestration/workflow", Family: "orchestration", Child: "JOB-168"},
		{Pkg: "./pkg/hooks", Target: "orchestration/hooks", Family: "orchestration", Child: "JOB-168"},

		// transport
		{Pkg: "./cmd/jobshop", Target: "stay", Family: "transport", Child: "-"},
		{Pkg: "./pkg/cli", Target: "transport/cli", Family: "transport", Child: "JOB-174"},
		{Pkg: "./pkg/desk", Target: "transport/desk", Family: "transport", Child: "JOB-174"},
		{Pkg: "./pkg/gateway", Target: "transport/gateway", Family: "transport", Child: "JOB-174"},

		// tooling
		{Pkg: "./pkg/runbook", Target: "tooling/runbook", Family: "tooling", Child: "JOB-174"},
		{Pkg: "./pkg/flaky", Target: "tooling/flaky", Family: "tooling", Child: "JOB-174"},
		{Pkg: "./pkg/flaky/cmd/ci-flaky", Target: "tooling/flaky/cmd/ci-flaky", Family: "tooling", Child: "JOB-174"},
		{Pkg: "./pkg/rpc/rpcschema/cmd/deskgen", Target: "transport/rpc/rpcschema/cmd/deskgen", Family: "tooling", Child: "JOB-174"},
		{Pkg: "./pkg/tooling/archcheck", Target: "stay", Family: "tooling", Child: "-"},
		{Pkg: "./pkg/tooling/archcheck/cmd/archcheck", Target: "stay", Family: "tooling", Child: "-"},

		// content with tests only
		{Pkg: "./deploy", Target: "stay (docs and dashboards tests)", Family: "content", Child: "-"},
		{Pkg: "./packs", Target: "stay (bundled pack content tests)", Family: "content", Child: "-"},
	}
}
