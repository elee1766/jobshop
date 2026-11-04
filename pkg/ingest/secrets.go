package ingest

import (
	"github.com/elee1766/jobshop/pkg/config"
)

// GitHubSecrets is the GitHub secret from cfg, nil when none is set.
func GitHubSecrets(r config.Root, cfg *config.Config, warn func(string, error)) []Secret {
	if cfg.Ingest.GitHubSecretFile == "" {
		return nil
	}
	s, err := ReadSecret("", r.Path(cfg.Ingest.GitHubSecretFile))
	if err != nil {
		warn("github", err)
		return nil
	}
	return []Secret{s}
}

// LinearSecrets is ingest.linear_secret_file plus the webhook secret of
// every linear-app identity, named after the identity.
func LinearSecrets(r config.Root, cfg *config.Config, warn func(string, error)) []Secret {
	var out []Secret
	if f := cfg.Ingest.LinearSecretFile; f != "" {
		if s, err := ReadSecret("", r.Path(f)); err != nil {
			warn("linear", err)
		} else {
			out = append(out, s)
		}
	}
	for _, name := range cfg.IdentityNames() {
		id := cfg.Identities[name]
		if id.Kind != config.IdentityLinearApp || id.WebhookSecretFile == "" {
			continue
		}
		// As given, like the identity's other files.
		s, err := ReadSecret(name, id.WebhookSecretFile)
		if err != nil {
			warn(name, err)
			continue
		}
		out = append(out, s)
	}
	return out
}
