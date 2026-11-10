package activity

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gfx-labs/jobshop/pkg/transport/rpc"
)

const bodyMax = 2000

// Truncate bounds s to the stored body size.
func Truncate(s string) string {
	if len(s) <= bodyMax {
		return s
	}
	return strings.ToValidUTF8(s[:bodyMax], "") + " [truncated]"
}

func ghBot(login string) bool { return strings.HasSuffix(login, "[bot]") }

// parseRepoEvents maps GET /repos/{r}/events to items. Unknown event
// types are skipped.
func parseRepoEvents(repo string, body []byte) ([]rpc.ActivityItem, error) {
	var evs []struct {
		ID    string `json:"id"`
		Type  string `json:"type"`
		Actor struct {
			Login string `json:"login"`
		} `json:"actor"`
		CreatedAt time.Time       `json:"created_at"`
		Payload   json.RawMessage `json:"payload"`
	}
	if err := json.Unmarshal(body, &evs); err != nil {
		return nil, fmt.Errorf("events: %w", err)
	}
	web := "https://github.com/" + repo
	var out []rpc.ActivityItem
	for _, e := range evs {
		var p struct {
			Action  string `json:"action"`
			Number  int    `json:"number"`
			Ref     string `json:"ref"`
			RefType string `json:"ref_type"`
			Head    string `json:"head"`
			Before  string `json:"before"`
			Commits []struct {
				Message string `json:"message"`
			} `json:"commits"`
			PullRequest *struct {
				Number  int    `json:"number"`
				Title   string `json:"title"`
				HTMLURL string `json:"html_url"`
				Merged  bool   `json:"merged"`
			} `json:"pull_request"`
			Issue *struct {
				Number      int    `json:"number"`
				Title       string `json:"title"`
				HTMLURL     string `json:"html_url"`
				PullRequest any    `json:"pull_request"`
			} `json:"issue"`
			Comment *struct {
				Body    string `json:"body"`
				HTMLURL string `json:"html_url"`
			} `json:"comment"`
			Review *struct {
				State   string `json:"state"`
				Body    string `json:"body"`
				HTMLURL string `json:"html_url"`
			} `json:"review"`
			Release *struct {
				TagName string `json:"tag_name"`
				Name    string `json:"name"`
				HTMLURL string `json:"html_url"`
			} `json:"release"`
		}
		if err := json.Unmarshal(e.Payload, &p); err != nil {
			continue
		}
		it := rpc.ActivityItem{
			ID: "gh:" + repo + ":" + e.ID, Source: rpc.ActivityGitHub, Repo: repo,
			Actor: e.Actor.Login, Bot: ghBot(e.Actor.Login), At: e.CreatedAt,
		}
		prRef := func() {
			n := p.Number
			if p.PullRequest != nil && p.PullRequest.Number > 0 {
				n = p.PullRequest.Number
			}
			it.Ref, it.PR, it.URL = strconv.Itoa(n), true, fmt.Sprintf("%s/pull/%d", web, n)
			if p.PullRequest != nil && p.PullRequest.HTMLURL != "" {
				it.URL = p.PullRequest.HTMLURL
			}
		}
		prTitle := func() string {
			if p.PullRequest != nil && p.PullRequest.Title != "" {
				return ": " + p.PullRequest.Title
			}
			return ""
		}
		switch e.Type {
		case "PushEvent":
			branch := strings.TrimPrefix(p.Ref, "refs/heads/")
			it.Kind, it.Title = "push", "pushed to "+branch
			if len(p.Commits) > 0 {
				it.Title = fmt.Sprintf("pushed %d commit(s) to %s", len(p.Commits), branch)
				msgs := make([]string, 0, len(p.Commits))
				for _, cm := range p.Commits {
					msgs = append(msgs, cm.Message)
				}
				it.Body = Truncate(strings.Join(msgs, "\n"))
			}
			if p.Head != "" {
				it.URL = web + "/commit/" + p.Head
				if p.Before != "" && strings.Trim(p.Before, "0") != "" {
					it.URL = fmt.Sprintf("%s/compare/%.12s...%.12s", web, p.Before, p.Head)
				}
			}
		case "PullRequestEvent":
			prRef()
			action := p.Action
			if action == "merged" || (action == "closed" && p.PullRequest != nil && p.PullRequest.Merged) {
				it.Kind, action = "merge", "merged"
			} else {
				it.Kind = "pr"
			}
			it.Title = fmt.Sprintf("PR #%s %s%s", it.Ref, action, prTitle())
		case "PullRequestReviewEvent":
			prRef()
			it.Kind = "review"
			state := ""
			if p.Review != nil {
				state = strings.ToLower(strings.ReplaceAll(p.Review.State, "_", " "))
				it.Body = Truncate(p.Review.Body)
				if p.Review.HTMLURL != "" {
					it.URL = p.Review.HTMLURL
				}
			}
			it.Title = strings.TrimSpace(fmt.Sprintf("review %s on PR #%s%s", state, it.Ref, prTitle()))
		case "PullRequestReviewCommentEvent":
			prRef()
			it.Kind, it.Title = "review_comment", fmt.Sprintf("review comment on PR #%s%s", it.Ref, prTitle())
			if p.Comment != nil {
				it.Body, it.URL = Truncate(p.Comment.Body), p.Comment.HTMLURL
			}
		case "IssueCommentEvent", "IssuesEvent":
			if p.Issue == nil {
				continue
			}
			it.Ref, it.PR, it.URL = strconv.Itoa(p.Issue.Number), p.Issue.PullRequest != nil, p.Issue.HTMLURL
			what := "issue"
			if it.PR {
				what = "PR"
			}
			if e.Type == "IssuesEvent" {
				it.Kind, it.Title = "issue", fmt.Sprintf("%s #%d %s: %s", what, p.Issue.Number, p.Action, p.Issue.Title)
			} else {
				if p.Action != "created" {
					continue // edits and deletes are noise
				}
				it.Kind, it.Title = "comment", fmt.Sprintf("comment on %s #%d: %s", what, p.Issue.Number, p.Issue.Title)
				if p.Comment != nil {
					it.Body, it.URL = Truncate(p.Comment.Body), p.Comment.HTMLURL
				}
			}
		case "CreateEvent", "DeleteEvent":
			if p.RefType != "branch" && p.RefType != "tag" {
				continue
			}
			verb := "created"
			if e.Type == "DeleteEvent" {
				verb = "deleted"
			}
			it.Kind, it.Title = "branch", fmt.Sprintf("%s %s %s", verb, p.RefType, p.Ref)
			if verb == "created" {
				it.URL = web + "/tree/" + p.Ref
			}
		case "ReleaseEvent":
			if p.Release == nil {
				continue
			}
			it.Kind, it.Title, it.URL = "release", fmt.Sprintf("release %s %s", p.Release.TagName, p.Action), p.Release.HTMLURL
		default:
			continue
		}
		out = append(out, it)
	}
	return out, nil
}

// parseWorkflowRuns maps completed actions runs to ci items.
func parseWorkflowRuns(repo string, body []byte) ([]rpc.ActivityItem, error) {
	var r struct {
		WorkflowRuns []struct {
			ID           int64     `json:"id"`
			RunAttempt   int       `json:"run_attempt"`
			Name         string    `json:"name"`
			DisplayTitle string    `json:"display_title"`
			Conclusion   string    `json:"conclusion"`
			HeadBranch   string    `json:"head_branch"`
			HTMLURL      string    `json:"html_url"`
			UpdatedAt    time.Time `json:"updated_at"`
			Actor        struct {
				Login string `json:"login"`
			} `json:"actor"`
		} `json:"workflow_runs"`
	}
	if err := json.Unmarshal(body, &r); err != nil {
		return nil, fmt.Errorf("actions runs: %w", err)
	}
	out := make([]rpc.ActivityItem, 0, len(r.WorkflowRuns))
	for _, w := range r.WorkflowRuns {
		if w.Conclusion == "" {
			continue
		}
		out = append(out, rpc.ActivityItem{
			ID:     fmt.Sprintf("gh:%s:run:%d:%d", repo, w.ID, w.RunAttempt),
			Source: rpc.ActivityGitHub, Repo: repo, Kind: "ci",
			Actor: w.Actor.Login, Bot: ghBot(w.Actor.Login), At: w.UpdatedAt,
			Title: fmt.Sprintf("%s %s on %s: %s", w.Name, w.Conclusion, w.HeadBranch, w.DisplayTitle),
			URL:   w.HTMLURL,
		})
	}
	return out, nil
}
