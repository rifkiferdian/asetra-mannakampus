package routes

import (
	"gobase-app/controllers"
	"gobase-app/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterWebRoutes(r *gin.Engine) {
	r.Use(middleware.UserMiddleware())

	r.GET("/", controllers.LoginPage)
	r.GET("/login", controllers.LoginPage)
	r.POST("/login", controllers.LoginPost)
	r.POST("/register", controllers.CreateUser)
	r.GET("/logout", controllers.Logout)

	auth := r.Group("/")
	auth.Use(middleware.AuthRequired(), middleware.PermissionContext())
	{
		auth.GET("/dashboard", controllers.DashboardIndex)

		auth.GET("/stores", controllers.StoreIndex)
		auth.POST("/stores", controllers.StoreStore)
		auth.POST("/stores/update", controllers.StoreUpdate)
		auth.GET("/stores/delete/:id", controllers.StoreDelete)
		auth.GET("/users", middleware.RequirePermission("user_management_access"), controllers.UserIndex)
		auth.POST("/users", middleware.RequirePermission("user_create"), controllers.UserStore)
		auth.POST("/users/update", middleware.RequirePermission("user_edit"), controllers.UserUpdate)
		auth.GET("/users/delete/:id", middleware.RequirePermission("user_delete"), controllers.UserDelete)
		auth.GET("/role", controllers.RoleIndex)
		auth.GET("/roleForm", controllers.RoleFormIndex)
		auth.GET("/role/:id/edit", middleware.RequirePermission("role_edit"), controllers.RoleEdit)
		auth.POST("/role", middleware.RequirePermission("role_create"), controllers.RoleStore)
		auth.POST("/role/update", middleware.RequirePermission("role_edit"), controllers.RoleUpdate)
		auth.GET("/role/delete/:id", middleware.RequirePermission("role_delete"), controllers.RoleDelete)
	}
}
