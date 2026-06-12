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

func VendorIndex(c *gin.Context) {
	vendorRepo := &repositories.VendorRepository{DB: config.DB}
	vendorService := &services.VendorService{Repo: vendorRepo}

	renderVendorPage(c, vendorService, "")
}

func VendorStore(c *gin.Context) {
	type vendorForm struct {
		Name     string `form:"name" binding:"required"`
		Phone    string `form:"phone"`
		Email    string `form:"email"`
		Address  string `form:"address"`
		IsActive string `form:"is_active"`
	}

	var (
		form          vendorForm
		vendorRepo    = &repositories.VendorRepository{DB: config.DB}
		vendorService = &services.VendorService{Repo: vendorRepo}
	)

	if err := c.ShouldBind(&form); err != nil {
		renderVendorPage(c, vendorService, "Form tidak lengkap")
		return
	}

	input := models.VendorCreateInput{
		Name:     form.Name,
		Phone:    form.Phone,
		Email:    form.Email,
		Address:  form.Address,
		IsActive: form.IsActive != "0",
	}

	if err := vendorService.CreateVendor(input); err != nil {
		renderVendorPage(c, vendorService, err.Error())
		return
	}

	c.Redirect(http.StatusSeeOther, "/vendors")
}

func VendorUpdate(c *gin.Context) {
	type vendorForm struct {
		ID       int64  `form:"id" binding:"required"`
		Name     string `form:"name" binding:"required"`
		Phone    string `form:"phone"`
		Email    string `form:"email"`
		Address  string `form:"address"`
		IsActive string `form:"is_active"`
	}

	var (
		form          vendorForm
		vendorRepo    = &repositories.VendorRepository{DB: config.DB}
		vendorService = &services.VendorService{Repo: vendorRepo}
	)

	if err := c.ShouldBind(&form); err != nil {
		renderVendorPage(c, vendorService, "Form tidak lengkap")
		return
	}

	input := models.VendorUpdateInput{
		ID:       form.ID,
		Name:     form.Name,
		Phone:    form.Phone,
		Email:    form.Email,
		Address:  form.Address,
		IsActive: form.IsActive != "0",
	}

	if err := vendorService.UpdateVendor(input); err != nil {
		renderVendorPage(c, vendorService, err.Error())
		return
	}

	c.Redirect(http.StatusSeeOther, "/vendors")
}

func VendorDelete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.String(http.StatusBadRequest, "invalid vendor id")
		return
	}

	vendorRepo := &repositories.VendorRepository{DB: config.DB}
	vendorService := &services.VendorService{Repo: vendorRepo}
	if err := vendorService.DeleteVendor(id); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.Redirect(http.StatusSeeOther, "/vendors")
}

func renderVendorPage(c *gin.Context, vendorService *services.VendorService, message string) {
	vendors, err := vendorService.GetVendors()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	Render(c, "vendor.html", gin.H{
		"Title":   "Vendor Management",
		"Page":    "vendor",
		"vendors": vendors,
		"Error":   message,
	})
}
