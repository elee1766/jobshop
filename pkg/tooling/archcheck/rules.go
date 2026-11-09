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
			Why:    "control owns orchestration. A subsystem, a store, a test support package or a client must not import it, including from a test, or the dependency points the wrong way and a cycle is one import away. Tests reach the controller through controltest. cmd/jobshop is the composition root: it passes control.Run to the CLI, so the CLI adapter never links the controller (JOB-174).",
			Scope:  ProdAndTest,
			From:   []string{"./pkg/...", "./cmd/..."},
			Except: slices.Concat(harness, []string{"./pkg/orchestration/control", "./cmd/jobshop"}),
			To:     []string{"./pkg/orchestration/control"},
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
				"./pkg/orchestration/hooks",
				"./pkg/execution/automation",
				"./pkg/execution/automation/...",
				"./pkg/execution/sandbox",
				"./pkg/orchestration/stepexec",
			},
			To: []string{
				"./pkg/orchestration/control",
				"./pkg/execution/session",
				"./pkg/conversation",
				"github.com/elee1766/acphost/pkg/acphost",
			},
		},
		{
			Name:  "command-paths-no-acp",
			Why:   "The ACP SDK is the agent conversation protocol. A hook or capability has no conversation, so reaching the SDK means it imports session-side types by accident.",
			Scope: Closure,
			From: []string{
				"./pkg/orchestration/hooks",
				"./pkg/execution/automation",
				"./pkg/execution/automation/...",
				"./pkg/execution/sandbox",
				"./pkg/orchestration/stepexec",
			},
			To: []string{acpSDK},
		},
		{
			Name:  "orchestration-engines-no-session-or-sandbox-runtime",
			Why:   "The controller starts and supervises sessions. The engines beside it (workflow, ingest, triggers, linearsvc, threads, activity, browse, scaleplan, slotprovider) take typed inputs and return decisions or call a provider, so none may link the session host mechanics or the process sandbox, directly or through what it imports. (The ACP SDK stays out of this rule: rpc carries it for several of them until JOB-174 splits the wire types.) stepexec and hooks run commands and have their own rule.",
			Scope: Closure,
			From: []string{
				"./pkg/orchestration/workflow", "./pkg/orchestration/ingest", "./pkg/orchestration/ingest/queue",
				"./pkg/orchestration/triggers", "./pkg/orchestration/linearsvc", "./pkg/orchestration/threads",
				"./pkg/orchestration/activity", "./pkg/orchestration/browse", "./pkg/orchestration/scaleplan",
				"./pkg/orchestration/slotprovider",
			},
			To: []string{"./pkg/execution/session", "./pkg/execution/sandbox"},
		},
		{
			Name:  "lower-layers-no-session-runtime",
			Why:   "Stores, knowledge, config, integrations, packs and support must be usable without the process host and the ACP SDK. A consumer of one of them must not link a session runtime.",
			Scope: Closure,
			From: []string{
				"./pkg/storage/store", "./pkg/storage/store/sqlite", "./pkg/storage/store/postgres", "./pkg/storage/store/storerules",
				"./pkg/storage/store/writekey", "./pkg/knowledge", "./pkg/knowledge/memory",
				"./pkg/config", "./pkg/config/cron", "./pkg/ids", "./pkg/pack", "./pkg/pack/cue", "./pkg/pack/tmpl",
				"./pkg/integrations/apicall", "./pkg/integrations/github", "./pkg/integrations/linear", "./pkg/integrations/identity",
				"./pkg/support/clock", "./pkg/support/redact", "./pkg/support/logctx", "./pkg/support/telemetry",
				"./pkg/execution/automation/...", "./pkg/execution/automation",
			},
			To: []string{acphostRuntime, acpSDK},
		},
		{
			Name:  "starlark-only-in-automation",
			Why:   "Starlark execution belongs to the automation family. Another package that imports it grows a second way to run scripts.",
			Scope: ProdAndTest,
			From:  []string{"./pkg/...", "./cmd/..."},
			Except: []string{
				"./pkg/execution/automation",
				"./pkg/execution/automation/...",
			},
			To: []string{"go.starlark.net/..."},
		},
		{
			Name:  "cue-only-in-pack-cue",
			Why:   "CUE evaluation is pack/cue's: pack loading and the step output schema check both go through it. A package that evaluates CUE itself bypasses its no-registry rule and grows a second evaluator.",
			Scope: ProdAndTest,
			From:  []string{"./pkg/...", "./cmd/..."},
			Except: []string{
				"./pkg/pack/cue",
			},
			To: []string{"cuelang.org/go/..."},
		},
		{
			Name:  "sqlite-driver-owners",
			Why:   "The SQLite driver is opened by the store backend, and by the ingest spool, which is its own database file on a separate host. No other package opens a database.",
			Scope: ProdAndTest,
			From:  []string{"./pkg/...", "./cmd/..."},
			Except: []string{
				"./pkg/storage/store/sqlite",
				"./pkg/orchestration/ingest",
			},
			To: []string{"modernc.org/sqlite"},
		},
		{
			Name:  "postgres-driver-owners",
			Why:   "Only the Postgres backend talks to Postgres.",
			Scope: ProdAndTest,
			From:  []string{"./pkg/...", "./cmd/..."},
			Except: []string{
				"./pkg/storage/store/postgres",
				// Its test opens an admin connection to create a scratch
				// database before copying into it.
				"./pkg/storage/store/storemigrate",
			},
			To: []string{"github.com/jackc/pgx/..."},
		},
		{
			Name:  "store-backends-are-siblings",
			Why:   "The SQLite and Postgres backends share only store types and storerules. Importing each other would couple their SQL.",
			Scope: ProdAndTest,
			From:  []string{"./pkg/storage/store/sqlite", "./pkg/storage/store/sqlite/..."},
			To:    []string{"./pkg/storage/store/postgres", "./pkg/storage/store/postgres/..."},
		},
		{
			Name:  "store-backends-are-siblings-reverse",
			Why:   "The SQLite and Postgres backends share only store types and storerules. Importing each other would couple their SQL.",
			Scope: ProdAndTest,
			From:  []string{"./pkg/storage/store/postgres", "./pkg/storage/store/postgres/..."},
			To:    []string{"./pkg/storage/store/sqlite", "./pkg/storage/store/sqlite/..."},
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
				"./pkg/config", "./pkg/conversation/question", "./pkg/orchestration/scaleplan",
				"./pkg/orchestration/workflow", "./pkg/pack", "./pkg/pack/cue", "./pkg/pack/tmpl",
			},
			To: []string{"./pkg/transport/rpc", "./pkg/storage/store"},
		},
		{
			Name:  "rpc-imports-no-adapter",
			Why:   "rpc is the wire contract Desk, the CLI and the gateway share. If it imported one of them the contract would depend on its own clients.",
			Scope: ProdAndTest,
			From:  []string{"./pkg/transport/rpc", "./pkg/transport/rpc/..."},
			To:    []string{"./pkg/transport/cli", "./pkg/transport/desk", "./pkg/transport/gateway"},
		},
		{
			Name:  "desk-not-from-cli-or-gateway",
			Why:   "Desk is its own adapter over the admin API. The CLI hosts the `desk` command, so the dependency runs from the CLI to Desk and never back, and Desk does not borrow from the gateway.",
			Scope: ProdAndTest,
			From:  []string{"./pkg/transport/desk"},
			To:    []string{"./pkg/transport/cli", "./pkg/transport/gateway"},
		},
		{
			Name:  "gateway-not-from-cli-or-desk",
			Why:   "The A2A gateway is its own adapter. The CLI hosts the `gateway` command, so the dependency runs from the CLI to the gateway and never back, and the gateway does not borrow from Desk.",
			Scope: ProdAndTest,
			From:  []string{"./pkg/transport/gateway"},
			To:    []string{"./pkg/transport/cli", "./pkg/transport/desk"},
		},
		{
			Name:  "transport-reaches-no-runtime",
			Why:   "rpc, Desk and the gateway carry and render requests. They must not link the controller, a database driver, a store backend or the session host (JOB-174): an adapter that imports SQL or a process runtime to render a response has absorbed a subsystem.",
			Scope: Closure,
			From: []string{
				"./pkg/transport/rpc", "./pkg/transport/rpc/rpcschema",
				"./pkg/transport/desk", "./pkg/transport/gateway",
			},
			To: []string{
				"./pkg/orchestration/control", "./pkg/execution/session", "./pkg/orchestration/hooks",
				"./pkg/storage/store/sqlite", "./pkg/storage/store/postgres",
				"modernc.org/sqlite", "github.com/jackc/pgx/v5",
				"github.com/elee1766/acphost/pkg/acphost",
			},
		},
		{
			Name:  "hooks-not-from-rpc",
			Why:   "hooks owns durable hook execution. Its host adapter in control maps the wire envelopes, so hooks takes plain inputs and returns plain results. Reaching rpc also reached the ACP SDK and the sandbox types.",
			Scope: Closure,
			From:  []string{"./pkg/orchestration/hooks"},
			To:    []string{"./pkg/transport/rpc"},
		},
		{
			Name:  "provider-clients-are-siblings",
			Why:   "GitHub and Linear differ (REST and ETags against GraphQL, redirect policy, error types) and share only apicall and internal/transport. One importing the other would grow the lowest-common-denominator tracker the family is not meant to have (JOB-169).",
			Scope: ProdAndTest,
			From:  []string{"./pkg/integrations/github"},
			To:    []string{"./pkg/integrations/linear"},
		},
		{
			Name:  "provider-clients-are-siblings-reverse",
			Why:   "GitHub and Linear differ (REST and ETags against GraphQL, redirect policy, error types) and share only apicall and internal/transport. One importing the other would grow the lowest-common-denominator tracker the family is not meant to have (JOB-169).",
			Scope: ProdAndTest,
			From:  []string{"./pkg/integrations/linear"},
			To:    []string{"./pkg/integrations/github"},
		},
		{
			Name:  "identity-only-in-control",
			Why:   "Minting and caching credentials is the controller's job (it holds the keys and token files). Every other package takes an apicall.Source, so a client, a hook or a capability never reads a key or knows how a token is minted.",
			Scope: ProdAndTest,
			From:  []string{"./pkg/...", "./cmd/..."},
			Except: []string{
				"./pkg/orchestration/control",
				"./pkg/integrations/identity",
			},
			To: []string{"./pkg/integrations/identity"},
		},
	}
}

// onlyRules are allow-list rules for leaves that must stay small.
func onlyRules() []Only {
	return []Only{
		{
			Name:  "pack-is-independent",
			Why:   "pack assembles files into a merged document and reports provenance. It is handed the few facts it needs as data, so it imports no service configuration, store, automation, provider or process package. config decodes its output and owns validation (JOB-172).",
			From:  []string{"./pkg/pack"},
			Allow: []string{"./pkg/pack/cue", "./pkg/pack/tmpl"},
			Std:   true,
			Third: true,
		},
		{
			Name: "independent-leaves",
			Why:  "These packages are used on their own or by many others: clock, cron, logctx, redact and tmpl by most of the module, knowledge/memory by the CLI and the prompt rule without a database (JOB-170: it imports neither the store nor the learning and edit records in pkg/knowledge), pack/cue as the standalone CUE evaluation, ids and apicall as generator and token mechanics, runbook and flaky as development tools with one caller each. If one imports another package of the module, every consumer takes the cost. Third-party imports are fine.",
			From: []string{
				"./pkg/support/clock", "./pkg/config/cron", "./pkg/support/logctx", "./pkg/support/redact", "./pkg/pack/tmpl",
				"./pkg/knowledge/memory", "./pkg/pack/cue", "./pkg/ids", "./pkg/integrations/apicall",
				"./pkg/tooling/runbook", "./pkg/tooling/flaky",
			},
			Std:   true,
			Third: true,
		},
		{
			Name:  "telemetry-takes-only-logctx",
			Why:   "telemetry installs the slog bridge and exporters and needs the ids logctx carries (JOB-49). It is a leaf apart from that: the support packages are separate primitives with no umbrella API, so telemetry must not reach the store, config or any adapter it is meant to observe.",
			From:  []string{"./pkg/support/telemetry"},
			Allow: []string{"./pkg/support/logctx"},
			Std:   true,
			Third: true,
		},
		{
			Name: "provider-clients-stay-small",
			Why:  "The GitHub and Linear clients are used by automation, the CLI, ingest and the controller. Each takes a token source and sends requests: it imports apicall, the family's shared transport and telemetry, and no config, store or other subsystem.",
			From: []string{"./pkg/integrations/github", "./pkg/integrations/linear"},
			Allow: []string{
				"./pkg/integrations/apicall",
				"./pkg/integrations/internal/transport",
				"./pkg/support/telemetry",
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
