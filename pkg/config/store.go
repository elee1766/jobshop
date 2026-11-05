package config

import (
	"bytes"
	"context"
	"fmt"
)

// Store drivers (#493).
const (
	StoreSQLite   = "sqlite"
	StorePostgres = "postgres"
)

// Store selects the ticket store backend (#493). The default is SQLite
// at Root.DB. The controller reads it at start; a change takes effect
// on restart.
type Store struct {
	// Driver is "sqlite" (default) or "postgres".
	Driver string `yaml:"driver"`
	// DSNFile holds the Postgres connection string (a URL or libpq
	// key=value string), kept out of the config because it may carry
	// a password. Required for postgres.
	DSNFile Secret `yaml:"dsn_file"`
}

// DriverName is Driver, or sqlite when unset.
func (s Store) DriverName() string {
	if s.Driver == "" {
		return StoreSQLite
	}
	return s.Driver
}

func (s Store) check() error {
	switch s.DriverName() {
	case StoreSQLite:
		if !s.DSNFile.IsZero() {
			return fmt.Errorf("store: dsn_file is for driver postgres")
		}
	case StorePostgres:
		if s.DSNFile.IsZero() {
			return fmt.Errorf("store: driver postgres needs dsn_file")
		}
	default:
		return fmt.Errorf("store: unknown driver %q (want sqlite or postgres)", s.Driver)
	}
	return nil
}

// DSN reads the Postgres connection string, with relative paths taken
// from dir.
func (s Store) DSN(ctx context.Context, dir string) (string, error) {
	b, err := ReadSecret(ctx, s.DSNFile.In(dir))
	if err != nil {
		return "", fmt.Errorf("store: dsn_file: %w", err)
	}
	dsn := string(bytes.TrimSpace(b))
	if dsn == "" {
		return "", fmt.Errorf("store: dsn_file %s is empty", s.DSNFile)
	}
	return dsn, nil
}
