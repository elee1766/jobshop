// Package hostspec is jobshop's session spec on top of acphost: the
// acphost.Spec plus the conversation export setup, which acphost leaves
// to the caller through acphost.ExporterFactory. The JSON layout matches
// the spec.json pkg/host wrote, so hosts started before the switch read
// and adopt cleanly.
package hostspec

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/elee1766/acphost"

	"github.com/elee1766/jobshop/pkg/telemetry"
)

// Spec is acphost.Spec plus jobshop's export setup, stored flat in spec.json.
type Spec struct {
	acphost.Spec
	// ConvExport, if set, turns on conversation export for this host.
	ConvExport *ConvExport `json:"conv_export,omitempty"`
}

// ConvExport is the conversation export setup the controller hands a host.
// Headers may hold tokens; spec.json is 0600 like the rest of the dir.
type ConvExport struct {
	LogsURL     string            `json:"logs_url"`
	Headers     map[string]string `json:"headers,omitempty"`
	Pool        string            `json:"pool"`
	Project     string            `json:"project"`
	TraceParent string            `json:"traceparent,omitempty"`
}

// Write stores s in dir as acphost.SpecFile.
func Write(dir string, s Spec) error {
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, acphost.SpecFile), b, 0o600)
}

// Read loads the spec from dir. acphost.ReadSpec checks the protocol.
func Read(dir string) (Spec, error) {
	base, err := acphost.ReadSpec(dir)
	if err != nil {
		return Spec{}, err
	}
	var s Spec
	b, err := os.ReadFile(filepath.Join(dir, acphost.SpecFile))
	if err != nil {
		return Spec{}, err
	}
	if err := json.Unmarshal(b, &s); err != nil {
		return Spec{}, fmt.Errorf("%s: %w", acphost.SpecFile, err)
	}
	s.Spec = base
	return s, nil
}

// HostOptions returns the acphost.Main options for the session in dir: a
// conversation exporter when spec.json sets conv_export. A spec it cannot
// read gives no options; acphost.Main then reports the error itself.
func HostOptions(dir string) []acphost.Option {
	s, err := Read(dir)
	if err != nil || s.ConvExport == nil {
		return nil
	}
	ce := *s.ConvExport
	return []acphost.Option{acphost.WithExporter(func(base acphost.Spec) (acphost.Exporter, error) {
		return telemetry.NewConvExporter(context.Background(), telemetry.ConvConfig{
			LogsURL: ce.LogsURL, Headers: ce.Headers, Session: base.Session, Pool: ce.Pool,
			Project: ce.Project, TraceParent: ce.TraceParent, Env: append(os.Environ(), base.Env...),
		})
	})}
}
