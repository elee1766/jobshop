package threads

import (
	"fmt"
	"strings"
	"time"
)

// Text is the message for a delivery of cs on th.
func Text(th Thread, cs []Comment) string {
	var b strings.Builder
	switch {
	case th.Linear != "":
		fmt.Fprintf(&b, "New on Linear issue %s%s:\n", th.Linear, linearHeader(cs))
	case th.PR:
		fmt.Fprintf(&b, "New on GitHub PR %s:\n", th.Key())
	default:
		fmt.Fprintf(&b, "New on GitHub issue %s:\n", th.Key())
	}
	for _, cm := range cs {
		body := cm.Body
		if len(body) > commentBodyMax {
			body = strings.ToValidUTF8(body[:commentBodyMax], "") + " [truncated]"
		}
		kind := cm.Kind
		if cm.State != "" {
			kind += " (" + strings.ToLower(cm.State) + ")"
		}
		fmt.Fprintf(&b, "\n[%s] %s %s %s\n", cm.At.UTC().Format(time.RFC3339), cm.Author, kind, cm.URL)
		if cm.Reply != "" {
			b.WriteString(cm.Reply + "\n")
		}
		b.WriteString(body + "\n")
	}
	return b.String()
}

// linearHeader is ` "<title>" (<state>)` from the newest comment that
// carries the issue's title, "" when none does (an older inbox row).
func linearHeader(cs []Comment) string {
	for i := len(cs) - 1; i >= 0; i-- {
		if cs[i].Title == "" {
			continue
		}
		h := fmt.Sprintf(" %q", cs[i].Title)
		if cs[i].IssueState != "" {
			h += " (" + cs[i].IssueState + ")"
		}
		return h
	}
	return ""
}

// DeliveryAuthor is the ticket comment author for a delivery: the one
// author, or "several" when the batch mixes authors (#595). Each line
// of the text still names its own author.
func DeliveryAuthor(th Thread, cs []Comment) string {
	for _, cm := range cs[1:] {
		if commentWho(cm) != commentWho(cs[0]) {
			return th.AuthorPrefix() + severalAuthors
		}
	}
	if cs[0].Linear && cs[0].Handle != "" {
		return th.AuthorPrefix() + cs[0].Handle
	}
	return th.AuthorPrefix() + cs[0].Author
}

// commentWho identifies a comment's author: the Linear id, else the login.
func commentWho(cm Comment) string {
	if cm.AuthorID != "" {
		return "id:" + cm.AuthorID
	}
	return strings.ToLower(cm.Author)
}
