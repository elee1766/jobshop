package ingest

import (
	"fmt"
	"strings"
	"time"
)

// LinearMaxAge is how far webhookTimestamp may be from now.
const LinearMaxAge = 60 * time.Second

// LinearEvent is the part of a Linear webhook payload jobshop reads.
type LinearEvent struct {
	Action         string `json:"action"`
	Type           string `json:"type"`
	OrganizationID string `json:"organizationId"`
	// WebhookTimestamp is Unix milliseconds.
	WebhookTimestamp int64 `json:"webhookTimestamp"`
	Actor            *struct {
		ID   string `json:"id"`
		Type string `json:"type"`
	} `json:"actor"`
	Data struct {
		Identifier string    `json:"identifier"`
		Number     int       `json:"number"`
		Team       *teamRef  `json:"team"`
		UserID     string    `json:"userId"` // Comment events
		BotActor   any       `json:"botActor"`
		Issue      *issueRef `json:"issue"`
	} `json:"data"`
}

type teamRef struct {
	Key string `json:"key"`
}

type issueRef struct {
	Identifier string   `json:"identifier"`
	Number     int      `json:"number"`
	Team       *teamRef `json:"team"`
}

// Fresh reports whether the event's timestamp is within LinearMaxAge
// of now, which bounds replay of a captured delivery.
func (e LinearEvent) Fresh(now time.Time) bool {
	age := now.Sub(time.UnixMilli(e.WebhookTimestamp))
	return age <= LinearMaxAge && age >= -LinearMaxAge
}

// Wanted is the filter that needs only the payload (#300): Issue and
// Comment events by a human. Apps, bots and integrations are skipped:
// their writes need no push. The organization and self checks need a
// Linear token and stay with the consumer.
func (e LinearEvent) Wanted() bool {
	if e.Type != "Issue" && e.Type != "Comment" {
		return false
	}
	return e.Actor != nil && strings.EqualFold(e.Actor.Type, "user") && e.Data.BotActor == nil
}

// IssueKey is the issue the event is about, "" when the payload does
// not name it.
func (e LinearEvent) IssueKey() string {
	key := func(ident string, n int, team *teamRef) string {
		if ident != "" {
			return ident
		}
		if team != nil && team.Key != "" && n > 0 {
			return fmt.Sprintf("%s-%d", team.Key, n)
		}
		return ""
	}
	switch e.Type {
	case "Issue":
		return key(e.Data.Identifier, e.Data.Number, e.Data.Team)
	case "Comment":
		if e.Data.Issue != nil {
			return key(e.Data.Issue.Identifier, e.Data.Issue.Number, e.Data.Issue.Team)
		}
	}
	return ""
}
