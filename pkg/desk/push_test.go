package desk

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdh"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/hkdf"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

func unb64(t *testing.T, s string) []byte {
	t.Helper()
	b, err := b64.DecodeString(strings.Join(strings.Fields(s), ""))
	require.NoError(t, err)
	return b
}

// RFC 8291 Section 5 and Appendix A.
func TestEncryptRFC8291(t *testing.T) {
	as, err := ecdh.P256().NewPrivateKey(unb64(t, "yfWPiYE-n46HLnH0KqZOF1fJJU3MYrct3AELtAQ-oRw"))
	require.NoError(t, err)
	got, err := encrypt(
		unb64(t, "V2hlbiBJIGdyb3cgdXAsIEkgd2FudCB0byBiZSBhIHdhdGVybWVsb24"),
		unb64(t, "BCVxsr7N_eNgVRqvHtD0zTZsEc6-VV- JvLexhqUzORcx aOzi6-AYWXvTBHm4bjyPjs7Vd8pZGH6SRpkNtoIAiw4"),
		unb64(t, "BTBZMqHH6r4Tts7J_aSIgg"), as, unb64(t, "DGv6ra1nlYgDCS1FRnbzlw"))
	require.NoError(t, err)
	want := append(unb64(t, `DGv6ra1nlYgDCS1FRnbzlwAAEABBBP4z 9KsN6nGRTbVYI_c7VJSPQTBtkgcy27ml
		mlMoZIIgDll6e3vCYLocInmYWAmS6Tlz AC8wEqKK6PBru3jl7A8`),
		unb64(t, `8pfeW0KbunFT06SuDKoJH9Ql87S1QUrd irN6GcG7sFz1y1sqLgVi1VhjVkHsUoEs bI_0LpXMuGvnzQ`)...)
	assert.Equal(t, want, got)
}

// browser is a fake push service plus user agent: it checks VAPID and
// decrypts what it receives.
type browser struct {
	t    *testing.T
	key  *ecdh.PrivateKey
	auth []byte
	srv  *httptest.Server
	mu   sync.Mutex
	got  []PushMessage
	gone bool
}

func newBrowser(t *testing.T, vapidPub string) *browser {
	b := &browser{t: t, auth: make([]byte, 16)}
	var err error
	b.key, err = ecdh.P256().GenerateKey(rand.Reader)
	require.NoError(t, err)
	_, _ = rand.Read(b.auth)
	b.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if b.gone {
			w.WriteHeader(http.StatusGone)
			return
		}
		b.checkVAPID(r.Header.Get("Authorization"), vapidPub)
		body, _ := io.ReadAll(r.Body)
		var m PushMessage
		require.NoError(t, json.Unmarshal(b.decrypt(body), &m))
		b.mu.Lock()
		b.got = append(b.got, m)
		b.mu.Unlock()
		w.WriteHeader(http.StatusCreated)
	}))
	t.Cleanup(b.srv.Close)
	return b
}

func (b *browser) sub() Subscription {
	var s Subscription
	s.Endpoint = b.srv.URL + "/push/abc"
	s.Keys.P256dh = b64.EncodeToString(b.key.PublicKey().Bytes())
	s.Keys.Auth = b64.EncodeToString(b.auth)
	return s
}

func (b *browser) checkVAPID(h, pub string) {
	t := b.t
	rest, ok := strings.CutPrefix(h, "vapid t=")
	require.True(t, ok, h)
	jwt, k, ok := strings.Cut(rest, ", k=")
	require.True(t, ok)
	assert.Equal(t, pub, k)
	parts := strings.Split(jwt, ".")
	require.Len(t, parts, 3)
	var claims map[string]any
	require.NoError(t, json.Unmarshal(unb64(t, parts[1]), &claims))
	assert.Equal(t, b.srv.URL, claims["aud"])
	pk, err := ecdsa.ParseUncompressedPublicKey(elliptic.P256(), unb64(t, k))
	require.NoError(t, err)
	sig := unb64(t, parts[2])
	sum := sha256.Sum256([]byte(parts[0] + "." + parts[1]))
	assert.True(t, ecdsa.Verify(pk, sum[:], new(big.Int).SetBytes(sig[:32]), new(big.Int).SetBytes(sig[32:])))
}

func (b *browser) decrypt(body []byte) []byte {
	t := b.t
	salt, rs, n := body[:16], binary.BigEndian.Uint32(body[16:20]), int(body[20])
	assert.EqualValues(t, 4096, rs)
	asPub, ct := body[21:21+n], body[21+n:]
	pk, err := ecdh.P256().NewPublicKey(asPub)
	require.NoError(t, err)
	shared, err := b.key.ECDH(pk)
	require.NoError(t, err)
	info := "WebPush: info\x00" + string(b.key.PublicKey().Bytes()) + string(asPub)
	ikm, _ := hkdf.Key(sha256.New, shared, b.auth, info, 32)
	prk, _ := hkdf.Extract(sha256.New, ikm, salt)
	cek, _ := hkdf.Expand(sha256.New, prk, "Content-Encoding: aes128gcm\x00", 16)
	nonce, _ := hkdf.Expand(sha256.New, prk, "Content-Encoding: nonce\x00", 12)
	blk, _ := aes.NewCipher(cek)
	gcm, _ := cipher.NewGCM(blk)
	pt, err := gcm.Open(nil, nonce, ct, nil)
	require.NoError(t, err)
	require.Equal(t, byte(2), pt[len(pt)-1])
	return pt[:len(pt)-1]
}

// TestPushAttention: an item already waiting is not pushed; a new one is,
// once; a gone subscription is dropped; state survives a restart.
func TestPushAttention(t *testing.T) {
	ctx := context.Background()
	file := filepath.Join(t.TempDir(), "push.json")
	p := &Push{File: file}
	require.NoError(t, p.open())
	br := newBrowser(t, p.publicKey())
	require.NoError(t, p.subscribe(br.sub()))

	old := AttentionItem{Kind: "ticket", ID: "js-1", Title: "old", Why: "escalation"}
	assert.Empty(t, p.diff([]AttentionItem{old}))
	fresh := AttentionItem{Kind: "run", ID: "js-2", Title: "deploy blocked", Why: "run needs a human"}
	got := p.diff([]AttentionItem{old, fresh})
	require.Len(t, got, 1)
	p.send(ctx, message(got[0]), t.Logf)
	assert.Equal(t, []PushMessage{{Title: "deploy blocked", Body: "run needs a human", Tag: "run:js-2", URL: "./?item=run%3Ajs-2"}}, br.got)
	assert.Empty(t, p.diff([]AttentionItem{old, fresh}))

	// Restart keeps the key, subscription and sent list.
	p2 := &Push{File: file}
	require.NoError(t, p2.open())
	assert.Equal(t, p.publicKey(), p2.publicKey())
	assert.Empty(t, p2.diff([]AttentionItem{old, fresh}))

	br.gone = true
	p2.send(ctx, PushMessage{Title: "x"}, t.Logf)
	assert.Empty(t, p2.state.Subs)
}

// TestPushRoutes: routes exist only with push on, and need the CSRF header.
func TestPushRoutes(t *testing.T) {
	off := New(Config{}, &fakeCtl{}, nil).Handler()
	assert.Equal(t, http.StatusOK, get(t, off, "/api/push/key", nil).Code) // falls through to the app
	p := &Push{File: filepath.Join(t.TempDir(), "push.json")}
	s := New(Config{Push: p}, &fakeCtl{results: map[string]any{rpc.MethodRunList + ":" + store.RunNeedsHuman: []store.Ticket{}}}, nil)
	require.NoError(t, p.open())
	var k map[string]string
	get(t, s.Handler(), "/api/push/key", &k)
	assert.Equal(t, p.publicKey(), k["key"])
	req := httptest.NewRequest(http.MethodPost, "/api/push/subscribe", strings.NewReader(`{}`))
	req.Host = "localhost"
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
}
