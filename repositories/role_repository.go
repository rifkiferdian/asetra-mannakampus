package repositories

import (
	"database/sql"
	"gobase-app/models"
	"strings"
)

type RoleRepository struct {
	DB *sql.DB
}

// RoleCreateParams menampung data yang diperlukan untuk menyimpan role baru.
type RoleCreateParams struct {
	Name          string
	GuardName     string
	IsAdmin       bool
	PermissionIDs []int64
}

// RoleUpdateParams menampung data yang diperlukan untuk memperbarui role.
type RoleUpdateParams struct {
	ID            int
	Name          string
	GuardName     string
	IsAdmin       bool
	PermissionIDs []int64
}

// GetAll mengambil seluruh data role beserta jumlah permission dan user yang terkait.
func (r *RoleRepository) GetAll() ([]models.Role, error) {
	rows, err := r.DB.Query(`
		SELECT 
			r.id,
			r.name,
			COUNT(DISTINCT rhp.permission_id) AS permission_count,
			COUNT(DISTINCT mhr.model_id) AS user_count,
			r.updated_at
		FROM roles r
		LEFT JOIN role_has_permissions rhp ON rhp.role_id = r.id
		LEFT JOIN model_has_roles mhr ON mhr.role_id = r.id
		GROUP BY r.id, r.name, r.updated_at
		ORDER BY r.updated_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []models.Role

	for rows.Next() {
		var (
			role      models.Role
			updatedAt sql.NullTime
		)

		if err := rows.Scan(
			&role.ID,
			&role.Name,
			&role.PermissionCount,
			&role.UserCount,
			&updatedAt,
		); err != nil {
			return nil, err
		}

		if updatedAt.Valid {
			role.UpdatedAt = updatedAt.Time.Format("01-02-2006 15:04:05")
		} else {
			role.UpdatedAt = "-"
		}

		roles = append(roles, role)
	}

	return roles, rows.Err()
}

// ExistsByNameAndGuard mengecek apakah kombinasi name + guard_name sudah ada.
func (r *RoleRepository) ExistsByNameAndGuard(name, guardName string) (bool, error) {
	var count int
	err := r.DB.QueryRow(`SELECT COUNT(1) FROM roles WHERE name = ? AND guard_name = ?`, name, guardName).Scan(&count)
	return count > 0, err
}

// ExistsByNameAndGuardExceptID mengecek apakah kombinasi name + guard_name sudah ada di role lain.
func (r *RoleRepository) ExistsByNameAndGuardExceptID(name, guardName string, excludeID int) (bool, error) {
	var count int
	err := r.DB.QueryRow(`SELECT COUNT(1) FROM roles WHERE name = ? AND guard_name = ? AND id <> ?`, name, guardName, excludeID).Scan(&count)
	return count > 0, err
}

// GetByID mengambil detail role dan permission yang dimilikinya.
func (r *RoleRepository) GetByID(id int) (*models.RoleDetail, error) {
	var role models.RoleDetail
	if err := r.DB.QueryRow(`SELECT id, name, guard_name, is_admin FROM roles WHERE id = ?`, id).
		Scan(&role.ID, &role.Name, &role.GuardName, &role.IsAdmin); err != nil {
		return nil, err
	}

	rows, err := r.DB.Query(`SELECT permission_id FROM role_has_permissions WHERE role_id = ?`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var permID int64
		if err := rows.Scan(&permID); err != nil {
			return nil, err
		}
		role.PermissionIDs = append(role.PermissionIDs, permID)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &role, nil
}

// FindExistingPermissionIDs mengembalikan map id permission yang ditemukan di database.
func (r *RoleRepository) FindExistingPermissionIDs(ids []int64) (map[int64]bool, error) {
	result := make(map[int64]bool)
	if len(ids) == 0 {
		return result, nil
	}

	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}

	query := `SELECT id FROM permissions WHERE id IN (` + strings.Join(placeholders, ",") + `)`
	rows, err := r.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		result[id] = true
	}

	return result, rows.Err()
}

// CreateRoleWithPermissions menyimpan role baru beserta relasi permission dalam satu transaksi.
func (r *RoleRepository) CreateRoleWithPermissions(params RoleCreateParams) (int64, error) {
	tx, err := r.DB.Begin()
	if err != nil {
		return 0, err
	}

	res, err := tx.Exec(`
		INSERT INTO roles (name, guard_name, is_admin)
		VALUES (?, ?, ?)
	`, params.Name, params.GuardName, params.IsAdmin)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	roleID, err := res.LastInsertId()
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	if len(params.PermissionIDs) > 0 {
		stmt, err := tx.Prepare(`INSERT INTO role_has_permissions (permission_id, role_id) VALUES (?, ?)`)
		if err != nil {
			tx.Rollback()
			return 0, err
		}
		defer stmt.Close()

		for _, permID := range params.PermissionIDs {
			if _, err := stmt.Exec(permID, roleID); err != nil {
				tx.Rollback()
				return 0, err
			}
		}
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return 0, err
	}

	return roleID, nil
}

// UpdateRoleWithPermissions memperbarui data role beserta relasi permission dalam satu transaksi.
func (r *RoleRepository) UpdateRoleWithPermissions(params RoleUpdateParams) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}

	if _, err := tx.Exec(`
		UPDATE roles 
		SET name = ?, guard_name = ?, is_admin = ?
		WHERE id = ?
	`, params.Name, params.GuardName, params.IsAdmin, params.ID); err != nil {
		tx.Rollback()
		return err
	}

	if _, err := tx.Exec(`DELETE FROM role_has_permissions WHERE role_id = ?`, params.ID); err != nil {
		tx.Rollback()
		return err
	}

	if len(params.PermissionIDs) > 0 {
		stmt, err := tx.Prepare(`INSERT INTO role_has_permissions (permission_id, role_id) VALUES (?, ?)`)
		if err != nil {
			tx.Rollback()
			return err
		}
		defer stmt.Close()

		for _, permID := range params.PermissionIDs {
			if _, err := stmt.Exec(permID, params.ID); err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	return tx.Commit()
}

// DeleteByID removes a role and its related mappings in a single transaction.
func (r *RoleRepository) DeleteByID(id int) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}

	if _, err := tx.Exec(`DELETE FROM role_has_permissions WHERE role_id = ?`, id); err != nil {
		tx.Rollback()
		return err
	}

	if _, err := tx.Exec(`DELETE FROM model_has_roles WHERE role_id = ?`, id); err != nil {
		tx.Rollback()
		return err
	}

	if _, err := tx.Exec(`DELETE FROM roles WHERE id = ?`, id); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

