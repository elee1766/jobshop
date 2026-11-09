// Package github is the GitHub API client every GitHub caller shares: the
// controller's reads and writes, the ingest poller, activity and the
// github built-in. A Client is one identity: it asks its apicall.Source for
// the token on every request, so one Client serves every caller, and a
// renewed token is picked up without rebuilding it.
package github

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/elee1766/jobshop/pkg/integrations/apicall"
	"github.com/elee1766/jobshop/pkg/integrations/internal/transport"
	"github.com/elee1766/jobshop/pkg/telemetry"
)

// DefaultAPIURL is the GitHub REST API base.
const DefaultAPIURL = "https://api.github.com"

// DefaultMaxBody bounds a response body unless Request.MaxBody is set.
const DefaultMaxBody = 16 << 20

var linkNext = regexp.MustCompile(`<([^>]+)>;\s*rel="next"`)

// Client calls the GitHub API as one identity.
type Client struct {
	// Base is the REST base: https://api.github.com, or GHES's /api/v3.
	Base string
	Auth apicall.Source
	HTTP *http.Client
}

// New returns a client for auth. base "" is DefaultAPIURL. A redirect is
// followed only within the API host, so the token never goes elsewhere.
func New(base string, auth apicall.Source) *Client {
	if base == "" {
		base = DefaultAPIURL
	}
	base = strings.TrimRight(base, "/")
	hc := telemetry.HTTPClient("github", time.Minute)
	if u, err := url.Parse(base); err == nil {
		hc.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 || req.URL.Host != u.Host || req.URL.Scheme != u.Scheme {
				return errors.New("redirect away from the API refused")
			}
			return nil
		}
	}
	return &Client{Base: base, Auth: auth, HTTP: hc}
}

// Request is one REST call.
type Request struct {
	Method string // GET when empty
	// Path is an API path such as /repos/o/r/pulls, or an absolute URL on
	// the API host (a Link next page).
	Path   string
	Query  url.Values
	Header map[string]string
	Body   any // marshalled as JSON when non-nil; []byte is sent as is
	// ETag is sent as If-None-Match; a 304 is then a response, not an error.
	ETag string
	// Route names the call in traces (telemetry.WithRoute); paths can hold
	// names that do not belong in a span.
	Route string
	// MaxBody bounds the body; 0 is DefaultMaxBody.
	MaxBody int
}

// Response is a REST response with its body read.
type Response struct {
	Status int
	Header http.Header
	Body   []byte
}

// NotModified reports a 304 for the request's ETag.
func (r *Response) NotModified() bool { return r.Status == http.StatusNotModified }

// Next is the rel="next" URL of the Link header, "" on the last page.
func (r *Response) Next() string {
	if m := linkNext.FindStringSubmatch(r.Header.Get("Link")); m != nil {
		return m[1]
	}
	return ""
}

// StatusError is a response outside 2xx (and 304 when an ETag was sent).
// A GraphQL rate limit or authentication failure is one too, with the
// status that means.
type StatusError struct {
	Method, Path string
	Status       int
	Body         string
	// RetryAt is when GitHub says to ask again: Retry-After, or the reset
	// time of an exhausted rate limit. Zero when it says nothing.
	RetryAt time.Time
}

func (e *StatusError) Error() string {
	return fmt.Sprintf("github %s %s: %d %s: %s", e.Method, e.Path, e.Status, http.StatusText(e.Status), e.Body)
}

// RateLimited reports whether the response is GitHub's rate limit: a 429,
// or a 403 that says the limit is used up (any other 403 is a permission).
func (e *StatusError) RateLimited() bool {
	if e.Status == http.StatusTooManyRequests {
		return true
	}
	body := strings.ToLower(e.Body)
	return e.Status == http.StatusForbidden && (strings.Contains(body, "rate limit") || strings.Contains(body, "abuse") || !e.RetryAt.IsZero())
}

// retryAt reads when GitHub allows the next request from a response's
// headers: Retry-After (seconds), else the rate limit reset when none is
// left.
func retryAt(h http.Header, now time.Time) time.Time {
	if s, err := strconv.Atoi(h.Get("Retry-After")); err == nil && s >= 0 {
		return now.Add(time.Duration(s) * time.Second)
	}
	if h.Get("X-RateLimit-Remaining") == "0" {
		if sec, err := strconv.ParseInt(h.Get("X-RateLimit-Reset"), 10, 64); err == nil && sec > 0 {
			return time.Unix(sec, 0)
		}
	}
	return time.Time{}
}

// IsNotFound reports a 404 from the API.
func IsNotFound(err error) bool {
	var se *StatusError
	return errors.As(err, &se) && se.Status == http.StatusNotFound
}

// Do sends r. A status outside 2xx (bar a 304 for r.ETag) is a
// *StatusError; Raw returns those responses instead.
func (c *Client) Do(ctx context.Context, r Request) (*Response, error) {
	resp, err := c.Raw(ctx, r)
	if err != nil {
		return nil, err
	}
	if resp.Status/100 != 2 && (!resp.NotModified() || r.ETag == "") {
		return nil, &StatusError{Method: r.method(), Path: r.Path, Status: resp.Status, Body: strings.TrimSpace(string(resp.Body[:min(len(resp.Body), 300)])), RetryAt: retryAt(resp.Header, time.Now())}
	}
	return resp, nil
}

// Get is Do for a GET of path, decoding a 2xx body into out.
func (c *Client) Get(ctx context.Context, path string, out any) error {
	resp, err := c.Do(ctx, Request{Path: path})
	if err != nil {
		return err
	}
	return json.Unmarshal(resp.Body, out)
}

// List GETs r and its next pages, at most pages, and concatenates the JSON
// arrays. The first request carries r.ETag; a 304 returns nil items and
// the response. The response is the first page's.
func (c *Client) List(ctx context.Context, r Request, pages int) ([]json.RawMessage, *Response, error) {
	first, err := c.Do(ctx, r)
	if err != nil || first.NotModified() {
		return nil, first, err
	}
	var all []json.RawMessage
	resp := first
	for page := 1; ; page++ {
		var items []json.RawMessage
		if err := json.Unmarshal(resp.Body, &items); err != nil {
			return nil, nil, fmt.Errorf("github GET %s: %w", r.Path, err)
		}
		all = append(all, items...)
		next := resp.Next()
		if next == "" || page >= pages {
			return all, first, nil
		}
		if resp, err = c.Do(ctx, Request{Path: next, Header: r.Header, Route: r.Route, MaxBody: r.MaxBody}); err != nil {
			return nil, nil, err
		}
	}
}

// Raw sends r and returns any response GitHub gives, renewing the token
// once on a 401.
func (c *Client) Raw(ctx context.Context, r Request) (*Response, error) {
	u, err := c.url(r)
	if err != nil {
		return nil, err
	}
	var body []byte
	switch b := r.Body.(type) {
	case nil:
	case []byte:
		body = b
	default:
		if body, err = json.Marshal(b); err != nil {
			return nil, err
		}
	}
	tok, err := c.Auth.Token(ctx)
	if err != nil {
		return nil, fmt.Errorf("github: %w", err)
	}
	return transport.Authorized(ctx, c.Auth, tok, func(tok string) (*Response, int, error) {
		resp, err := c.send(ctx, r, u, body, tok)
		if err != nil {
			return nil, 0, err
		}
		return resp, resp.Status, nil
	})
}

func (r Request) method() string {
	if r.Method == "" {
		return http.MethodGet
	}
	return r.Method
}

// url resolves r.Path against Base. An absolute URL must be on the API's
// host: a Link header is the only source of one.
func (c *Client) url(r Request) (string, error) {
	u := r.Path
	if strings.HasPrefix(u, "/") {
		u = c.Base + u
	} else {
		pu, err := url.Parse(u)
		bu, berr := url.Parse(c.Base)
		if err != nil || berr != nil || pu.Host != bu.Host || pu.Scheme != bu.Scheme {
			return "", errors.New("github: URL leaves the API")
		}
	}
	if len(r.Query) > 0 {
		sep := "?"
		if strings.Contains(u, "?") {
			sep = "&"
		}
		u += sep + r.Query.Encode()
	}
	return u, nil
}

func (c *Client) send(ctx context.Context, r Request, u string, body []byte, tok string) (*Response, error) {
	ctx, cancel := apicall.Request(ctx)
	defer cancel()
	if r.Route != "" {
		ctx = telemetry.WithRoute(ctx, r.Route)
	}
	var rd io.Reader
	if body != nil {
		rd = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, r.method(), u, rd)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	for k, v := range r.Header {
		req.Header.Set(k, v)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if r.ETag != "" {
		req.Header.Set("If-None-Match", r.ETag)
	}
	req.Header.Set("Authorization", "Bearer "+tok)
	apicall.Sending(ctx, tok)
	resp, err := c.HTTP.Do(req)
	if err != nil {
		// url.Error repeats the full URL; the path is enough.
		var ue *url.Error
		if errors.As(err, &ue) {
			err = ue.Err
		}
		return nil, fmt.Errorf("github %s %s: %w", r.method(), r.Path, err)
	}
	defer func() { _ = resp.Body.Close() }()
	limit := r.MaxBody
	if limit <= 0 {
		limit = DefaultMaxBody
	}
	b, over, err := transport.ReadLimited(resp.Body, limit)
	if err != nil {
		return nil, err
	}
	if over {
		return nil, fmt.Errorf("github %s %s: response over %d bytes", r.method(), r.Path, limit)
	}
	return &Response{Status: resp.StatusCode, Header: resp.Header, Body: b}, nil
}

// GraphQL runs query with vars and decodes data into out (a
// *json.RawMessage keeps it as sent). A non-200 status or a response with
// errors is an error carrying GitHub's messages.
func (c *Client) GraphQL(ctx context.Context, query string, vars map[string]any, out any) error {
	// GraphQL lives at /graphql on api.github.com and /api/graphql on GHES (/api/v3).
	endpoint := c.Base + "/graphql"
	if strings.HasSuffix(c.Base, "/api/v3") {
		endpoint = strings.TrimSuffix(c.Base, "/v3") + "/graphql"
	}
	if vars == nil {
		vars = map[string]any{}
	}
	resp, err := c.Raw(ctx, Request{Method: http.MethodPost, Path: endpoint, Body: map[string]any{"query": query, "variables": vars}, Route: "/graphql"})
	if err != nil {
		return err
	}
	var r struct {
		// Message is the body of a non-GraphQL error, such as a 403.
		Message string          `json:"message"`
		Data    json.RawMessage `json:"data"`
		Errors  []struct {
			Type    string `json:"type"`
			Message string `json:"message"`
		} `json:"errors"`
	}
	at := retryAt(resp.Header, time.Now())
	if err := json.Unmarshal(resp.Body, &r); err != nil {
		if resp.Status != http.StatusOK {
			return &StatusError{Method: http.MethodPost, Path: "/graphql", Status: resp.Status, RetryAt: at}
		}
		return fmt.Errorf("github graphql: HTTP %d: invalid JSON response", resp.Status)
	}
	if len(r.Errors) > 0 {
		msgs := make([]string, len(r.Errors))
		for i, e := range r.Errors {
			msgs[i] = e.Message
		}
		// GraphQL reports a used-up rate limit as a 200 with a RATE_LIMITED
		// error type.
		if r.Errors[0].Type == "RATE_LIMITED" {
			return &StatusError{Method: http.MethodPost, Path: "/graphql", Status: http.StatusTooManyRequests, Body: strings.Join(msgs, "; "), RetryAt: at}
		}
		return fmt.Errorf("github graphql: %s", strings.Join(msgs, "; "))
	}
	if resp.Status != http.StatusOK || len(r.Data) == 0 || string(r.Data) == "null" {
		if resp.Status != http.StatusOK {
			return &StatusError{Method: http.MethodPost, Path: "/graphql", Status: resp.Status, Body: r.Message, RetryAt: at}
		}
		return fmt.Errorf("github graphql: HTTP %d", resp.Status)
	}
	return json.Unmarshal(r.Data, out)
}
