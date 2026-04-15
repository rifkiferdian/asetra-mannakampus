package controllers

import (
	"github.com/gin-gonic/gin"
)

func DashboardIndex(c *gin.Context) {
	Render(c, "dashboard.html", gin.H{
		"Title": "Dashboard",
		"Page":  "dashboard",
	})

}
