package control

import (
	"context"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/gfx-labs/jobshop/pkg/storage/store/sqlite"
)

// Slow tick probe (JOB-212). One selfdev tick took 7.3s while every store
// call, even a 4ms reader-pool list, took about 90ms or a multiple of it.
// That pattern fits a stalled process (CPU starvation, cgroup throttling,
// GC) as well as a stalled database (WAL checkpoint, writer queue, a
// housekeeping step holding the writer), and the existing spans cannot say
// which. A tick at or over slowTickAfter now logs one "slow tick" record
// with the evidence for each candidate and a list of hints; the same
// attributes go on the tick span. Quiet ticks cost two cheap snapshots.

// slowTickAfter is the default tick length that gets a report. A quiet
// tick is under 0.1s (docs/verify/tick-latency.md).
const slowTickAfter = time.Second

// stallEvery is how often the stall monitor expects to run, and
// stallFloor how late a run must be to count as a stall of the process.
const (
	stallEvery = 20 * time.Millisecond
	stallFloor = 40 * time.Millisecond
)

// probeKeep bounds the stalls and housekeeping runs kept for later ticks.
const probeKeep = 256

type stall struct {
	at   time.Time // when the late run happened
	late time.Duration
}

type hkRun struct {
	name       string
	start, end time.Time // end is zero while the step runs
}

// probeSnap is what a tick compares at its start and end.
type probeSnap struct {
	at       time.Time
	numGC    uint32
	gcPause  time.Duration
	psi      map[string]uint64 // "cpu", "io", "memory": some total, microseconds
	throttle cpuThrottle
	dbMod    time.Time
	dbSize   int64
	walSize  int64
	pool     map[string]sqlite.PoolWait
}

// cpuThrottle is the cgroup cpu.stat throttling counters.
type cpuThrottle struct {
	periods, throttled uint64
	usec               uint64
}

// tickProbe collects the evidence. The zero value is not usable; build it
// with newTickProbe.
type tickProbe struct {
	slow time.Duration
	proc string // /proc, a test points it at a fixture
	// cgroupDir is the controller's cgroup directory, "" when unknown.
	cgroupDir string
	dbPath    string                            // "" for a store that is not a file
	pool      func() map[string]sqlite.PoolWait // nil for Postgres

	mu     sync.Mutex
	stalls []stall
	hk     []hkRun
}

func newTickProbe(slow time.Duration, dbPath string, pool func() map[string]sqlite.PoolWait) *tickProbe {
	if slow <= 0 {
		slow = slowTickAfter
	}
	p := &tickProbe{slow: slow, proc: "/proc", dbPath: dbPath, pool: pool}
	p.cgroupDir = cgroupDirOf(p.proc)
	return p
}

// cgroupDirOf returns the cgroup v2 directory of this process.
func cgroupDirOf(proc string) string {
	b, err := os.ReadFile(filepath.Join(proc, "self", "cgroup"))
	if err != nil {
		return ""
	}
	for line := range strings.SplitSeq(string(b), "\n") {
		if rest, ok := strings.CutPrefix(line, "0::"); ok {
			return filepath.Join("/sys/fs/cgroup", rest)
		}
	}
	return ""
}

// monitor records every time it ran late, until ctx ends. A late run means
// the process was not scheduled (CPU starvation, cgroup throttling, a GC
// stop-the-world, a stopped VM), whatever the store was doing.
func (p *tickProbe) monitor(ctx context.Context) {
	t := time.NewTicker(stallEvery)
	defer t.Stop()
	prev := time.Now()
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-t.C:
			if late := now.Sub(prev) - stallEvery; late >= stallFloor {
				p.addStall(now, late)
			}
			prev = now
		}
	}
}

func (p *tickProbe) addStall(at time.Time, late time.Duration) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.stalls = append(p.stalls, stall{at: at, late: late})
	if len(p.stalls) > probeKeep {
		p.stalls = slices.Clone(p.stalls[len(p.stalls)-probeKeep:])
	}
}

// houseStart notes that housekeeping step name began and returns the call
// that notes its end.
func (p *tickProbe) houseStart(name string, start time.Time) func(end time.Time) {
	p.mu.Lock()
	p.hk = append(p.hk, hkRun{name: name, start: start})
	i := len(p.hk) - 1
	p.mu.Unlock()
	return func(end time.Time) {
		p.mu.Lock()
		defer p.mu.Unlock()
		// The ring may have moved on; find the run again.
		for j := min(i, len(p.hk)-1); j >= 0; j-- {
			if p.hk[j].name == name && p.hk[j].start.Equal(start) {
				p.hk[j].end = end
				break
			}
		}
		if len(p.hk) > probeKeep {
			p.hk = slices.Clone(p.hk[len(p.hk)-probeKeep:])
		}
	}
}

func (p *tickProbe) snap() probeSnap {
	s := probeSnap{at: time.Now()}
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	s.numGC, s.gcPause = ms.NumGC, time.Duration(ms.PauseTotalNs)
	s.psi = map[string]uint64{}
	for _, k := range []string{"cpu", "io", "memory"} {
		if v, ok := readPSI(filepath.Join(p.proc, "pressure", k)); ok {
			s.psi[k] = v
		}
	}
	if p.cgroupDir != "" {
		s.throttle = readThrottle(filepath.Join(p.cgroupDir, "cpu.stat"))
	}
	if p.dbPath != "" {
		if fi, err := os.Stat(p.dbPath); err == nil {
			s.dbMod, s.dbSize = fi.ModTime(), fi.Size()
		}
		if fi, err := os.Stat(p.dbPath + "-wal"); err == nil {
			s.walSize = fi.Size()
		}
	}
	if p.pool != nil {
		s.pool = p.pool()
	}
	return s
}

// readPSI returns the "some" total of a pressure file, in microseconds.
func readPSI(path string) (uint64, bool) {
	b, err := os.ReadFile(path)
	if err != nil {
		return 0, false
	}
	for line := range strings.SplitSeq(string(b), "\n") {
		if !strings.HasPrefix(line, "some ") {
			continue
		}
		for f := range strings.FieldsSeq(line) {
			if v, ok := strings.CutPrefix(f, "total="); ok {
				n, err := strconv.ParseUint(v, 10, 64)
				return n, err == nil
			}
		}
	}
	return 0, false
}

func readThrottle(path string) cpuThrottle {
	var t cpuThrottle
	b, err := os.ReadFile(path)
	if err != nil {
		return t
	}
	for line := range strings.SplitSeq(string(b), "\n") {
		k, v, ok := strings.Cut(line, " ")
		if !ok {
			continue
		}
		n, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			continue
		}
		switch k {
		case "nr_periods":
			t.periods = n
		case "nr_throttled":
			t.throttled = n
		case "throttled_usec":
			t.usec = n
		}
	}
	return t
}

// tickReport is the evidence for one slow tick.
type tickReport struct {
	Seconds     float64
	SlowestStep string
	StepSeconds float64
	// GCCycles and GCPause are the Go collector's work during the tick.
	GCCycles uint32
	GCPause  time.Duration
	// Stalls are the monitor's late runs inside the tick: the process was
	// not running. StallMax is the longest, StallTotal the sum.
	Stalls     int
	StallMax   time.Duration
	StallTotal time.Duration
	// PSI is the time some task of the host waited for the resource during
	// the tick, per resource.
	PSI map[string]time.Duration
	// Throttled is how long the controller's cgroup was CPU-throttled.
	Throttled        time.Duration
	ThrottledPeriods uint64
	// DBWritten is set when the database file was written during the tick:
	// a WAL checkpoint copies pages into it. WALBytes is the WAL size at
	// the end, WALDelta its change.
	DBWritten bool
	DBGrew    int64
	WALBytes  int64
	WALDelta  int64
	// Pool is the connection wait during the tick, per pool.
	Pool map[string]sqlite.PoolWait
	// Housekeeping names the steps that ran during the tick.
	Housekeeping []string
	Hints        []string
}

// report diffs the two snapshots. steps is the tick's seconds per step.
func (p *tickProbe) report(a, b probeSnap, steps map[string]float64) tickReport {
	d := b.at.Sub(a.at)
	r := tickReport{
		GCCycles: b.numGC - a.numGC, GCPause: b.gcPause - a.gcPause,
		PSI:       map[string]time.Duration{},
		Throttled: time.Duration(b.throttle.usec-a.throttle.usec) * time.Microsecond, ThrottledPeriods: b.throttle.throttled - a.throttle.throttled,
		WALBytes: b.walSize, WALDelta: b.walSize - a.walSize,
		DBWritten: !b.dbMod.IsZero() && b.dbMod.After(a.dbMod),
		DBGrew:    b.dbSize - a.dbSize,
	}
	for _, name := range slices.Sorted(maps.Keys(steps)) {
		if steps[name] > r.StepSeconds {
			r.SlowestStep, r.StepSeconds = name, steps[name]
		}
	}
	for k, v := range b.psi {
		if w, ok := a.psi[k]; ok && v >= w {
			r.PSI[k] = time.Duration(v-w) * time.Microsecond
		}
	}
	if len(b.pool) > 0 {
		r.Pool = map[string]sqlite.PoolWait{}
		for k, v := range b.pool {
			w := a.pool[k]
			r.Pool[k] = sqlite.PoolWait{Count: v.Count - w.Count, Duration: v.Duration - w.Duration}
		}
	}
	p.mu.Lock()
	for _, s := range p.stalls {
		if !s.at.Before(a.at) && !s.at.After(b.at) {
			r.Stalls++
			r.StallTotal += s.late
			r.StallMax = max(r.StallMax, s.late)
		}
	}
	for _, h := range p.hk {
		// Ran at some point inside [a.at, b.at]; a step still running has no end.
		if h.start.After(b.at) || (!h.end.IsZero() && h.end.Before(a.at)) {
			continue
		}
		r.Housekeeping = append(r.Housekeeping, h.name)
	}
	p.mu.Unlock()
	r.Hints = hints(r, d)
	return r
}

// hints names each candidate cause the evidence supports, so a reader
// starts from them. None is proof: the numbers are in the record.
func hints(r tickReport, d time.Duration) []string {
	var h []string
	frac := func(x time.Duration, f float64) bool { return x > 0 && float64(x) >= f*float64(d) }
	if r.StallMax >= stallFloor {
		h = append(h, "process_stalled")
	}
	if frac(r.PSI["cpu"], 0.1) {
		h = append(h, "cpu_pressure")
	}
	if frac(r.Throttled, 0.05) {
		h = append(h, "cgroup_throttled")
	}
	if frac(r.PSI["io"], 0.1) {
		h = append(h, "io_pressure")
	}
	if frac(r.PSI["memory"], 0.05) {
		h = append(h, "memory_pressure")
	}
	if r.GCPause >= 50*time.Millisecond {
		h = append(h, "gc")
	}
	if r.DBWritten {
		h = append(h, "db_file_written")
	}
	if w, ok := r.Pool["writer"]; ok && frac(w.Duration, 0.2) {
		h = append(h, "writer_queue")
	}
	for _, n := range slices.Compact(slices.Sorted(slices.Values(r.Housekeeping))) {
		h = append(h, "housekeeping:"+n)
	}
	if len(h) == 0 {
		h = append(h, "none")
	}
	return h
}

// attrs flattens the report into log and span attributes.
func (r tickReport) attrs() []attribute.KeyValue {
	a := []attribute.KeyValue{
		attribute.String("jobshop.tick.slowest_step", r.SlowestStep),
		attribute.Float64("jobshop.tick.slowest_step_seconds", r.StepSeconds),
		attribute.Int64("jobshop.tick.gc_cycles", int64(r.GCCycles)),
		attribute.Float64("jobshop.tick.gc_pause_seconds", r.GCPause.Seconds()),
		attribute.Int("jobshop.tick.stalls", r.Stalls),
		attribute.Float64("jobshop.tick.stall_max_seconds", r.StallMax.Seconds()),
		attribute.Float64("jobshop.tick.stall_total_seconds", r.StallTotal.Seconds()),
		attribute.Float64("jobshop.tick.throttled_seconds", r.Throttled.Seconds()),
		attribute.Bool("jobshop.tick.db_written", r.DBWritten),
		attribute.Int64("jobshop.tick.wal_bytes", r.WALBytes),
		attribute.Int64("jobshop.tick.wal_delta_bytes", r.WALDelta),
		attribute.StringSlice("jobshop.tick.hints", r.Hints),
	}
	for _, k := range slices.Sorted(maps.Keys(r.PSI)) {
		a = append(a, attribute.Float64("jobshop.tick.psi_"+k+"_seconds", r.PSI[k].Seconds()))
	}
	for _, k := range slices.Sorted(maps.Keys(r.Pool)) {
		a = append(a,
			attribute.Int64("jobshop.tick.pool_"+k+"_waits", r.Pool[k].Count),
			attribute.Float64("jobshop.tick.pool_"+k+"_wait_seconds", r.Pool[k].Duration.Seconds()))
	}
	return a
}

// logArgs is attrs for slog.
func (r tickReport) logArgs() []any {
	var out []any
	for _, kv := range r.attrs() {
		out = append(out, strings.TrimPrefix(string(kv.Key), "jobshop.tick."), kv.Value.String())
	}
	return out
}

// beginTick snapshots before a tick; the result goes to endTick.
func (c *Controller) beginTick() probeSnap {
	if c.probe == nil {
		return probeSnap{}
	}
	return c.probe.snap()
}

// endTick logs the slow tick report when the tick took the probe's
// threshold or more. It runs before finishTick, still under tickMu.
func (c *Controller) endTick(ctx context.Context, span trace.Span, a probeSnap, took time.Duration) {
	if c.probe == nil || took < c.probe.slow {
		return
	}
	c.statusMu.Lock()
	steps := maps.Clone(c.tickSteps)
	c.statusMu.Unlock()
	r := c.probe.report(a, c.probe.snap(), steps)
	span.SetAttributes(append(r.attrs(), attribute.Bool("jobshop.tick.slow", true))...)
	args := append([]any{"seconds", fmt.Sprintf("%.3f", took.Seconds())}, r.logArgs()...)
	c.log.WarnContext(ctx, "slow tick", args...)
}
