package hooks

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/goccy/go-yaml"

	"github.com/elee1766/jobshop/pkg/automation"
	"github.com/elee1766/jobshop/pkg/config"
)

// Program pins loaded source, credentials, authority, and the prompt asset root.
type Program struct {
	engine *automation.Engine
	entry  string
	root   string
	digest string
	// authority is digest without the configuration around the callback:
	// a run pins it so unrelated reloads do not strand the run.
	authority string
	secrets   []string
}

// LoadPrograms loads callbacks and binds source identity to configured host authority.
func LoadPrograms(ctx context.Context, cfg *config.Config, root config.Root) (map[string]*Program, error) {
	options := automation.Options{AllowHTTP: cfg.Automation.HTTP, AllowShell: cfg.Automation.Shell, ShellDir: root.Dir, Auth: map[string]automation.Credential{}, Observe: observeHookCall, Replay: replayHookCall}
	secrets := []string{}
	options.Limits.Steps = cfg.Automation.MaxSteps
	if cfg.Automation.GitHub {
		options.GitHub = hookGitHubToken
	}
	options.Limits.OutputBytes = cfg.Automation.OutputBytes
	if cfg.Automation.Timeout != "" {
		options.Limits.Duration, _ = time.ParseDuration(cfg.Automation.Timeout)
	}
	for name, credential := range cfg.Automation.Credentials {
		value, err := config.ReadSecret(ctx, credential.Secret.In(root.Dir))
		if err != nil {
			return nil, fmt.Errorf("automation credential %s: could not read secret", name)
		}
		secret := strings.TrimSpace(string(value))
		if secret == "" {
			return nil, fmt.Errorf("automation credential %s: empty secret", name)
		}
		options.Auth[name] = automation.Credential{Origins: credential.Origins, Headers: map[string]string{credential.Header: credential.Prefix + secret}, Secrets: []string{secret}}
		secrets = append(secrets, secret)
	}
	if cfg.Jev.Enabled {
		value, err := config.ReadSecret(ctx, cfg.Jev.APIKey.In(root.Dir))
		if err != nil {
			return nil, errors.New("jev: could not read API key")
		}
		key := strings.TrimSpace(string(value))
		if key == "" {
			return nil, errors.New("jev: empty API key")
		}
		options.Jev = automation.JevOptions{Enabled: true, Endpoint: cfg.Jev.Endpoint, APIKey: key}
		secrets = append(secrets, key)
	}
	entries := []string{}
	for _, w := range cfg.Workflows {
		entries = append(entries, w.OnStart, w.OnStepCompleted)
	}
	for _, tr := range cfg.Triggers {
		entries = append(entries, tr.Handler)
	}
	programs := map[string]*Program{}
	for _, entry := range slices.Sorted(slices.Values(entries)) {
		if entry == "" || programs[entry] != nil {
			continue
		}
		file, fn, ok := strings.Cut(entry, ":")
		if !ok || fn == "" {
			return nil, fmt.Errorf("hook %q: expected file.star:function", entry)
		}
		if !filepath.IsAbs(file) {
			file = filepath.Join(root.Dir, file)
		}
		dir, local := filepath.Dir(file), filepath.Base(file)+":"+fn
		engine, err := automation.Load(ctx, dir, []string{local}, options)
		if err != nil {
			return nil, fmt.Errorf("hook %s: %w", entry, err)
		}
		digest, err := hookManifestDigest(cfg, root, entry, dir, engine.Digest(), options, true)
		if err != nil {
			return nil, err
		}
		authority, err := hookManifestDigest(cfg, root, entry, dir, engine.Digest(), options, false)
		if err != nil {
			return nil, err
		}
		programs[entry] = &Program{engine: engine, entry: local, root: dir, digest: digest, authority: authority, secrets: secrets}
	}
	return programs, nil
}

// Source bytes alone do not identify a callback's scope or host authority.
// Without surroundings, owners, instances and the pool and workflow maps
// are left out; a run pins its reference scope separately (RunScope).
func hookManifestDigest(cfg *config.Config, root config.Root, entry, dir, modules string, options automation.Options, surroundings bool) (string, error) {
	rootDir, err := filepath.Abs(root.Dir)
	if err != nil {
		return "", err
	}
	assetDir, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	resolvedRoot, err := filepath.EvalSymlinks(rootDir)
	if err != nil {
		return "", err
	}
	resolvedAsset, err := filepath.EvalSymlinks(assetDir)
	if err != nil {
		return "", err
	}
	owners := map[string]any{}
	for name, w := range cfg.Workflows {
		if w.OnStart == entry || w.OnStepCompleted == entry {
			owners["workflows."+name] = w
		}
	}
	for name, tr := range cfg.Triggers {
		if tr.Handler == entry {
			owners["triggers."+name] = tr
		}
	}
	credentials := map[string]string{}
	pools, workflows := map[string]string{}, map[string]string{}
	for name, pool := range cfg.Pools {
		pools[name] = pool.Project
	}
	for name, workflow := range cfg.Workflows {
		workflows[name] = workflow.Project
	}
	for name, credential := range options.Auth {
		data, err := json.Marshal(credential)
		if err != nil {
			return "", err
		}
		credentials[name] = hookKey(string(data))
	}
	manifest := struct {
		Version           int
		Root              string
		AssetRoot         string
		ResolvedRoot      string
		ResolvedAssetRoot string
		Entry             string
		Modules           string
		Owners            map[string]any
		Instances         []config.PackInstance
		Pools             map[string]string
		Workflows         map[string]string
		Automation        config.Automation
		Jev               config.Jev
		Credentials       map[string]string
		JevKey            string
	}{1, rootDir, assetDir, resolvedRoot, resolvedAsset, entry, modules, owners, append([]config.PackInstance{}, cfg.Instances...), pools, workflows, cfg.Automation, cfg.Jev, credentials, hookKey(options.Jev.APIKey)}
	if !surroundings {
		manifest.Version, manifest.Owners, manifest.Instances, manifest.Pools, manifest.Workflows = 2, nil, nil, nil, nil
	}
	// Match the last-good serializer's empty collection normalization.
	encoded, err := yaml.Marshal(manifest)
	if err != nil {
		return "", err
	}
	var canonical map[string]any
	if err := yaml.Unmarshal(encoded, &canonical); err != nil {
		return "", err
	}
	data, err := json.Marshal(canonical)
	if err != nil {
		return "", err
	}
	return hookKey(string(data)), nil
}

// WorkflowDigest is the full manifest pin of runs started before RunPin.
func WorkflowDigest(programs map[string]*Program, w config.Workflow) string {
	a, b := programs[w.OnStart], programs[w.OnStepCompleted]
	if a == nil || b == nil {
		return ""
	}
	return a.digest + ":" + b.digest
}

// runPinPrefix marks a RunPin, so a pinned run cannot match the legacy form.
const runPinPrefix = "run2:"

// ScopedRunPin reports whether pinned is a RunPin, whose run must carry
// the reference scope it accepted at start.
func ScopedRunPin(pinned string) bool { return strings.HasPrefix(pinned, runPinPrefix) }

// RunPin pins a callback workflow run to its owner and the owner's pack
// instance, selectors, code, roots and host grants. Its reference scope is
// pinned separately in the run.
func RunPin(programs map[string]*Program, cfg *config.Config, name string) string {
	w := cfg.Workflows[name]
	a, b := programs[w.OnStart], programs[w.OnStepCompleted]
	if a == nil || b == nil {
		return ""
	}
	// Host-level workflows belong to no instance: both stay empty.
	var path, prefix string
	for _, in := range cfg.Instances {
		if slices.Contains(in.Workflows, name) {
			path, prefix = in.Path, in.Prefix
			break
		}
	}
	data, err := json.Marshal([]string{name, w.Project, path, prefix, w.OnStart, w.OnStepCompleted, a.authority, b.authority})
	if err != nil {
		return ""
	}
	return runPinPrefix + hookKey(string(data))
}

// RunPinned reports whether the active configuration still runs pinned for
// workflow name: the RunPin, or the full manifest of an older run.
func RunPinned(programs map[string]*Program, cfg *config.Config, name, pinned string) bool {
	w, ok := cfg.Workflows[name]
	if !ok || pinned == "" {
		return false
	}
	if ScopedRunPin(pinned) {
		return RunPin(programs, cfg, name) == pinned
	}
	return WorkflowDigest(programs, w) == pinned
}
