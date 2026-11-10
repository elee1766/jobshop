package command_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gfx-labs/jobshop/pkg/execution/automation"
)

func hook(t *testing.T, source string, options automation.Options) *automation.Engine {
	t.Helper()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "main.star"), []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	engine, err := automation.Load(t.Context(), root, []string{"main.star:handle"}, options)
	if err != nil {
		t.Fatal(err)
	}
	return engine
}

// TestProcessHelper is a real child process used to exercise shell boundaries.
func TestProcessHelper(_ *testing.T) {
	mode := os.Getenv("AUTOMATION_TEST_PROCESS")
	if mode == "" {
		return
	}
	switch mode {
	case "echo":
		cwd, err := os.Getwd()
		if err != nil {
			os.Exit(1)
		}
		_ = json.NewEncoder(os.Stdout).Encode(map[string]any{"cwd": cwd, "env": os.Getenv("HOOK_VALUE"), "args": os.Args[len(os.Args)-1]})
		os.Exit(7)
	case "large":
		fmt.Print(strings.Repeat("x", 100))
		os.Exit(0)
	case "wait":
		executable, err := os.Executable()
		if err != nil {
			os.Exit(2)
		}
		child := exec.Command(executable, "-test.run=^TestProcessHelper$")
		child.Env = append(os.Environ(), "AUTOMATION_TEST_PROCESS=child")
		if err := child.Run(); err != nil {
			os.Exit(2)
		}
		os.Exit(0)
	case "child":
		conn, err := net.Dial("tcp", os.Getenv("AUTOMATION_TEST_ADDRESS"))
		if err != nil {
			os.Exit(2)
		}
		_, _ = conn.Write([]byte("ready"))
		_, _ = io.Copy(io.Discard, conn)
		os.Exit(0)
	}
}

func TestShellArgvEnvironmentLimitsAndCancellation(t *testing.T) {
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	options := automation.Options{AllowShell: true, ShellDir: dir, ShellEnv: []string{"AUTOMATION_TEST_PROCESS=echo"}}
	engine := hook(t, `def handle(argv):
    result = shell(argv, env={"HOOK_VALUE": "literal"})
    return {"code": result.exit_code, "data": result.stdout}
`, options)
	argv := []string{executable, "-test.run=^TestProcessHelper$", "--", "$(false); untrusted"}
	result, err := engine.Invoke(t.Context(), "main.star:handle", argv)
	if err != nil {
		t.Fatal(err)
	}
	data := result.Output
	if data["code"] != int64(7) {
		t.Fatalf("wrong exit status %#v", data)
	}
	var output map[string]any
	if err := json.Unmarshal([]byte(data["data"].(string)), &output); err != nil {
		t.Fatal(err)
	}
	if output["args"] != "$(false); untrusted" || output["env"] != "literal" || output["cwd"] != dir {
		t.Fatalf("shell did not preserve explicit argv/env/cwd: %#v", output)
	}
	options.ShellEnv = []string{"AUTOMATION_TEST_PROCESS=large"}
	options.Limits.OutputBytes = 10
	large := hook(t, "def handle(argv):\n    return shell(argv).stdout\n", options)
	if _, err := large.Invoke(t.Context(), "main.star:handle", argv); err == nil || !strings.Contains(err.Error(), "output byte limit") {
		t.Fatalf("unbounded shell output: %v", err)
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = listener.Close() }()
	options.ShellEnv = []string{"AUTOMATION_TEST_PROCESS=wait", "AUTOMATION_TEST_ADDRESS=" + listener.Addr().String()}
	options.Limits.OutputBytes = 0
	waiting := hook(t, "def handle(argv):\n    return shell(argv).exit_code\n", options)
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	done := make(chan error, 1)
	go func() { _, err := waiting.Invoke(ctx, "main.star:handle", argv); done <- err }()
	connection, err := listener.Accept()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = connection.Close() }()
	ready := make([]byte, 5)
	if _, err := io.ReadFull(connection, ready); err != nil {
		t.Fatal(err)
	}
	cancel()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("shell cancellation succeeded")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("shell cancellation blocked")
	}
	if err := connection.SetReadDeadline(time.Now().Add(5 * time.Second)); err != nil {
		t.Fatal(err)
	}
	if _, err := connection.Read(make([]byte, 1)); !errors.Is(err, io.EOF) {
		t.Fatalf("shell descendant survived cancellation: %v", err)
	}
}
