package control

import (
	"time"

	"github.com/elee1766/jobshop/pkg/telemetry"
)

// outboundCeiling bounds an outbound call whose caller sets no deadline.
// Every caller below sets its own, shorter ctx timeout (a step's timeout,
// a hook's timeout, a minute for the GitHub reads), so this only stops a
// call that would otherwise hang for good, as http.DefaultClient did.
const outboundCeiling = 24 * time.Hour

// Traced clients for the outbound calls that used http.DefaultClient
// (JOB-105). A call's span shows the host and a route template, never a
// path or query that may hold a token: set the template with
// telemetry.WithRoute.
var (
	githubHTTP   = telemetry.HTTPClient("github", outboundCeiling)
	hookHTTP     = telemetry.HTTPClient("hook", outboundCeiling)
	stepHTTP     = telemetry.HTTPClient("step", outboundCeiling)
	providerHTTP = telemetry.HTTPClient("provider", outboundCeiling)
)
