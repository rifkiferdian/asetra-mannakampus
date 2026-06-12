package controllers

import (
	"gobase-app/config"
	"gobase-app/models"
	"gobase-app/repositories"
	"gobase-app/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GLAccountIndex(c *gin.Context) {
	repo := &repositories.GLAccountRepository{DB: config.DB}
	service := &services.GLAccountService{Repo: repo}

	renderGLAccountPage(c, service, "")
}

func GLAccountStore(c *gin.Context) {
	type glAccountForm struct {
		GLCode    string `form:"gl_code" binding:"required"`
		GLName    string `form:"gl_name" binding:"required"`
		SpendType string `form:"spend_type" binding:"required"`
		IsActive  string `form:"is_active"`
	}

	var (
		form    glAccountForm
		repo    = &repositories.GLAccountRepository{DB: config.DB}
		service = &services.GLAccountService{Repo: repo}
	)

	if err := c.ShouldBind(&form); err != nil {
		renderGLAccountPage(c, service, "Form tidak lengkap")
		return
	}

	input := models.GLAccountCreateInput{
		GLCode:    form.GLCode,
		GLName:    form.GLName,
		SpendType: form.SpendType,
		IsActive:  form.IsActive != "0",
	}

	if err := service.CreateGLAccount(input); err != nil {
		renderGLAccountPage(c, service, err.Error())
		return
	}

	c.Redirect(http.StatusSeeOther, "/gl-accounts")
}

func GLAccountUpdate(c *gin.Context) {
	type glAccountForm struct {
		ID        int    `form:"id" binding:"required"`
		GLCode    string `form:"gl_code" binding:"required"`
		GLName    string `form:"gl_name" binding:"required"`
		SpendType string `form:"spend_type" binding:"required"`
		IsActive  string `form:"is_active"`
	}

	var (
		form    glAccountForm
		repo    = &repositories.GLAccountRepository{DB: config.DB}
		service = &services.GLAccountService{Repo: repo}
	)

	if err := c.ShouldBind(&form); err != nil {
		renderGLAccountPage(c, service, "Form tidak lengkap")
		return
	}

	input := models.GLAccountUpdateInput{
		ID:        form.ID,
		GLCode:    form.GLCode,
		GLName:    form.GLName,
		SpendType: form.SpendType,
		IsActive:  form.IsActive != "0",
	}

	if err := service.UpdateGLAccount(input); err != nil {
		renderGLAccountPage(c, service, err.Error())
		return
	}

	c.Redirect(http.StatusSeeOther, "/gl-accounts")
}

func GLAccountDelete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.String(http.StatusBadRequest, "invalid gl account id")
		return
	}

	repo := &repositories.GLAccountRepository{DB: config.DB}
	service := &services.GLAccountService{Repo: repo}
	if err := service.DeleteGLAccount(id); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.Redirect(http.StatusSeeOther, "/gl-accounts")
}

func renderGLAccountPage(c *gin.Context, service *services.GLAccountService, message string) {
	accounts, err := service.GetGLAccounts()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	Render(c, "gl_account.html", gin.H{
		"Title":      "GL Account Management",
		"Page":       "gl_account",
		"glAccounts": accounts,
		"Error":      message,
	})
}
