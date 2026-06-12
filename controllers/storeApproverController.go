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

func StoreApproverIndex(c *gin.Context) {
	repo := &repositories.StoreApproverRepository{DB: config.DB}
	service := &services.StoreApproverService{Repo: repo}

	renderStoreApproverPage(c, service, "")
}

func StoreApproverStore(c *gin.Context) {
	type storeApproverForm struct {
		StoreID  int    `form:"store_id" binding:"required"`
		RoleID   int64  `form:"role_id" binding:"required"`
		UserID   int    `form:"user_id" binding:"required"`
		IsActive string `form:"is_active"`
	}

	var (
		form    storeApproverForm
		repo    = &repositories.StoreApproverRepository{DB: config.DB}
		service = &services.StoreApproverService{Repo: repo}
	)

	if err := c.ShouldBind(&form); err != nil {
		renderStoreApproverPage(c, service, "Form tidak lengkap")
		return
	}

	input := models.StoreApproverCreateInput{
		StoreID:  form.StoreID,
		RoleID:   form.RoleID,
		UserID:   form.UserID,
		IsActive: form.IsActive != "0",
	}

	if err := service.CreateStoreApprover(input); err != nil {
		renderStoreApproverPage(c, service, err.Error())
		return
	}

	c.Redirect(http.StatusSeeOther, "/store-approvers")
}

func StoreApproverUpdate(c *gin.Context) {
	type storeApproverForm struct {
		ID       int64  `form:"id" binding:"required"`
		StoreID  int    `form:"store_id" binding:"required"`
		RoleID   int64  `form:"role_id" binding:"required"`
		UserID   int    `form:"user_id" binding:"required"`
		IsActive string `form:"is_active"`
	}

	var (
		form    storeApproverForm
		repo    = &repositories.StoreApproverRepository{DB: config.DB}
		service = &services.StoreApproverService{Repo: repo}
	)

	if err := c.ShouldBind(&form); err != nil {
		renderStoreApproverPage(c, service, "Form tidak lengkap")
		return
	}

	input := models.StoreApproverUpdateInput{
		ID:       form.ID,
		StoreID:  form.StoreID,
		RoleID:   form.RoleID,
		UserID:   form.UserID,
		IsActive: form.IsActive != "0",
	}

	if err := service.UpdateStoreApprover(input); err != nil {
		renderStoreApproverPage(c, service, err.Error())
		return
	}

	c.Redirect(http.StatusSeeOther, "/store-approvers")
}

func StoreApproverDelete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.String(http.StatusBadRequest, "invalid store approver id")
		return
	}

	repo := &repositories.StoreApproverRepository{DB: config.DB}
	service := &services.StoreApproverService{Repo: repo}
	if err := service.DeleteStoreApprover(id); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.Redirect(http.StatusSeeOther, "/store-approvers")
}

func renderStoreApproverPage(c *gin.Context, service *services.StoreApproverService, message string) {
	items, err := service.GetStoreApprovers()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	storeRepo := &repositories.StoreRepository{DB: config.DB}
	stores, err := storeRepo.GetAll()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	roleRepo := &repositories.RoleRepository{DB: config.DB}
	roles, err := roleRepo.GetAll()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	userRepo := &repositories.UserRepository{DB: config.DB}
	users, err := userRepo.GetAll()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	Render(c, "store_approver.html", gin.H{
		"Title":          "Store Approver Management",
		"Page":           "store_approver",
		"storeApprovers": items,
		"stores":         stores,
		"roles":          roles,
		"users":          users,
		"Error":          message,
	})
}
