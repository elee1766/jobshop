package ingest

// SetInboxPruneBatch sets the prune batch size until the returned func
// restores it. Tests that change it must not run in parallel with other
// prunes.
func SetInboxPruneBatch(n int) (restore func()) {
	old := inboxPruneBatch
	inboxPruneBatch = n
	return func() { inboxPruneBatch = old }
}
