// Package ids generates ticket ids: a per-project prefix plus a Sonyflake id
// in fixed-width lowercase Crockford base32, e.g. "js-0031w02j0r012".
// Fixed width means ids sort by creation time as strings.
package ids

import (
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"strings"
	"time"

	"github.com/sony/sonyflake/v2"
)

// Crockford's alphabet, lowercase: no i, l, o, u.
var encoding = base32.NewEncoding("0123456789abcdefghjkmnpqrstvwxyz").WithPadding(base32.NoPadding)

// epoch is fixed forever: changing it would break ordering of existing ids.
var epoch = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

// Generator mints ids. Only the controller holds one, and the controller lock
// guarantees a single writer, so the machine id is constant.
type Generator struct {
	sf *sonyflake.Sonyflake
}

// New returns a Generator.
func New() (*Generator, error) {
	sf, err := sonyflake.New(sonyflake.Settings{
		TimeUnit:      time.Millisecond, // 47 time bits: about 4,400 years from epoch
		BitsSequence:  12,               // 4,096 ids per ms
		BitsMachineID: 4,
		StartTime:     epoch,
		MachineID:     func() (int, error) { return 1, nil },
	})
	if err != nil {
		return nil, fmt.Errorf("ids: %w", err)
	}
	return &Generator{sf: sf}, nil
}

// SessionTag marks session ids: "<prefix>-s-<sonyflake>" (#615). Ids minted
// before #615 have no tag, so the shape only proves a new session id.
const SessionTag = "s"

// Next returns a new id with the given project prefix and kind tag. An empty
// tag gives "<prefix>-<sonyflake>"; a tag gives "<prefix>-<tag>-<sonyflake>".
func (g *Generator) Next(prefix, tag string) (string, error) {
	n, err := g.sf.NextID()
	if err != nil {
		return "", fmt.Errorf("ids: %w", err)
	}
	if tag != "" {
		return prefix + "-" + tag + "-" + encode(n), nil
	}
	return prefix + "-" + encode(n), nil
}

// IsSession reports whether id has the session tag. False for ticket ids and
// for session ids minted before #615: callers fall back to a store lookup.
func IsSession(id string) bool {
	// Match from the end: a project prefix may itself contain dashes.
	n := len(id) - encodedLen
	return n > len(SessionTag)+2 && strings.HasSuffix(id[:n], "-"+SessionTag+"-")
}

// encodedLen is the width of an encoded 64-bit id.
const encodedLen = 13

func encode(n int64) string {
	var b [8]byte
	binary.BigEndian.PutUint64(b[:], uint64(n))
	return encoding.EncodeToString(b[:])
}
