package controller

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRefreshCookieFlags(t *testing.T) {
	for _, secure := range []bool{true, false} {
		t.Run(map[bool]string{true: "production", false: "local"}[secure], func(t *testing.T) {
			if secure {
				t.Setenv("AUTH_COOKIE_SECURE", "true")
			} else {
				t.Setenv("AUTH_COOKIE_SECURE", "false")
			}
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			setRefreshCookie(c, "opaque-refresh", 3600)
			cookies := w.Result().Cookies()
			if len(cookies) != 1 {
				t.Fatal("missing cookie")
			}
			cookie := cookies[0]
			if !cookie.HttpOnly || cookie.Secure != secure || cookie.Path != "/api/auth" || cookie.Domain != "" || cookie.MaxAge != 3600 {
				t.Fatalf("unsafe cookie: %+v", cookie)
			}
			if secure && cookie.SameSite != http.SameSiteNoneMode {
				t.Fatal("wrong SameSite")
			}
			if w.Header().Get("Cache-Control") != "no-store" {
				t.Fatal("session response is cacheable")
			}
		})
	}
}
func TestSessionEndpointsRequireCSRFHeader(t *testing.T) {
	ctrl := New(nil)
	for _, handler := range []gin.HandlerFunc{ctrl.Refresh, ctrl.Logout} {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/auth/refresh", nil)
		handler(c)
		if w.Code != http.StatusForbidden {
			t.Fatalf("missing CSRF protection: %d", w.Code)
		}
	}
}
