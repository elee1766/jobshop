package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"golang.org/x/sys/unix"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/store/postgres"
	"github.com/elee1766/jobshop/pkg/store/sqlite"
	"github.com/elee1766/jobshop/pkg/store/storemigrate"
)

// adminCmd holds offline maintenance that runs without a controller.
type adminCmd struct {
	Store adminStoreCmd `cmd:"" help:"Ticket store maintenance."`
}

type adminStoreCmd struct {
	Migrate storeMigrateCmd `cmd:"" help:"Copy a SQLite store into Postgres (#632). The controller must be stopped."`
}

// storeMigrateCmd opens the databases itself: it runs while the
// controller is stopped, so there is no controller to go through. It
// takes the controller lock so a running controller makes it fail.
type storeMigrateCmd struct {
	From    string `help:"Source store, sqlite:<path>. Default: the root's jobshop.db." placeholder:"sqlite:PATH"`
	To      string `help:"Target driver." enum:"postgres" default:"postgres"`
	DSNFile string `name:"dsn-file" help:"File holding the Postgres DSN (URL or key=value). Default: the root config's store.dsn_file." type:"path"`
	Force   bool   `help:"Replace the target's rows instead of refusing a non-empty target."`
}

func (c storeMigrateCmd) Run(g *Globals) error {
	r, err := g.root()
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	src := r.DB()
	if c.From != "" {
		p, ok := strings.CutPrefix(c.From, "sqlite:")
		if !ok || p == "" {
			return fmt.Errorf("--from %q: want sqlite:<path>", c.From)
		}
		src = r.Path(p)
	}
	if _, err := os.Stat(src); err != nil {
		return fmt.Errorf("source: %w", err)
	}
	sc := config.Store{Driver: config.StorePostgres, DSNFile: config.SecretPath(c.DSNFile)}
	if c.DSNFile == "" {
		cfg, err := config.Load(r)
		if err != nil {
			return fmt.Errorf("no --dsn-file and the config does not load: %w", err)
		}
		if cfg.Store.DSNFile.IsZero() {
			return errors.New("no --dsn-file and the config has no store.dsn_file")
		}
		sc.DSNFile = cfg.Store.DSNFile
	}
	dsn, err := sc.DSN(ctx, r.Dir)
	if err != nil {
		return err
	}

	unlock, err := lockController(r.Lock())
	if err != nil {
		return err
	}
	defer unlock()

	from, err := sqlite.Open(ctx, src, sqlite.Options{})
	if err != nil {
		return fmt.Errorf("open %s: %w", src, err)
	}
	defer func() { _ = from.Close() }()
	to, err := postgres.Open(ctx, dsn, postgres.Options{})
	if err != nil {
		return err
	}
	defer func() { _ = to.Close() }()
	res, err := storemigrate.Copy(ctx, from.DB(), to.DB(), storemigrate.Options{Force: c.Force})
	if err != nil {
		return err
	}
	for _, t := range storemigrate.Tables {
		fmt.Printf("%-20s %d\n", t, res[t])
	}
	fmt.Println("ok: row counts match. Set store.driver: postgres and start the controller.")
	return nil
}

// lockController takes the controller lock without blocking, as the
// controller does at start, so a migrate and a controller never run at
// once.
func lockController(path string) (func(), error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, err
	}
	if err := unix.Flock(int(f.Fd()), unix.LOCK_EX|unix.LOCK_NB); err != nil {
		_ = f.Close()
		if errors.Is(err, unix.EWOULDBLOCK) {
			return nil, fmt.Errorf("a controller holds %s: stop it first", path)
		}
		return nil, err
	}
	return func() {
		_ = unix.Flock(int(f.Fd()), unix.LOCK_UN)
		_ = f.Close()
	}, nil
}
