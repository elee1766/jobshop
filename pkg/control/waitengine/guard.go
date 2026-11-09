package waitengine

import (
	"fmt"
	"regexp"

	"github.com/elee1766/jobshop/pkg/domain"
)

// eventRe matches a release --until reason that names an event a wait
// kind covers (JOB-11). Such a timer guesses: the step sits idle after the
// event, or wakes early and burns an attempt.
var eventRe = regexp.MustCompile(`(?i)\b(ci|checks?|pr|pull request|deploy\w*|tags?|merge\w*)\b`)

// TimerHint is the error for a timed release whose reason is an event,
// naming the wait add line to use instead. It is nil when the reason
// names no event.
func TimerHint(reason string) error {
	m := eventRe.FindString(reason)
	if m == "" {
		return nil
	}
	return fmt.Errorf("%w: release --until for %q waits on an event, not a time. Use a wait instead:\n"+
		"  CI:     jobshop wait add 'github.checks:pr=<o/r>#<n>,state=success' 'github.checks:pr=<o/r>#<n>,state=failure' --any --resume -r %q\n"+
		"  master: jobshop wait add 'github.checks:repo=<o/r>,sha=<sha>,state=success' 'github.checks:repo=<o/r>,sha=<sha>,state=failure' --any --resume -r %q\n"+
		"  merge:  jobshop wait add 'github.pr:pr=<o/r>#<n>,state=merged' --resume -r %q\n"+
		"  deploy: jobshop wait add 'deploy:repo=<o/r>,sha=<sha>' --resume -r %q\n"+
		"  tag:    jobshop wait add 'github.tag:repo=<o/r>,tag=<tag>' --resume -r %q\n"+
		"A real timer (a window, a rate limit backoff): add --time-wait", domain.ErrInvalid, m, reason, reason, reason, reason, reason)
}
