package repositories

import (
	"database/sql"
	"gobase-app/models"
)

type GLAccountRepository struct {
	DB *sql.DB
}

func (r *GLAccountRepository) GetAll() ([]models.GLAccount, error) {
	rows, err := r.DB.Query(`
		SELECT id, gl_code, gl_name, spend_type, is_active, created_at, updated_at
		FROM gl_accounts
		ORDER BY id ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []models.GLAccount
	for rows.Next() {
		var (
			a         models.GLAccount
			isActive  int
			createdAt sql.NullTime
			updatedAt sql.NullTime
		)
		if err := rows.Scan(&a.ID, &a.GLCode, &a.GLName, &a.SpendType, &isActive, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		a.IsActive = isActive == 1
		if a.IsActive {
			a.IsActiveLabel = "Aktif"
		} else {
			a.IsActiveLabel = "Non Aktif"
		}
		if createdAt.Valid {
			a.CreatedAt = createdAt.Time.Format("2006-01-02 15:04:05")
			a.CreatedAtDisplay = createdAt.Time.Format("02 Jan 2006 15:04:05")
		} else {
			a.CreatedAt = "-"
			a.CreatedAtDisplay = "-"
		}
		if updatedAt.Valid {
			a.UpdatedAt = updatedAt.Time.Format("2006-01-02 15:04:05")
			a.UpdatedAtDisplay = updatedAt.Time.Format("02 Jan 2006 15:04:05")
		} else {
			a.UpdatedAt = "-"
			a.UpdatedAtDisplay = "-"
		}
		accounts = append(accounts, a)
	}

	return accounts, rows.Err()
}

func (r *GLAccountRepository) ExistsByID(id int) (bool, error) {
	var count int
	err := r.DB.QueryRow(`SELECT COUNT(1) FROM gl_accounts WHERE id = ?`, id).Scan(&count)
	return count > 0, err
}

func (r *GLAccountRepository) ExistsByCode(code string) (bool, error) {
	var count int
	err := r.DB.QueryRow(`SELECT COUNT(1) FROM gl_accounts WHERE gl_code = ?`, code).Scan(&count)
	return count > 0, err
}

func (r *GLAccountRepository) ExistsByCodeExceptID(code string, id int) (bool, error) {
	var count int
	err := r.DB.QueryRow(`SELECT COUNT(1) FROM gl_accounts WHERE gl_code = ? AND id <> ?`, code, id).Scan(&count)
	return count > 0, err
}

func (r *GLAccountRepository) Create(input models.GLAccountCreateInput) error {
	isActive := 0
	if input.IsActive {
		isActive = 1
	}

	_, err := r.DB.Exec(`
		INSERT INTO gl_accounts (gl_code, gl_name, spend_type, is_active)
		VALUES (?, ?, ?, ?)
	`, input.GLCode, input.GLName, input.SpendType, isActive)

	return err
}

func (r *GLAccountRepository) Update(input models.GLAccountUpdateInput) error {
	isActive := 0
	if input.IsActive {
		isActive = 1
	}

	_, err := r.DB.Exec(`
		UPDATE gl_accounts
		SET gl_code = ?, gl_name = ?, spend_type = ?, is_active = ?
		WHERE id = ?
	`, input.GLCode, input.GLName, input.SpendType, isActive, input.ID)

	return err
}

func (r *GLAccountRepository) DeleteByID(id int) error {
	_, err := r.DB.Exec(`DELETE FROM gl_accounts WHERE id = ?`, id)
	return err
}
