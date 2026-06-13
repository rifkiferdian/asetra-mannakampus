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

func ApprovalRuleIndex(c *gin.Context) {
	repo := &repositories.ApprovalRuleRepository{DB: config.DB}
	service := &services.ApprovalRuleService{Repo: repo}

	rules, err := service.GetApprovalRules()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	Render(c, "approval_rule.html", gin.H{
		"Title": "Approval Rules",
		"Page":  "approval_rule",
		"Rules": rules,
	})
}

func ApprovalRuleFormIndex(c *gin.Context) {
	renderApprovalRuleForm(c, models.ApprovalRuleDetail{}, "", nil)
}

func ApprovalRuleEdit(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.String(http.StatusBadRequest, "invalid approval rule id")
		return
	}

	repo := &repositories.ApprovalRuleRepository{DB: config.DB}
	service := &services.ApprovalRuleService{Repo: repo}
	detail, err := service.GetApprovalRuleDetail(id)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	renderApprovalRuleEditForm(c, *detail, "", nil)
}

func ApprovalRuleStore(c *gin.Context) {
	input, detail, errMessage := bindApprovalRuleInput(c)
	if errMessage != "" {
		renderApprovalRuleForm(c, detail, errMessage, nil)
		return
	}

	repo := &repositories.ApprovalRuleRepository{DB: config.DB}
	service := &services.ApprovalRuleService{Repo: repo}
	if err := service.CreateApprovalRule(input); err != nil {
		renderApprovalRuleForm(c, detail, err.Error(), nil)
		return
	}

	c.Redirect(http.StatusSeeOther, "/approval-rules")
}

func ApprovalRuleUpdate(c *gin.Context) {
	id, err := strconv.ParseInt(strings.TrimSpace(c.PostForm("rule_id")), 10, 64)
	if err != nil || id <= 0 {
		renderApprovalRuleEditForm(c, models.ApprovalRuleDetail{}, "approval rule tidak valid", nil)
		return
	}

	createInput, detail, errMessage := bindApprovalRuleInput(c)
	if errMessage != "" {
		detail.ID = id
		renderApprovalRuleEditForm(c, detail, errMessage, nil)
		return
	}

	repo := &repositories.ApprovalRuleRepository{DB: config.DB}
	service := &services.ApprovalRuleService{Repo: repo}
	if err := service.UpdateApprovalRule(models.ApprovalRuleUpdateInput{
		ID:            id,
		Name:          createInput.Name,
		IsActive:      createInput.IsActive,
		MinAmount:     createInput.MinAmount,
		MaxAmount:     createInput.MaxAmount,
		LocationScope: createInput.LocationScope,
		SpendType:     createInput.SpendType,
		UrgentLevel:   createInput.UrgentLevel,
		Steps:         createInput.Steps,
	}); err != nil {
		detail.ID = id
		renderApprovalRuleEditForm(c, detail, err.Error(), nil)
		return
	}

	c.Redirect(http.StatusSeeOther, "/approval-rules")
}

func ApprovalRuleDelete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.String(http.StatusBadRequest, "invalid approval rule id")
		return
	}

	repo := &repositories.ApprovalRuleRepository{DB: config.DB}
	service := &services.ApprovalRuleService{Repo: repo}
	if err := service.DeleteApprovalRule(id); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.Redirect(http.StatusSeeOther, "/approval-rules")
}

func renderApprovalRuleForm(c *gin.Context, detail models.ApprovalRuleDetail, message string, statusCode *int) {
	renderApprovalRuleFormPage(c, "approval_rule_form.html", "approvalRuleForm", "Form Approval Rule", detail, message, statusCode)
}

func renderApprovalRuleEditForm(c *gin.Context, detail models.ApprovalRuleDetail, message string, statusCode *int) {
	renderApprovalRuleFormPage(c, "approval_rule_form_edit.html", "approvalRuleEdit", "Edit Approval Rule", detail, message, statusCode)
}

func renderApprovalRuleFormPage(c *gin.Context, templateName, page, title string, detail models.ApprovalRuleDetail, message string, statusCode *int) {
	roleRepo := &repositories.RoleRepository{DB: config.DB}
	roles, err := roleRepo.GetAll()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	if len(detail.Steps) == 0 {
		detail.Steps = []models.ApprovalRuleStep{{StepOrder: 1, Scope: "STORE", IsRequired: true}}
	}
	selectedMaxAmount := ""
	if detail.MaxAmount != nil {
		selectedMaxAmount = strconv.FormatFloat(*detail.MaxAmount, 'f', -1, 64)
	}

	data := gin.H{
		"Title":             title,
		"Page":              page,
		"Rule":              detail,
		"Roles":             roles,
		"Error":             message,
		"SelectedMaxAmount": selectedMaxAmount,
	}

	Render(c, templateName, data)
}

func bindApprovalRuleInput(c *gin.Context) (models.ApprovalRuleCreateInput, models.ApprovalRuleDetail, string) {
	name := strings.TrimSpace(c.PostForm("name"))
	locationScope := strings.TrimSpace(c.PostForm("location_scope"))
	spendType := strings.TrimSpace(c.PostForm("spend_type"))
	urgentLevel := strings.TrimSpace(c.PostForm("urgent_level"))
	isActive := c.PostForm("is_active") != "0"

	minAmount, err := strconv.ParseFloat(strings.TrimSpace(c.PostForm("min_amount")), 64)
	if err != nil {
		return models.ApprovalRuleCreateInput{}, models.ApprovalRuleDetail{}, "minimum amount harus berupa angka"
	}

	var maxAmount *float64
	maxAmountStr := strings.TrimSpace(c.PostForm("max_amount"))
	if maxAmountStr != "" {
		val, err := strconv.ParseFloat(maxAmountStr, 64)
		if err != nil {
			return models.ApprovalRuleCreateInput{}, models.ApprovalRuleDetail{}, "maximum amount harus berupa angka"
		}
		maxAmount = &val
	}

	stepOrders := c.PostFormArray("step_order[]")
	roleIDs := c.PostFormArray("role_id[]")
	scopes := c.PostFormArray("scope[]")
	parallelFlags := c.PostFormArray("is_parallel[]")
	requiredFlags := c.PostFormArray("is_required[]")

	if len(stepOrders) == 0 || len(roleIDs) == 0 || len(scopes) == 0 || len(parallelFlags) == 0 || len(requiredFlags) == 0 {
		return models.ApprovalRuleCreateInput{}, models.ApprovalRuleDetail{}, "minimal harus ada 1 approval step"
	}
	if len(stepOrders) != len(roleIDs) || len(stepOrders) != len(scopes) || len(stepOrders) != len(parallelFlags) || len(stepOrders) != len(requiredFlags) {
		return models.ApprovalRuleCreateInput{}, models.ApprovalRuleDetail{}, "data approval steps tidak sinkron"
	}

	steps := make([]models.ApprovalRuleStepInput, 0, len(stepOrders))
	detailSteps := make([]models.ApprovalRuleStep, 0, len(stepOrders))

	for i := range stepOrders {
		stepOrder, err := strconv.Atoi(strings.TrimSpace(stepOrders[i]))
		if err != nil {
			return models.ApprovalRuleCreateInput{}, models.ApprovalRuleDetail{}, "step order harus berupa angka"
		}
		roleID, err := strconv.ParseInt(strings.TrimSpace(roleIDs[i]), 10, 64)
		if err != nil {
			return models.ApprovalRuleCreateInput{}, models.ApprovalRuleDetail{}, "role approver tidak valid"
		}

		isParallel := strings.TrimSpace(parallelFlags[i]) == "1"
		isRequired := strings.TrimSpace(requiredFlags[i]) != "0"
		scope := strings.TrimSpace(scopes[i])

		steps = append(steps, models.ApprovalRuleStepInput{
			StepOrder:  stepOrder,
			RoleID:     roleID,
			Scope:      scope,
			IsParallel: isParallel,
			IsRequired: isRequired,
		})
		detailSteps = append(detailSteps, models.ApprovalRuleStep{
			StepOrder:  stepOrder,
			RoleID:     roleID,
			Scope:      scope,
			IsParallel: isParallel,
			IsRequired: isRequired,
		})
	}

	return models.ApprovalRuleCreateInput{
			Name:          name,
			IsActive:      isActive,
			MinAmount:     minAmount,
			MaxAmount:     maxAmount,
			LocationScope: locationScope,
			SpendType:     spendType,
			UrgentLevel:   urgentLevel,
			Steps:         steps,
		}, models.ApprovalRuleDetail{
			Name:          name,
			IsActive:      isActive,
			MinAmount:     minAmount,
			MaxAmount:     maxAmount,
			LocationScope: locationScope,
			SpendType:     spendType,
			UrgentLevel:   urgentLevel,
			Steps:         detailSteps,
		}, ""
}
