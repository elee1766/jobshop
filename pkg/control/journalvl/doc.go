// Package journalvl tests journal shipping and search against a real
// VictoriaLogs (#530) through the controller's RPC API, using
// controltest (JOB-47). Every test file carries the victorialogs build
// tag; CI's victorialogs job runs them.
package journalvl
