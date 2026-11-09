package archcheck

// known lists the violations that exist today, each with the child that
// removes it. It is the measured baseline and only shrinks: an entry whose
// edge is gone fails the check as stale, and a new violation is not covered by
// an old entry because each entry names one source and one target. A package
// pattern is for a subtree that moves together, never for "everything".
func known() []Known {
	return []Known{}
}
