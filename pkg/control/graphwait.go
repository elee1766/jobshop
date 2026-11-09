package control

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/storage/store"
)

// Wait edges in the ticket graph (#1091). A parked ticket shows an edge to
// what it waits on: another ticket or run, or an external node (a Linear
// issue, a PR, a tag, a deploy) that links to its canonical page. The live
// wait gives state "waiting" or "met" per condition. A wait that ended is
// gone from fields, so its edges come from the history (the fields diffs
// that set and removed it) and the "Wait (...) ended: ..." comment the
// controller wrote with the removal.

// Wait edge states.
const (
	waitStateWaiting     = "waiting"
	waitStateMet         = "met"
	waitStateInvalidated = "invalidated"
	waitStateCleared     = "cleared"
	waitStateEnded       = "ended"
)

// maxEndedWaits bounds the finished waits shown per ticket: the newest.
const maxEndedWaits = 10

// graphWait is one wait condition of a ticket: the relation, and the
// external node it points at (nil when it points at a ticket).
type graphWait struct {
	rel rpc.GraphRelation
	ext *rpc.GraphExternal
}

// waitEpisode is one wait the ticket had: the live one has end zero.
type waitEpisode struct {
	wait domain.Wait
	end  time.Time
}

// graphWaits lists the wait edges of t, oldest wait first, so a later wait
// on the same target replaces an earlier one when the caller dedupes.
func graphWaits(cfg *config.Config, t domain.Ticket, th store.Thread) []graphWait {
	var out []graphWait
	for _, ep := range waitEpisodes(t, th) {
		state := func(cd domain.WaitCond) string {
			if ep.end.IsZero() {
				if cd.Met {
					return waitStateMet
				}
				return waitStateWaiting
			}
			return endedState(ep.wait, cd, endReason(ep, th.Comments))
		}
		for _, cd := range ep.wait.Conds {
			target, ext := waitTarget(cfg, t.Project, cd)
			if target == "" {
				continue
			}
			out = append(out, graphWait{
				rel: rpc.GraphRelation{
					From: t.ID, To: target, Kind: "wait", Type: cd.Type,
					Reason: ep.wait.Reason, State: state(cd), Since: ep.wait.Since,
				},
				ext: ext,
			})
		}
	}
	return out
}

// waitEpisodes is the ticket's finished waits (newest maxEndedWaits) then
// its live one.
func waitEpisodes(t domain.Ticket, th store.Thread) []waitEpisode {
	var out []waitEpisode
	var cur *domain.Wait
	for _, h := range th.History {
		if h.Op != "update" || len(h.Diff) == 0 {
			continue
		}
		var d struct {
			Fields *struct {
				Wait *domain.Wait `json:"wait"`
			} `json:"fields"`
		}
		if json.Unmarshal(h.Diff, &d) != nil || d.Fields == nil {
			continue
		}
		next := d.Fields.Wait
		if cur != nil && (next == nil || !next.Since.Equal(cur.Since)) {
			out = append(out, waitEpisode{wait: *cur, end: h.At})
		}
		cur = next
	}
	if len(out) > maxEndedWaits {
		out = out[len(out)-maxEndedWaits:]
	}
	if live := waitOf(t); live != nil {
		out = append(out, waitEpisode{wait: *live})
	}
	return out
}

// endReason is the text after "ended: " in the comment that closed ep.
func endReason(ep waitEpisode, comments []domain.Comment) string {
	prefix := "Wait (" + ep.wait.Reason + ") ended: "
	for _, cm := range comments {
		if cm.CreatedAt.Before(ep.end) || !strings.HasPrefix(cm.Body, prefix) {
			continue
		}
		why, _, _ := strings.Cut(strings.TrimPrefix(cm.Body, prefix), "\n\n")
		return why
	}
	return ""
}

// endedState is how a condition stood when its wait ended. The final
// evaluation that met the wait is not saved, so with mode any and several
// conditions, one not saved as met may or may not have held: "ended".
func endedState(w domain.Wait, cd domain.WaitCond, why string) string {
	switch {
	case strings.HasPrefix(why, "invalidated"):
		return waitStateInvalidated
	case strings.HasPrefix(why, "cleared"):
		return waitStateCleared
	case strings.HasPrefix(why, "condition met"):
		if cd.Met || w.Mode != domain.WaitAny || len(w.Conds) == 1 {
			return waitStateMet
		}
	}
	return waitStateEnded
}

// waitTarget names what a condition waits on: a ticket id (ext nil) or an
// external node's id. Empty for a condition with nothing to point at, such
// as a time.
func waitTarget(cfg *config.Config, project string, cd domain.WaitCond) (string, *rpc.GraphExternal) {
	ext := func(label, link string) (string, *rpc.GraphExternal) {
		e := rpc.GraphExternal{ID: "ext:" + cd.Type + ":" + label, Type: cd.Type, Label: label, URL: link}
		return e.ID, &e
	}
	pr := fmt.Sprintf("%s PR #%d", cd.Repo, cd.PR)
	prURL := fmt.Sprintf("https://github.com/%s/pull/%d", cd.Repo, cd.PR)
	switch cd.Type {
	case domain.WaitRun:
		return cd.Run, nil
	case domain.WaitLinearState:
		return ext("Linear "+cd.Thread+": "+cd.State, linearIssueURL(cfg, project, cd.Thread, ""))
	case domain.WaitLinearLabel:
		label := "Linear " + cd.Thread + " label " + cd.Label
		if cd.Absent {
			label = "Linear " + cd.Thread + " without label " + cd.Label
		}
		return ext(label, linearIssueURL(cfg, project, cd.Thread, ""))
	case domain.WaitPR:
		return ext(pr+": "+cd.State, prURL)
	case domain.WaitChecks:
		label := pr + " checks: " + cd.State
		link := prURL
		if cd.PR == 0 {
			short := cd.SHA[:min(7, len(cd.SHA))]
			label = fmt.Sprintf("%s@%s checks: %s", cd.Repo, short, cd.State)
			link = fmt.Sprintf("https://github.com/%s/commit/%s", cd.Repo, cd.SHA)
		}
		if cd.Workflow != "" {
			label += " (" + cd.Workflow + ")"
		}
		return ext(label, link)
	case domain.WaitReview:
		return ext(pr+" review by "+cd.Login, prURL)
	case domain.WaitLabel:
		return ext(fmt.Sprintf("%s #%d label %s", cd.Repo, cd.Issue, cd.Label), fmt.Sprintf("https://github.com/%s/issues/%d", cd.Repo, cd.Issue))
	case domain.WaitTag:
		return ext(cd.Repo+" tag "+cd.Tag, "https://github.com/"+cd.Repo+"/releases/tag/"+url.PathEscape(cd.Tag))
	case domain.WaitDeploy:
		return ext(fmt.Sprintf("deploy %s@%s", cd.Repo, cd.SHA[:min(7, len(cd.SHA))]), "https://github.com/"+cd.Repo+"/commit/"+cd.SHA)
	case domain.WaitReply:
		return ext("reply on "+cd.Thread, threadURL(cfg, project, cd.Thread))
	}
	return "", nil
}

// threadURL is the page of a thread.reply thread: owner/name#n or a Linear key.
func threadURL(cfg *config.Config, project, thread string) string {
	if repo, n, ok := strings.Cut(thread, "#"); ok {
		return "https://github.com/" + repo + "/issues/" + n
	}
	return linearIssueURL(cfg, project, thread, "")
}

// linearIssueURL is the Linear page of an issue key. A workspace the caller
// does not know comes from the project's configured identities, and only
// when one workspace covers the key's team. Empty when there is none.
func linearIssueURL(cfg *config.Config, project, key, workspace string) string {
	if workspace == "" {
		team, _, _ := strings.Cut(key, "-")
		matches := map[string]bool{}
		for _, name := range cfg.Projects[project].Identities {
			identity := cfg.Identities[name]
			if config.IsLinear(identity.Kind) && identity.CoversTeam(team) {
				matches[identity.Workspace] = true
			}
		}
		if len(matches) == 1 {
			for match := range matches {
				workspace = match
			}
		}
	}
	if workspace == "" {
		return ""
	}
	return "https://linear.app/" + url.PathEscape(workspace) + "/issue/" + url.PathEscape(key)
}
