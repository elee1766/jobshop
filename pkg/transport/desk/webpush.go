package desk

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdh"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/hkdf"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Web Push (RFC 8030) with aes128gcm payload encryption (RFC 8291) and
// VAPID (RFC 8292). Browsers' push services (FCM, Mozilla, Apple) all
// take this, so no vendor SDK or account is needed.

var b64 = base64.RawURLEncoding

// Subscription is a browser PushSubscription as JSON.
type Subscription struct {
	Endpoint string `json:"endpoint"`
	Keys     struct {
		P256dh string `json:"p256dh"`
		Auth   string `json:"auth"`
	} `json:"keys"`
}

// VAPID is the application server key pair browsers subscribe against.
type VAPID struct {
	Private string `json:"private"` // base64url raw scalar
	key     *ecdsa.PrivateKey
}

// NewVAPID makes a new key pair.
func NewVAPID() (*VAPID, error) {
	k, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}
	b, err := k.Bytes()
	if err != nil {
		return nil, err
	}
	return &VAPID{Private: b64.EncodeToString(b), key: k}, nil
}

func (v *VAPID) load() error {
	if v.key != nil {
		return nil
	}
	b, err := b64.DecodeString(v.Private)
	if err != nil {
		return err
	}
	k, err := ecdsa.ParseRawPrivateKey(elliptic.P256(), b)
	if err != nil {
		return err
	}
	v.key = k
	return nil
}

// Public is the uncompressed public key, base64url: the browser's
// applicationServerKey.
func (v *VAPID) Public() (string, error) {
	if err := v.load(); err != nil {
		return "", err
	}
	b, err := v.key.PublicKey.Bytes()
	if err != nil {
		return "", err
	}
	return b64.EncodeToString(b), nil
}

// auth is the VAPID Authorization header for a push endpoint.
func (v *VAPID) auth(endpoint, subject string, now time.Time) (string, error) {
	u, err := url.Parse(endpoint)
	if err != nil {
		return "", err
	}
	pub, err := v.Public()
	if err != nil {
		return "", err
	}
	hdr := b64.EncodeToString([]byte(`{"typ":"JWT","alg":"ES256"}`))
	claims, _ := json.Marshal(map[string]any{
		"aud": u.Scheme + "://" + u.Host,
		"exp": now.Add(12 * time.Hour).Unix(),
		"sub": subject,
	})
	in := hdr + "." + b64.EncodeToString(claims)
	sum := sha256.Sum256([]byte(in))
	r, s, err := ecdsa.Sign(rand.Reader, v.key, sum[:])
	if err != nil {
		return "", err
	}
	sig := make([]byte, 64)
	r.FillBytes(sig[:32])
	s.FillBytes(sig[32:])
	return "vapid t=" + in + "." + b64.EncodeToString(sig) + ", k=" + pub, nil
}

// errGone means the push service dropped the subscription (404, 410).
var errGone = errors.New("subscription gone")

// Send encrypts payload for sub and posts it. errGone means delete sub.
func (v *VAPID) Send(ctx context.Context, c *http.Client, sub Subscription, subject string, payload []byte) error {
	uaPub, err := b64.DecodeString(sub.Keys.P256dh)
	if err != nil {
		return fmt.Errorf("p256dh: %w", err)
	}
	authSecret, err := b64.DecodeString(sub.Keys.Auth)
	if err != nil {
		return fmt.Errorf("auth: %w", err)
	}
	as, err := ecdh.P256().GenerateKey(rand.Reader)
	if err != nil {
		return err
	}
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return err
	}
	body, err := encrypt(payload, uaPub, authSecret, as, salt)
	if err != nil {
		return err
	}
	authz, err := v.auth(sub.Endpoint, subject, time.Now())
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, sub.Endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", authz)
	req.Header.Set("Content-Encoding", "aes128gcm")
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("TTL", "86400")
	req.Header.Set("Urgency", "high")
	res, err := c.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = res.Body.Close() }()
	msg, _ := io.ReadAll(io.LimitReader(res.Body, 512))
	switch {
	case res.StatusCode == http.StatusNotFound || res.StatusCode == http.StatusGone:
		return errGone
	case res.StatusCode/100 != 2:
		return fmt.Errorf("push service: %s: %s", res.Status, bytes.TrimSpace(msg))
	}
	return nil
}

// encrypt is RFC 8291 aes128gcm in one record of 4096.
func encrypt(plain, uaPubBytes, authSecret []byte, as *ecdh.PrivateKey, salt []byte) ([]byte, error) {
	uaPub, err := ecdh.P256().NewPublicKey(uaPubBytes)
	if err != nil {
		return nil, fmt.Errorf("p256dh: %w", err)
	}
	shared, err := as.ECDH(uaPub)
	if err != nil {
		return nil, err
	}
	asPub := as.PublicKey().Bytes()
	info := append(append([]byte("WebPush: info\x00"), uaPubBytes...), asPub...)
	ikm, err := hkdf.Key(sha256.New, shared, authSecret, string(info), 32)
	if err != nil {
		return nil, err
	}
	prk, err := hkdf.Extract(sha256.New, ikm, salt)
	if err != nil {
		return nil, err
	}
	cek, err := hkdf.Expand(sha256.New, prk, "Content-Encoding: aes128gcm\x00", 16)
	if err != nil {
		return nil, err
	}
	nonce, err := hkdf.Expand(sha256.New, prk, "Content-Encoding: nonce\x00", 12)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(cek)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	const rs = 4096
	if len(plain)+1+16 > rs {
		return nil, errors.New("push payload too large")
	}
	out := make([]byte, 0, 16+4+1+len(asPub)+len(plain)+17)
	out = append(out, salt...)
	out = binary.BigEndian.AppendUint32(out, rs)
	out = append(out, byte(len(asPub)))
	out = append(out, asPub...)
	return gcm.Seal(out, nonce, append(plain, 2), nil), nil
}
