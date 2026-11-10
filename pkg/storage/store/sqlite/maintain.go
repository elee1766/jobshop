package sqlite

import "context"

// optimizeSQL refreshes planner statistics. 0x10000 checks every table,
// not only those this connection queried (a fresh connection has queried
// none), and 0x02 runs the ANALYZE it decides is due. Without statistics
// SQLite takes records_created to satisfy ORDER BY created_at and scans a
// whole collection to find no row (JOB-191). After the first run it
// analyzes only what changed a lot, so repeating it is cheap.
const optimizeSQL = "PRAGMA optimize=0x10002"

// Maintain refreshes the planner statistics of the database. Open runs it
// once; the controller repeats it on a slow interval so a collection that
// grew since is analyzed too.
func (s *Store) Maintain(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, optimizeSQL)
	return err
}
