package repositories

import (
	"database/sql"
	"gobase-app/models"
	"strings"
)

type PermissionRepository struct {
	DB *sql.DB
}

// GetGrouped returns permissions grouped by their group column.
func (r *PermissionRepository) GetGrouped() ([]models.PermissionGroup, error) {
	rows, err := r.DB.Query(`
		SELECT 
			id,
			name,
			COALESCE(` + "`group`" + `, '') AS group_name,
			guard_name
		FROM permissions
		ORDER BY group_name, name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	groupPermissions := make(map[string][]models.Permission)
	groupOrder := []string{}

	for rows.Next() {
		var perm models.Permission
		if err := rows.Scan(&perm.ID, &perm.Name, &perm.GroupName, &perm.GuardName); err != nil {
			return nil, err
		}

		groupKey := perm.GroupName
		if _, exists := groupPermissions[groupKey]; !exists {
			groupOrder = append(groupOrder, groupKey)
		}

		groupPermissions[groupKey] = append(groupPermissions[groupKey], perm)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	groups := make([]models.PermissionGroup, 0, len(groupPermissions))
	for _, key := range groupOrder {
		groups = append(groups, models.PermissionGroup{
			Key:         key,
			Label:       formatGroupLabel(key),
			Permissions: groupPermissions[key],
		})
	}

	return groups, nil
}

func formatGroupLabel(groupKey string) string {
	if groupKey == "" {
		return "Others"
	}

	// Replace separators with space and Title-case the group name for display.
	normalized := strings.ReplaceAll(groupKey, "_", " ")
	return strings.Title(normalized)
}

func (r *PermissionRepository) EnsurePermissions(defs []models.PermissionDefinition) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}

	for _, def := range defs {
		if _, err := tx.Exec(`
			INSERT INTO permissions (name, `+"`group`"+`, guard_name, created_at, updated_at)
			SELECT ?, ?, ?, NOW(), NOW()
			WHERE NOT EXISTS (
				SELECT 1 FROM permissions WHERE name = ? AND guard_name = ?
			)
		`, def.Name, def.Group, def.GuardName, def.Name, def.GuardName); err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func (r *PermissionRepository) GrantPermissionsToRoles(permissionNames []string, roleNames []string, guardName string) error {
	if len(permissionNames) == 0 || len(roleNames) == 0 {
		return nil
	}

	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}

	for _, roleName := range roleNames {
		for _, permName := range permissionNames {
			if _, err := tx.Exec(`
				INSERT INTO role_has_permissions (permission_id, role_id)
				SELECT p.id, r.id
				FROM permissions p
				JOIN roles r ON r.guard_name = p.guard_name
				WHERE p.name = ? AND p.guard_name = ? AND r.name = ?
				AND NOT EXISTS (
					SELECT 1
					FROM role_has_permissions rhp
					WHERE rhp.permission_id = p.id AND rhp.role_id = r.id
				)
			`, permName, guardName, roleName); err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	return tx.Commit()
}

