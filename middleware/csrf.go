package middleware

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

const (
	csrfSessionKey = "csrf_token"
	csrfCookieName = "asetra_csrf"
	csrfFormField  = "_csrf"
)

// CSRFProtection implements a session-bound synchronizer token for every
// state-changing request. A readable companion cookie lets the browser helper
// add the token to regular forms and fetch requests; the authoritative token
// remains inside the signed HttpOnly session cookie.
func CSRFProtection(secureCookie bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		token, _ := session.Get(csrfSessionKey).(string)
		if token == "" {
			generated, err := newCSRFToken()
			if err != nil {
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}
			token = generated
			session.Set(csrfSessionKey, token)
			if err := session.Save(); err != nil {
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}
		}

		http.SetCookie(c.Writer, &http.Cookie{
			Name:     csrfCookieName,
			Value:    token,
			Path:     "/",
			MaxAge:   60 * 60 * 8,
			Secure:   secureCookie,
			HttpOnly: false,
			SameSite: http.SameSiteStrictMode,
		})
		c.Set("CSRFToken", token)

		if isSafeMethod(c.Request.Method) {
			c.Next()
			return
		}

		provided := strings.TrimSpace(c.GetHeader("X-CSRF-Token"))
		if provided == "" {
			provided = strings.TrimSpace(c.PostForm(csrfFormField))
		}
		if !sameToken(token, provided) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "CSRF token tidak valid atau sudah kedaluwarsa"})
			return
		}

		c.Next()
	}
}

func newCSRFToken() (string, error) {
	value := make([]byte, 32)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(value), nil
}

func isSafeMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace:
		return true
	default:
		return false
	}
}

func sameToken(expected, actual string) bool {
	if expected == "" || len(expected) != len(actual) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(expected), []byte(actual)) == 1
}
