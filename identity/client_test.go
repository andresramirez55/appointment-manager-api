package identity

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestValidateTrustBoundaryAndKeyCache(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		json.NewEncoder(w).Encode(map[string]any{"keys": []map[string]string{{"kid": "trusted", "kty": "RSA", "use": "sig", "alg": "RS256", "n": base64.RawURLEncoding.EncodeToString(key.N.Bytes()), "e": base64.RawURLEncoding.EncodeToString(big.NewInt(int64(key.E)).Bytes())}}})
	}))
	defer server.Close()
	client, err := New(server.URL, "trusted-issuer")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, issuer, subject, kind, kid string
		expiry                           int
		hmac                             bool
		valid                            bool
	}{
		{"valid", "trusted-issuer", "user-1", "access", "trusted", 1, false, true},
		{"cached", "trusted-issuer", "user-2", "access", "trusted", 1, false, true},
		{"wrong issuer", "evil", "user-1", "access", "trusted", 1, false, false},
		{"expired", "trusted-issuer", "user-1", "access", "trusted", -1, false, false},
		{"no expiry", "trusted-issuer", "user-1", "access", "trusted", 0, false, false},
		{"missing subject", "trusted-issuer", "", "access", "trusted", 1, false, false},
		{"refresh JWT", "trusted-issuer", "user-1", "refresh", "trusted", 1, false, false},
		{"unknown key", "trusted-issuer", "user-1", "access", "evil", 1, false, false},
		{"algorithm confusion", "trusted-issuer", "user-1", "access", "trusted", 1, true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			claims := jwt.MapClaims{"iss": tc.issuer, "sub": tc.subject, "token_type": tc.kind}
			if tc.expiry != 0 {
				claims["exp"] = time.Now().Add(time.Duration(tc.expiry) * time.Hour).Unix()
			}
			var signing any = key
			method := jwt.SigningMethod(jwt.SigningMethodRS256)
			if tc.hmac {
				method = jwt.SigningMethodHS256
				signing = []byte("public-key-not-a-secret")
			}
			token := jwt.NewWithClaims(method, claims)
			token.Header["kid"] = tc.kid
			raw, err := token.SignedString(signing)
			if err != nil {
				t.Fatal(err)
			}
			sub, err := client.Validate(context.Background(), raw)
			if tc.valid && (err != nil || sub != tc.subject) {
				t.Fatalf("valid token rejected: %v", err)
			}
			if !tc.valid && err == nil {
				t.Fatal("untrusted token accepted")
			}
		})
	}
	if requests.Load() != 1 {
		t.Fatalf("unexpected JWKS requests: %d", requests.Load())
	}
	// An attacker-controlled private key cannot impersonate a linked subject.
	attacker, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	forged := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{"iss": "trusted-issuer", "sub": "user-1", "token_type": "access", "exp": time.Now().Add(time.Hour).Unix()})
	forged.Header["kid"] = "trusted"
	raw, err := forged.SignedString(attacker)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.Validate(context.Background(), raw); err == nil {
		t.Fatal("forged signature accepted")
	}
	// After cache expiry, a provider failure must not fall back to stale keys.
	server.Close()
	client.loadedAt = time.Now().Add(-time.Hour)
	client.attemptedAt = time.Now().Add(-time.Minute)
	if _, err := client.key(context.Background(), "trusted"); err == nil {
		t.Fatal("accepted stale key during outage")
	}
}

func TestClientRejectsUnsafeConfiguration(t *testing.T) {
	for _, u := range []string{"", "http://auth.example.com", "https://user:pass@auth.example.com", "https://auth.example.com?redirect=evil"} {
		if _, err := New(u, "issuer"); err == nil {
			t.Fatalf("accepted %q", u)
		}
	}
	if _, err := New("https://auth.example.com", ""); err == nil {
		t.Fatal("accepted empty issuer")
	}
}

func TestRefreshAndLogoutContract(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("wrong method %s", r.Method)
		}
		var body map[string]string
		json.NewDecoder(r.Body).Decode(&body)
		if body["refresh_token"] != "opaque-session" {
			t.Error("refresh token not forwarded")
		}
		if r.URL.Path == "/v1/auth/logout" {
			w.WriteHeader(204)
			return
		}
		json.NewEncoder(w).Encode(Tokens{AccessToken: "access", RefreshToken: "rotated"})
	}))
	defer server.Close()
	client, _ := New(server.URL, "issuer")
	pair, err := client.Refresh(context.Background(), "opaque-session")
	if err != nil || pair.RefreshToken != "rotated" {
		t.Fatalf("bad refresh: %v", err)
	}
	if err := client.Logout(context.Background(), "opaque-session"); err != nil {
		t.Fatal(err)
	}
}
