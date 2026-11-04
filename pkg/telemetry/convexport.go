package telemetry

import (
	"context"
	"encoding/json"
	"errors"
	"regexp"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unicode/utf8"

	"github.com/zricethezav/gitleaks/v8/detect"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	otellog "go.opentelemetry.io/otel/log"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
)

// Conversation export defaults.
const (
	DefaultFieldCap = 8 << 10 // bytes per exported text field
	convQueueFrames = 1024
	convQueueBytes  = 16 << 20
	redacted        = "[REDACTED]"
)

// Conversation record attributes.
const (
	AttrConvEvent     = attribute.Key("jobshop.conversation.event")
	AttrConvTruncated = attribute.Key("jobshop.conversation.truncated")
	AttrToolCall      = attribute.Key("jobshop.tool.call_id")
	AttrToolTitle     = attribute.Key("jobshop.tool.title")
	AttrToolKind      = attribute.Key("jobshop.tool.kind")
	AttrToolStatus    = attribute.Key("jobshop.tool.status")
	AttrToolInput     = attribute.Key("jobshop.tool.input")
	AttrToolOutput    = attribute.Key("jobshop.tool.output")
	AttrStopReason    = attribute.Key("jobshop.turn.stop_reason")
	AttrModel         = attribute.Key("jobshop.session.model")
	AttrEffort        = attribute.Key("jobshop.session.effort")
)

// MetaTicket is the session/prompt _meta key carrying the ticket the turn
// works on.
const MetaTicket = "jobshop/ticket"

// MetaModel and MetaEffort carry the session's model and effort (#304).
const (
	MetaModel  = "jobshop/model"
	MetaEffort = "jobshop/effort"
)

// ConvConfig configures one session's conversation export.
type ConvConfig struct {
	LogsURL     string
	Headers     map[string]string
	Session     string
	Pool        string
	Project     string
	TraceParent string
	// Env is the agent's launch env (KEY=VALUE); values of secret-named
	// keys are redacted wherever they appear.
	Env      []string
	FieldCap int // default DefaultFieldCap
	// Exporter replaces the OTLP exporter (tests).
	Exporter sdklog.Exporter
}

// ConvExporter turns ACP frames into OTLP log records. Observe never
// blocks: frames go to a bounded queue and a worker maps, redacts and
// emits them. Overflow drops the frame and counts it.
type ConvExporter struct {
	cfg    ConvConfig
	lp     *sdklog.LoggerProvider
	logger otellog.Logger
	ctx    context.Context
	env    []string // secret values from the launch env, longest first

	q      chan convFrame
	qBytes atomic.Int64
	done   chan struct{}
	closed sync.Once

	dropped, truncated atomic.Int64

	// worker state
	ticket string
	// model and effort come with each prompt's _meta (#304).
	model, effort string
	text          strings.Builder // pending agent message chunks
	kind          string          // "agent_message" or "agent_thought"
}

type convFrame struct {
	from string // "agent" or "controller"
	line []byte
}

// NewConvExporter starts the exporter. It fails only on bad config; an
// unreachable endpoint shows up as export errors in the background.
func NewConvExporter(ctx context.Context, cfg ConvConfig) (*ConvExporter, error) {
	if cfg.FieldCap <= 0 {
		cfg.FieldCap = DefaultFieldCap
	}
	exp := cfg.Exporter
	if exp == nil {
		if cfg.LogsURL == "" {
			return nil, errors.New("conversation export: no logs endpoint")
		}
		var err error
		exp, err = otlploghttp.New(ctx, otlploghttp.WithEndpointURL(cfg.LogsURL), otlploghttp.WithHeaders(cfg.Headers),
			otlploghttp.WithTimeout(10*time.Second), otlploghttp.WithRetry(otlploghttp.RetryConfig{Enabled: false}))
		if err != nil {
			return nil, err
		}
	}
	res := resource.NewSchemaless(attribute.String("service.name", "jobshop-host"))
	lp := sdklog.NewLoggerProvider(sdklog.WithResource(res), sdklog.WithProcessor(sdklog.NewBatchProcessor(exp)))
	e := &ConvExporter{
		cfg: cfg, lp: lp, logger: lp.Logger(Scope + "/conversation"),
		ctx: WithTraceParent(context.WithoutCancel(ctx), cfg.TraceParent),
		env: secretEnvValues(cfg.Env),
		q:   make(chan convFrame, convQueueFrames), done: make(chan struct{}),
	}
	go e.run()
	return e, nil
}

// Observe queues one frame without blocking. line is copied.
func (e *ConvExporter) Observe(from string, line []byte) {
	if e == nil {
		return
	}
	n := int64(len(line))
	if e.qBytes.Add(n) > convQueueBytes {
		e.qBytes.Add(-n)
		e.dropped.Add(1)
		return
	}
	select {
	case e.q <- convFrame{from: from, line: append([]byte(nil), line...)}:
	default:
		e.qBytes.Add(-n)
		e.dropped.Add(1)
	}
}

// TakeCounts returns and resets the dropped and truncated counts.
func (e *ConvExporter) TakeCounts() (dropped, truncated int64) {
	if e == nil {
		return 0, 0
	}
	return e.dropped.Swap(0), e.truncated.Swap(0)
}

// Close drains the queue and flushes, giving up when ctx ends.
func (e *ConvExporter) Close(ctx context.Context) error {
	if e == nil {
		return nil
	}
	e.closed.Do(func() { close(e.q) })
	select {
	case <-e.done:
	case <-ctx.Done():
	}
	return e.lp.Shutdown(ctx)
}

func (e *ConvExporter) run() {
	defer close(e.done)
	for f := range e.q {
		e.qBytes.Add(-int64(len(f.line)))
		e.handle(f)
	}
	e.flushText()
}

type acpFrame struct {
	ID     json.RawMessage `json:"id"`
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
	Result json.RawMessage `json:"result"`
}

func (e *ConvExporter) handle(f convFrame) {
	var m acpFrame
	if json.Unmarshal(f.line, &m) != nil {
		return
	}
	switch {
	case f.from == "controller" && m.Method == "session/prompt":
		e.flushText()
		var p struct {
			Meta   map[string]any `json:"_meta"`
			Prompt []struct {
				Text string `json:"text"`
			} `json:"prompt"`
		}
		_ = json.Unmarshal(m.Params, &p)
		// Message turns carry no ticket; do not keep the last one.
		e.ticket, _ = p.Meta[MetaTicket].(string)
		e.model, _ = p.Meta[MetaModel].(string)
		e.effort, _ = p.Meta[MetaEffort].(string)
		var b strings.Builder
		for _, c := range p.Prompt {
			b.WriteString(c.Text)
		}
		e.emit("prompt", b.String())
	case f.from == "agent" && m.Method == "session/update":
		e.update(m.Params)
	case f.from == "agent" && m.Method == "" && len(m.Result) > 0:
		var r struct {
			StopReason string `json:"stopReason"`
		}
		if json.Unmarshal(m.Result, &r) == nil && r.StopReason != "" {
			e.flushText()
			e.emit("turn_end", "", AttrStopReason.String(r.StopReason))
		}
	}
}

func (e *ConvExporter) update(params json.RawMessage) {
	var u struct {
		Update struct {
			Kind    string `json:"sessionUpdate"`
			Content struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"content"`
			ToolCallID string          `json:"toolCallId"`
			Title      string          `json:"title"`
			ToolKind   string          `json:"kind"`
			Status     string          `json:"status"`
			RawInput   json.RawMessage `json:"rawInput"`
			RawOutput  json.RawMessage `json:"rawOutput"`
		} `json:"update"`
	}
	if json.Unmarshal(params, &u) != nil {
		return
	}
	up := u.Update
	switch up.Kind {
	case "agent_message_chunk", "agent_thought_chunk":
		kind := strings.TrimSuffix(up.Kind, "_chunk")
		if kind != e.kind {
			e.flushText()
			e.kind = kind
		}
		e.text.WriteString(up.Content.Text)
	case "tool_call":
		e.flushText()
		e.emit("tool_call", "", AttrToolCall.String(up.ToolCallID), AttrToolTitle.String(up.Title),
			AttrToolKind.String(up.ToolKind), AttrToolStatus.String(up.Status), AttrToolInput.String(redactJSON(up.RawInput)))
	case "tool_call_update":
		// Only state changes; status-less updates are output deltas.
		if up.Status == "" {
			return
		}
		e.flushText()
		attrs := []attribute.KeyValue{AttrToolCall.String(up.ToolCallID), AttrToolStatus.String(up.Status)}
		if len(up.RawInput) > 0 {
			attrs = append(attrs, AttrToolInput.String(redactJSON(up.RawInput)))
		}
		if len(up.RawOutput) > 0 {
			attrs = append(attrs, AttrToolOutput.String(redactJSON(up.RawOutput)))
		}
		e.emit("tool_result", "", attrs...)
	}
}

func (e *ConvExporter) flushText() {
	if e.text.Len() == 0 {
		return
	}
	s := e.text.String()
	e.text.Reset()
	e.emit(e.kind, s)
}

func (e *ConvExporter) emit(event, body string, attrs ...attribute.KeyValue) {
	var r otellog.Record
	r.SetTimestamp(time.Now())
	r.SetEventName("jobshop.conversation." + event)
	r.SetSeverity(otellog.SeverityInfo)
	trunc := false
	if body != "" {
		var t bool
		body, t = e.clean(body)
		trunc = trunc || t
		r.SetBody(attribute.StringValue(body))
	}
	kvs := []attribute.KeyValue{
		AttrConvEvent.String(event),
		AttrSession.String(e.cfg.Session),
		AttrPool.String(e.cfg.Pool),
		AttrProject.String(e.cfg.Project),
	}
	if e.model != "" {
		kvs = append(kvs, AttrModel.String(e.model))
	}
	if e.effort != "" {
		kvs = append(kvs, AttrEffort.String(e.effort))
	}
	if e.ticket != "" {
		kvs = append(kvs, AttrTicket.String(e.ticket))
	}
	for _, a := range attrs {
		v := a.Value.AsString()
		if v == "" {
			continue
		}
		if a.Key == AttrToolInput || a.Key == AttrToolOutput || a.Key == AttrToolTitle {
			var t bool
			v, t = e.clean(v)
			trunc = trunc || t
		}
		kvs = append(kvs, a.Key.String(v))
	}
	if trunc {
		kvs = append(kvs, AttrConvTruncated.Bool(true))
	}
	r.AddAttributes(kvs...)
	e.logger.Emit(e.ctx, r)
}

// clean redacts secrets in s and caps it at FieldCap bytes, reporting
// whether it cut. Redaction runs on the whole text first, so a secret
// straddling the cap is still caught.
func (e *ConvExporter) clean(s string) (string, bool) {
	s = e.Redact(s)
	if len(s) <= e.cfg.FieldCap {
		return s, false
	}
	e.truncated.Add(1)
	cut := e.cfg.FieldCap
	for cut > 0 && !utf8.RuneStart(s[cut]) {
		cut--
	}
	return s[:cut], true
}

// Redact replaces launch-env secret values and gitleaks findings in s.
func (e *ConvExporter) Redact(s string) string {
	for _, v := range e.env {
		s = strings.ReplaceAll(s, v, redacted)
	}
	d := leaks()
	if d == nil || s == "" {
		return s
	}
	for _, f := range d.DetectString(s) {
		if f.Secret != "" {
			s = strings.ReplaceAll(s, f.Secret, redacted)
		}
	}
	return s
}

var (
	leakOnce sync.Once
	leakDet  *detect.Detector
)

// leaks returns the shared gitleaks detector with the default rules. Its
// config load uses viper's global state, so it is built once. DetectString
// keeps no findings on the detector, so sharing it is safe.
func leaks() *detect.Detector {
	leakOnce.Do(func() {
		if d, err := detect.NewDetectorDefaultConfig(); err == nil {
			leakDet = d
		}
	})
	return leakDet
}

var secretKey = regexp.MustCompile(`(?i)(secret|token|passw(or)?d|api[_-]?key|auth|credential|private[_-]?key|session[_-]?key|cookie)`)

// secretEnvValues returns values of secret-named env vars, longest first
// so a value containing another is replaced whole. Short values are
// skipped: replacing "1" everywhere would wreck the text.
func secretEnvValues(env []string) []string {
	var out []string
	for _, kv := range env {
		k, v, ok := strings.Cut(kv, "=")
		if ok && len(v) >= 8 && secretKey.MatchString(k) {
			out = append(out, v)
		}
	}
	sortByLenDesc(out)
	return out
}

func sortByLenDesc(s []string) {
	slices.SortFunc(s, func(a, b string) int { return len(b) - len(a) })
}

// redactJSON re-encodes raw JSON with values under secret-named keys
// replaced. Non-JSON input is returned as is.
func redactJSON(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	var v any
	if json.Unmarshal(raw, &v) != nil {
		return string(raw)
	}
	b, err := json.Marshal(redactValue(v))
	if err != nil {
		return string(raw)
	}
	return string(b)
}

func redactValue(v any) any {
	switch t := v.(type) {
	case map[string]any:
		for k, x := range t {
			if secretKey.MatchString(k) {
				t[k] = redacted
			} else {
				t[k] = redactValue(x)
			}
		}
	case []any:
		for i, x := range t {
			t[i] = redactValue(x)
		}
	}
	return v
}
