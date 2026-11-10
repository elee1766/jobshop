package ingest

import (
	"context"

	"github.com/gfx-labs/jobshop/pkg/config"
)

// SecretFor returns a func that reads an endpoint's secret from cfg: its
// secret_file relative to the root, or its identity's
// webhook_secret_file as given, like the identity's other files.
func SecretFor(r config.Root, cfg *config.Config) func(config.IngestEndpoint) (Secret, error) {
	return func(e config.IngestEndpoint) (Secret, error) {
		if e.Identity != "" {
			return ReadSecret(context.Background(), e.Identity, cfg.Identities[e.Identity].WebhookSecretFile)
		}
		return ReadSecret(context.Background(), "", e.SecretFile.In(r.Dir))
	}
}
