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

func DivisionIndex(c *gin.Context) {
	service := &services.DivisionService{Repo: &repositories.DivisionRepository{DB: config.DB}}
	renderDivisionPage(c, service, "")
}

func DivisionStore(c *gin.Context) {
	type divisionForm struct {
		DivisionCode string `form:"division_code" binding:"required"`
		DivisionName string `form:"division_name" binding:"required"`
	}

	var form divisionForm
	service := &services.DivisionService{Repo: &repositories.DivisionRepository{DB: config.DB}}
	if err := c.ShouldBind(&form); err != nil {
		renderDivisionPage(c, service, "Form tidak lengkap")
		return
	}

	if err := service.CreateDivision(models.DivisionCreateInput{DivisionCode: form.DivisionCode, DivisionName: form.DivisionName}); err != nil {
		renderDivisionPage(c, service, err.Error())
		return
	}

	c.Redirect(http.StatusSeeOther, "/divisions")
}

func DivisionUpdate(c *gin.Context) {
	type divisionForm struct {
		ID           int    `form:"id" binding:"required"`
		DivisionCode string `form:"division_code" binding:"required"`
		DivisionName string `form:"division_name" binding:"required"`
	}

	var form divisionForm
	service := &services.DivisionService{Repo: &repositories.DivisionRepository{DB: config.DB}}
	if err := c.ShouldBind(&form); err != nil {
		renderDivisionPage(c, service, "Form tidak lengkap")
		return
	}

	if err := service.UpdateDivision(models.DivisionUpdateInput{ID: form.ID, DivisionCode: form.DivisionCode, DivisionName: form.DivisionName}); err != nil {
		renderDivisionPage(c, service, err.Error())
		return
	}

	c.Redirect(http.StatusSeeOther, "/divisions")
}

func DivisionDelete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.String(http.StatusBadRequest, "invalid division id")
		return
	}

	service := &services.DivisionService{Repo: &repositories.DivisionRepository{DB: config.DB}}
	if err := service.DeleteDivision(id); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.Redirect(http.StatusSeeOther, "/divisions")
}

func renderDivisionPage(c *gin.Context, service *services.DivisionService, message string) {
	divisions, err := service.GetDivisions()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	Render(c, "division.html", gin.H{
		"Title":     "Division Management",
		"Page":      "division",
		"Divisions": divisions,
		"Error":     message,
	})
}
