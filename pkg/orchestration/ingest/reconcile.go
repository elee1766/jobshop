package ingest

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gfx-labs/jobshop/pkg/integrations/github"
)

// Reconcile asks GitHub to redeliver App webhook deliveries that failed
// (#335): a delivery the receiver never got (down, proxy error) is lost
// otherwise, since GitHub does not retry on its own. Only App hooks have
// a deliveries API reachable with the App's JWT, so this is off for
// github-token identities.
type Reconcile struct {
	// Client is the App's client, authenticated with its JWT.
	Client *github.Client
	// Max bounds redeliveries per pass; 0 is DefaultReconcileMax.
	Max int
}

// DefaultReconcileMax is the redelivery bound per pass.
const DefaultReconcileMax = 20

// HookDelivery is one entry of GET /app/hook/deliveries.
type HookDelivery struct {
	ID          int64     `json:"id"`
	GUID        string    `json:"guid"`
	DeliveredAt time.Time `json:"delivered_at"`
	Redelivery  bool      `json:"redelivery"`
	StatusCode  int       `json:"status_code"`
	Event       string    `json:"event"`
}

// ok reports whether the receiver accepted the delivery: any 2xx. The
// handler answers 202 or 204.
func (d HookDelivery) ok() bool { return d.StatusCode >= 200 && d.StatusCode < 300 }

// ReconcileResult counts one pass.
type ReconcileResult struct {
	Failed      int `json:"failed"`      // failed deliveries since the cutoff, not yet redelivered OK
	Redelivered int `json:"redelivered"` // redeliver requests accepted
}

// Pass lists deliveries newer than since (newest first, paginated) and
// asks for a redelivery of each failed GUID that has no later success,
// oldest first, at most Max. A GUID is redelivered once per pass.
func (r *Reconcile) Pass(ctx context.Context, since time.Time) (ReconcileResult, error) {
	var res ReconcileResult
	byGUID := map[string][]HookDelivery{}
	var order []string
	u := "/app/hook/deliveries?per_page=100"
	for u != "" {
		var page []HookDelivery
		next, err := r.do(ctx, http.MethodGet, u, &page)
		if err != nil {
			return res, err
		}
		done := false
		for _, d := range page {
			if !d.DeliveredAt.After(since) {
				done = true
				continue
			}
			if _, ok := byGUID[d.GUID]; !ok {
				order = append(order, d.GUID)
			}
			byGUID[d.GUID] = append(byGUID[d.GUID], d)
		}
		if done {
			break
		}
		u = next
	}
	limit := r.Max
	if limit <= 0 {
		limit = DefaultReconcileMax
	}
	// order is newest first; redeliver oldest first.
	for i := len(order) - 1; i >= 0; i-- {
		ds := byGUID[order[i]]
		failed := true
		var last HookDelivery
		for _, d := range ds {
			if d.ok() {
				failed = false
			}
			if d.DeliveredAt.After(last.DeliveredAt) || last.ID == 0 {
				last = d
			}
		}
		if !failed {
			continue
		}
		res.Failed++
		if res.Redelivered >= limit {
			continue
		}
		if _, err := r.do(ctx, http.MethodPost, fmt.Sprintf("/app/hook/deliveries/%d/attempts", last.ID), nil); err != nil {
			return res, fmt.Errorf("redeliver %s: %w", last.GUID, err)
		}
		res.Redelivered++
	}
	return res, nil
}

// do sends one request and decodes a JSON body into out (if non-nil).
// It returns the Link rel="next" URL.
func (r *Reconcile) do(ctx context.Context, method, u string, out any) (string, error) {
	resp, err := r.Client.Do(ctx, github.Request{Method: method, Path: u, Route: "/app/hook/deliveries", MaxBody: 8 << 20})
	if err != nil {
		return "", err
	}
	if out != nil {
		if err := json.Unmarshal(resp.Body, out); err != nil {
			return "", fmt.Errorf("%s %s: %w", method, u, err)
		}
	}
	return resp.Next(), nil
}
