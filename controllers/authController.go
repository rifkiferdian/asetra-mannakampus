package controllers

import (
	"database/sql"
	"errors"
	"net/http"

	"gobase-app/config"
	"gobase-app/repositories"
	"gobase-app/services"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func LoginPage(c *gin.Context) {
	session := sessions.Default(c)
	user := session.Get("user")
	if user != nil {
		c.Redirect(302, "/dashboard")
		return
	}
	c.HTML(http.StatusOK, "login.html", gin.H{
		"Title": "Login User",
	})
}

func LoginPost(c *gin.Context) {
	authService := buildAuthService()
	user, err := authService.Authenticate(c.PostForm("username"), c.PostForm("password"))
	if err != nil {
		status := http.StatusOK
		message := err.Error()
		if !isAuthenticationError(err) {
			status = http.StatusInternalServerError
			message = "Terjadi kesalahan saat mengambil data user"
		}

		renderLogin(c, status, message)
		return
	}

	session := sessions.Default(c)
	session.Set("user", user)
	session.Set("user_id", user.UserID)
	if err := session.Save(); err != nil {
		renderLogin(c, http.StatusInternalServerError, "Gagal menyimpan sesi")
		return
	}

	c.Redirect(http.StatusFound, "/dashboard")
}

func buildAuthService() *services.AuthService {
	return &services.AuthService{
		Repo: &repositories.AuthRepository{DB: config.DB},
	}
}

func isAuthenticationError(err error) bool {
	return errors.Is(err, services.ErrUsernameRequired) ||
		errors.Is(err, services.ErrPasswordRequired) ||
		errors.Is(err, services.ErrInactiveOrMissing) ||
		errors.Is(err, services.ErrInvalidPassword)
}

func renderLogin(c *gin.Context, status int, message string) {
	c.HTML(status, "login.html", gin.H{
		"Title": "Login User",
		"Error": message,
	})
}

func Logout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	session.Save()
	c.Redirect(302, "/")
}

func CreateUser(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	// Check if username already exists
	var existingUser string
	err := config.DB.QueryRow("SELECT username FROM users WHERE username = ?", username).Scan(&existingUser)
	if err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username already exists"})
		return
	} else if err != sql.ErrNoRows {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Insert new user
	_, err = config.DB.Exec("INSERT INTO users (username, password) VALUES (?, ?)", username, string(hashedPassword))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User created successfully"})
}
