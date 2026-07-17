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
		{Name: "purchase_request_edit", Group: "purchase_request", GuardName: "web"},
		{Name: "purchase_order_management_access", Group: "purchase_order", GuardName: "web"},
		{Name: "purchase_order_view", Group: "purchase_order", GuardName: "web"},
		{Name: "purchase_order_create", Group: "purchase_order", GuardName: "web"},
		{Name: "purchase_order_edit", Group: "purchase_order", GuardName: "web"},
		{Name: "asset_type_management_access", Group: "asset_type", GuardName: "web"},
		{Name: "asset_type_view", Group: "asset_type", GuardName: "web"},
		{Name: "asset_type_create", Group: "asset_type", GuardName: "web"},
		{Name: "asset_type_edit", Group: "asset_type", GuardName: "web"},
		{Name: "asset_type_delete", Group: "asset_type", GuardName: "web"},
		{Name: "component_type_management_access", Group: "component_type", GuardName: "web"},
		{Name: "component_type_view", Group: "component_type", GuardName: "web"},
		{Name: "component_type_create", Group: "component_type", GuardName: "web"},
		{Name: "component_type_edit", Group: "component_type", GuardName: "web"},
		{Name: "component_type_delete", Group: "component_type", GuardName: "web"},
		{Name: "asset_location_management_access", Group: "asset_location", GuardName: "web"},
		{Name: "asset_location_view", Group: "asset_location", GuardName: "web"},
		{Name: "asset_location_create", Group: "asset_location", GuardName: "web"},
		{Name: "asset_location_edit", Group: "asset_location", GuardName: "web"},
		{Name: "asset_location_delete", Group: "asset_location", GuardName: "web"},
		{Name: "asset_management_access", Group: "asset", GuardName: "web"},
		{Name: "asset_view", Group: "asset", GuardName: "web"},
		{Name: "asset_create", Group: "asset", GuardName: "web"},
		{Name: "asset_edit", Group: "asset", GuardName: "web"},
		{Name: "asset_delete", Group: "asset", GuardName: "web"},
		{Name: "asset_component_management_access", Group: "asset_component", GuardName: "web"},
		{Name: "asset_component_view", Group: "asset_component", GuardName: "web"},
		{Name: "asset_component_create", Group: "asset_component", GuardName: "web"},
		{Name: "asset_component_edit", Group: "asset_component", GuardName: "web"},
		{Name: "asset_component_delete", Group: "asset_component", GuardName: "web"},
		{Name: "asset_movement_management_access", Group: "asset_movement", GuardName: "web"},
		{Name: "asset_movement_view", Group: "asset_movement", GuardName: "web"},
		{Name: "asset_movement_create", Group: "asset_movement", GuardName: "web"},
		{Name: "asset_component_movement_management_access", Group: "asset_component_movement", GuardName: "web"},
		{Name: "asset_component_movement_view", Group: "asset_component_movement", GuardName: "web"},
		{Name: "asset_component_movement_create", Group: "asset_component_movement", GuardName: "web"},
		{Name: "asset_depreciation_management_access", Group: "asset_depreciation", GuardName: "web"},
		{Name: "asset_depreciation_view", Group: "asset_depreciation", GuardName: "web"},
		{Name: "asset_depreciation_generate", Group: "asset_depreciation", GuardName: "web"},
		{Name: "asset_depreciation_post", Group: "asset_depreciation", GuardName: "web"},
		{Name: "asset_depreciation_profile_management_access", Group: "asset_depreciation", GuardName: "web"},
		{Name: "asset_depreciation_profile_edit", Group: "asset_depreciation", GuardName: "web"},
		{Name: "asset_depreciation_posting_history_access", Group: "asset_depreciation", GuardName: "web"},
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

	if err := s.Repo.GrantPermissionsToRoles(
		[]string{"approval_task_management_access", "approval_task_view", "approval_task_approve", "approval_task_reject"},
		[]string{"manager", "finance-manager", "gm", "IT-manager", "it-manager"},
		"web",
	); err != nil {
		return err
	}

	if err := s.Repo.GrantPermissionsToRoles(
		[]string{"purchase_order_management_access", "purchase_order_view", "purchase_order_create", "purchase_order_edit"},
		[]string{"procurement"},
		"web",
	); err != nil {
		return err
	}

	if err := s.Repo.GrantPermissionsToRoles(
		[]string{
			"asset_depreciation_management_access", "asset_depreciation_view", "asset_depreciation_generate", "asset_depreciation_post",
			"asset_depreciation_profile_management_access", "asset_depreciation_profile_edit", "asset_depreciation_posting_history_access",
		},
		[]string{"finance-manager"},
		"web",
	); err != nil {
		return err
	}

	return s.Repo.GrantPermissionsToRoles(
		[]string{
			"asset_type_management_access", "asset_type_view", "asset_type_create", "asset_type_edit", "asset_type_delete",
			"component_type_management_access", "component_type_view", "component_type_create", "component_type_edit", "component_type_delete",
			"asset_location_management_access", "asset_location_view", "asset_location_create", "asset_location_edit", "asset_location_delete",
			"asset_management_access", "asset_view", "asset_create", "asset_edit", "asset_delete",
			"asset_component_management_access", "asset_component_view", "asset_component_create", "asset_component_edit", "asset_component_delete",
			"asset_movement_management_access", "asset_movement_view", "asset_movement_create",
			"asset_component_movement_management_access", "asset_component_movement_view", "asset_component_movement_create",
		},
		[]string{"IT-manager", "it-manager", "warehouse", "procurement"},
		"web",
	)
}
