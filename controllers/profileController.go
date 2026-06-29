package controllers

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"gobase-app/config"
	helpers "gobase-app/helper"
	"gobase-app/models"
	"gobase-app/repositories"
	"gobase-app/services"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func ProfileIndex(c *gin.Context) {
	userID := currentSessionUserID(c)
	if userID <= 0 {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	userSvc := profileUserService()
	profile, err := userSvc.GetUserProfile(userID)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	Render(c, "profile.html", gin.H{
		"Title":           "Profile",
		"Page":            "profile",
		"Profile":         profile,
		"Success":         c.Query("success"),
		"Error":           c.Query("error"),
		"PasswordSuccess": c.Query("password_success"),
		"PasswordError":   c.Query("password_error"),
	})
}

func ProfileUpdate(c *gin.Context) {
	userID := currentSessionUserID(c)
	if userID <= 0 {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	input := models.UserProfileUpdateInput{
		ID:       userID,
		Name:     strings.TrimSpace(c.PostForm("name")),
		Username: strings.TrimSpace(c.PostForm("username")),
		Email:    strings.TrimSpace(c.PostForm("email")),
	}

	userSvc := profileUserService()
	if err := userSvc.UpdateOwnProfile(input); err != nil {
		c.Redirect(http.StatusSeeOther, "/profile?error="+url.QueryEscape(err.Error()))
		return
	}

	profile, err := userSvc.GetUserProfile(userID)
	if err == nil {
		refreshProfileSession(c, profile)
	}

	c.Redirect(http.StatusSeeOther, "/profile?success="+url.QueryEscape("Profile berhasil diperbarui"))
}

func ProfilePasswordUpdate(c *gin.Context) {
	userID := currentSessionUserID(c)
	if userID <= 0 {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	input := models.UserPasswordUpdateInput{
		ID:              userID,
		CurrentPassword: c.PostForm("current_password"),
		NewPassword:     c.PostForm("new_password"),
		ConfirmPassword: c.PostForm("confirm_password"),
	}

	userSvc := profileUserService()
	if err := userSvc.UpdateOwnPassword(input); err != nil {
		c.Redirect(http.StatusSeeOther, "/profile?password_error="+url.QueryEscape(err.Error()))
		return
	}

	c.Redirect(http.StatusSeeOther, "/profile?password_success="+url.QueryEscape("Password berhasil diganti"))
}

func profileUserService() *services.UserService {
	return &services.UserService{
		Repo: &repositories.UserRepository{DB: config.DB},
	}
}

func currentSessionUserID(c *gin.Context) int {
	session := sessions.Default(c)
	if v := session.Get("user_id"); v != nil {
		return normalizeSessionID(v)
	}

	if u := session.Get("user"); u != nil {
		switch val := u.(type) {
		case models.SessionUser:
			return val.UserID
		case map[string]interface{}:
			if id, ok := val["user_id"]; ok {
				return normalizeSessionID(id)
			}
			if id, ok := val["UserID"]; ok {
				return normalizeSessionID(id)
			}
		case gin.H:
			if id, ok := val["user_id"]; ok {
				return normalizeSessionID(id)
			}
			if id, ok := val["UserID"]; ok {
				return normalizeSessionID(id)
			}
		}
	}

	return 0
}

func normalizeSessionID(val interface{}) int {
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

func refreshProfileSession(c *gin.Context, profile models.User) {
	session := sessions.Default(c)
	storeIDs := make([]string, 0, len(profile.StoreIDs))
	for _, id := range profile.StoreIDs {
		storeIDs = append(storeIDs, strconv.Itoa(id))
	}

	session.Set("user", models.SessionUser{
		UserID:          profile.ID,
		NIP:             strconv.Itoa(profile.NIP),
		Name:            profile.Name,
		Initials:        helpers.Initials(profile.Name),
		Username:        profile.Username,
		Role:            profile.RoleDisplay,
		StoreID:         strings.Join(storeIDs, ","),
		IsAuthenticated: true,
	})
	session.Set("user_id", profile.ID)
	_ = session.Save()
}
