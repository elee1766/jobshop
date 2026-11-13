package session

import (
	"errors"
	"fmt"
	"strings"

	acp "github.com/coder/acp-go-sdk"
)

// requestError is the ACP request error in err and its message and data,
// lowercased. Agents report provider failures as generic errors, so callers match the
// text. re is nil when err is not an ACP request error.
func requestError(err error) (re *acp.RequestError, text string) {
	if !errors.As(err, &re) {
		return nil, ""
	}
	return re, strings.ToLower(fmt.Sprint(re.Message, " ", re.Data))
}

// codeResourceNotFound is ACP's "Resource not found". acp-go-sdk has no
// constant for it.
const codeResourceNotFound = -32002

// UnknownSession reports whether a prompt failed because the agent does
// not know the session id. jcode answers -32602 "Unknown ACP session id"
// when its process lost the session (seen 2026-11-05 after restarts,
// JOB-65). towa answers -32002 "Resource not found" with data
// {"sessionId": ...} (JOB-214). -32002 counts only with that key: an
// agent may use the code for another missing resource, and a new
// conversation would drop context for nothing. A host restart would only
// repeat either.
func UnknownSession(err error) bool {
	re, text := requestError(err)
	if re == nil {
		return false
	}
	switch re.Code {
	case -32602:
		return strings.Contains(text, "unknown acp session")
	case codeResourceNotFound:
		data, ok := re.Data.(map[string]any)
		if !ok {
			return false
		}
		_, ok = data["sessionId"]
		return ok
	}
	return false
}

// RateLimited reports whether a prompt error is the provider refusing
// work for now (account limit) rather than a broken agent. Agents report
// it as a generic ACP error, so match the message.
func RateLimited(err error) bool {
	re, text := requestError(err)
	if re == nil {
		return false
	}
	for _, k := range []string{"429", "rate_limit", "rate limit", "is unavailable"} {
		if strings.Contains(text, k) {
			return true
		}
	}
	return false
}

// ProviderUnreachable reports whether a prompt error is the agent failing
// to reach its provider (seen 2026-11-05 on acphost-dev: -32000 "Failed
// to send request ... Connect").
func ProviderUnreachable(err error) bool {
	re, text := requestError(err)
	return re != nil && (strings.Contains(text, "failed to send request") || strings.Contains(text, "error sending request"))
}

// AgentBusy reports whether the agent refused a prompt because it still
// considers an earlier one in flight. jcode does this after a prompt fails
// mid-stream (seen 2026-11-01); a session/cancel clears it, a host restart
// does not, so restarting only burns the crash budget.
func AgentBusy(err error) bool {
	re, text := requestError(err)
	return re != nil && strings.Contains(text, "already processing")
}
