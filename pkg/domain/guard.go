package domain

// Guard decisions (#188): what a guardrail hook answers before an outside
// action such as push, pr-create, merge, comment or message. The wire
// envelopes that carry them (rpc.GuardParams, rpc.GuardResult) stay in rpc.
const (
	GuardAllow    = "allow"
	GuardDeny     = "deny"
	GuardAskHuman = "ask_human"
)
