package config

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Automation grants capabilities to trusted Starlark hooks, not to agent sessions.
type Automation struct {
	HTTP  bool `yaml:"http,omitempty"`
	Shell bool `yaml:"shell,omitempty"`
	// GitHub enables the github built-in, authenticated by the GitHub
	// identity covering the repo of each call.
	GitHub bool `yaml:"github,omitempty"`
	// Linear enables the linear built-in, authenticated by the Linear
	// identity covering the team of each call.
	Linear      bool                            `yaml:"linear,omitempty"`
	Timeout     string                          `yaml:"timeout,omitempty"`
	MaxSteps    uint64                          `yaml:"max_steps,omitempty"`
	OutputBytes int                             `yaml:"output_bytes,omitempty"`
	Credentials map[string]AutomationCredential `yaml:"credentials,omitempty"`
}

// AutomationCredential binds a host-managed header secret to exact HTTP origins.
type AutomationCredential struct {
	Origins []string `yaml:"origins"`
	Header  string   `yaml:"header"`
	Prefix  string   `yaml:"prefix,omitempty"`
	Secret  Secret   `yaml:"secret"`
}

// Jev configures the optional model-decision capability for Starlark hooks.
type Jev struct {
	Enabled  bool   `yaml:"enabled,omitempty"`
	Endpoint string `yaml:"endpoint,omitempty"`
	APIKey   Secret `yaml:"api_key,omitempty"`
}

func (c *Config) checkAutomation() error {
	if c.Automation.Timeout != "" {
		d, err := time.ParseDuration(c.Automation.Timeout)
		if err != nil || d <= 0 {
			return errors.New("automation.timeout must be a positive Go duration")
		}
	}
	if c.Automation.OutputBytes < 0 {
		return errors.New("automation.output_bytes must not be negative")
	}
	for name, credential := range c.Automation.Credentials {
		if !namePattern.MatchString(name) {
			return fmt.Errorf("automation.credentials: invalid name %q", name)
		}
		if len(credential.Origins) == 0 || credential.Secret.IsZero() {
			return fmt.Errorf("automation.credentials.%s: origins and secret are required", name)
		}
		if !validHeaderName(credential.Header) {
			return fmt.Errorf("automation.credentials.%s.header must be an HTTP header name", name)
		}
		if strings.ContainsAny(credential.Prefix, "\r\n") {
			return fmt.Errorf("automation.credentials.%s.prefix must not contain newlines", name)
		}
		for _, origin := range credential.Origins {
			u, err := automationURL(origin)
			if err != nil || (u.Path != "" && u.Path != "/") || u.RawQuery != "" {
				return fmt.Errorf("automation.credentials.%s.origins: expected an HTTP origin without path or query", name)
			}
		}
	}
	if c.Jev.Enabled {
		if _, err := automationURL(c.Jev.Endpoint); err != nil {
			return errors.New("jev.endpoint must be an absolute HTTP(S) URL without credentials or fragment")
		}
		if c.Jev.APIKey.IsZero() {
			return errors.New("jev.api_key is required when Jev is enabled")
		}
	}
	return nil
}

func automationURL(value string) (*url.URL, error) {
	u, err := url.Parse(value)
	if err != nil {
		return nil, err
	}
	if (u.Scheme != "https" && u.Scheme != "http") || u.Hostname() == "" || u.User != nil || u.Fragment != "" {
		return nil, errors.New("expected an absolute HTTP(S) URL without credentials or fragment")
	}
	return u, nil
}

func validHeaderName(name string) bool {
	if name == "" || http.CanonicalHeaderKey(name) == "" {
		return false
	}
	for _, r := range name {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || strings.ContainsRune("!#$%&'*+-.^_`|~", r) {
			continue
		}
		return false
	}
	return true
}
