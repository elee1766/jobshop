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

// sqlcQueries returns every named query text in sqldb.
func sqlcQueries(t *testing.T) []string {
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
						out = append(out, s)
					}
				}
			}
		}
	}
	return out
}

// TestEveryQueryPrepares prepares each sqlc query, as rewritten, on a
// migrated database: a sqlite query or column Postgres cannot run fails
// here by name instead of deep in the conformance suite.
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
	rw := newRewriter(parseOverrides(overridesSQL))
	for _, q := range sqlcQueries(t) {
		if _, err := c.Prepare(ctx, "", rw.rewrite(q)); err != nil {
			t.Errorf("%s: %v", queryName(q), err)
		}
	}
}

func TestOverridesNameQueries(t *testing.T) {
	names := map[string]bool{}
	for _, q := range sqlcQueries(t) {
		names[queryName(q)] = true
	}
	for name := range parseOverrides(overridesSQL) {
		require.True(t, names[name], "override %s names no sqldb query", name)
	}
}

func TestTranslate(t *testing.T) {
	for in, want := range map[string]string{
		"SELECT ?1, ?2, ?1":                                    "SELECT $1, $2, $1",
		"WHERE a = ? AND b IN (?,?) AND ?":                     "WHERE a = $1 AND b IN ($2,$3) AND $4",
		"WHERE a = ?2 AND b = ?":                               "WHERE a = $2 AND b = $3",
		"SELECT '?' -- what?\n, ?":                             "SELECT '?' -- what?\n, $1",
		"LIMIT CAST(?3 AS INTEGER)":                            "LIMIT CAST($3 AS BIGINT)",
		"SELECT 'it''s ?', ?":                                  "SELECT 'it''s ?', $1",
		"WHERE (?1 IS NULL OR a = ?1)":                         "WHERE (CAST($1 AS TEXT) IS NULL OR a = $1)",
		"SELECT CAST(EXISTS (SELECT f(?)) AS INTEGER) AS x, ?": "SELECT CAST(CAST(EXISTS (SELECT f($1)) AS INT) AS BIGINT) AS x, $2",
	} {
		require.Equal(t, want, translate(in), in)
	}
}
