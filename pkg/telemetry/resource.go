package telemetry

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
)

// Resource attribute keys jobshop sets on every signal (JOB-106).
const (
	AttrInstance       = attribute.Key("jobshop.instance")
	AttrCommit         = attribute.Key("jobshop.commit")
	attrServiceName    = attribute.Key("service.name")
	attrServiceVersion = attribute.Key("service.version")
	attrServiceInst    = attribute.Key("service.instance.id")
)

// Identity says which jobshop a process is. NewResource turns it into the
// resource shared by traces, logs and metrics.
type Identity struct {
	// Instance names the deployment (telemetry.instance). Empty:
	// DefaultInstance(Root).
	Instance string
	// Root is the jobshop root directory.
	Root string
	// Version and Commit identify the binary.
	Version, Commit string
	// Attributes are extra resource attributes (telemetry.resource),
	// already expanded.
	Attributes map[string]string
}

// NewResource builds the process resource. Precedence, lowest first:
// jobshop's own attributes (service.name=jobshop, service.version,
// jobshop.commit, service.instance.id and jobshop.instance), then
// id.Attributes, then OTEL_RESOURCE_ATTRIBUTES and OTEL_SERVICE_NAME.
// The environment wins so an operator can relabel one process without
// editing the config. A malformed OTEL_RESOURCE_ATTRIBUTES keeps what
// parsed.
func NewResource(ctx context.Context, id Identity) (*resource.Resource, error) {
	inst := id.Instance
	if inst == "" {
		inst = DefaultInstance(id.Root)
	}
	attrs := []attribute.KeyValue{
		attrServiceName.String("jobshop"),
		attrServiceInst.String(inst),
		AttrInstance.String(inst),
	}
	if v := serviceVersion(id.Version, id.Commit); v != "" {
		attrs = append(attrs, attrServiceVersion.String(v))
	}
	if id.Commit != "" && id.Commit != "unknown" {
		attrs = append(attrs, AttrCommit.String(id.Commit))
	}
	keys := make([]string, 0, len(id.Attributes))
	for k := range id.Attributes {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		attrs = append(attrs, attribute.String(k, id.Attributes[k]))
	}
	res, err := resource.New(ctx, resource.WithAttributes(attrs...), resource.WithFromEnv())
	if err != nil && !errors.Is(err, resource.ErrPartialResource) {
		return nil, err
	}
	return res, nil
}

// serviceVersion is version with the commit as semver build metadata,
// unless the version already carries it (the Nix build, JOB-79).
func serviceVersion(version, commit string) string {
	if commit == "" || commit == "unknown" || strings.Contains(version, commit) {
		return version
	}
	if version == "" {
		return commit
	}
	return version + "+" + commit
}

// DefaultInstance is the instance name when telemetry.instance is unset:
// the root directory's base name, or hostname:root when the base name
// says nothing (the module default dirs both end in "jobshop").
func DefaultInstance(root string) string {
	base := filepath.Base(root)
	switch base {
	case "jobshop", ".", "/", "":
	default:
		return base
	}
	host, err := os.Hostname()
	if err != nil || host == "" {
		host = "localhost"
	}
	if abs, err := filepath.Abs(root); err == nil {
		root = abs
	}
	return host + ":" + root
}
