package main

import (
	"encoding/gob"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"gobase-app/config"
	"gobase-app/models"
	"gobase-app/routes"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env if present
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, falling back to OS environment")
	}

	// Initialize database / config
	config.Connect()

	// Initialize Gin engine // menampilkan logger di terminal
	// r := gin.Default()

	// 🔥 Set Gin release mode (biar tidak ada log debug)
	gin.SetMode(gin.ReleaseMode)

	// Initialize Gin tanpa logger
	r := gin.New()
	r.Use(gin.Recovery())

	// Custom template functions tambah
	r.SetFuncMap(template.FuncMap{
		"no": func(a, b int) int {
			return a + b
		},
		"baseURL": func(path string) string {
			base := strings.TrimRight(os.Getenv("BASE_URL"), "/")
			p := "/" + strings.TrimLeft(path, "/")
			return base + p
		},
	})

	// Templates & static files
	r.LoadHTMLGlob("templates/**/*")
	r.Static("/assets", "./assets")

	useSecureCookie := strings.ToLower(os.Getenv("APP_SECURE_COOKIE")) == "true"

	// Register custom session payload for gob encoder used by cookie store.
	gob.Register(models.SessionUser{})

	// SESSION - must be registered BEFORE routes that use sessions
	store := cookie.NewStore([]byte("secret-key"))
	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   60 * 60 * 8, // 8 jam
		HttpOnly: true,
		// Secure harus false saat akses lokal HTTP; aktifkan otomatis jika APP_ENV=production atau APP_SECURE_COOKIE=true.
		Secure:   useSecureCookie,
		SameSite: http.SameSiteLaxMode,
	})
	r.Use(sessions.Sessions("mysession", store))

	// Register application routes
	routes.RegisterWebRoutes(r)

	// Render custom 404 page
	r.NoRoute(func(c *gin.Context) {
		c.HTML(http.StatusNotFound, "error.html", gin.H{
			"code_error": http.StatusNotFound,
			"error":      "Page not found",
		})
	})

	// Determine port (fallback to 8080 if APP_PORT is not set)
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	// ===============================
	// 🔥 BANNER DI SINI (POSISI BENAR)
	// ===============================

	fmt.Println("🚀 Server is running at http://localhost:" + port)
	fmt.Println("⚠️  DO NOT CLOSE THIS SERVER!")

	// Start HTTP server and log fatal on error
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}

