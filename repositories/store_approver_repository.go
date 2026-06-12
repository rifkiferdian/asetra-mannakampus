package repositories

import (
	"database/sql"
	"gobase-app/models"
)

type StoreApproverRepository struct {
	DB *sql.DB
}

func (r *StoreApproverRepository) GetAll() ([]models.StoreApprover, error) {
	rows, err := r.DB.Query(`
		SELECT
			sa.id,
			sa.store_id,
			COALESCE(s.store_name, '') AS store_name,
			sa.role_id,
			COALESCE(r.name, '') AS role_name,
			sa.user_id,
			COALESCE(u.name, '') AS user_name,
			COALESCE(u.username, '') AS username,
			sa.is_active
		FROM store_approvers sa
		LEFT JOIN stores s ON s.store_id = sa.store_id
		LEFT JOIN roles r ON r.id = sa.role_id
		LEFT JOIN users u ON u.id = sa.user_id
		ORDER BY sa.id ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.StoreApprover
	for rows.Next() {
		var (
			item     models.StoreApprover
			isActive int
		)
		if err := rows.Scan(
			&item.ID,
			&item.StoreID,
			&item.StoreName,
			&item.RoleID,
			&item.RoleName,
			&item.UserID,
			&item.UserName,
			&item.Username,
			&isActive,
		); err != nil {
			return nil, err
		}
		item.IsActive = isActive == 1
		if item.IsActive {
			item.IsActiveLabel = "Aktif"
		} else {
			item.IsActiveLabel = "Non Aktif"
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

func (r *StoreApproverRepository) ExistsByID(id int64) (bool, error) {
	var count int
	err := r.DB.QueryRow(`SELECT COUNT(1) FROM store_approvers WHERE id = ?`, id).Scan(&count)
	return count > 0, err
}

func (r *StoreApproverRepository) ExistsDuplicate(storeID int, roleID int64, userID int, excludeID int64) (bool, error) {
	var count int
	query := `SELECT COUNT(1) FROM store_approvers WHERE store_id = ? AND role_id = ? AND user_id = ?`
	args := []interface{}{storeID, roleID, userID}
	if excludeID > 0 {
		query += ` AND id <> ?`
		args = append(args, excludeID)
	}

	err := r.DB.QueryRow(query, args...).Scan(&count)
	return count > 0, err
}

func (r *StoreApproverRepository) ExistsStore(storeID int) (bool, error) {
	var count int
	err := r.DB.QueryRow(`SELECT COUNT(1) FROM stores WHERE store_id = ?`, storeID).Scan(&count)
	return count > 0, err
}

func (r *StoreApproverRepository) ExistsRole(roleID int64) (bool, error) {
	var count int
	err := r.DB.QueryRow(`SELECT COUNT(1) FROM roles WHERE id = ?`, roleID).Scan(&count)
	return count > 0, err
}

func (r *StoreApproverRepository) ExistsUser(userID int) (bool, error) {
	var count int
	err := r.DB.QueryRow(`SELECT COUNT(1) FROM users WHERE id = ?`, userID).Scan(&count)
	return count > 0, err
}

func (r *StoreApproverRepository) Create(input models.StoreApproverCreateInput) error {
	isActive := 0
	if input.IsActive {
		isActive = 1
	}

	_, err := r.DB.Exec(`
		INSERT INTO store_approvers (store_id, role_id, user_id, is_active)
		VALUES (?, ?, ?, ?)
	`, input.StoreID, input.RoleID, input.UserID, isActive)

	return err
}

func (r *StoreApproverRepository) Update(input models.StoreApproverUpdateInput) error {
	isActive := 0
	if input.IsActive {
		isActive = 1
	}

	_, err := r.DB.Exec(`
		UPDATE store_approvers
		SET store_id = ?, role_id = ?, user_id = ?, is_active = ?
		WHERE id = ?
	`, input.StoreID, input.RoleID, input.UserID, isActive, input.ID)

	return err
}

func (r *StoreApproverRepository) DeleteByID(id int64) error {
	_, err := r.DB.Exec(`DELETE FROM store_approvers WHERE id = ?`, id)
	return err
}
