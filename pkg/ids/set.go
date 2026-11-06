package ids

import (
	"fmt"
	"sync"
)

// Set opens generators by spec and keeps them, so a generator's
// monotonic state survives across calls and a config reload that keeps
// the spec.
type Set struct {
	mu   sync.Mutex
	gens map[Spec]Generator
}

// Get returns the generator for s.
func (c *Set) Get(s Spec) (Generator, error) {
	s = s.Normal()
	c.mu.Lock()
	defer c.mu.Unlock()
	if g, ok := c.gens[s]; ok {
		return g, nil
	}
	g, err := Open(s)
	if err != nil {
		return nil, err
	}
	if c.gens == nil {
		c.gens = map[Spec]Generator{}
	}
	c.gens[s] = g
	return g, nil
}

// mintTries bounds collision retries: words ids that are taken, and seq
// numbers already used by an explicit id.
const mintTries = 1000

// Mint returns a new id from g. The store calls it inside the insert
// transaction: seq bumps the prefix counter there (one writer holds the
// lock, so numbers are not handed out twice), and exists checks the id is
// free so a words or seq id is retried instead of colliding.
func Mint(g Generator, prefix, tag string, seq func(prefix string) (int64, error), exists func(id string) (bool, error)) (string, error) {
	check := IsSeq(g) || Retries(g)
	for range mintTries {
		var id string
		if IsSeq(g) {
			n, err := seq(prefix)
			if err != nil {
				return "", err
			}
			if n > SeqMax {
				return "", fmt.Errorf("ids: seq counter for %q is past %d", prefix, int64(SeqMax))
			}
			id = SeqID(prefix, tag, n)
		} else {
			var err error
			if id, err = g.Next(prefix, tag); err != nil {
				return "", err
			}
		}
		if !check {
			return id, nil
		}
		taken, err := exists(id)
		if err != nil {
			return "", err
		}
		if !taken {
			return id, nil
		}
	}
	return "", fmt.Errorf("ids: no free id for %q after %d tries", prefix, mintTries)
}
