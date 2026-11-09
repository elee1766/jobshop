package cli

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/elee1766/jobshop/pkg/transport/rpc"
)

// Outbound secret check (#574): jobshop gh and jobshop linear writes are
// scanned before anything is sent. The controller scans them (secret.scan)
// so the host's redact rules apply (#578); the CLI loads no config. A
// finding refuses the write with the rule and line, never the secret, and
// nothing is rewritten: the agent fixes its text. Only read when the
// command writes.

// allowSecretsFlag skips the check; only admin callers may pass it.
const allowSecretsFlag = "--allow-secret-findings"

// cutAllowSecrets removes allowSecretsFlag from args (before "--").
func cutAllowSecrets(args []string) (bool, []string) {
	var rest []string
	found := false
	for i, a := range args {
		if a == "--" {
			return found, append(rest, args[i:]...)
		}
		if a == allowSecretsFlag {
			found = true
			continue
		}
		rest = append(rest, a)
	}
	return found, rest
}

// outText is one piece of outbound text and where it came from, for the
// refusal message.
type outText struct {
	from string
	text string
}

// secretFindings asks the controller to scan texts and returns one line
// per finding, naming the rule and where; "" when clean.
func secretFindings(g *Globals, texts []outText) (string, error) {
	p := rpc.SecretScanParams{Texts: make([]rpc.SecretText, 0, len(texts))}
	for _, t := range texts {
		if t.text != "" {
			p.Texts = append(p.Texts, rpc.SecretText{From: t.from, Text: t.text})
		}
	}
	if len(p.Texts) == 0 {
		return "", nil
	}
	var r rpc.SecretScanResult
	if err := g.call(rpc.MethodSecretScan, p, &r); err != nil {
		return "", fmt.Errorf("secret scan: %w", err)
	}
	var b strings.Builder
	for _, f := range r.Findings {
		fmt.Fprintf(&b, "  %s: rule %s, line %d\n", f.From, f.Rule, f.Line)
	}
	return b.String(), nil
}

// allowSecrets checks that the caller is admin, for allowSecretsFlag.
func allowSecrets(g *Globals) error {
	var w rpc.Whoami
	if err := g.call(rpc.MethodWhoami, nil, &w); err != nil {
		return fmt.Errorf("%s: %w", allowSecretsFlag, err)
	}
	if !w.Admin {
		return fmt.Errorf("%s is for admin callers only", allowSecretsFlag)
	}
	return nil
}

// checkOut refuses a write whose texts hold a secret finding, unless
// allow is set by an admin caller. what names the command for the error.
func checkOut(g *Globals, what string, allow bool, texts ...outText) error {
	found, err := secretFindings(g, texts)
	if err != nil {
		return fmt.Errorf("%s: %w", what, err)
	}
	if found == "" {
		return nil
	}
	if !allow {
		return fmt.Errorf("%s: refused, the text looks like it holds a secret; remove it and retry (not sent):\n%s", what, strings.TrimRight(found, "\n"))
	}
	return allowSecrets(g)
}

// ghOutText is the text a gh write sends: -b/--body, -t/--title,
// -F/--body-file (pr, issue), -f/-F/--field/--raw-field values (api) and
// --input. "-" (stdin) is read from in and returned so it can be passed on.
func ghOutText(args []string, in io.Reader) ([]outText, []byte, error) {
	words, flags := ghWords(args)
	api := len(words) > 0 && words[0] == "api"
	var out []outText
	add := func(from string, vals []string) {
		for _, v := range vals {
			out = append(out, outText{from, v})
		}
	}
	add("--body", append(flags["-b"], flags["--body"]...))
	add("--title", append(flags["-t"], flags["--title"]...))
	var stdin []byte
	read := func(from, path string) error {
		var b []byte
		var err error
		if path == "-" {
			if stdin == nil {
				if stdin, err = io.ReadAll(in); err != nil {
					return err
				}
			}
			b = stdin
		} else if b, err = os.ReadFile(path); err != nil {
			return fmt.Errorf("%s %s: %w", from, path, err)
		}
		out = append(out, outText{from, string(b)})
		return nil
	}
	files := flags["--body-file"]
	fields := append(append([]string{}, flags["--field"]...), flags["--raw-field"]...)
	if api {
		// -f is --fill elsewhere, so ghWords does not take its value.
		fields = append(fields, apiShortFields(args)...)
	} else {
		files = append(files, flags["-F"]...)
	}
	for _, f := range files {
		if err := read("--body-file", f); err != nil {
			return nil, nil, err
		}
	}
	for _, f := range fields {
		_, v, _ := strings.Cut(f, "=")
		if p, ok := strings.CutPrefix(v, "@"); ok && api {
			if err := read("-F "+f[:len(f)-len(v)], p); err != nil {
				return nil, nil, err
			}
			continue
		}
		out = append(out, outText{"field " + strings.TrimSuffix(f[:len(f)-len(v)], "="), v})
	}
	for _, f := range flags["--input"] {
		if err := read("--input", f); err != nil {
			return nil, nil, err
		}
	}
	return out, stdin, nil
}

// apiShortFields is the values of -f and -F in gh api args ("-f k=v",
// "-fk=v").
func apiShortFields(args []string) []string {
	var out []string
	for i := 0; i < len(args); i++ {
		a := args[i]
		if a == "--" {
			break
		}
		if (a == "-f" || a == "-F") && i+1 < len(args) {
			out = append(out, args[i+1])
			i++
		} else if len(a) > 2 && (strings.HasPrefix(a, "-f") || strings.HasPrefix(a, "-F")) {
			out = append(out, a[2:])
		}
	}
	return out
}
