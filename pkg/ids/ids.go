// Package ids generates ticket ids: a per-project prefix, an optional kind
// tag and a body from a configurable generator (JOB-57), e.g.
// "js-0031w02j0r012" (sonyflake, the default), "js-12" (seq) or
// "js-s-brisk-otter-lamp" (words, a session).
//
// Every generator's body has its own shape (Generator.Match), and the
// shapes are disjoint, so ids from different generators under one prefix
// never collide and a root can switch generator: only new ids change.
package ids

import (
	"fmt"
	"strings"
)

// Generator mints id bodies.
type Generator interface {
	// Next returns "<prefix>-<body>", or "<prefix>-<tag>-<body>" with a tag.
	Next(prefix, tag string) (string, error)
	// Sorted reports whether ids sort by creation time as strings.
	Sorted() bool
	// Match reports whether body has this generator's shape.
	Match(body string) bool
}

// Generator names.
const (
	Sonyflake = "sonyflake"
	Seq       = "seq"
	ULID      = "ulid"
	UUIDv7    = "uuidv7"
	UUID      = "uuid"
	XID       = "xid"
	Words     = "words"
)

// Names lists the built-in generators.
var Names = []string{Sonyflake, Seq, ULID, UUIDv7, UUID, XID, Words}

// Spec picks a generator; it is the `ids` config block.
type Spec struct {
	// Generator is one of Names. Default sonyflake.
	Generator string `yaml:"generator,omitempty"`
	// Words is the word count for words. Default 3, 2 to 8.
	Words int `yaml:"words,omitempty"`
	// Separator joins words: "-" (default), "_" or ".".
	Separator string `yaml:"separator,omitempty"`
}

// Normal returns s with defaults filled in, so equal settings compare equal.
func (s Spec) Normal() Spec {
	if s.Generator == "" {
		s.Generator = Sonyflake
	}
	if s.Generator != Words {
		return Spec{Generator: s.Generator}
	}
	if s.Words == 0 {
		s.Words = 3
	}
	if s.Separator == "" {
		s.Separator = "-"
	}
	return s
}

// Check validates s.
func (s Spec) Check() error {
	n := s.Normal()
	switch n.Generator {
	case Sonyflake, Seq, ULID, UUIDv7, UUID, XID:
		if s.Words != 0 || s.Separator != "" {
			return fmt.Errorf("ids: words and separator only apply to the words generator")
		}
	case Words:
		if n.Words < 2 || n.Words > 8 {
			return fmt.Errorf("ids: words must be 2 to 8, not %d", n.Words)
		}
		if n.Separator != "-" && n.Separator != "_" && n.Separator != "." {
			return fmt.Errorf(`ids: separator must be "-", "_" or ".", not %q`, n.Separator)
		}
	default:
		return fmt.Errorf("ids: unknown generator %q (one of %s)", s.Generator, strings.Join(Names, ", "))
	}
	return nil
}

// New returns the default (sonyflake) generator.
func New() (Generator, error) { return newSonyflake() }

// Open returns the generator for s.
func Open(s Spec) (Generator, error) {
	if err := s.Check(); err != nil {
		return nil, err
	}
	s = s.Normal()
	switch s.Generator {
	case Seq:
		return seqGen{}, nil
	case ULID:
		return &ulidGen{}, nil
	case UUIDv7:
		return &uuidGen{v7: true}, nil
	case UUID:
		return &uuidGen{}, nil
	case XID:
		return newXID(), nil
	case Words:
		return wordsGen{n: s.Words, sep: s.Separator}, nil
	}
	return newSonyflake()
}

// all has one of each generator, for Match.
var all = func() []Generator {
	var out []Generator
	for _, n := range Names {
		g, err := Open(Spec{Generator: n})
		if err != nil {
			panic(err) // built-ins open without error
		}
		out = append(out, g)
	}
	return out
}()

// SessionTag marks session ids: "<prefix>-s-<body>" (#615). Ids minted
// before #615 have no tag, so the shape only proves a new session id.
const SessionTag = "s"

func join(prefix, tag, body string) string {
	if tag != "" {
		return prefix + "-" + tag + "-" + body
	}
	return prefix + "-" + body
}

// IsSession reports whether id has the session tag. False for ticket ids and
// for session ids minted before #615: callers fall back to a store lookup.
func IsSession(id string) bool {
	// A project prefix may itself contain dashes: try every "-s-".
	mark := "-" + SessionTag + "-"
	for i := strings.Index(id, mark); i >= 0; {
		if i > 0 && isBody(id[i+len(mark):]) {
			return true
		}
		j := strings.Index(id[i+1:], mark)
		if j < 0 {
			break
		}
		i += 1 + j
	}
	return false
}

func isBody(b string) bool {
	for _, g := range all {
		if g.Match(b) {
			return true
		}
	}
	return false
}
