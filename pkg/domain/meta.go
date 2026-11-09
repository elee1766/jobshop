package domain

import "encoding/json"

// Readers for the meta packs on a ticket (see NormalizeMeta). They never
// fail: a ticket with no or malformed meta has no packs.

// MetaString is meta[pack][key] when it is a string, else "".
func MetaString(raw json.RawMessage, pack, key string) string {
	var m map[string]map[string]any
	if json.Unmarshal(raw, &m) != nil {
		return ""
	}
	s, _ := m[pack][key].(string)
	return s
}

// MetaMap is the packs of raw, an empty map when there are none. Writers
// edit one pack and marshal the map back, keeping the others as they were.
func MetaMap(raw json.RawMessage) map[string]json.RawMessage {
	m := map[string]json.RawMessage{}
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &m)
	}
	return m
}

// Names of the built-in meta packs the controller and its subsystem
// packages read by name (config.BuiltinMetaPacks lists all of them).
const (
	MetaRun        = "run"        // an escalation ticket's link to its run
	MetaPool       = "pool"       // a pool quarantine escalation
	MetaHook       = "hook"       // a guardrail approval
	MetaDesk       = "desk"       // Desk's attention chat, summary and context
	MetaQuestion   = "question"   // a question ticket
	MetaEscalation = "escalation" // an escalation ticket's chat
)
