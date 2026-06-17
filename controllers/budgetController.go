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

func BudgetIndex(c *gin.Context) {
	renderBudgetPage(c, buildBudgetService(), "")
}

func BudgetStore(c *gin.Context) {
	input, errMessage := bindBudgetCreateInput(c)
	service := buildBudgetService()
	if errMessage != "" {
		renderBudgetPage(c, service, errMessage)
		return
	}

	if err := service.CreateBudget(input); err != nil {
		renderBudgetPage(c, service, err.Error())
		return
	}

	c.Redirect(http.StatusSeeOther, "/budgets")
}

func BudgetUpdate(c *gin.Context) {
	input, errMessage := bindBudgetUpdateInput(c)
	service := buildBudgetService()
	if errMessage != "" {
		renderBudgetPage(c, service, errMessage)
		return
	}

	if err := service.UpdateBudget(input); err != nil {
		renderBudgetPage(c, service, err.Error())
		return
	}

	c.Redirect(http.StatusSeeOther, "/budgets")
}

func BudgetDelete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.String(http.StatusBadRequest, "invalid budget id")
		return
	}

	if err := buildBudgetService().DeleteBudget(id); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.Redirect(http.StatusSeeOther, "/budgets")
}

func renderBudgetPage(c *gin.Context, service *services.BudgetService, message string) {
	budgets, err := service.GetBudgets()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	stores, err := (&repositories.StoreRepository{DB: config.DB}).GetAll()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	divisions, err := (&repositories.DivisionRepository{DB: config.DB}).GetAll()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	glAccounts, err := (&repositories.GLAccountRepository{DB: config.DB}).GetAll()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	Render(c, "budget.html", gin.H{
		"Title":      "Budget Management",
		"Page":       "budget",
		"Budgets":    budgets,
		"Stores":     stores,
		"Divisions":  divisions,
		"GLAccounts": glAccounts,
		"Error":      message,
	})
}

func buildBudgetService() *services.BudgetService {
	return &services.BudgetService{
		Repo:         &repositories.BudgetRepository{DB: config.DB},
		StoreRepo:    &repositories.StoreRepository{DB: config.DB},
		DivisionRepo: &repositories.DivisionRepository{DB: config.DB},
		GLRepo:       &repositories.GLAccountRepository{DB: config.DB},
	}
}

func bindBudgetCreateInput(c *gin.Context) (models.BudgetCreateInput, string) {
	fiscalYear, err := strconv.Atoi(strings.TrimSpace(c.PostForm("fiscal_year")))
	if err != nil {
		return models.BudgetCreateInput{}, "fiscal year tidak valid"
	}
	glAccountID, err := strconv.Atoi(strings.TrimSpace(c.PostForm("gl_account_id")))
	if err != nil {
		return models.BudgetCreateInput{}, "GL account wajib dipilih"
	}
	amount, err := strconv.ParseFloat(strings.ReplaceAll(strings.TrimSpace(c.PostForm("amount")), ",", ""), 64)
	if err != nil {
		return models.BudgetCreateInput{}, "amount tidak valid"
	}

	return models.BudgetCreateInput{
		FiscalYear:  fiscalYear,
		PeriodType:  c.PostForm("period_type"),
		PeriodKey:   c.PostForm("period_key"),
		StoreID:     optionalFormInt(c.PostForm("store_id")),
		DivisionID:  optionalFormInt(c.PostForm("division_id")),
		GLAccountID: glAccountID,
		Amount:      amount,
	}, ""
}

func bindBudgetUpdateInput(c *gin.Context) (models.BudgetUpdateInput, string) {
	id, err := strconv.ParseInt(strings.TrimSpace(c.PostForm("id")), 10, 64)
	if err != nil {
		return models.BudgetUpdateInput{}, "budget tidak valid"
	}

	createInput, message := bindBudgetCreateInput(c)
	if message != "" {
		return models.BudgetUpdateInput{}, message
	}

	return models.BudgetUpdateInput{
		ID:          id,
		FiscalYear:  createInput.FiscalYear,
		PeriodType:  createInput.PeriodType,
		PeriodKey:   createInput.PeriodKey,
		StoreID:     createInput.StoreID,
		DivisionID:  createInput.DivisionID,
		GLAccountID: createInput.GLAccountID,
		Amount:      createInput.Amount,
	}, ""
}

func optionalFormInt(value string) int {
	id, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return 0
	}
	return id
}
