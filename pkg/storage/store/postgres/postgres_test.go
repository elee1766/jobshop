package postgres_test

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"strings"
	"sync/atomic"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib" // database/sql driver "pgx"
	"github.com/stretchr/testify/require"

	"github.com/gfx-labs/jobshop/pkg/domain"
	"github.com/gfx-labs/jobshop/pkg/storage/store"
	"github.com/gfx-labs/jobshop/pkg/storage/store/postgres"
	"github.com/gfx-labs/jobshop/pkg/storage/store/storetest"
)

// dsnEnv names a server the tests may create databases on. Unset, the
// tests skip; CI sets it to a Postgres service container (#493).
const dsnEnv = "JOBSHOP_TEST_POSTGRES"

const templateDB = "jobshop_test_template"

var (
	admin *sql.DB
	seq   atomic.Int64
)

func TestMain(m *testing.M) {
	code, err := run(m)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	os.Exit(code)
}

// run migrates a template database once; each test clones it with
// CREATE DATABASE ... TEMPLATE, which is faster than migrating.
func run(m *testing.M) (int, error) {
	dsn := os.Getenv(dsnEnv)
	if dsn == "" {
		return m.Run(), nil
	}
	ctx := context.Background()
	var err error
	if admin, err = sql.Open("pgx", dsn); err != nil {
		return 0, err
	}
	defer func() { _ = admin.Close() }()
	if _, err := admin.ExecContext(ctx, "DROP DATABASE IF EXISTS "+templateDB+" WITH (FORCE)"); err != nil {
		return 0, err
	}
	if _, err := admin.ExecContext(ctx, "CREATE DATABASE "+templateDB); err != nil {
		return 0, err
	}
	s, err := postgres.Open(ctx, withDB(dsn, templateDB), postgres.Options{})
	if err != nil {
		return 0, err
	}
	if err := s.Close(); err != nil {
		return 0, err
	}
	return m.Run(), nil
}

// withDB returns dsn (a URL) pointing at database name.
func withDB(dsn, name string) string {
	u, err := url.Parse(dsn)
	if err != nil || u.Scheme == "" {
		return dsn + " dbname=" + name
	}
	u.Path = "/" + name
	return u.String()
}

// openTest returns a store on a fresh database cloned from the template.
func openTest(t *testing.T, opt postgres.Options) *postgres.Store {
	t.Helper()
	s, _ := openNamed(t, opt)
	return s
}

// openNamed is openTest that also returns the database name.
func openNamed(t *testing.T, opt postgres.Options) (*postgres.Store, string) {
	t.Helper()
	dsn := os.Getenv(dsnEnv)
	if dsn == "" {
		t.Skip(dsnEnv + " is not set")
	}
	name := fmt.Sprintf("jobshop_test_%d_%d", os.Getpid(), seq.Add(1))
	ctx := context.WithoutCancel(t.Context())
	_, err := admin.ExecContext(ctx, "CREATE DATABASE "+name+" TEMPLATE "+templateDB)
	require.NoError(t, err)
	s, err := postgres.Open(ctx, withDB(dsn, name), opt)
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = s.Close()
		_, _ = admin.ExecContext(ctx, "DROP DATABASE IF EXISTS "+name+" WITH (FORCE)")
	})
	return s, name
}

func TestConformance(t *testing.T) {
	t.Parallel()
	storetest.Run(t, func(t *testing.T, c *storetest.Clock) store.Store {
		return openTest(t, postgres.Options{Now: c.Now})
	})
}

func TestRecall(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	s := openTest(t, postgres.Options{})
	tk, err := store.Create(ctx, s, "admin", store.NewTicket{Project: "p", Kind: domain.KindTask, Title: "t"})
	require.NoError(t, err)
	_, err = store.AddComment(ctx, s, "a", tk.ID, "CI failed: golangci-lint gofumpt formatting in pkg/cli")
	require.NoError(t, err)
	_, err = store.AddComment(ctx, s, "a", tk.ID, "unrelated note about the weather")
	require.NoError(t, err)

	hits, err := s.Recall(ctx, store.RecallQuery{Text: "gofumpt formatting"})
	require.NoError(t, err)
	require.Len(t, hits, 1)
	require.Equal(t, store.RecallComment, hits[0].Kind)
	require.Equal(t, tk.ID, hits[0].TicketID)
	require.Equal(t, "p", hits[0].Project)
	require.Contains(t, hits[0].Snippet, "[gofumpt]")
	require.False(t, hits[0].At.IsZero())

	// Any-word fallback and query syntax in the text do not fail.
	hits, err = s.Recall(ctx, store.RecallQuery{Text: `weather "nothing-matches" AND: & |`})
	require.NoError(t, err)
	require.Len(t, hits, 1)
	require.True(t, strings.Contains(hits[0].Snippet, "weather"))

	hits, err = s.Recall(ctx, store.RecallQuery{Text: "gofumpt", Project: "other"})
	require.NoError(t, err)
	require.Empty(t, hits)
}
