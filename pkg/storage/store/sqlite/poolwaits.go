package sqlite

import "time"

// PoolWait is how often, and for how long in total, callers waited for a
// free connection since the store opened.
type PoolWait struct {
	Count    int64
	Duration time.Duration
}

// PoolWaits returns the connection waits of the writer pool ("writer") and
// the reader pool ("reader"). A writer wait is a call queued behind another
// write; the controller diffs it across a slow tick (JOB-212) to tell a
// busy writer from a stalled process.
func (s *Store) PoolWaits() map[string]PoolWait {
	w := s.db.Stats()
	out := map[string]PoolWait{"writer": {Count: w.WaitCount, Duration: w.WaitDuration}}
	if s.rdb != s.db {
		r := s.rdb.Stats()
		out["reader"] = PoolWait{Count: r.WaitCount, Duration: r.WaitDuration}
	}
	return out
}
