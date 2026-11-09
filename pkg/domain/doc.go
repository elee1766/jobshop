// Package domain holds the types and invariants that more than one part of
// Jobshop shares: tickets and their typed fields, waits, learnings, resources
// and slots, inbox items, guard decisions and question options (JOB-166).
//
// It imports the standard library only. A consumer that needs a ticket or a
// wait takes it from here instead of from store (persistence), rpc (the wire
// protocol) or control, so it does not link their dependencies. Persisted and
// wire representations are unchanged: the JSON tags, the stored vocabulary and
// the error messages are the ones the types had before they moved, and
// domain_test.go pins them.
//
// What stays with its owner: the store interface, its query and update shapes
// and collections (store), request and result envelopes (rpc), configuration
// (config), provider values (github, linear) and session runtime types.
package domain
