package controllers

import (
	"gobase-app/config"
	"gobase-app/models"
	"gobase-app/repositories"
	"gobase-app/services"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func StoreIndex(c *gin.Context) {
	storeRepo := &repositories.StoreRepository{DB: config.DB}
	storeService := &services.StoreService{Repo: storeRepo}

	renderStorePage(c, storeService, "")
}

func StoreStore(c *gin.Context) {
	type storeForm struct {
		StoreID      string `form:"store_id" binding:"required"`
		StoreCode    string `form:"store_code" binding:"required"`
		StoreName    string `form:"store_name" binding:"required"`
		StoreAddress string `form:"store_address" binding:"required"`
		IsActive     string `form:"is_active"`
	}

	var (
		form         storeForm
		storeRepo    = &repositories.StoreRepository{DB: config.DB}
		storeService = &services.StoreService{Repo: storeRepo}
	)

	if err := c.ShouldBind(&form); err != nil {
		renderStorePage(c, storeService, "Form tidak lengkap")
		return
	}

	storeID, err := strconv.Atoi(strings.TrimSpace(form.StoreID))
	if err != nil {
		renderStorePage(c, storeService, "Store ID harus berupa angka")
		return
	}

	input := models.StoreCreateInput{
		StoreID:      storeID,
		StoreCode:    form.StoreCode,
		StoreName:    form.StoreName,
		StoreAddress: form.StoreAddress,
		IsActive:     form.IsActive != "0",
	}

	if err := storeService.CreateStore(input); err != nil {
		renderStorePage(c, storeService, err.Error())
		return
	}

	c.Redirect(http.StatusSeeOther, "/stores")
}

func StoreUpdate(c *gin.Context) {
	type storeForm struct {
		StoreID      int    `form:"store_id" binding:"required"`
		StoreCode    string `form:"store_code" binding:"required"`
		StoreName    string `form:"store_name" binding:"required"`
		StoreAddress string `form:"store_address" binding:"required"`
		IsActive     string `form:"is_active"`
	}

	var (
		form         storeForm
		storeRepo    = &repositories.StoreRepository{DB: config.DB}
		storeService = &services.StoreService{Repo: storeRepo}
	)

	if err := c.ShouldBind(&form); err != nil {
		renderStorePage(c, storeService, "Form tidak lengkap")
		return
	}

	input := models.StoreUpdateInput{
		StoreID:      form.StoreID,
		StoreCode:    form.StoreCode,
		StoreName:    form.StoreName,
		StoreAddress: form.StoreAddress,
		IsActive:     form.IsActive != "0",
	}

	if err := storeService.UpdateStore(input); err != nil {
		renderStorePage(c, storeService, err.Error())
		return
	}

	c.Redirect(http.StatusSeeOther, "/stores")
}

func StoreDelete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		c.String(http.StatusBadRequest, "invalid store id")
		return
	}

	storeRepo := &repositories.StoreRepository{DB: config.DB}
	storeService := &services.StoreService{Repo: storeRepo}

	if err := storeService.DeleteStore(id); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.Redirect(http.StatusSeeOther, "/stores")
}

func renderStorePage(c *gin.Context, storeService *services.StoreService, message string) {
	stores, err := storeService.GetStores()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	Render(c, "store.html", gin.H{
		"Title":  "Daftar Store",
		"Page":   "store",
		"stores": stores,
		"Error":  message,
	})
}
