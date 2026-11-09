package archcheck

import "slices"

const (
	acphostRuntime = "github.com/elee1766/acphost/..."
	acpSDK         = "github.com/coder/acp-go-sdk"
)

// rules are the named boundary rules. Each says in Why what it protects, so
// a failure explains the intent, not only the edge. Rules for packages that
// do not exist yet (the shared domain package, the pack and storage
// families) arrive with the child that creates them: a rule whose pattern
// matches nothing fails the check.
func rules() []Rule {
	harness := harnessPackages()
	return []Rule{
		{
			Name:   "control-is-top",
			Why:    "control owns orchestration. A subsystem, a store, a test support package or a client must not import it, including from a test, or the dependency points the wrong way and a cycle is one import away. Tests reach the controller through controltest.",
			Scope:  ProdAndTest,
			From:   []string{"./pkg/...", "./cmd/..."},
			Except: slices.Concat(harness, []string{"./pkg/control"}),
			To:     []string{"./pkg/control"},
		},
		{
			Name:   "no-harness-in-production",
			Why:    "A test support package (storetest, controltest, vltest, journalvl) is for test files. Production code importing one would link test dependencies into the binary.",
			Scope:  Prod,
			From:   []string{"./pkg/...", "./cmd/..."},
			Except: harness,
			To:     harness,
		},
		{
			Name:  "command-paths-no-session",
			Why:   "A hook, an automation capability or a controller-run step runs a command. It may use process and isolation mechanics, but must not depend on session lifecycle (host adoption, liveness, slots, ACP conversations), directly or through what it imports.",
			Scope: Closure,
			From: []string{
				"./pkg/hooks",
				"./pkg/automation",
				"./pkg/automation/...",
				"./pkg/control/stepexec",
			},
			To: []string{
				"./pkg/control",
				"./pkg/control/sessionrt",
				"./pkg/control/chatlife",
				"./pkg/hostspec",
				"github.com/elee1766/acphost/pkg/acphost",
			},
		},
		{
			Name:  "command-paths-no-acp",
			Why:   "The ACP SDK is the agent conversation protocol. A hook or capability has no conversation, so reaching the SDK means it imports session-side types by accident.",
			Scope: Closure,
			From: []string{
				"./pkg/hooks",
				"./pkg/automation",
				"./pkg/automation/...",
				"./pkg/control/stepexec",
			},
			To: []string{acpSDK},
		},
		{
			Name:  "lower-layers-no-session-runtime",
			Why:   "Stores, knowledge, config, integrations, packs and support must be usable without the process host and the ACP SDK. A consumer of one of them must not link a session runtime.",
			Scope: Closure,
			From: []string{
				"./pkg/store", "./pkg/store/sqlite", "./pkg/store/postgres", "./pkg/store/storerules",
				"./pkg/writekey", "./pkg/learning", "./pkg/memedit", "./pkg/memory",
				"./pkg/config", "./pkg/config/tmpl", "./pkg/cron", "./pkg/ids", "./pkg/packcue",
				"./pkg/apicall", "./pkg/github", "./pkg/linear", "./pkg/identity",
				"./pkg/clock", "./pkg/redact", "./pkg/logctx", "./pkg/telemetry",
				"./pkg/automation/...", "./pkg/automation",
			},
			To: []string{acphostRuntime, acpSDK},
		},
		{
			Name:  "starlark-only-in-automation",
			Why:   "Starlark execution belongs to the automation family. Another package that imports it grows a second way to run scripts.",
			Scope: ProdAndTest,
			From:  []string{"./pkg/...", "./cmd/..."},
			Except: []string{
				"./pkg/automation",
				"./pkg/automation/...",
			},
			To: []string{"go.starlark.net/..."},
		},
		{
			Name:  "cue-only-in-packcue",
			Why:   "CUE evaluation is the pack subsystem's. A package that evaluates CUE itself bypasses its no-registry rule.",
			Scope: ProdAndTest,
			From:  []string{"./pkg/...", "./cmd/..."},
			Except: []string{
				"./pkg/packcue",
			},
			To: []string{"cuelang.org/go/..."},
		},
		{
			Name:  "sqlite-driver-owners",
			Why:   "The SQLite driver is opened by the store backend, and by the ingest spool, which is its own database file on a separate host. No other package opens a database.",
			Scope: ProdAndTest,
			From:  []string{"./pkg/...", "./cmd/..."},
			Except: []string{
				"./pkg/store/sqlite",
				"./pkg/ingest",
			},
			To: []string{"modernc.org/sqlite"},
		},
		{
			Name:  "postgres-driver-owners",
			Why:   "Only the Postgres backend talks to Postgres.",
			Scope: ProdAndTest,
			From:  []string{"./pkg/...", "./cmd/..."},
			Except: []string{
				"./pkg/store/postgres",
				// Its test opens an admin connection to create a scratch
				// database before copying into it.
				"./pkg/store/storemigrate",
			},
			To: []string{"github.com/jackc/pgx/..."},
		},
		{
			Name:  "store-backends-are-siblings",
			Why:   "The SQLite and Postgres backends share only store types and storerules. Importing each other would couple their SQL.",
			Scope: ProdAndTest,
			From:  []string{"./pkg/store/sqlite", "./pkg/store/sqlite/..."},
			To:    []string{"./pkg/store/postgres", "./pkg/store/postgres/..."},
		},
		{
			Name:  "store-backends-are-siblings-reverse",
			Why:   "The SQLite and Postgres backends share only store types and storerules. Importing each other would couple their SQL.",
			Scope: ProdAndTest,
			From:  []string{"./pkg/store/postgres", "./pkg/store/postgres/..."},
			To:    []string{"./pkg/store/sqlite", "./pkg/store/sqlite/..."},
		},
		{
			Name:  "domain-is-a-leaf",
			Why:   "domain is imported by every family for its types. Its production code reaching any other package of the module (store, rpc, config, control) would hand every consumer that package's dependencies again, which is what the package exists to avoid. The family rule keeps third-party modules out too.",
			Scope: ProdAndTest,
			From:  []string{"./pkg/domain"},
			To:    []string{"./pkg/...", "./cmd/..."},
		},
		{
			Name:  "types-not-from-store-or-rpc",
			Why:   "These packages need ticket, wait, learning and question types, not a persistence interface or the wire protocol. Reaching store or rpc, even through another package, drags SQL builders, the ACP connection and the sandbox types into a consumer of a struct. A package that needs the store itself is not listed.",
			Scope: Closure,
			From: []string{
				"./pkg/config", "./pkg/control/question", "./pkg/control/scaleplan",
				"./pkg/control/waitengine",
			},
			To: []string{"./pkg/rpc", "./pkg/store"},
		},
		{
			Name:  "hooks-not-from-rpc",
			Why:   "hooks owns durable hook execution. Its host adapter in control maps the wire envelopes, so hooks takes plain inputs and returns plain results. Reaching rpc also reached the ACP SDK and the sandbox types.",
			Scope: Closure,
			From:  []string{"./pkg/hooks"},
			To:    []string{"./pkg/rpc"},
		},
	}
}

// onlyRules are allow-list rules for leaves that must stay small.
func onlyRules() []Only {
	return []Only{
		{
			Name: "independent-leaves",
			Why:  "These packages are used on their own or by many others: clock, cron, logctx, redact and tmpl by most of the module, memory by the CLI without a database (JOB-164: knowledge/memory stays usable without the database-backed learning machinery), packcue as the standalone CUE evaluation, ids and apicall as generator and token mechanics, hostspec as the spec file format. If one imports another package of the module, every consumer takes the cost. Third-party imports are fine.",
			From: []string{
				"./pkg/clock", "./pkg/cron", "./pkg/logctx", "./pkg/redact", "./pkg/config/tmpl",
				"./pkg/memory", "./pkg/packcue", "./pkg/ids", "./pkg/apicall", "./pkg/hostspec",
			},
			Std:   true,
			Third: true,
		},
	}
}

// harnessPackages are the packages the ownership table marks as test support.
func harnessPackages() []string {
	var out []string
	for _, o := range owners() {
		if o.Harness {
			out = append(out, o.Pkg)
		}
	}
	return out
}
