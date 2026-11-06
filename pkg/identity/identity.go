// Package identity mints and caches credentials for configured identities
// (#200): GitHub App installation tokens, Linear App client_credentials
// tokens (#278), and GitHub tokens (#340) and Linear API keys read from
// files. Secrets never come
// from yaml.
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
	"net/url"
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

	mu     sync.Mutex
	cache  map[string]Token  // identity/owner -> token
	logins map[string]string // github-token identity -> its login
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
	case config.IdentityLinearApp:
		scopes := id.LinearScopes()
		key := name + "/linear-app/" + strings.Join(scopes, ",")
		m.mu.Lock()
		defer m.mu.Unlock()
		now := m.Now()
		if t, ok := m.cache[key]; ok && now.Add(refreshBefore).Before(t.Expires) {
			return t, nil
		}
		t, err := m.mintLinear(ctx, id, scopes, now)
		if err != nil {
			return Token{}, fmt.Errorf("identity %s: %w", name, err)
		}
		t.Identity, t.Kind = name, id.Kind
		m.cache[key] = t
		return t, nil
	case config.IdentityLinear, config.IdentityGitHubToken:
		b, err := config.ReadSecret(ctx, id.TokenFile)
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

// Forget drops cached tokens and logins, e.g. after a reload changed keys.
func (m *Minter) Forget() {
	m.mu.Lock()
	m.cache = map[string]Token{}
	m.logins = nil
	m.mu.Unlock()
}

// Drop forgets name's cached tokens, e.g. after the API refused one.
func (m *Minter) Drop(name string) {
	m.mu.Lock()
	for k := range m.cache {
		if strings.HasPrefix(k, name+"/") {
			delete(m.cache, k)
		}
	}
	m.mu.Unlock()
}

// Renew answers a caller whose token (stale) the API refused. For a
// linear-app the cache is checked first: when it already holds a
// different token, another caller renewed and that one is returned, so
// sessions that all hold the same dead token cause one mint, not one
// each (#866). Otherwise the cache is dropped and a token minted. Other
// kinds return their current token, which asking again does not change.
func (m *Minter) Renew(ctx context.Context, name string, id config.Identity, target, stale string) (Token, error) {
	if id.Kind != config.IdentityLinearApp {
		return m.Token(ctx, name, id, target)
	}
	scopes := id.LinearScopes()
	key := name + "/linear-app/" + strings.Join(scopes, ",")
	m.mu.Lock()
	defer m.mu.Unlock()
	if t, ok := m.cache[key]; ok && t.Token != stale {
		return t, nil
	}
	t, err := m.mintLinear(ctx, id, scopes, m.Now())
	if err != nil {
		return Token{}, fmt.Errorf("identity %s: %w", name, err)
	}
	t.Identity, t.Kind = name, id.Kind
	m.cache[key] = t
	return t, nil
}

// Login returns the GitHub login an identity acts as, cached until
// Forget: GET /user for a github-token (#340), "<slug>[bot]" from GET
// /app for a github-app (#402). Its comments are filtered so they are not
// delivered back to itself.
func (m *Minter) Login(ctx context.Context, name string, id config.Identity) (string, error) {
	if !config.IsGitHub(id.Kind) {
		return "", fmt.Errorf("identity %s: login is for GitHub identities, not %s", name, id.Kind)
	}
	m.mu.Lock()
	l, ok := m.logins[name]
	m.mu.Unlock()
	if ok {
		return l, nil
	}
	path, auth := "/user", ""
	if id.Kind == config.IdentityGitHubApp {
		jwt, err := JWTFor(id, m.Now())
		if err != nil {
			return "", fmt.Errorf("identity %s: %w", name, err)
		}
		path, auth = "/app", jwt
	} else {
		tok, err := m.Token(ctx, name, id, "")
		if err != nil {
			return "", err
		}
		auth = tok.Token
	}
	base := id.APIURL
	if base == "" {
		base = DefaultAPIURL
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(base, "/")+path, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+auth)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	resp, err := m.HTTP.Do(req)
	if err != nil {
		return "", fmt.Errorf("identity %s: %w", name, err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("identity %s: GET %s: %s: %s", name, path, resp.Status, strings.TrimSpace(string(body)))
	}
	var u struct {
		Login string `json:"login"`
		Slug  string `json:"slug"`
	}
	if err := json.Unmarshal(body, &u); err != nil {
		return "", fmt.Errorf("identity %s: GET %s: %w", name, path, err)
	}
	if id.Kind == config.IdentityGitHubApp && u.Slug != "" {
		u.Login = u.Slug + "[bot]"
	}
	if u.Login == "" {
		return "", fmt.Errorf("identity %s: GET %s: no login", name, path)
	}
	m.mu.Lock()
	if m.logins == nil {
		m.logins = map[string]string{}
	}
	m.logins[name] = u.Login
	m.mu.Unlock()
	return u.Login, nil
}

func (m *Minter) mint(ctx context.Context, id config.Identity, inst int64, now time.Time) (Token, error) {
	pemBytes, err := config.ReadSecret(ctx, id.PrivateKeyFile)
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

// DefaultLinearTokenURL is Linear's OAuth token endpoint.
const DefaultLinearTokenURL = "https://api.linear.app/oauth/token"

// mintLinear gets an app token with the OAuth client_credentials grant
// (#278). Error bodies are included (Linear returns error and
// error_description), never the secret or a token.
func (m *Minter) mintLinear(ctx context.Context, id config.Identity, scopes []string, now time.Time) (Token, error) {
	cid, err := config.ReadSecret(ctx, id.ClientIDFile)
	if err != nil {
		return Token{}, err
	}
	secret, err := config.ReadSecret(ctx, id.ClientSecretFile)
	if err != nil {
		return Token{}, err
	}
	form := url.Values{
		"grant_type":    {"client_credentials"},
		"client_id":     {strings.TrimSpace(string(cid))},
		"client_secret": {strings.TrimSpace(string(secret))},
		"scope":         {strings.Join(scopes, ",")},
	}
	u := id.TokenURL
	if u == "" {
		u = DefaultLinearTokenURL
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, strings.NewReader(form.Encode()))
	if err != nil {
		return Token{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	resp, err := m.HTTP.Do(req)
	if err != nil {
		return Token{}, err
	}
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	var r struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int64  `json:"expires_in"`
		Error       string `json:"error"`
		Description string `json:"error_description"`
	}
	_ = json.Unmarshal(body, &r)
	if resp.StatusCode != http.StatusOK {
		msg := strings.TrimSpace(r.Error + " " + r.Description)
		if msg == "" {
			msg = resp.Status
		}
		return Token{}, fmt.Errorf("linear app token: %s: %s", resp.Status, msg)
	}
	if r.AccessToken == "" {
		return Token{}, errors.New("linear app token: empty token")
	}
	t := Token{Token: r.AccessToken}
	if r.ExpiresIn > 0 {
		t.Expires = now.Add(time.Duration(r.ExpiresIn) * time.Second)
	}
	return t, nil
}

// JWTFor signs an App JWT for a github-app identity, reading its key
// file. Used for App-level APIs such as webhook deliveries (#335).
func JWTFor(id config.Identity, now time.Time) (string, error) {
	if id.Kind != config.IdentityGitHubApp {
		return "", fmt.Errorf("app jwt: kind %s is not %s", id.Kind, config.IdentityGitHubApp)
	}
	b, err := config.ReadSecret(context.Background(), id.PrivateKeyFile)
	if err != nil {
		return "", err
	}
	key, err := parseKey(b)
	if err != nil {
		return "", err
	}
	return AppJWT(id.AppID, key, now)
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
