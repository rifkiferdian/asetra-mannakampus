package middleware

import (
	"net/http"
	"gobase-app/models"
	"gobase-app/services"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		user := session.Get("user")

		if user == nil {
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		c.Next()
	}
}

func UserMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		user := session.Get("user")

		// Kalau user sudah login, simpan ke context (bisa dipakai di handler / template)
		if user != nil {
			c.Set("User", user)
		}

		c.Next()
	}
}

func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		sess := sessions.Default(c)
		role, _ := sess.Get("role").(string)

		for _, r := range roles {
			if role == r {
				c.Next()
				return
			}
		}

		// c.AbortWithStatusJSON(403, gin.H{
		// 	"error": "Akses ditolak: role tidak diizinkan",
		// })
		c.HTML(403, "error.html", gin.H{
			"code_error": 3,
			"error":      "Anda Tidak punya Akses di Halaman ini",
		})
		c.Abort()
	}
}

func RequirePermission(perm string) gin.HandlerFunc {
	return func(c *gin.Context) {
		sess := sessions.Default(c)
		userID := extractUserID(sess)

		if userID == 0 {
			c.AbortWithStatus(401)
			return
		}

		ok, err := services.UserHasPermission(userID, perm)
		if err != nil || !ok {
			c.HTML(403, "error.html", gin.H{
				// "error": "Tidak punya Aksess di Halaman ini: " + perm,
				"code_error": 3,
				"error":      "Anda Tidak punya Akses di Halaman ini",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// extractUserID mencoba mengambil user_id dari session (baik dari key "user_id" maupun payload "user").
func extractUserID(sess sessions.Session) int {
	if v := sess.Get("user_id"); v != nil {
		switch id := v.(type) {
		case int:
			return id
		case int64:
			return int(id)
		case float64:
			return int(id)
		}
	}

	if u := sess.Get("user"); u != nil {
		switch val := u.(type) {
		case models.SessionUser:
			return val.UserID
		case map[string]interface{}:
			if id, ok := val["user_id"]; ok {
				return normalizeID(id)
			}
			if id, ok := val["UserID"]; ok {
				return normalizeID(id)
			}
		case gin.H:
			if id, ok := val["user_id"]; ok {
				return normalizeID(id)
			}
			if id, ok := val["UserID"]; ok {
				return normalizeID(id)
			}
		}
	}

	return 0
}

func normalizeID(val interface{}) int {
	switch id := val.(type) {
	case int:
		return id
	case int64:
		return int(id)
	case float64:
		return int(id)
	default:
		return 0
	}
}

func PermissionContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		sess := sessions.Default(c)
		userID := extractUserID(sess)

		if userID > 0 {
			perms, err := services.GetUserPermissions(userID)
			if err == nil {
				c.Set("Permissions", perms) // simpan di context (dipakai di render)
			}
		}

		c.Next()
	}
}



