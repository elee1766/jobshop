// Package identity mints and caches credentials for configured identities
// (#200): GitHub App installation tokens, and Linear API keys read from
// files. Secrets never come from yaml.
package identity

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/elee1766/jobshop/pkg/config"
)

// DefaultAPIURL is the GitHub REST API base.
const DefaultAPIURL = "https://api.github.com"

// refreshBefore is how long before expiry a cached token is replaced.
// GitHub tokens live an hour; a session gets at least this much use.
const refreshBefore = 10 * time.Minute

// Token is a credential for one target.
type Token struct {
	Identity string    `json:"identity"`
	Kind     string    `json:"kind"`
	Token    string    `json:"token"`
	Expires  time.Time `json:"expires,omitzero"`
}

// Minter caches installation tokens per identity and owner.
type Minter struct {
	HTTP *http.Client
	Now  func() time.Time

	mu    sync.Mutex
	cache map[string]Token // identity/owner -> token
}

// NewMinter returns a Minter with a 30s HTTP timeout.
func NewMinter() *Minter {
	return &Minter{HTTP: &http.Client{Timeout: 30 * time.Second}, Now: time.Now, cache: map[string]Token{}}
}

// Token returns a credential for target (owner/repo for github-app, team
// key for linear) as identity name. The caller has checked the identity
// covers target.
func (m *Minter) Token(ctx context.Context, name string, id config.Identity, target string) (Token, error) {
	switch id.Kind {
	case config.IdentityLinear:
		b, err := os.ReadFile(id.TokenFile)
		if err != nil {
			return Token{}, fmt.Errorf("identity %s: %w", name, err)
		}
		return Token{Identity: name, Kind: id.Kind, Token: strings.TrimSpace(string(b))}, nil
	case config.IdentityGitHubApp:
		owner, _, _ := strings.Cut(target, "/")
		inst, ok := id.Installations[owner]
		if !ok {
			return Token{}, fmt.Errorf("identity %s: no installation for %s", name, owner)
		}
		key := name + "/" + owner
		m.mu.Lock()
		defer m.mu.Unlock()
		now := m.Now()
		if t, ok := m.cache[key]; ok && now.Add(refreshBefore).Before(t.Expires) {
			return t, nil
		}
		t, err := m.mint(ctx, id, inst, now)
		if err != nil {
			return Token{}, fmt.Errorf("identity %s: %w", name, err)
		}
		t.Identity, t.Kind = name, id.Kind
		m.cache[key] = t
		return t, nil
	}
	return Token{}, fmt.Errorf("identity %s: unknown kind %q", name, id.Kind)
}

// Forget drops cached tokens, e.g. after a reload changed keys.
func (m *Minter) Forget() {
	m.mu.Lock()
	m.cache = map[string]Token{}
	m.mu.Unlock()
}

func (m *Minter) mint(ctx context.Context, id config.Identity, inst int64, now time.Time) (Token, error) {
	pemBytes, err := os.ReadFile(id.PrivateKeyFile)
	if err != nil {
		return Token{}, err
	}
	key, err := parseKey(pemBytes)
	if err != nil {
		return Token{}, err
	}
	jwt, err := AppJWT(id.AppID, key, now)
	if err != nil {
		return Token{}, err
	}
	base := id.APIURL
	if base == "" {
		base = DefaultAPIURL
	}
	url := fmt.Sprintf("%s/app/installations/%d/access_tokens", strings.TrimRight(base, "/"), inst)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return Token{}, err
	}
	req.Header.Set("Authorization", "Bearer "+jwt)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	resp, err := m.HTTP.Do(req)
	if err != nil {
		return Token{}, err
	}
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode != http.StatusCreated {
		return Token{}, fmt.Errorf("installation %d token: %s: %s", inst, resp.Status, strings.TrimSpace(string(body)))
	}
	var r struct {
		Token     string    `json:"token"`
		ExpiresAt time.Time `json:"expires_at"`
	}
	if err := json.Unmarshal(body, &r); err != nil {
		return Token{}, fmt.Errorf("installation %d token: %w", inst, err)
	}
	if r.Token == "" {
		return Token{}, fmt.Errorf("installation %d token: empty token", inst)
	}
	return Token{Token: r.Token, Expires: r.ExpiresAt}, nil
}

func parseKey(b []byte) (*rsa.PrivateKey, error) {
	blk, _ := pem.Decode(b)
	if blk == nil {
		return nil, errors.New("private key: no PEM block")
	}
	if k, err := x509.ParsePKCS1PrivateKey(blk.Bytes); err == nil {
		return k, nil
	}
	k, err := x509.ParsePKCS8PrivateKey(blk.Bytes)
	if err != nil {
		return nil, fmt.Errorf("private key: %w", err)
	}
	rk, ok := k.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("private key: not RSA")
	}
	return rk, nil
}

// AppJWT signs the RS256 JWT a GitHub App uses to request installation
// tokens. iat is backdated 60s for clock drift; GitHub caps exp at 10m.
func AppJWT(appID int64, key *rsa.PrivateKey, now time.Time) (string, error) {
	enc := base64.RawURLEncoding
	head := enc.EncodeToString([]byte(`{"alg":"RS256","typ":"JWT"}`))
	claims, err := json.Marshal(map[string]any{
		"iat": now.Add(-60 * time.Second).Unix(),
		"exp": now.Add(9 * time.Minute).Unix(),
		"iss": fmt.Sprint(appID),
	})
	if err != nil {
		return "", err
	}
	signing := head + "." + enc.EncodeToString(claims)
	sum := sha256.Sum256([]byte(signing))
	sig, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, sum[:])
	if err != nil {
		return "", err
	}
	return signing + "." + enc.EncodeToString(sig), nil
}
