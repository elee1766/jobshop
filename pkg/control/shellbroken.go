package control

import (
	"fmt"
	"net"
	"strings"

	acp "github.com/coder/acp-go-sdk"
)

// A session whose host was respawned after a controller restart and whose
// jcode conversation was reloaded (session/load) could lose its tool
// shell: every bash call failed in 0ms with "Permission denied (os error
// 13)" while read tools still worked, and no later call recovered
// (JOB-94, gasserver 2026-11-06). Sessions that started a new
// conversation kept working. The controller cannot run a command in the
// agent's shell, so it watches the agent's tool updates instead: after
// shellDeniedLimit shell calls in a row fail that way it cancels the turn
// and respawns the host with a new conversation seeded from the journal,
// keeping the claims, as a restart does.

// shellDeniedLimit is how many shell calls in a row must fail to spawn
// before the shell counts as broken. One alone can be a real EACCES.
const shellDeniedLimit = 2

// shellDeniedText is how jcode reports a shell call it could not start.
const shellDeniedText = "(os error 13)"

// shellState is the runner's view of the agent's tool shell. Guarded by
// runner.mu.
type shellState struct {
	// exec marks the open calls of kind execute: a later update may omit
	// the kind.
	exec map[acp.ToolCallId]bool
	// denied counts shell calls in a row that failed to start.
	denied int
	// broken is set at shellDeniedLimit until the runner respawns.
	broken bool
	// respawned is set by a shell respawn and cleared by the next shell
	// call that completes: a repeat before then counts as a crash.
	respawned bool
	// newConversation makes the next host start a new conversation
	// instead of loading the old one.
	newConversation bool
}

// noteShellCall records a tool call start or update. It reports whether
// this update made the shell broken.
func (r *runner) noteShellCall(id acp.ToolCallId, kind *acp.ToolKind, st acp.ToolCallStatus, out any, content []acp.ToolCallContent) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	s := &r.shell
	if kind != nil {
		if *kind == acp.ToolKindExecute {
			if s.exec == nil {
				s.exec = map[acp.ToolCallId]bool{}
			}
			s.exec[id] = true
		} else {
			delete(s.exec, id)
		}
	}
	if !s.exec[id] {
		return false
	}
	switch st {
	case acp.ToolCallStatusCompleted:
		delete(s.exec, id)
		s.denied, s.respawned = 0, false
	case acp.ToolCallStatusFailed:
		delete(s.exec, id)
		if !spawnDenied(out, content) {
			s.denied = 0
			return false
		}
		s.denied++
		if s.denied >= shellDeniedLimit && !s.broken {
			s.broken = true
			return true
		}
	}
	return false
}

// spawnDenied reports whether a failed tool call's output is the agent
// failing to start the command, not the command's own output.
func spawnDenied(out any, content []acp.ToolCallContent) bool {
	var text string
	switch v := out.(type) {
	case map[string]any:
		text, _ = v["error"].(string)
		if text == "" {
			text, _ = v["output"].(string)
		}
	case string:
		text = v
	}
	if text == "" {
		for _, c := range content {
			if c.Content != nil && c.Content.Content.Text != nil {
				text = c.Content.Content.Text.Text
				break
			}
		}
	}
	text = strings.TrimSpace(text)
	// The whole output is the error: a command printing it among other
	// output is not a spawn failure.
	return strings.HasSuffix(text, "Permission denied "+shellDeniedText) && !strings.Contains(text, "\n")
}

// resetShellCalls forgets the calls and failures counted so far, once a
// host's conversation is set up: anything a load replayed is history.
// respawned and newConversation stay.
func (r *runner) resetShellCalls() {
	r.mu.Lock()
	r.shell.exec, r.shell.denied, r.shell.broken = nil, 0, false
	r.mu.Unlock()
	select {
	case <-r.shellDead:
	default:
	}
}

func (r *runner) shellBroken() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.shell.broken
}

// takeNewConversation reports and clears a pending new conversation.
func (r *runner) takeNewConversation() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	v := r.shell.newConversation
	r.shell.newConversation = false
	return v
}

// respawnShell stops the host after its shell broke. Like restart it keeps
// the claims and skips afterTurn; the new host starts a new conversation.
// A shell that broke again without a working call in between is a crash,
// so a host that cannot run commands at all ends crashed, not looping.
func (r *runner) respawnShell(nc net.Conn) error {
	r.mu.Lock()
	again := r.shell.respawned
	r.shell = shellState{respawned: true, newConversation: true}
	r.mu.Unlock()
	r.log.WarnContext(r.ctx, "agent shell cannot start commands, respawning host with a new conversation",
		"host_pid", hostPid(r.dir), "denied", shellDeniedLimit, "again", again)
	r.lifecycle("shell_denied")
	_ = nc.Close()
	r.c.stopHost(r.dir)
	if again {
		return fmt.Errorf("agent shell: %d shell calls in a row failed with Permission denied %s again after a respawn", shellDeniedLimit, shellDeniedText)
	}
	return errRefresh
}
