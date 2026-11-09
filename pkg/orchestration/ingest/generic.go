package ingest

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/domain"
)

// SourceGeneric is any other HMAC-signed JSON sender (#399).
const SourceGeneric = "generic"

// MapHeader carries a generic endpoint's map with each delivery, so the
// drain maps it the way the endpoint was configured when it arrived and
// needs no endpoint config of its own.
const MapHeader = "Jobshop-Map"

// GenericDelivery builds the spool delivery for a generic payload on ep.
// id is the sender's delivery id ("" uses the map's delivery_id path,
// else the body hash); kind overrides the mapped kind.
func GenericDelivery(ep config.IngestEndpoint, id, kind string, body []byte, now time.Time, hdr map[string]string) (Delivery, error) {
	var p any
	if err := json.Unmarshal(body, &p); err != nil {
		return Delivery{}, fmt.Errorf("%w: payload is not JSON: %w", ErrBadDelivery, err)
	}
	if id == "" && ep.Map.DeliveryID != "" {
		id = lookup(p, ep.Map.DeliveryID)
	}
	if kind == "" {
		kind = lookup(p, ep.Map.Kind)
	}
	m, err := json.Marshal(ep.Map)
	if err != nil {
		return Delivery{}, err
	}
	h := map[string]string{MapHeader: string(m)}
	for k, v := range hdr {
		h[k] = v
	}
	return Delivery{
		Endpoint: ep.Name, Source: SourceGeneric, DeliveryID: deliveryID(id, body),
		Kind: kind, ReceivedAt: now.UTC(), Headers: h, Payload: body,
	}, nil
}

func (h *Handler) serveGeneric(w http.ResponseWriter, r *http.Request, ep config.IngestEndpoint, body []byte) {
	var id string
	if ep.DeliveryIDHeader != "" {
		id = r.Header.Get(ep.DeliveryIDHeader)
	}
	var names []string
	for _, n := range []string{ep.SignatureHeader, ep.DeliveryIDHeader} {
		if n != "" {
			names = append(names, n)
		}
	}
	d, err := GenericDelivery(ep, id, "", body, h.now(), pick(r, names...))
	if err != nil {
		http.Error(w, "bad payload", http.StatusBadRequest)
		return
	}
	h.write(w, r, d)
}

// Put writes one generic delivery straight to the spool (`jobshop ingest
// put`, #399): for trusted local senders, no HTTP or signature. It
// reports whether the delivery was new; a duplicate is not an error.
func Put(ctx context.Context, sp *Spool, ep config.IngestEndpoint, id, kind string, body []byte, now time.Time) (bool, error) {
	if ep.Source != SourceGeneric {
		return false, fmt.Errorf("endpoint %q is %s; ingest put takes generic endpoints", ep.Name, ep.Source)
	}
	d, err := GenericDelivery(ep, id, kind, body, now, nil)
	if err != nil {
		return false, err
	}
	return sp.Write(ctx, d)
}

func genericSubject(it *domain.InboxItem, d Delivery) error {
	var m config.IngestMap
	if s := d.Headers[MapHeader]; s != "" {
		if err := json.Unmarshal([]byte(s), &m); err != nil {
			return fmt.Errorf("generic map: %w", err)
		}
	}
	var p any
	if err := json.Unmarshal(d.Payload, &p); err != nil {
		return fmt.Errorf("generic payload: %w", err)
	}
	if it.Kind == "" {
		it.Kind = lookup(p, m.Kind)
	}
	it.RepoOrTeam = lookup(p, m.RepoOrTeam)
	it.Actor = lookup(p, m.Actor)
	it.CommentID = lookup(p, m.CommentID)
	if n := lookup(p, m.Number); n != "" {
		v, err := strconv.ParseInt(n, 10, 64)
		if err != nil {
			return fmt.Errorf("generic map number %q: %q is not an integer", m.Number, n)
		}
		it.Number = v
	}
	if m.NaturalKey != "" {
		// Prefixed with the endpoint so a generic key never clashes with
		// a github: or linear: key or another endpoint's.
		it.NaturalKey = "generic:" + d.Endpoint + ":" + keyPart.ReplaceAllStringFunc(m.NaturalKey, func(s string) string {
			return lookup(p, s[1:len(s)-1])
		})
	}
	return nil
}

var keyPart = regexp.MustCompile(`\{[^{}]*\}`)

// lookup follows a dotted path (`a.b.0.c`) through decoded JSON and
// returns the value as text: strings as is, numbers without exponent,
// other scalars by fmt, objects and lists as JSON. Missing is "".
func lookup(v any, path string) string {
	if path == "" {
		return ""
	}
	for _, part := range strings.Split(path, ".") {
		switch x := v.(type) {
		case map[string]any:
			var ok bool
			if v, ok = x[part]; !ok {
				return ""
			}
		case []any:
			i, err := strconv.Atoi(part)
			if err != nil || i < 0 || i >= len(x) {
				return ""
			}
			v = x[i]
		default:
			return ""
		}
	}
	switch x := v.(type) {
	case nil:
		return ""
	case string:
		return x
	case float64:
		return strconv.FormatFloat(x, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(x)
	}
	b, _ := json.Marshal(v)
	return string(b)
}
