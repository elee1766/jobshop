package postgres

import (
	"strconv"
	"strings"
	"sync"
)

// rewriteMax bounds the rewrite cache. List queries (LabelsFor) expand to one
// text per list length, so the set is finite but not fixed.
const rewriteMax = 1024

// rewriter turns sqldb's SQLite SQL into Postgres SQL
// (#493): ?N and ? placeholders become $N, INTEGER casts become BIGINT
// (times are milliseconds), and queries named in overrides are replaced
// whole.
type rewriter struct {
	overrides map[string]string
	mu        sync.Mutex
	m         map[string]string
}

func newRewriter(overrides map[string]string) *rewriter {
	return &rewriter{overrides: overrides, m: map[string]string{}}
}

func (r *rewriter) rewrite(q string) string {
	r.mu.Lock()
	out, ok := r.m[q]
	r.mu.Unlock()
	if ok {
		return out
	}
	if o, ok := r.overrides[queryName(q)]; ok {
		out = o
	} else {
		out = translate(q)
	}
	r.mu.Lock()
	if len(r.m) < rewriteMax {
		r.m[q] = out
	}
	r.mu.Unlock()
	return out
}

// queryName returns X from sqldb's leading "-- name: X :kind" line.
func queryName(q string) string {
	const p = "-- name: "
	if !strings.HasPrefix(q, p) {
		return ""
	}
	rest := q[len(p):]
	if i := strings.IndexAny(rest, " \n"); i >= 0 {
		rest = rest[:i]
	}
	return rest
}

// translate rewrites placeholders and casts outside string literals and
// comments. A bare ? takes the number after the highest seen so far, as
// SQLite numbers it, so sqldb's mix of ?N and expanded slices binds the
// same arguments.
func translate(q string) string {
	var b strings.Builder
	b.Grow(len(q) + 16)
	high := 0
	// depth counts open parentheses; exists holds the depth of each
	// open CAST(EXISTS's outer CAST.
	depth, exists := 0, []int(nil)
	for i := 0; i < len(q); {
		c := q[i]
		switch {
		case c == '\'':
			j := i + 1
			for j < len(q) {
				if q[j] == '\'' {
					if j+1 < len(q) && q[j+1] == '\'' {
						j += 2
						continue
					}
					break
				}
				j++
			}
			j = min(j+1, len(q))
			b.WriteString(q[i:j])
			i = j
		case c == '-' && i+1 < len(q) && q[i+1] == '-':
			j := strings.IndexByte(q[i:], '\n')
			if j < 0 {
				j = len(q) - i
			}
			b.WriteString(q[i : i+j])
			i += j
		case c == '?':
			j := i + 1
			for j < len(q) && q[j] >= '0' && q[j] <= '9' {
				j++
			}
			n := high + 1
			if j > i+1 {
				n, _ = strconv.Atoi(q[i+1 : j])
			}
			high = max(high, n)
			// A nullable arg IS NULL: Postgres cannot type a parameter
			// that is only tested for NULL. Every such filter is text.
			if hasFoldPrefix(q[j:], " IS NULL") {
				b.WriteString("CAST($" + strconv.Itoa(n) + " AS TEXT)")
			} else {
				b.WriteString("$" + strconv.Itoa(n))
			}
			i = j
		// CAST(EXISTS (...) AS INTEGER): Postgres casts boolean to INT
		// but not to BIGINT, so the EXISTS gets an inner cast to INT,
		// closed where its parenthesis closes.
		case hasFoldPrefix(q[i:], "CAST(EXISTS"):
			b.WriteString("CAST(CAST(EXISTS")
			depth++
			exists = append(exists, depth)
			i += len("CAST(EXISTS")
		case c == '(':
			depth++
			b.WriteByte(c)
			i++
		case c == ')':
			b.WriteByte(c)
			depth--
			if n := len(exists); n > 0 && exists[n-1] == depth {
				b.WriteString(" AS INT)")
				exists = exists[:n-1]
			}
			i++
		case hasFoldPrefix(q[i:], " AS INTEGER)"):
			b.WriteString(" AS BIGINT)")
			depth--
			i += len(" AS INTEGER)")
		default:
			b.WriteByte(c)
			i++
		}
	}
	return b.String()
}

func hasFoldPrefix(s, p string) bool {
	return len(s) >= len(p) && strings.EqualFold(s[:len(p)], p)
}

// parseOverrides splits a file of "-- name: X" blocks into X -> SQL.
func parseOverrides(src string) map[string]string {
	out := map[string]string{}
	const p = "-- name: "
	blocks := strings.Split(src, "\n"+p)
	for _, blk := range blocks[1:] {
		name := blk
		if i := strings.IndexAny(name, " \n"); i >= 0 {
			name = name[:i]
		}
		out[name] = p + blk
	}
	return out
}
