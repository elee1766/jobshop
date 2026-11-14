package control

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace/noop"

	"github.com/gfx-labs/jobshop/pkg/storage/store/sqlite"
)

const quietYAML = `projects:
  js: {}
agents:
  fake: {command: /bin/false}
pools:
  dev: {agent: fake, dir: %[1]q}
`

// writeFile writes name under dir, replacing it.
func writeFile(t *testing.T, dir, name, body string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Dir(filepath.Join(dir, name)), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(body), 0o600))
}

func psi(total int) string {
	return "some avg10=0.00 avg60=0.00 avg300=0.00 total=" + strconv.Itoa(total) + "\nfull avg10=0.00 avg60=0.00 avg300=0.00 total=0\n"
}

// TestTickProbeReport (JOB-212): the evidence for a slow tick comes from
// the difference of two snapshots, and each candidate cause sets its hint.
func TestTickProbeReport(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	p := newTickProbe(time.Second, filepath.Join(dir, "db"), nil)
	p.proc = filepath.Join(dir, "proc")
	p.cgroupDir = filepath.Join(dir, "cg")
	writeFile(t, dir, "db", "x")
	writeFile(t, dir, "db-wal", "xx")
	writeFile(t, dir, "proc/pressure/cpu", psi(1_000_000))
	writeFile(t, dir, "proc/pressure/io", psi(0))
	writeFile(t, dir, "cg/cpu.stat", "nr_periods 10\nnr_throttled 1\nthrottled_usec 5000\n")
	p.pool = func() map[string]sqlite.PoolWait {
		return map[string]sqlite.PoolWait{"writer": {Count: 1, Duration: time.Second}}
	}

	a := p.snap()
	// Quiet: nothing moved, so no hint.
	quiet := p.report(a, a, map[string]float64{"sweep": 0.01})
	assert.Equal(t, []string{"none"}, quiet.Hints)
	assert.Equal(t, "sweep", quiet.SlowestStep)

	// Evidence of a stalled process and a busy writer inside the tick.
	writeFile(t, dir, "proc/pressure/cpu", psi(1_000_000+3_000_000))
	writeFile(t, dir, "cg/cpu.stat", "nr_periods 40\nnr_throttled 20\nthrottled_usec 2005000\n")
	writeFile(t, dir, "db-wal", "xxxxxx")
	p.pool = func() map[string]sqlite.PoolWait {
		return map[string]sqlite.PoolWait{"writer": {Count: 4, Duration: 3 * time.Second}}
	}
	b := p.snap()
	b.at = a.at.Add(7 * time.Second)
	b.dbMod = a.dbMod.Add(time.Second)
	p.addStall(a.at.Add(time.Second), 90*time.Millisecond)
	p.addStall(a.at.Add(-time.Hour), 900*time.Millisecond) // before the tick: not counted
	end := p.houseStart("maintain_store", a.at.Add(time.Second))
	end(a.at.Add(3 * time.Second))
	p.houseStart("compress_journals", a.at.Add(-time.Hour))(a.at.Add(-time.Hour + time.Second)) // earlier

	r := p.report(a, b, map[string]float64{"sweep": 0.01, "supervise": 7.1})
	assert.Equal(t, "supervise", r.SlowestStep)
	assert.Equal(t, 1, r.Stalls)
	assert.Equal(t, 90*time.Millisecond, r.StallMax)
	assert.Equal(t, 3*time.Second, r.PSI["cpu"])
	assert.Equal(t, 2*time.Second, r.Throttled)
	assert.Equal(t, uint64(19), r.ThrottledPeriods)
	assert.Equal(t, int64(4), r.WALDelta)
	assert.True(t, r.DBWritten)
	assert.Equal(t, int64(3), r.Pool["writer"].Count)
	assert.Equal(t, 2*time.Second, r.Pool["writer"].Duration)
	assert.Equal(t, []string{"maintain_store"}, r.Housekeeping)
	assert.ElementsMatch(t, []string{
		"process_stalled", "cpu_pressure", "cgroup_throttled", "db_file_written", "writer_queue", "housekeeping:maintain_store",
	}, r.Hints)
}

// TestSlowTickLogged (JOB-212): a pass at or over the threshold logs one
// "slow tick" record with the hints, and a faster one logs nothing.
func TestSlowTickLogged(t *testing.T) {
	t.Parallel()
	var log lockedBuf
	c, _, dir := offTickController(t, quietYAML, &log)
	c.tracer = noop.NewTracerProvider().Tracer("test")
	c.probe = newTickProbe(time.Hour, filepath.Join(dir, "db"), nil)
	c.tickSteps = map[string]float64{}
	c.tick()
	assert.NotContains(t, log.String(), "slow tick")

	c.probe.slow = time.Nanosecond
	c.tick()
	out := log.String()
	assert.Contains(t, out, "slow tick")
	assert.Contains(t, out, "hints=")
	assert.Contains(t, out, "slowest_step=")
}
