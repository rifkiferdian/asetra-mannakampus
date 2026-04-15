package controllers

import (
	"net/http"
	"gobase-app/config"
	"gobase-app/models"
	"gobase-app/repositories"
	"gobase-app/services"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func UserIndex(c *gin.Context) {
	userRepo := &repositories.UserRepository{DB: config.DB}
	userService := &services.UserService{Repo: userRepo}

	renderUserPage(c, userService, "")

}

func UserStore(c *gin.Context) {
	type userForm struct {
		Name     string `form:"name" binding:"required"`
		Username string `form:"username" binding:"required"`
		Password string `form:"password" binding:"required"`
		Email    string `form:"email"`
		NIP      string `form:"nip" binding:"required"`
		Status   string `form:"status"`
	}

	var (
		form     userForm
		userRepo = &repositories.UserRepository{DB: config.DB}
		userSvc  = &services.UserService{Repo: userRepo}
	)

	if err := c.ShouldBind(&form); err != nil {
		renderUserPage(c, userSvc, "Form tidak lengkap")
		return
	}

	nip, err := strconv.Atoi(strings.TrimSpace(form.NIP))
	if err != nil {
		renderUserPage(c, userSvc, "NIP harus berupa angka")
		return
	}

	storeIDs := []int{}
	for _, val := range c.PostFormArray("store_id") {
		if val == "" {
			continue
		}
		id, err := strconv.Atoi(val)
		if err != nil {
			renderUserPage(c, userSvc, "Store ID tidak valid")
			return
		}
		storeIDs = append(storeIDs, id)
	}

	input := models.UserCreateInput{
		NIP:       nip,
		Name:      strings.TrimSpace(form.Name),
		Username:  strings.TrimSpace(form.Username),
		Password:  form.Password,
		Email:     strings.TrimSpace(form.Email),
		Status:    form.Status,
		StoreIDs:  storeIDs,
		RoleNames: c.PostFormArray("roles"),
	}

	if err := userSvc.CreateUser(input); err != nil {
		renderUserPage(c, userSvc, err.Error())
		return
	}

	c.Redirect(http.StatusSeeOther, "/users")
}

// UserUpdate memperbarui data user yang sudah ada.
func UserUpdate(c *gin.Context) {
	type userUpdateForm struct {
		ID       int    `form:"user_id" binding:"required"`
		Name     string `form:"name" binding:"required"`
		Username string `form:"username" binding:"required"`
		Password string `form:"password"`
		Email    string `form:"email"`
		NIP      string `form:"nip" binding:"required"`
		Status   string `form:"status"`
	}

	var (
		form     userUpdateForm
		userRepo = &repositories.UserRepository{DB: config.DB}
		userSvc  = &services.UserService{Repo: userRepo}
	)

	if err := c.ShouldBind(&form); err != nil {
		renderUserPage(c, userSvc, "Form tidak lengkap")
		return
	}

	nip, err := strconv.Atoi(strings.TrimSpace(form.NIP))
	if err != nil {
		renderUserPage(c, userSvc, "NIP harus berupa angka")
		return
	}

	storeIDs := []int{}
	for _, val := range c.PostFormArray("store_id") {
		if val == "" {
			continue
		}
		id, err := strconv.Atoi(val)
		if err != nil {
			renderUserPage(c, userSvc, "Store ID tidak valid")
			return
		}
		storeIDs = append(storeIDs, id)
	}

	input := models.UserUpdateInput{
		ID:        form.ID,
		NIP:       nip,
		Name:      strings.TrimSpace(form.Name),
		Username:  strings.TrimSpace(form.Username),
		Password:  form.Password,
		Email:     strings.TrimSpace(form.Email),
		Status:    form.Status,
		StoreIDs:  storeIDs,
		RoleNames: c.PostFormArray("roles"),
	}

	if err := userSvc.UpdateUser(input); err != nil {
		renderUserPage(c, userSvc, err.Error())
		return
	}

	c.Redirect(http.StatusSeeOther, "/users")
}

// UserDelete menghapus data user berdasarkan ID.
func UserDelete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		c.String(http.StatusBadRequest, "invalid user id")
		return
	}

	userRepo := &repositories.UserRepository{DB: config.DB}
	userService := &services.UserService{Repo: userRepo}

	if err := userService.DeleteUser(id); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.Redirect(http.StatusSeeOther, "/users")
}

func renderUserPage(c *gin.Context, userService *services.UserService, message string) {
	users, err := userService.GetUsers()
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

	storeRepo := &repositories.StoreRepository{DB: config.DB}
	stores, err := storeRepo.GetAll()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	Render(c, "user.html", gin.H{
		"Title":  "Daftar User",
		"Page":   "user",
		"users":  users,
		"roles":  roles,
		"stores": stores,
		"Error":  message,
	})
}

