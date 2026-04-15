package controllers

import (
	"net/http"
	"gobase-app/models"

	helpers "gobase-app/helper"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func Render(c *gin.Context, name string, data gin.H) {

	if data == nil {
		data = gin.H{}
	}

	session := sessions.Default(c)

	setInitials := func(initials, name string) {
		if existing, ok := data["Initials"].(string); ok && existing != "" {
			return
		}

		if initials == "" {
			initials = helpers.Initials(name)
		}

		if initials != "" {
			data["Initials"] = initials
		}
	}

	if u := session.Get("user"); u != nil {
		switch val := u.(type) {
		case models.SessionUser:
			setInitials(val.Initials, val.Name)
			data["User"] = gin.H{
				"user_id":          val.UserID,
				"nip":              val.NIP,
				"name":             val.Name,
				"initials":         val.Initials,
				"username":         val.Username,
				"role":             val.Role,
				"store_id":         val.StoreID,
				"is_authenticated": val.IsAuthenticated,
			}
		case map[string]interface{}:
			data["User"] = val
			if initials, ok := val["initials"].(string); ok {
				setInitials(initials, "")
			} else if nameVal, ok := val["name"].(string); ok {
				setInitials("", nameVal)
			}
		case gin.H:
			data["User"] = map[string]interface{}(val)
			if initials, ok := val["initials"].(string); ok {
				setInitials(initials, "")
			} else if nameVal, ok := val["name"].(string); ok {
				setInitials("", nameVal)
			}
		default:
			data["User"] = gin.H{
				"name":     u,
				"username": u,
			}
			if nameVal, ok := u.(string); ok {
				setInitials("", nameVal)
			}
		}
	}

	// ambil permissions dari context
	permsAny, _ := c.Get("Permissions")
	perms, _ := permsAny.(map[string]bool)

	// inject global data (biar semua halaman dapat)
	data["Permissions"] = perms

	c.HTML(http.StatusOK, name, data)
}

