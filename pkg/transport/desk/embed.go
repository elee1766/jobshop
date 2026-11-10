package desk

import (
	"context"
	"html"
	"net"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/gfx-labs/jobshop/pkg/config"
)

// EmbedView is one desk.embeds entry as the browser needs it (JOB-151). The
// controller's validated config is the only source: no request parameter,
// ticket or agent output adds an entry.
type EmbedView struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	URL   string `json:"url"`
	// Sandbox is the complete iframe sandbox attribute value.
	Sandbox string `json:"sandbox"`
	// Allow is the iframe allow attribute value, empty for none.
	Allow string `json:"allow,omitempty"`
}

// EmbedList is GET /api/embeds and the live "embeds" section.
type EmbedList struct {
	Embeds []EmbedView `json:"embeds"`
}

// embedOriginsMeta names the <meta> that tells the page which origins its
// frame-src allowed when it loaded, so it can say a reload is needed after
// the config adds an embed.
const embedOriginsMeta = "jobshop-embed-origins"

// baseCSP is Desk's policy for every response.
const baseCSP = "default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self' data: blob:; connect-src 'self'; worker-src 'self'; manifest-src 'self'; base-uri 'self'; frame-ancestors 'none'; form-action 'none'"

// embeds returns the valid configured embeds for the request. An entry that
// fails the controller's own checks (a config from another version) is
// dropped. An embed on Desk's own host never gets allow-same-origin: with
// allow-scripts that lets the page remove its sandbox and act as Desk.
func (s *Server) embeds(r *http.Request) ([]EmbedView, error) {
	dc, err := s.deskConfig(r)
	if err != nil {
		return nil, err
	}
	out := make([]EmbedView, 0, len(dc.Embeds))
	for _, e := range dc.Embeds {
		if e.Check() != nil {
			continue
		}
		tokens := e.SandboxTokens()
		if sameHost(e, r.Host) {
			tokens = slices.DeleteFunc(tokens, func(t string) bool { return t == "allow-same-origin" })
		}
		out = append(out, EmbedView{ID: e.ID, Label: e.Label, URL: e.URL, Sandbox: strings.Join(tokens, " "), Allow: strings.Join(e.Allow, "; ")})
	}
	return out, nil
}

// sameHost reports whether the embed's host name is the one the browser
// used for Desk. Ports and schemes are ignored on purpose: another port on
// the same host is another origin, but it shares Desk's cookies, so the
// conservative answer is to treat it as Desk's own.
func sameHost(e config.DeskEmbed, host string) bool {
	u, err := url.Parse(e.URL)
	if err != nil {
		return false
	}
	deskHost, _, err := net.SplitHostPort(host)
	if err != nil {
		deskHost = host
	}
	return strings.EqualFold(u.Hostname(), strings.Trim(deskHost, "[]"))
}

func (s *Server) embedList(r *http.Request) (any, error) {
	list, err := s.embeds(r)
	return EmbedList{Embeds: list}, err
}

// embedOrigins are the distinct origins of the configured embeds, in order.
func (s *Server) embedOrigins(ctx context.Context, r *http.Request) []string {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	dc, err := s.deskConfig(r.WithContext(ctx))
	if err != nil {
		s.log.WarnContext(ctx, "desk: embeds: config unavailable, serving without frame-src", "err", err)
		return nil
	}
	var origins []string
	for _, e := range dc.Embeds {
		if e.Check() != nil {
			continue
		}
		if o := e.Origin(); !slices.Contains(origins, o) {
			origins = append(origins, o)
		}
	}
	return origins
}

// indexEmbeds adds the frame-src for the configured embeds to the index
// page's policy and a meta tag naming them. With none configured the page and
// its policy are unchanged. Only the index needs frame-src: the iframe lives
// in the document, and assets and the API never frame anything.
func (s *Server) indexEmbeds(w http.ResponseWriter, r *http.Request, page string) string {
	origins := s.embedOrigins(r.Context(), r)
	if len(origins) == 0 {
		return page
	}
	w.Header().Set("Content-Security-Policy", baseCSP+"; frame-src "+strings.Join(origins, " "))
	meta := `<meta name="` + embedOriginsMeta + `" content="` + html.EscapeString(strings.Join(origins, " ")) + `">`
	return strings.Replace(page, "<head>", "<head>\n    "+meta, 1)
}
