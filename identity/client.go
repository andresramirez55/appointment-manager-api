// Package identity integrates the external identity provider. It owns no app data.
package identity

import (
	"bytes"
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var ErrUnavailable = errors.New("servicio de autenticación no disponible")
var ErrRejected = errors.New("credenciales o solicitud de autenticación inválidas")

type User struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}
type Tokens struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	ExpiresAt    time.Time `json:"expires_at"`
}
type Session struct {
	User   User   `json:"user"`
	Tokens Tokens `json:"tokens"`
}
type Client struct {
	baseURL, issuer       string
	http                  *http.Client
	mu                    sync.Mutex
	keys                  map[string]*rsa.PublicKey
	loadedAt, attemptedAt time.Time
}

func New(baseURL, issuer string) (*Client, error) {
	u, err := url.Parse(baseURL)
	if err != nil || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || (u.Scheme != "https" && !(u.Scheme == "http" && (u.Hostname() == "localhost" || u.Hostname() == "127.0.0.1"))) || strings.TrimSpace(issuer) == "" {
		return nil, fmt.Errorf("AUTH_SERVICE_URL must use HTTPS (HTTP allowed on localhost); AUTH_ISSUER is required")
	}
	return &Client{baseURL: strings.TrimRight(baseURL, "/"), issuer: issuer, http: &http.Client{Timeout: 10 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}}, nil
}
func (c *Client) request(ctx context.Context, method, path string, input, output any) error {
	var body io.Reader
	if input != nil {
		b, err := json.Marshal(input)
		if err != nil {
			return err
		}
		body = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return ErrUnavailable
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return ErrUnavailable
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 500 || resp.StatusCode == 429 {
		return ErrUnavailable
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return ErrRejected
	}
	if output == nil {
		return nil
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(output); err != nil {
		return ErrUnavailable
	}
	return nil
}
func (c *Client) Authenticate(ctx context.Context, register bool, email, password, name string) (*Session, error) {
	path := "/v1/auth/login"
	if register {
		path = "/v1/auth/register"
	}
	var s Session
	err := c.request(ctx, http.MethodPost, path, map[string]string{"email": email, "password": password, "name": name}, &s)
	if err != nil {
		return nil, err
	}
	sub, err := c.Validate(ctx, s.Tokens.AccessToken)
	if err != nil || sub != s.User.ID || s.User.Email == "" || s.Tokens.RefreshToken == "" {
		return nil, ErrUnavailable
	}
	return &s, nil
}
func (c *Client) Refresh(ctx context.Context, refresh string) (*Tokens, error) {
	var t Tokens
	if err := c.request(ctx, http.MethodPost, "/v1/auth/refresh", map[string]string{"refresh_token": refresh}, &t); err != nil {
		return nil, err
	}
	if t.RefreshToken == "" {
		return nil, ErrUnavailable
	}
	return &t, nil
}
func (c *Client) Logout(ctx context.Context, refresh string) error {
	return c.request(ctx, http.MethodPost, "/v1/auth/logout", map[string]string{"refresh_token": refresh}, nil)
}
func (c *Client) Validate(ctx context.Context, raw string) (string, error) {
	claims := struct {
		TokenType string `json:"token_type"`
		jwt.RegisteredClaims
	}{}
	token, err := jwt.ParseWithClaims(raw, &claims, func(t *jwt.Token) (any, error) {
		kid, ok := t.Header["kid"].(string)
		if !ok || kid == "" {
			return nil, ErrRejected
		}
		return c.key(ctx, kid)
	}, jwt.WithValidMethods([]string{"RS256"}), jwt.WithIssuer(c.issuer), jwt.WithExpirationRequired(), jwt.WithIssuedAt())
	if errors.Is(err, ErrUnavailable) {
		return "", ErrUnavailable
	}
	if err != nil || !token.Valid || claims.TokenType != "access" || strings.TrimSpace(claims.Subject) == "" {
		return "", ErrRejected
	}
	return claims.Subject, nil
}
func (c *Client) key(ctx context.Context, kid string) (*rsa.PublicKey, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if key := c.keys[kid]; key != nil && time.Since(c.loadedAt) < 5*time.Minute {
		return key, nil
	}
	// Limit network requests for attacker-controlled unknown key IDs.
	if time.Since(c.attemptedAt) < 5*time.Second {
		return nil, ErrRejected
	}
	c.attemptedAt = time.Now()
	var document struct {
		Keys []struct{ Kty, Use, Alg, Kid, N, E string } `json:"keys"`
	}
	if err := c.request(ctx, http.MethodGet, "/.well-known/jwks.json", nil, &document); err != nil {
		return nil, err
	}
	keys := map[string]*rsa.PublicKey{}
	for _, k := range document.Keys {
		if k.Kty != "RSA" || k.Alg != "RS256" || k.Use != "sig" || k.Kid == "" {
			continue
		}
		n, e1 := base64.RawURLEncoding.DecodeString(k.N)
		e, e2 := base64.RawURLEncoding.DecodeString(k.E)
		if e1 != nil || e2 != nil || len(e) > 4 {
			continue
		}
		exponent := new(big.Int).SetBytes(e).Int64()
		modulus := new(big.Int).SetBytes(n)
		if exponent < 3 || exponent%2 == 0 || modulus.BitLen() < 2048 {
			continue
		}
		keys[k.Kid] = &rsa.PublicKey{N: modulus, E: int(exponent)}
	}
	c.keys = keys
	c.loadedAt = time.Now()
	if key := keys[kid]; key != nil {
		return key, nil
	}
	return nil, ErrRejected
}
