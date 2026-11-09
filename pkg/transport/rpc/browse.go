package rpc

import "time"

// Browse is Desk's read-only view of Linear and GitHub (JOB-163): lists,
// one issue or PR with its discussion, and the Linear catalog behind the
// property dropdowns. The controller reads through the identities it already
// holds and only for the teams and repos it is configured for (the activity
// scope), so a request never names a URL and the browser never sees a
// credential. There are no write methods: the designed write controls in
// Desk are disabled (BrowseWrites says why).

// Browse providers.
const (
	BrowseLinear = "linear"
	BrowseGitHub = "github"
)

// Browse status states. A list or detail is Fresh when this call read the
// provider, Cached when a recent copy answered, Stale when the provider
// failed and an older copy answered, and Error when there is nothing to
// show.
const (
	BrowseFresh  = "fresh"
	BrowseCached = "cached"
	BrowseStale  = "stale"
	BrowseError  = "error"
)

// Browse error codes in BrowseStatus.Code.
const (
	BrowseCodeAuth          = "auth"
	BrowseCodeRateLimited   = "rate_limited"
	BrowseCodeUnavailable   = "unavailable"
	BrowseCodeNotConfigured = "not_configured"
	BrowseCodeNotFound      = "not_found"
	BrowseCodeError         = "error"
)

// BrowseStatus says where a result came from and, when the provider failed,
// why. The message never holds a credential.
type BrowseStatus struct {
	State     string    `json:"state"`
	Code      string    `json:"code,omitempty"`
	Message   string    `json:"message,omitempty"`
	FetchedAt time.Time `json:"fetched_at,omitzero"`
	// RetryAt is when a rate-limited provider may be asked again.
	RetryAt time.Time `json:"retry_at,omitzero"`
}

// BrowseWrites says whether Desk may change anything at the providers.
type BrowseWrites struct {
	Enabled bool   `json:"enabled"`
	Reason  string `json:"reason"`
}

// BrowseLinearScope is one Linear team Desk may read.
type BrowseLinearScope struct {
	Team      string `json:"team"`
	Workspace string `json:"workspace"`
}

// BrowseGitHubScope is one GitHub repo Desk may read.
type BrowseGitHubScope struct {
	Repo string `json:"repo"`
}

// BrowseLimits are the bounds a client can rely on.
type BrowseLimits struct {
	PageSize    int `json:"page_size"`
	MaxPageSize int `json:"max_page_size"`
	// ThreadPage is how many comments one detail or older-activity page
	// reads.
	ThreadPage int `json:"thread_page"`
}

// BrowseScopeResult is browse.scope: what is configured, with a revision
// that changes when the controller learns the providers changed. Desk's
// live stream sends it, so a changed revision tells an open view to read
// again without polling.
type BrowseScopeResult struct {
	Revision int64               `json:"revision"`
	Linear   []BrowseLinearScope `json:"linear"`
	GitHub   []BrowseGitHubScope `json:"github"`
	Writes   BrowseWrites        `json:"writes"`
	Limits   BrowseLimits        `json:"limits"`
}

// BrowseListParams is browse.list. Scope is a team key or owner/name and
// must be configured. View is, for Linear, open, active, backlog, closed or
// all, for GitHub open, merged, closed or all. Project and Owner are Linear
// ids from browse.catalog, Author a GitHub login. Query is plain words:
// provider search qualifiers in it are dropped, so it cannot widen the
// scope. Cursor continues a previous result's Next.
type BrowseListParams struct {
	Provider string `json:"provider"`
	Scope    string `json:"scope"`
	Kind     string `json:"kind,omitempty"`
	View     string `json:"view,omitempty"`
	Project  string `json:"project,omitempty"`
	Owner    string `json:"owner,omitempty"`
	Author   string `json:"author,omitempty"`
	Query    string `json:"query,omitempty"`
	Cursor   string `json:"cursor,omitempty"`
	Limit    int    `json:"limit,omitempty"`
	Refresh  bool   `json:"refresh,omitempty"`
}

// BrowseItem is one row of a list and the head of a detail. Ref is the
// exact reference (JOB-150, owner/name#986), ID what the id column shows
// (JOB-150, #986).
type BrowseItem struct {
	Provider string `json:"provider"`
	Scope    string `json:"scope"`
	Ref      string `json:"ref"`
	ID       string `json:"id"`
	Kind     string `json:"kind"`
	Title    string `json:"title"`
	// State is the Linear state name, or open, draft, merged or closed on
	// GitHub. StateType is the Linear state type (backlog, unstarted,
	// started, completed, canceled, triage) or the GitHub state again.
	State     string `json:"state"`
	StateType string `json:"state_type"`
	// Owner is the Linear assignee or the GitHub author. OwnerBot is set
	// for a GitHub App author, which GitHub searches as app/<login>.
	Owner    string `json:"owner,omitempty"`
	OwnerBot bool   `json:"owner_bot,omitempty"`
	// Checks is the GitHub CI rollup: pass, fail, run or empty. Review is
	// approved, changes, required or empty.
	Checks    string    `json:"checks,omitempty"`
	Review    string    `json:"review,omitempty"`
	Project   string    `json:"project,omitempty"`
	Priority  string    `json:"priority,omitempty"`
	Labels    []string  `json:"labels"`
	UpdatedAt time.Time `json:"updated_at,omitzero"`
	URL       string    `json:"url"`
}

// BrowseListResult is one page of a list. Total is what GitHub reports for
// a search, 0 when the provider reports none. HasMore with Next continues
// it: a list is never complete just because it is loaded.
type BrowseListResult struct {
	Items   []BrowseItem `json:"items"`
	Next    string       `json:"next,omitempty"`
	HasMore bool         `json:"has_more"`
	Total   int          `json:"total,omitempty"`
	Status  BrowseStatus `json:"status"`
}

// BrowseShowParams is browse.show. Ref is JOB-150 or owner/name#986; Kind
// (pr or issue) picks between a PR and an issue of the same number. Cursor
// is a previous detail's Older, and then only older discussion returns.
// Brief skips the discussion, for a context reference.
type BrowseShowParams struct {
	Provider string `json:"provider"`
	Ref      string `json:"ref"`
	Kind     string `json:"kind,omitempty"`
	Cursor   string `json:"cursor,omitempty"`
	Brief    bool   `json:"brief,omitempty"`
	Refresh  bool   `json:"refresh,omitempty"`
}

// BrowseProps are the properties shown beside the title. The Linear ones
// are what the dropdowns show; the GitHub ones the PR header.
type BrowseProps struct {
	Assignee      string   `json:"assignee,omitempty"`
	Priority      string   `json:"priority,omitempty"`
	PriorityValue int      `json:"priority_value,omitempty"`
	Project       string   `json:"project,omitempty"`
	Labels        []string `json:"labels"`

	Author       string    `json:"author,omitempty"`
	Base         string    `json:"base,omitempty"`
	Head         string    `json:"head,omitempty"`
	Additions    int       `json:"additions,omitempty"`
	Deletions    int       `json:"deletions,omitempty"`
	ChangedFiles int       `json:"changed_files,omitempty"`
	Mergeable    string    `json:"mergeable,omitempty"`
	CreatedAt    time.Time `json:"created_at,omitzero"`
}

// BrowseLinked is one real association of an item: a Linear relation,
// parent, child or attached PR, or a GitHub closing reference or a mention
// of a configured Linear key. A run is a Jobshop run whose meta names the
// item; Ref is then the run id and URL is empty.
type BrowseLinked struct {
	Kind     string `json:"kind"` // issue, pr or run
	Ref      string `json:"ref"`
	Title    string `json:"title,omitempty"`
	State    string `json:"state,omitempty"`
	URL      string `json:"url,omitempty"`
	Relation string `json:"relation"`
}

// BrowseComment is one message. ReplyTo is set only when the provider
// names the message this one answers.
type BrowseComment struct {
	ID            string    `json:"id"`
	Author        string    `json:"author"`
	Bot           bool      `json:"bot,omitempty"`
	At            time.Time `json:"at"`
	EditedAt      time.Time `json:"edited_at,omitzero"`
	URL           string    `json:"url,omitempty"`
	Body          string    `json:"body"`
	Truncated     bool      `json:"truncated,omitempty"`
	ReplyTo       string    `json:"reply_to,omitempty"`
	ReplyToAuthor string    `json:"reply_to_author,omitempty"`
	// State is a GitHub review's verdict: approved, changes, commented.
	State string `json:"state,omitempty"`
}

// BrowseThread is a Linear comment with its replies, a GitHub review
// thread on a line of a file, or a single GitHub comment or review.
// Comments run oldest first. Hidden counts replies the provider has that
// this response leaves out.
type BrowseThread struct {
	ID       string          `json:"id"`
	Kind     string          `json:"kind"` // comment, review or diff
	Path     string          `json:"path,omitempty"`
	Line     int             `json:"line,omitempty"`
	Resolved bool            `json:"resolved,omitempty"`
	Outdated bool            `json:"outdated,omitempty"`
	URL      string          `json:"url,omitempty"`
	Comments []BrowseComment `json:"comments"`
	Hidden   int             `json:"hidden,omitempty"`
}

// BrowseFile is one changed file of a PR.
type BrowseFile struct {
	Path      string `json:"path"`
	Additions int    `json:"additions"`
	Deletions int    `json:"deletions"`
	Change    string `json:"change"`
}

// BrowseCheck is one CI check or status of a PR's head commit. State is
// pass, fail, run or skip.
type BrowseCheck struct {
	Name      string    `json:"name"`
	State     string    `json:"state"`
	URL       string    `json:"url,omitempty"`
	StartedAt time.Time `json:"started_at,omitzero"`
	Seconds   int       `json:"seconds,omitempty"`
}

// BrowseDetail is browse.show. Older, when set, continues the discussion
// with browse.show Cursor. Files and Checks are bounded; the Total fields
// say how many the provider has.
type BrowseDetail struct {
	Item          BrowseItem     `json:"item"`
	Body          string         `json:"body"`
	BodyTruncated bool           `json:"body_truncated,omitempty"`
	Props         BrowseProps    `json:"props"`
	Linked        []BrowseLinked `json:"linked"`
	Threads       []BrowseThread `json:"threads"`
	Older         string         `json:"older,omitempty"`
	Files         []BrowseFile   `json:"files"`
	FilesTotal    int            `json:"files_total,omitempty"`
	Checks        []BrowseCheck  `json:"checks"`
	ChecksTotal   int            `json:"checks_total,omitempty"`
	Status        BrowseStatus   `json:"status"`
}

// BrowseCatalogParams is browse.catalog for one Linear team.
type BrowseCatalogParams struct {
	Team    string `json:"team"`
	Refresh bool   `json:"refresh,omitempty"`
}

// BrowseOption is one catalog entry.
type BrowseOption struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type,omitempty"`
}

// BrowseCatalogResult holds the lists behind the Linear dropdowns.
type BrowseCatalogResult struct {
	States   []BrowseOption `json:"states"`
	Projects []BrowseOption `json:"projects"`
	Users    []BrowseOption `json:"users"`
	Labels   []BrowseOption `json:"labels"`
	Status   BrowseStatus   `json:"status"`
}
