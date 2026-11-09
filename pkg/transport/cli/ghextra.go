package cli

import (
	"encoding/json"
	"os"
	"slices"
	"strings"
)

// ghExtra is the guard event's extra for a jobshop gh write (JOB-113).
// detail only holds the body, so a hook could not see who a write asks
// for review, labels, assigns or what an api call sends. Argv is the
// whole command; the other fields are parsed from it for hooks that do
// not want to parse gh flags themselves.
type ghExtra struct {
	Argv    []string `json:"argv"`
	Command string   `json:"command"` // subcommand words, e.g. "pr edit"

	Title            string   `json:"title,omitempty"`
	ReviewersAdded   []string `json:"reviewers_added,omitempty"`
	ReviewersRemoved []string `json:"reviewers_removed,omitempty"`
	LabelsAdded      []string `json:"labels_added,omitempty"`
	LabelsRemoved    []string `json:"labels_removed,omitempty"`
	AssigneesAdded   []string `json:"assignees_added,omitempty"`
	AssigneesRemoved []string `json:"assignees_removed,omitempty"`

	// gh api only.
	Method string    `json:"method,omitempty"`
	Path   string    `json:"path,omitempty"`
	Fields []ghField `json:"fields,omitempty"`
	Input  string    `json:"input,omitempty"` // --input file contents
}

// ghField is one -f/-F/--raw-field/--field of gh api, in argument order.
// A typed field's @file or @- value is replaced by what it reads.
type ghField struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// buildGHExtra parses a write's args into ghExtra. stdin is what the
// secret check already read from stdin (nil if nothing asked for it).
func buildGHExtra(c ghCall, args []string, stdin []byte) json.RawMessage {
	words, flags := ghWords(args)
	e := ghExtra{Argv: args}
	if e.Argv == nil {
		e.Argv = []string{}
	}
	if len(words) > 0 && words[0] == "api" {
		e.Command = "api"
		e.Method = c.method
		e.Path = c.apiPath
		e.Fields = apiFields(args, stdin)
		if f, ok := flagValue(flags, "--input"); ok {
			e.Input = readArg(f, stdin)
		}
	} else {
		e.Command = strings.Join(words[:ghCmdLen(words)], " ")
		fillGHEdit(&e, words, flags, args)
	}
	b, err := json.Marshal(e)
	if err != nil {
		return nil
	}
	return b
}

// ghCmdLen is how many leading words of a ghWrites command are the
// subcommand, so positional arguments stay out of Command.
func ghCmdLen(words []string) int {
	for _, w := range ghWrites {
		if len(words) >= len(w.cmd) && slices.Equal(words[:len(w.cmd)], w.cmd) {
			return len(w.cmd)
		}
	}
	return len(words)
}

// fillGHEdit sets title, reviewers, labels and assignees for pr, issue
// and release create and edit. Short flags mean different things across
// gh subcommands (-r is --reviewer in pr create, --rebase in pr merge),
// so each one is read only where it is that flag.
func fillGHEdit(e *ghExtra, words []string, flags map[string][]string, args []string) {
	if len(words) < 2 {
		return
	}
	cmd := words[0] + " " + words[1]
	switch cmd {
	case "pr create", "pr edit", "issue create", "issue edit", "release create":
		e.Title, _ = flagValue(flags, "-t", "--title")
	}
	switch cmd {
	case "pr create":
		e.ReviewersAdded = splitList(flags["--reviewer"], shortValues(args, "-r"))
		fallthrough
	case "issue create":
		e.LabelsAdded = splitList(flags["--label"], flags["-l"])
		e.AssigneesAdded = splitList(flags["--assignee"], shortValues(args, "-a"))
	case "pr edit", "issue edit":
		e.ReviewersAdded = splitList(flags["--add-reviewer"])
		e.ReviewersRemoved = splitList(flags["--remove-reviewer"])
		e.LabelsAdded = splitList(flags["--add-label"])
		e.LabelsRemoved = splitList(flags["--remove-label"])
		e.AssigneesAdded = splitList(flags["--add-assignee"])
		e.AssigneesRemoved = splitList(flags["--remove-assignee"])
	}
}

// splitList flattens flag values the way gh reads them: each may hold a
// comma separated list.
func splitList(vals ...[]string) []string {
	var out []string
	for _, vs := range vals {
		for _, v := range vs {
			for p := range strings.SplitSeq(v, ",") {
				if p = strings.TrimSpace(p); p != "" {
					out = append(out, p)
				}
			}
		}
	}
	return out
}

// shortValues is the values of a short flag that ghWords leaves alone
// because it is a boolean in other subcommands: "-r v" and "-rv".
func shortValues(args []string, name string) []string {
	var out []string
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "--":
			return out
		case a == name && i+1 < len(args):
			out = append(out, args[i+1])
			i++
		case len(a) > 2 && strings.HasPrefix(a, name):
			out = append(out, a[2:])
		}
	}
	return out
}

// apiWords is the positional words of gh api args. ghWords keeps the
// value of -f and -F as a word (they are booleans elsewhere), so they
// are dropped first: `gh api -f a=b repos/o/n/x` has path repos/o/n/x.
func apiWords(args []string) []string {
	var rest []string
	for i := 0; i < len(args); i++ {
		a := args[i]
		if a == "--" {
			rest = append(rest, args[i:]...)
			break
		}
		if a == "-f" || a == "-F" {
			i++
			continue
		}
		rest = append(rest, a)
	}
	words, _ := ghWords(rest)
	return words
}

// apiFields is the request fields of gh api args in order. -F/--field
// values "@path" and "@-" are read like gh does.
func apiFields(args []string, stdin []byte) []ghField {
	var out []ghField
	for i := 0; i < len(args); i++ {
		a := args[i]
		if a == "--" {
			break
		}
		var name, val string
		switch {
		case a == "-f" || a == "-F" || a == "--field" || a == "--raw-field":
			if i+1 >= len(args) {
				continue
			}
			name, val = a, args[i+1]
			i++
		case strings.HasPrefix(a, "--field=") || strings.HasPrefix(a, "--raw-field="):
			name, val, _ = strings.Cut(a, "=")
		case len(a) > 2 && (strings.HasPrefix(a, "-f") || strings.HasPrefix(a, "-F")) && a[1] != '-':
			name, val = a[:2], a[2:]
		default:
			continue
		}
		k, v, _ := strings.Cut(val, "=")
		if p, ok := strings.CutPrefix(v, "@"); ok && (name == "-F" || name == "--field") {
			v = readArg(p, stdin)
		}
		out = append(out, ghField{Key: k, Value: v})
	}
	return out
}

// readArg reads a file argument; "-" is stdin. An unreadable file gives
// "": the secret check already refused the write if it could not read it.
func readArg(path string, stdin []byte) string {
	if path == "-" {
		return string(stdin)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(b)
}
