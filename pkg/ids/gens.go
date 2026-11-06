package ids

import (
	"crypto/rand"
	"encoding/base32"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/sony/sonyflake/v2"
)

// Crockford's alphabet, lowercase: no i, l, o, u.
const crockford = "0123456789abcdefghjkmnpqrstvwxyz"

var encoding = base32.NewEncoding(crockford).WithPadding(base32.NoPadding)

func inAlphabet(s, alpha string, n int) bool {
	if len(s) != n {
		return false
	}
	for i := range len(s) {
		if !strings.ContainsRune(alpha, rune(s[i])) {
			return false
		}
	}
	return true
}

// sonyflake: fixed epoch 2026-01-01, ms units, 12 sequence bits, machine
// id 1, 13 chars of Crockford base32. Sorts by time.

// epoch is fixed forever: changing it would break ordering of existing ids.
var epoch = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

type sonyflakeGen struct{ sf *sonyflake.Sonyflake }

func newSonyflake() (Generator, error) {
	// Only the controller mints ids and its lock guarantees a single
	// writer, so the machine id is constant.
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
	return &sonyflakeGen{sf: sf}, nil
}

func (g *sonyflakeGen) Next(prefix, tag string) (string, error) {
	n, err := g.sf.NextID()
	if err != nil {
		return "", fmt.Errorf("ids: %w", err)
	}
	var b [8]byte
	binary.BigEndian.PutUint64(b[:], uint64(n))
	return join(prefix, tag, encoding.EncodeToString(b[:])), nil
}

func (*sonyflakeGen) Sorted() bool           { return true }
func (*sonyflakeGen) Match(body string) bool { return inAlphabet(body, crockford, 13) }

// seq: a per-prefix counter, "js-1". The store mints the number inside
// the insert transaction (SeqID); Next alone cannot. Numbers sort
// numerically, not as strings ("js-10" < "js-9").

type seqGen struct{}

// SeqMax bounds seq bodies to 12 digits, so they never have the
// 13-char sonyflake shape.
const SeqMax = 999_999_999_999

func (seqGen) Next(string, string) (string, error) {
	return "", fmt.Errorf("ids: seq ids are minted by the store")
}
func (seqGen) Sorted() bool { return false }
func (seqGen) Match(body string) bool {
	_, ok := SeqNumber(body)
	return ok
}

// SeqID formats the n-th seq id.
func SeqID(prefix, tag string, n int64) string {
	return join(prefix, tag, strconv.FormatInt(n, 10))
}

// SeqStart is the first number a new seq counter for prefix hands out:
// one past the highest seq-shaped id already under it (any tag), so
// switching an existing root to seq never reuses a number.
func SeqStart(prefix string, existing []string) int64 {
	var top int64
	for _, id := range existing {
		rest, ok := strings.CutPrefix(id, prefix+"-")
		if !ok {
			continue
		}
		if i := strings.LastIndexByte(rest, '-'); i >= 0 {
			rest = rest[i+1:]
		}
		if n, ok := SeqNumber(rest); ok && n > top {
			top = n
		}
	}
	return top + 1
}

// SeqNumber parses a seq body: 1 to SeqMax with no leading zero.
func SeqNumber(body string) (int64, bool) {
	if body == "" || len(body) > 12 || body[0] == '0' {
		return 0, false
	}
	for i := range len(body) {
		if body[i] < '0' || body[i] > '9' {
			return 0, false
		}
	}
	n, err := strconv.ParseInt(body, 10, 64)
	return n, err == nil
}

// IsSeq reports whether g is the seq generator.
func IsSeq(g Generator) bool {
	_, ok := g.(seqGen)
	return ok
}

// ulid: 48-bit ms time plus 80 random bits, 26 chars of Crockford base32,
// monotonic within a millisecond. Sorts by time.

type ulidGen struct {
	mu   sync.Mutex
	ms   uint64
	rand [10]byte
}

func (g *ulidGen) Next(prefix, tag string) (string, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	ms := uint64(time.Now().UnixMilli())
	if ms <= g.ms {
		// Same (or earlier) ms: increment the random part.
		ms = g.ms
		i := len(g.rand) - 1
		for ; i >= 0; i-- {
			g.rand[i]++
			if g.rand[i] != 0 {
				break
			}
		}
		if i < 0 {
			return "", fmt.Errorf("ids: ulid overflow in one ms")
		}
	} else if _, err := rand.Read(g.rand[:]); err != nil {
		return "", err
	}
	g.ms = ms
	// 128 bits as 26 chars: 2 leading pad bits, then 5 bits per char.
	var b [16]byte
	binary.BigEndian.PutUint16(b[0:], uint16(ms>>32))
	binary.BigEndian.PutUint32(b[2:], uint32(ms))
	copy(b[6:], g.rand[:])
	hi, lo := binary.BigEndian.Uint64(b[:8]), binary.BigEndian.Uint64(b[8:])
	var out [26]byte
	for i := 25; i >= 0; i-- {
		out[i] = crockford[lo&31]
		lo = lo>>5 | hi<<59
		hi >>= 5
	}
	return join(prefix, tag, string(out[:])), nil
}

func (*ulidGen) Sorted() bool { return true }
func (*ulidGen) Match(body string) bool {
	// The first char holds 3 bits.
	return inAlphabet(body, crockford, 26) && body[0] <= '7'
}

// uuid: RFC 9562 text form. v7 (time-ordered, monotonic within a ms like
// ulid) sorts by time; v4 (random) does not.

type uuidGen struct {
	v7 bool
	mu sync.Mutex
	ms uint64
	sq uint16
}

func (g *uuidGen) Next(prefix, tag string) (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	if g.v7 {
		g.mu.Lock()
		ms := uint64(time.Now().UnixMilli())
		if ms <= g.ms {
			ms = g.ms
			g.sq++
			if g.sq > 0xfff {
				ms++ // borrow the next ms rather than repeat
				g.sq = 0
			}
		} else {
			g.sq = 0
		}
		g.ms = ms
		sq := g.sq
		g.mu.Unlock()
		binary.BigEndian.PutUint16(b[0:], uint16(ms>>32))
		binary.BigEndian.PutUint32(b[2:], uint32(ms))
		// rand_a holds a counter so ids in one ms stay ordered.
		b[6] = 0x70 | byte(sq>>8)
		b[7] = byte(sq)
	} else {
		b[6] = b[6]&0x0f | 0x40
	}
	b[8] = b[8]&0x3f | 0x80
	h := hex.EncodeToString(b[:])
	return join(prefix, tag, h[0:8]+"-"+h[8:12]+"-"+h[12:16]+"-"+h[16:20]+"-"+h[20:]), nil
}

func (g *uuidGen) Sorted() bool { return g.v7 }
func (*uuidGen) Match(body string) bool {
	if len(body) != 36 {
		return false
	}
	for i := range len(body) {
		c := body[i]
		switch i {
		case 8, 13, 18, 23:
			if c != '-' {
				return false
			}
		default:
			if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
				return false
			}
		}
	}
	return true
}

// xid: rs/xid layout, 4 bytes of Unix seconds, 3 machine, 2 pid, 3
// counter, as 20 chars of lowercase base32hex. Sorts by time (to the
// second, then by counter).

const base32hex = "0123456789abcdefghijklmnopqrstuv"

var hexEncoding = base32.NewEncoding(base32hex).WithPadding(base32.NoPadding)

type xidGen struct {
	mu   sync.Mutex
	head [5]byte // machine and pid
	n    uint32
}

func newXID() *xidGen {
	g := &xidGen{}
	var r [7]byte
	_, _ = rand.Read(r[:])
	copy(g.head[:3], r[:3])
	binary.BigEndian.PutUint16(g.head[3:], uint16(os.Getpid()))
	g.n = binary.BigEndian.Uint32(r[3:])
	return g
}

func (g *xidGen) Next(prefix, tag string) (string, error) {
	g.mu.Lock()
	g.n++
	n := g.n
	g.mu.Unlock()
	var b [12]byte
	binary.BigEndian.PutUint32(b[0:], uint32(time.Now().Unix()))
	copy(b[4:9], g.head[:])
	b[9], b[10], b[11] = byte(n>>16), byte(n>>8), byte(n)
	return join(prefix, tag, hexEncoding.EncodeToString(b[:])), nil
}

func (*xidGen) Sorted() bool           { return true }
func (*xidGen) Match(body string) bool { return inAlphabet(body, base32hex, 20) }
