package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	authcontroller "github.com/andresramirez/psych-appointments/controllers/auth"
)

func TestSessionCORSAndProtectedRoutes(t *testing.T) {
	t.Setenv("FRONTEND_URL", "https://appointments.example.com/")
	r := NewRouter(nil, authcontroller.New(nil), nil, nil, nil, nil, nil, nil, nil)
	for _, tc := range []struct {
		method, path, origin string
		status               int
	}{
		{"OPTIONS", "/api/auth/refresh", "https://appointments.example.com", 204},
		{"OPTIONS", "/api/auth/refresh", "https://attacker.example.com", 403},
		{"POST", "/api/auth/refresh", "https://attacker.example.com", 403},
		{"POST", "/api/auth/refresh", "https://appointments.example.com", 403}, // missing custom header
		{"GET", "/api/patients", "https://appointments.example.com", 401},
	} {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(tc.method, tc.path, nil)
		req.Header.Set("Origin", tc.origin)
		r.engine.ServeHTTP(w, req)
		if w.Code != tc.status {
			t.Fatalf("%s %s: got %d want %d", tc.method, tc.path, w.Code, tc.status)
		}
		if tc.method == http.MethodOptions && tc.status == 204 {
			if w.Header().Get("Access-Control-Allow-Origin") != tc.origin || w.Header().Get("Access-Control-Allow-Credentials") != "true" {
				t.Fatal("credentialed preflight is not origin-bound")
			}
		}
		if tc.path == "/api/patients" && w.Header().Get("Cache-Control") != "no-store" {
			t.Fatal("protected response can be cached")
		}
	}
}
