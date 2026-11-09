package session

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Background jobs (JOB-70). An agent that starts a long command in the
// background and then releases its ticket looks idle to the controller,
// and ending the session (scale down, idle timeout) kills the command
// with the host's process tree. Before ending an idle session the
// runner checks the agent for such jobs and keeps the session while
// they run, for at most one lease.

// helperComms are sandbox and network helpers that sit between the
// agent pid and the agent; they are not jobs.
var helperComms = map[string]bool{"bwrap": true, "pasta": true, "passt": true, "pasta.avx2": true}

// BackgroundJobs lists the command lines of agentPid's descendants that
// are not the agent itself: processes whose executable differs from
// agentPid's and is not a sandbox helper or the host binary. Between
// turns an agent runs no tools, so anything left is a job it started
// and did not wait for. Nil when agentPid is unknown or /proc is
// unreadable.
func BackgroundJobs(agentPid int, hostExe string) []string {
	if agentPid <= 0 {
		return nil
	}
	children := childMap()
	self := exeOf(agentPid)
	var out []string
	stack := []int{agentPid}
	for len(stack) > 0 {
		p := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		for _, c := range children[p] {
			stack = append(stack, c)
			if isAgentProc(c, self, hostExe) {
				continue
			}
			out = append(out, cmdlineOf(c))
		}
	}
	return out
}

// isAgentProc reports whether pid is part of the agent rather than a job:
// the agent's own executable (a wrapper and its child), the host binary
// (sandbox init) or a sandbox helper.
func isAgentProc(pid int, agentExe, hostExe string) bool {
	exe := exeOf(pid)
	if exe != "" && (exe == agentExe || (hostExe != "" && filepath.Base(exe) == filepath.Base(hostExe))) {
		return true
	}
	return helperComms[commOf(pid)]
}

// childMap reads every process's parent from /proc.
func childMap() map[int][]int {
	ents, err := os.ReadDir("/proc")
	if err != nil {
		return nil
	}
	m := map[int][]int{}
	for _, e := range ents {
		pid, err := strconv.Atoi(e.Name())
		if err != nil {
			continue
		}
		if pp := ppidOf(pid); pp > 0 {
			m[pp] = append(m[pp], pid)
		}
	}
	return m
}

func ppidOf(pid int) int {
	b, err := os.ReadFile("/proc/" + strconv.Itoa(pid) + "/stat")
	if err != nil {
		return 0
	}
	// The command name is in parentheses and may hold spaces.
	s := string(b)
	i := strings.LastIndexByte(s, ')')
	if i < 0 {
		return 0
	}
	f := strings.Fields(s[i+1:])
	if len(f) < 2 {
		return 0
	}
	pp, _ := strconv.Atoi(f[1])
	return pp
}

func exeOf(pid int) string {
	p, _ := os.Readlink("/proc/" + strconv.Itoa(pid) + "/exe")
	return strings.TrimSuffix(p, " (deleted)")
}

func commOf(pid int) string {
	b, _ := os.ReadFile("/proc/" + strconv.Itoa(pid) + "/comm")
	return strings.TrimSpace(string(b))
}

func cmdlineOf(pid int) string {
	b, _ := os.ReadFile("/proc/" + strconv.Itoa(pid) + "/cmdline")
	s := strings.TrimSpace(strings.ReplaceAll(string(b), "\x00", " "))
	if len(s) > 200 {
		s = s[:200]
	}
	if s == "" {
		s = commOf(pid)
	}
	return strconv.Itoa(pid) + " " + s
}
