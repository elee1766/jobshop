package config

import (
	"fmt"
	"net"
	"net/url"
	"regexp"
	"slices"
	"strings"
	"unicode"
)

// DeskEmbed is one desk.embeds entry (JOB-151): an operator-chosen web app
// that Desk shows in an iframe tab. The URL is a trust decision made in the
// service config. Nothing a ticket, an agent or a payload carries can add or
// change one, and Desk never fetches the URL itself.
type DeskEmbed struct {
	// ID is the stable route name: ?view=embed&embed=<id>.
	ID string `yaml:"id" json:"id"`
	// Label is the top bar text.
	Label string `yaml:"label" json:"label"`
	// URL is the absolute http(s) address the iframe loads.
	URL string `yaml:"url" json:"url"`
	// Sandbox adds iframe sandbox tokens to the baseline allow-scripts.
	Sandbox []string `yaml:"sandbox,omitempty" json:"sandbox,omitempty"`
	// Allow delegates browser features (the iframe allow attribute).
	Allow []string `yaml:"allow,omitempty" json:"allow,omitempty"`
	// AllowHTTP permits a plain http URL on a host that is not loopback.
	AllowHTTP bool `yaml:"allow_http,omitempty" json:"allow_http,omitempty"`
}

const (
	// MaxDeskEmbeds bounds the top bar.
	MaxDeskEmbeds = 16
	maxEmbedLabel = 32
	// EmbedBaseSandbox is always present: an app without scripts is a
	// document, not an app, and the sandbox still gives it an opaque origin.
	EmbedBaseSandbox = "allow-scripts"
)

// EmbedSandboxTokens are the sandbox tokens a config may add. Not offered:
// allow-top-navigation* (the app could replace Desk), allow-orientation-lock,
// allow-pointer-lock and allow-storage-access-by-user-activation.
var EmbedSandboxTokens = []string{
	"allow-downloads", "allow-forms", "allow-modals", "allow-popups",
	"allow-popups-to-escape-sandbox", "allow-presentation", "allow-same-origin", "allow-scripts",
}

// EmbedAllowFeatures are the browser features a config may delegate. Camera,
// microphone, geolocation and screen capture are not offered.
var EmbedAllowFeatures = []string{"autoplay", "clipboard-read", "clipboard-write", "fullscreen", "picture-in-picture"}

// embedReserved are ids that would read as a built-in view or action.
var embedReserved = []string{"home", "trace", "improvements", "learn", "system", "sys", "catch-up", "embed", "waiting", "inflight"}

var embedHost = regexp.MustCompile(`^[A-Za-z0-9]([A-Za-z0-9.-]*[A-Za-z0-9])?$`)

// secretKeys mark a query or fragment key that carries a credential. Desk
// shows every embed URL to every Desk user and to config.show, so a secret
// in a URL is not supported.
var secretKeys = []string{"token", "secret", "password", "passwd", "apikey", "api_key", "api-key", "access_key", "credential"}

// Origin returns scheme://host[:port] of a validated URL, the form a CSP
// source and the browser's URL.origin both take: lowercase host, no default
// port.
func (e DeskEmbed) Origin() string {
	u, err := url.Parse(e.URL)
	if err != nil {
		return ""
	}
	host := strings.ToLower(u.Hostname())
	if strings.Contains(host, ":") {
		host = "[" + host + "]"
	}
	port := u.Port()
	if port == "" || (u.Scheme == "http" && port == "80") || (u.Scheme == "https" && port == "443") {
		return u.Scheme + "://" + host
	}
	return u.Scheme + "://" + host + ":" + port
}

// SandboxTokens is the full sandbox attribute list: the baseline, then the
// added tokens in config order without duplicates.
func (e DeskEmbed) SandboxTokens() []string {
	out := []string{EmbedBaseSandbox}
	for _, t := range e.Sandbox {
		if !slices.Contains(out, t) {
			out = append(out, t)
		}
	}
	return out
}

// Check validates one entry. It does not check uniqueness across entries.
func (e DeskEmbed) Check() error {
	if !namePattern.MatchString(e.ID) {
		return fmt.Errorf("id %q must match %s", e.ID, namePattern)
	}
	if slices.Contains(embedReserved, e.ID) {
		return fmt.Errorf("id %q is reserved for a built-in view", e.ID)
	}
	if label := strings.TrimSpace(e.Label); label == "" || label != e.Label {
		return fmt.Errorf("label must be nonempty and not padded")
	}
	if len([]rune(e.Label)) > maxEmbedLabel || strings.ContainsFunc(e.Label, unicode.IsControl) {
		return fmt.Errorf("label must be one line of at most %d characters", maxEmbedLabel)
	}
	if err := e.checkURL(); err != nil {
		return err
	}
	for _, t := range e.Sandbox {
		if !slices.Contains(EmbedSandboxTokens, t) {
			return fmt.Errorf("sandbox %q is not allowed (one of %s)", t, strings.Join(EmbedSandboxTokens, ", "))
		}
	}
	if slices.Contains(e.Sandbox, "allow-popups-to-escape-sandbox") && !slices.Contains(e.Sandbox, "allow-popups") {
		return fmt.Errorf("sandbox allow-popups-to-escape-sandbox needs allow-popups")
	}
	for _, f := range e.Allow {
		if !slices.Contains(EmbedAllowFeatures, f) {
			return fmt.Errorf("allow %q is not allowed (one of %s)", f, strings.Join(EmbedAllowFeatures, ", "))
		}
	}
	return nil
}

func (e DeskEmbed) checkURL() error {
	if e.URL == "" {
		return fmt.Errorf("url is required")
	}
	if strings.ContainsFunc(e.URL, func(r rune) bool { return unicode.IsSpace(r) || unicode.IsControl(r) || r > unicode.MaxASCII }) {
		return fmt.Errorf("url must be ASCII without spaces (encode it)")
	}
	u, err := url.Parse(e.URL)
	if err != nil {
		return fmt.Errorf("url: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("url scheme must be http or https, not %q", u.Scheme)
	}
	if u.User != nil {
		return fmt.Errorf("url must not carry credentials (userinfo)")
	}
	host := u.Hostname()
	if host == "" || u.Opaque != "" {
		return fmt.Errorf("url must be absolute with a host")
	}
	ip := net.ParseIP(host)
	if ip == nil && !embedHost.MatchString(host) {
		return fmt.Errorf("url host %q is not a plain host name or address", host)
	}
	for _, part := range []string{u.RawQuery, u.Fragment} {
		for key := range parseLooseQuery(part) {
			for _, s := range secretKeys {
				if strings.Contains(strings.ToLower(key), s) {
					return fmt.Errorf("url parameter %q looks like a credential: secrets in a URL are not supported", key)
				}
			}
		}
	}
	if u.Scheme == "http" && !e.AllowHTTP && !loopbackHost(host, ip) {
		return fmt.Errorf("url %q is plain http on a non-loopback host: use https or set allow_http: true", e.URL)
	}
	return nil
}

func loopbackHost(host string, ip net.IP) bool {
	if ip != nil {
		return ip.IsLoopback()
	}
	host = strings.ToLower(host)
	return host == "localhost" || strings.HasSuffix(host, ".localhost")
}

// parseLooseQuery splits k=v&k2 pairs without failing on odd escapes.
func parseLooseQuery(s string) map[string]string {
	out := map[string]string{}
	for _, pair := range strings.FieldsFunc(s, func(r rune) bool { return r == '&' || r == ';' }) {
		k, v, _ := strings.Cut(pair, "=")
		if dk, err := url.QueryUnescape(k); err == nil {
			k = dk
		}
		out[k] = v
	}
	return out
}

// checkEmbeds validates the list: each entry, unique ids, and the count.
func checkEmbeds(list []DeskEmbed) error {
	if len(list) > MaxDeskEmbeds {
		return fmt.Errorf("desk.embeds: at most %d entries", MaxDeskEmbeds)
	}
	seen := map[string]bool{}
	for i, e := range list {
		if err := e.Check(); err != nil {
			return fmt.Errorf("desk.embeds[%d]: %w", i, err)
		}
		if seen[e.ID] {
			return fmt.Errorf("desk.embeds[%d]: duplicate id %q", i, e.ID)
		}
		seen[e.ID] = true
	}
	return nil
}
