// Package vltest starts a real VictoriaLogs for tests (#530). The binary
// is $VICTORIALOGS_BIN or victoria-logs on PATH. Without one the test is
// skipped, unless JOBSHOP_REQUIRE_VICTORIALOGS is set (CI), where it fails.
package vltest

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// Server is one VictoriaLogs process.
type Server struct {
	URL  string
	addr string
	bin  string
	dir  string
	cmd  *exec.Cmd
	t    testing.TB
}

// Bin finds the binary or skips (fails in CI).
func Bin(t testing.TB) string {
	t.Helper()
	bin := os.Getenv("VICTORIALOGS_BIN")
	if bin == "" {
		var err error
		if bin, err = exec.LookPath("victoria-logs"); err != nil {
			bin = ""
		}
	}
	if bin == "" {
		if os.Getenv("JOBSHOP_REQUIRE_VICTORIALOGS") != "" {
			t.Fatal("victoria-logs not found: set VICTORIALOGS_BIN or put it on PATH")
		}
		t.Skip("victoria-logs not found (VICTORIALOGS_BIN or PATH)")
	}
	return bin
}

// Start runs VictoriaLogs on a free port with a temp data dir and stops
// it at cleanup.
func Start(t testing.TB) *Server {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := ln.Addr().String()
	_ = ln.Close()
	s := &Server{URL: "http://" + addr, addr: addr, bin: Bin(t), dir: filepath.Join(t.TempDir(), "vl"), t: t}
	s.Up()
	t.Cleanup(s.Down)
	return s
}

// Up starts the process (again, after Down) and waits for /health.
func (s *Server) Up() {
	s.t.Helper()
	if s.cmd != nil {
		return
	}
	cmd := exec.Command(s.bin, "-storageDataPath="+s.dir, "-httpListenAddr="+s.addr, "-loggerLevel=ERROR", "-retentionPeriod=100y")
	cmd.Stdout, cmd.Stderr = os.Stderr, os.Stderr
	if err := cmd.Start(); err != nil {
		s.t.Fatal(err)
	}
	s.cmd = cmd
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := waitHealthy(ctx, s.URL); err != nil {
		s.Down()
		s.t.Fatalf("victoria-logs did not come up: %v", err)
	}
}

// Down stops the process and keeps its data.
func (s *Server) Down() {
	if s.cmd == nil {
		return
	}
	_ = s.cmd.Process.Signal(os.Interrupt)
	done := make(chan struct{})
	go func() { _ = s.cmd.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		_ = s.cmd.Process.Kill()
		<-done
	}
	s.cmd = nil
}

func waitHealthy(ctx context.Context, url string) error {
	tick := time.NewTicker(50 * time.Millisecond)
	defer tick.Stop()
	for {
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url+"/health", nil)
		if resp, err := http.DefaultClient.Do(req); err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}
		select {
		case <-ctx.Done():
			return errors.Join(ctx.Err(), fmt.Errorf("no healthy answer from %s", url))
		case <-tick.C:
		}
	}
}
