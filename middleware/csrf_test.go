package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

func TestCSRFProtectionRejectsMissingTokenAndAcceptsValidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	store := cookie.NewStore([]byte("test-session-secret-with-at-least-32-bytes"))
	router.Use(sessions.Sessions("test_session", store), CSRFProtection(false))
	router.GET("/form", func(c *gin.Context) { c.Status(http.StatusNoContent) })
	router.POST("/action", func(c *gin.Context) { c.Status(http.StatusNoContent) })

	getRecorder := httptest.NewRecorder()
	router.ServeHTTP(getRecorder, httptest.NewRequest(http.MethodGet, "/form", nil))
	cookies := getRecorder.Result().Cookies()
	var csrfToken string
	for _, item := range cookies {
		if item.Name == csrfCookieName {
			csrfToken = item.Value
		}
	}
	if csrfToken == "" {
		t.Fatal("expected CSRF companion cookie")
	}

	missingRequest := httptest.NewRequest(http.MethodPost, "/action", nil)
	for _, item := range cookies {
		missingRequest.AddCookie(item)
	}
	missingRecorder := httptest.NewRecorder()
	router.ServeHTTP(missingRecorder, missingRequest)
	if missingRecorder.Code != http.StatusForbidden {
		t.Fatalf("missing token status = %d, want %d", missingRecorder.Code, http.StatusForbidden)
	}

	validRequest := httptest.NewRequest(http.MethodPost, "/action", nil)
	for _, item := range cookies {
		validRequest.AddCookie(item)
	}
	validRequest.Header.Set("X-CSRF-Token", csrfToken)
	validRecorder := httptest.NewRecorder()
	router.ServeHTTP(validRecorder, validRequest)
	if validRecorder.Code != http.StatusNoContent {
		t.Fatalf("valid token status = %d, want %d", validRecorder.Code, http.StatusNoContent)
	}
}
