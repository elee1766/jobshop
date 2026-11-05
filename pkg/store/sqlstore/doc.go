// Package sqlstore is the backend-neutral store core: transactions,
// revision-checked mutate, history, claim rules and row mapping, run on
// sqldb's Querier. A backend supplies migrations and a Driver (#404).
// sqldb's SQL is written for SQLite; Postgres rewrites it as it runs.
package sqlstore
