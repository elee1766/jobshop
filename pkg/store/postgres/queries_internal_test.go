package postgres

import (
	"context"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

// sqldbQueries returns every named query text in sqldb, with the
// LabelsFor list filled in.
func sqldbQueries(t *testing.T) []string {
	t.Helper()
	files, err := filepath.Glob("../sqlstore/sqldb/*.go")
	require.NoError(t, err)
	require.NotEmpty(t, files)
	var out []string
	fset := token.NewFileSet()
	for _, f := range files {
		af, err := parser.ParseFile(fset, f, nil, 0)
		require.NoError(t, err)
		for _, d := range af.Decls {
			gd, ok := d.(*ast.GenDecl)
			if !ok || gd.Tok != token.CONST {
				continue
			}
			for _, sp := range gd.Specs {
				for _, v := range sp.(*ast.ValueSpec).Values {
					lit, ok := v.(*ast.BasicLit)
					if !ok || lit.Kind != token.STRING {
						continue
					}
					s, err := strconv.Unquote(lit.Value)
					require.NoError(t, err)
					if strings.HasPrefix(s, "-- name: ") {
						out = append(out, strings.ReplaceAll(s, "/*SLICE:ids*/", "$1,$2"))
					}
				}
			}
		}
	}
	require.NotEmpty(t, out)
	return out
}

// queryName returns X from a leading "-- name: X :kind" line.
func queryName(q string) string {
	rest := strings.TrimPrefix(q, "-- name: ")
	if i := strings.IndexAny(rest, " \n"); i >= 0 {
		rest = rest[:i]
	}
	return rest
}

// TestEveryQueryPrepares prepares each sqldb query as is on a migrated
// Postgres database (#694: no rewriting), so a query Postgres cannot run
// fails here by name instead of deep in the conformance suite. SQLite
// runs the same text in every sqlite test.
func TestEveryQueryPrepares(t *testing.T) {
	dsn := os.Getenv("JOBSHOP_TEST_POSTGRES")
	if dsn == "" {
		t.Skip("JOBSHOP_TEST_POSTGRES is not set")
	}
	ctx := context.Background()
	s, err := Open(ctx, dsn, Options{})
	require.NoError(t, err)
	require.NoError(t, s.Close())
	c, err := pgx.Connect(ctx, dsn)
	require.NoError(t, err)
	defer func() { _ = c.Close(ctx) }()
	for _, q := range append(sqldbQueries(t), recallSQL) {
		if _, err := c.Prepare(ctx, "", q); err != nil {
			t.Errorf("%s: %v", queryName(q), err)
		}
	}
}

// TestNoSQLitePlaceholders keeps sqldb on $N: Postgres rejects ? and ?N.
func TestNoSQLitePlaceholders(t *testing.T) {
	for _, q := range sqldbQueries(t) {
		for _, line := range strings.Split(q, "\n") {
			if i := strings.Index(line, "--"); i >= 0 {
				line = line[:i]
			}
			require.NotContains(t, line, "?", queryName(q))
			require.NotContains(t, strings.ToUpper(line), " AS INTEGER)", queryName(q))
		}
	}
}
