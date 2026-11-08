package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/goccy/go-yaml"
)

// Last good config (#227). After each successful load the controller
// saves the merged config, packs already applied, so a later start whose
// live packs need a newer binary can keep serving with it.

// LastGoodName is the saved merged config in the root.
const LastGoodName = "config.last-good.yaml"

// LastGood returns the saved last good config path.
func (r Root) LastGood() string { return filepath.Join(r.Dir, LastGoodName) }

type lastGoodSnapshot struct {
	Config   map[string]any    `yaml:"config"`
	Metadata *lastGoodMetadata `yaml:"metadata"`
}

type lastGoodMetadata struct {
	Version     int               `yaml:"version"`
	Instances   []PackInstance    `yaml:"instances"`
	HookDigests map[string]string `yaml:"hook_digests"`
}

// SaveLastGood writes c (merged, no packs list) to path atomically.
func SaveLastGood(path string, c *Config) error {
	metadata := &lastGoodMetadata{Version: 1, Instances: append([]PackInstance{}, c.Instances...), HookDigests: c.HookDigests}
	if err := validateLastGoodMetadata(c, metadata); err != nil {
		return fmt.Errorf("%s: %w", LastGoodName, err)
	}
	cc := *c
	cc.Pack = ""
	cc.Packs = nil
	merged, err := yaml.Marshal(&cc)
	if err != nil {
		return err
	}
	var definitions map[string]any
	if err := yaml.Unmarshal(merged, &definitions); err != nil {
		return err
	}
	b, err := yaml.Marshal(lastGoodSnapshot{Config: definitions, Metadata: metadata})
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// LoadLastGood restores merged definitions and their accepted hook authority.
func LoadLastGood(path string) (*Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	c, err := parseLastGood(b)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", LastGoodName, err)
	}
	return c, nil
}

func parseLastGood(b []byte) (*Config, error) {
	var keys map[string]yaml.RawMessage
	if err := yaml.Unmarshal(b, &keys); err != nil {
		return nil, err
	}
	var metadata *lastGoodMetadata
	if keys["config"] != nil || keys["metadata"] != nil {
		var snapshot lastGoodSnapshot
		if err := yaml.UnmarshalWithOptions(b, &snapshot, yaml.Strict()); err != nil {
			return nil, err
		}
		if len(snapshot.Config) == 0 || snapshot.Metadata == nil {
			return nil, errors.New("last-good snapshot requires config and metadata")
		}
		var err error
		b, err = yaml.Marshal(snapshot.Config)
		if err != nil {
			return nil, err
		}
		metadata = snapshot.Metadata
	}
	c, err := Parse(b)
	if err != nil {
		return nil, err
	}
	if err := validateLastGoodMetadata(c, metadata); err != nil {
		return nil, err
	}
	if metadata != nil {
		c.Instances, c.HookDigests = metadata.Instances, metadata.HookDigests
	}
	return c, nil
}

func validateLastGoodMetadata(c *Config, metadata *lastGoodMetadata) error {
	entries := []string{}
	for _, e := range c.ScriptEntries() {
		entries = append(entries, e.Entry)
	}
	if metadata != nil {
		if metadata.Version != 1 {
			return fmt.Errorf("unsupported last-good metadata version %d", metadata.Version)
		}
		if metadata.Instances == nil {
			return errors.New("last-good metadata requires authoritative instances")
		}
		for _, instance := range metadata.Instances {
			if strings.TrimSpace(instance.Path) == "" || instance.Prefix != "" && !prefixPattern.MatchString(instance.Prefix) {
				return errors.New("last-good metadata contains an invalid pack instance")
			}
		}
	}
	for _, entry := range entries {
		if entry == "" {
			continue
		}
		if metadata == nil || strings.TrimSpace(metadata.HookDigests[entry]) == "" {
			return fmt.Errorf("hook %q requires accepted last-good metadata and digest", entry)
		}
	}
	return nil
}

// LoadOrLastGood loads the root's config. If only a pack fails and a
// last good config exists, it returns that config and the pack error;
// any other failure returns a nil config.
func LoadOrLastGood(r Root) (*Config, error) {
	c, err := Load(r)
	if err == nil {
		return c, nil
	}
	if !errors.Is(err, ErrPack) {
		return nil, err
	}
	lg, lerr := LoadLastGood(r.LastGood())
	if lerr != nil {
		return nil, err
	}
	return lg, err
}
