package control

import (
	"context"
	"time"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/storage/journal/compress"
	"github.com/elee1766/jobshop/pkg/storage/store"
)

// compressor is the journal compression sweeper (#631,
// pkg/storage/journal/compress), made on first use so controllers built
// without Run get one too.
func (c *Controller) compressor() *compress.Sweeper {
	c.compressOnce.Do(func() {
		c.compress = compress.New(compressHost{c}, c.log)
	})
	return c.compress
}

// maybeCompressJournals is the tick step: it starts a sweep when due.
func (c *Controller) maybeCompressJournals(_ context.Context, passCfg *config.Config) {
	c.compressor().Maybe(passCfg.Journal.Compress)
}

// compressHost is the controller as compress.Host.
type compressHost struct{ c *Controller }

func (h compressHost) Now() time.Time              { return h.c.now() }
func (h compressHost) SessionDir(id string) string { return h.c.root.SessionDir(id) }
func (h compressHost) Go(f func())                 { h.c.wg.Go(f) }
func (h compressHost) Context() context.Context    { return h.c.ctx }

func (h compressHost) List(ctx context.Context, f store.Filter) ([]domain.Ticket, error) {
	return h.c.store.List(ctx, f)
}
