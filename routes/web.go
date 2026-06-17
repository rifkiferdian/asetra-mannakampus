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
		auth.GET("/purchase-requests", middleware.RequirePermission("purchase_request_management_access"), controllers.PurchaseRequestIndex)
		auth.GET("/purchase-requests/form", middleware.RequirePermission("purchase_request_create"), controllers.PurchaseRequestFormIndex)
		auth.POST("/purchase-requests", middleware.RequirePermission("purchase_request_create"), controllers.PurchaseRequestStore)

		auth.GET("/stores", controllers.StoreIndex)
		auth.POST("/stores", controllers.StoreStore)
		auth.POST("/stores/update", controllers.StoreUpdate)
		auth.GET("/stores/delete/:id", controllers.StoreDelete)
		auth.GET("/vendors", middleware.RequirePermission("vendor_management_access"), controllers.VendorIndex)
		auth.POST("/vendors", middleware.RequirePermission("vendor_create"), controllers.VendorStore)
		auth.POST("/vendors/update", middleware.RequirePermission("vendor_edit"), controllers.VendorUpdate)
		auth.GET("/vendors/delete/:id", middleware.RequirePermission("vendor_delete"), controllers.VendorDelete)
		auth.GET("/gl-accounts", middleware.RequirePermission("gl_account_management_access"), controllers.GLAccountIndex)
		auth.POST("/gl-accounts", middleware.RequirePermission("gl_account_create"), controllers.GLAccountStore)
		auth.POST("/gl-accounts/update", middleware.RequirePermission("gl_account_edit"), controllers.GLAccountUpdate)
		auth.GET("/gl-accounts/delete/:id", middleware.RequirePermission("gl_account_delete"), controllers.GLAccountDelete)
		auth.GET("/divisions", middleware.RequirePermission("division_management_access"), controllers.DivisionIndex)
		auth.POST("/divisions", middleware.RequirePermission("division_create"), controllers.DivisionStore)
		auth.POST("/divisions/update", middleware.RequirePermission("division_edit"), controllers.DivisionUpdate)
		auth.GET("/divisions/delete/:id", middleware.RequirePermission("division_delete"), controllers.DivisionDelete)
		auth.GET("/budgets", middleware.RequirePermission("budget_management_access"), controllers.BudgetIndex)
		auth.POST("/budgets", middleware.RequirePermission("budget_create"), controllers.BudgetStore)
		auth.POST("/budgets/update", middleware.RequirePermission("budget_edit"), controllers.BudgetUpdate)
		auth.GET("/budgets/delete/:id", middleware.RequirePermission("budget_delete"), controllers.BudgetDelete)
		auth.GET("/store-approvers", middleware.RequirePermission("store_approver_management_access"), controllers.StoreApproverIndex)
		auth.POST("/store-approvers", middleware.RequirePermission("store_approver_create"), controllers.StoreApproverStore)
		auth.POST("/store-approvers/update", middleware.RequirePermission("store_approver_edit"), controllers.StoreApproverUpdate)
		auth.GET("/store-approvers/delete/:id", middleware.RequirePermission("store_approver_delete"), controllers.StoreApproverDelete)
		auth.GET("/approval-rules", middleware.RequirePermission("approval_rule_management_access"), controllers.ApprovalRuleIndex)
		auth.GET("/approval-rules/form", middleware.RequirePermission("approval_rule_create"), controllers.ApprovalRuleFormIndex)
		auth.GET("/approval-rules/:id/edit", middleware.RequirePermission("approval_rule_edit"), controllers.ApprovalRuleEdit)
		auth.POST("/approval-rules", middleware.RequirePermission("approval_rule_create"), controllers.ApprovalRuleStore)
		auth.POST("/approval-rules/update", middleware.RequirePermission("approval_rule_edit"), controllers.ApprovalRuleUpdate)
		auth.GET("/approval-rules/delete/:id", middleware.RequirePermission("approval_rule_delete"), controllers.ApprovalRuleDelete)
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
