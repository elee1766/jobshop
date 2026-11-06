package rpc

import (
	"encoding/json"
	"strings"
)

// toolTargetKeys are rawInput keys that name what a tool call acts on,
// most specific first. ACP leaves rawInput to the agent; these are the
// names common agents use for a command, a file and a search.
var toolTargetKeys = []string{"command", "cmd", "file_path", "path", "filePath", "url", "query", "pattern"}

// MaxToolTarget bounds ToolTarget and ToolIntent, so a long command
// does not widen a list row.
const MaxToolTarget = 200

// ToolTarget is a one-line label for what a tool call acts on: a
// command, file path or query from its raw input, else the first of its
// locations. Empty when neither says.
func ToolTarget(rawInput json.RawMessage, locations []string) string {
	var in map[string]any
	if json.Unmarshal(rawInput, &in) == nil {
		for _, k := range toolTargetKeys {
			if s, ok := in[k].(string); ok && strings.TrimSpace(s) != "" {
				return oneLine(s)
			}
		}
	}
	if len(locations) > 0 {
		return oneLine(locations[0])
	}
	return ""
}

// ToolIntent is the raw input's intent field: jcode tools carry a short
// reason for every call there. Other agents leave it empty.
func ToolIntent(rawInput json.RawMessage) string {
	var in struct {
		Intent string `json:"intent"`
	}
	if json.Unmarshal(rawInput, &in) != nil {
		return ""
	}
	return oneLine(in.Intent)
}

func oneLine(s string) string {
	s = strings.Join(strings.Fields(s), " ")
	if r := []rune(s); len(r) > MaxToolTarget {
		s = string(r[:MaxToolTarget]) + "…"
	}
	return s
}
