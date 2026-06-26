package services

import (
	"gobase-app/models"
	"gobase-app/repositories"
)

type PermissionService struct {
	Repo *repositories.PermissionRepository
}

func (s *PermissionService) GetGroupedPermissions() ([]models.PermissionGroup, error) {
	return s.Repo.GetGrouped()
}

func (s *PermissionService) EnsureSystemPermissions() error {
	defs := []models.PermissionDefinition{
		{Name: "dashboard_access", Group: "dashboard", GuardName: "web"},
		{Name: "document_type_management_access", Group: "document_type", GuardName: "web"},
		{Name: "report_management_access", Group: "report", GuardName: "web"},
		{Name: "role_management_access", Group: "role", GuardName: "web"},
		{Name: "role_view", Group: "role", GuardName: "web"},
		{Name: "role_create", Group: "role", GuardName: "web"},
		{Name: "role_edit", Group: "role", GuardName: "web"},
		{Name: "role_delete", Group: "role", GuardName: "web"},
		{Name: "user_management_access", Group: "user", GuardName: "web"},
		{Name: "user_view", Group: "user", GuardName: "web"},
		{Name: "user_create", Group: "user", GuardName: "web"},
		{Name: "user_edit", Group: "user", GuardName: "web"},
		{Name: "user_delete", Group: "user", GuardName: "web"},
		{Name: "store_management_access", Group: "store", GuardName: "web"},
		{Name: "store_view", Group: "store", GuardName: "web"},
		{Name: "store_create", Group: "store", GuardName: "web"},
		{Name: "store_edit", Group: "store", GuardName: "web"},
		{Name: "store_delete", Group: "store", GuardName: "web"},
		{Name: "vendor_management_access", Group: "vendor", GuardName: "web"},
		{Name: "vendor_view", Group: "vendor", GuardName: "web"},
		{Name: "vendor_create", Group: "vendor", GuardName: "web"},
		{Name: "vendor_edit", Group: "vendor", GuardName: "web"},
		{Name: "vendor_delete", Group: "vendor", GuardName: "web"},
		{Name: "gl_account_management_access", Group: "gl_account", GuardName: "web"},
		{Name: "gl_account_view", Group: "gl_account", GuardName: "web"},
		{Name: "gl_account_create", Group: "gl_account", GuardName: "web"},
		{Name: "gl_account_edit", Group: "gl_account", GuardName: "web"},
		{Name: "gl_account_delete", Group: "gl_account", GuardName: "web"},
		{Name: "division_management_access", Group: "division", GuardName: "web"},
		{Name: "division_view", Group: "division", GuardName: "web"},
		{Name: "division_create", Group: "division", GuardName: "web"},
		{Name: "division_edit", Group: "division", GuardName: "web"},
		{Name: "division_delete", Group: "division", GuardName: "web"},
		{Name: "budget_management_access", Group: "budget", GuardName: "web"},
		{Name: "budget_view", Group: "budget", GuardName: "web"},
		{Name: "budget_create", Group: "budget", GuardName: "web"},
		{Name: "budget_edit", Group: "budget", GuardName: "web"},
		{Name: "budget_delete", Group: "budget", GuardName: "web"},
		{Name: "store_approver_management_access", Group: "store_approver", GuardName: "web"},
		{Name: "store_approver_view", Group: "store_approver", GuardName: "web"},
		{Name: "store_approver_create", Group: "store_approver", GuardName: "web"},
		{Name: "store_approver_edit", Group: "store_approver", GuardName: "web"},
		{Name: "store_approver_delete", Group: "store_approver", GuardName: "web"},
		{Name: "approval_rule_management_access", Group: "approval_rule", GuardName: "web"},
		{Name: "approval_rule_view", Group: "approval_rule", GuardName: "web"},
		{Name: "approval_rule_create", Group: "approval_rule", GuardName: "web"},
		{Name: "approval_rule_edit", Group: "approval_rule", GuardName: "web"},
		{Name: "approval_rule_delete", Group: "approval_rule", GuardName: "web"},
		{Name: "approval_task_management_access", Group: "approval_task", GuardName: "web"},
		{Name: "approval_task_view", Group: "approval_task", GuardName: "web"},
		{Name: "approval_task_approve", Group: "approval_task", GuardName: "web"},
		{Name: "approval_task_reject", Group: "approval_task", GuardName: "web"},
		{Name: "purchase_request_management_access", Group: "purchase_request", GuardName: "web"},
		{Name: "purchase_request_view", Group: "purchase_request", GuardName: "web"},
		{Name: "purchase_request_create", Group: "purchase_request", GuardName: "web"},
	}

	if err := s.Repo.EnsurePermissions(defs); err != nil {
		return err
	}

	permNames := make([]string, 0, len(defs))
	for _, def := range defs {
		permNames = append(permNames, def.Name)
	}

	if err := s.Repo.GrantPermissionsToRoles(permNames, []string{"super-admin", "admin"}, "web"); err != nil {
		return err
	}

	if err := s.Repo.GrantPermissionsToRoles(
		[]string{"dashboard_access"},
		[]string{"super-admin", "admin", "requester", "manager", "staff-counter", "ga-manager", "it-manager", "IT-manager", "finance-manager", "gm", "procurement", "warehouse"},
		"web",
	); err != nil {
		return err
	}

	return s.Repo.GrantPermissionsToRoles(
		[]string{"approval_task_management_access", "approval_task_view", "approval_task_approve", "approval_task_reject"},
		[]string{"manager", "finance-manager", "gm", "IT-manager"},
		"web",
	)
}
