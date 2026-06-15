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

	return s.Repo.GrantPermissionsToRoles(permNames, []string{"super-admin", "admin"}, "web")
}

