// Package sqlstore is the backend-neutral store core: transactions,
// revision-checked mutate, history, claim rules and row mapping, run on
// sqlc's Querier. A backend supplies SQL files, migrations and a Driver
// (#404). sqlcdb is generated from pkg/store/sqlite's queries; a second
// engine generates the same method set and types.
package sqlstore
