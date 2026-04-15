package repositories

import (
	"database/sql"
	"gobase-app/models"
	"strconv"
	"strings"
	"time"
)

type UserRepository struct {
	DB *sql.DB
}

const userModelType = "Models\\User"

type UserCreateParams struct {
	NIP            int
	Username       string
	HashedPassword string
	Name           string
	Email          string
	Status         string
	StoreIDs       []int
}

type UserUpdateParams struct {
	ID             int
	NIP            int
	Username       string
	HashedPassword string
	Name           string
	Email          string
	Status         string
	StoreIDs       []int
}

// GetAll mengambil seluruh data user beserta outlet dari tabel user_stores.
func (r *UserRepository) GetAll() ([]models.User, error) {
	rows, err := r.DB.Query(`
		SELECT 
			u.id, 
			u.nip,
			u.username, 
			u.name, 
			u.email, 
			u.status, 
			COALESCE(GROUP_CONCAT(DISTINCT us.store_id ORDER BY us.store_id SEPARATOR ','), '') AS store_ids,
			COALESCE(GROUP_CONCAT(DISTINCT s.store_name ORDER BY s.store_name SEPARATOR ', '), '') AS store_display,
			u.created_at,
			COALESCE(GROUP_CONCAT(DISTINCT r2.name ORDER BY r2.name SEPARATOR ', '), '') AS role_display
		FROM users u
		LEFT JOIN user_stores us ON us.user_id = u.id
		LEFT JOIN stores s ON s.store_id = us.store_id
		LEFT JOIN model_has_roles mhr ON mhr.model_id = u.id AND mhr.model_type = ?
		LEFT JOIN roles r2 ON r2.id = mhr.role_id
		GROUP BY 
			u.id, u.nip, u.username, u.name, u.email, u.status, u.created_at
		ORDER BY u.created_at DESC
	`, userModelType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User

	for rows.Next() {
		var (
			u            models.User
			storeIDsCSV  string
			storeDisplay string
			createdAt    time.Time
		)

		if err := rows.Scan(
			&u.ID,
			&u.NIP,
			&u.Username,
			&u.Name,
			&u.Email,
			&u.Status,
			&storeIDsCSV,
			&storeDisplay,
			&createdAt,
			&u.RoleDisplay,
		); err != nil {
			return nil, err
		}

		u.CreatedAt = createdAt.Format("2006-01-02 15:04:05")
		u.CreatedAtDisplay = createdAt.Format("02 Jan 2006 15:04:05")

		if u.Status == "active" {
			u.StatusLabel = "Aktif"
		} else {
			u.StatusLabel = "Non Aktif"
		}

		u.StoreIDs = splitIntCSV(storeIDsCSV)
		u.StoreDisplay = storeDisplay

		if u.StoreDisplay == "" {
			u.StoreDisplay = "-"
		}

		if u.RoleDisplay == "" {
			u.RoleDisplay = "-"
		}
		u.RoleNames = splitAndTrimCSV(u.RoleDisplay)

		users = append(users, u)
	}

	return users, rows.Err()
}

// CreateUserWithRoles menyimpan data user baru beserta assignment rolenya dalam satu transaksi.
func (r *UserRepository) CreateUserWithRoles(params UserCreateParams, roleIDs []int64) (int64, error) {
	tx, err := r.DB.Begin()
	if err != nil {
		return 0, err
	}

	var emailVal interface{}
	if strings.TrimSpace(params.Email) == "" {
		emailVal = nil // simpan NULL jika email kosong
	} else {
		emailVal = params.Email
	}

	res, err := tx.Exec(`
		INSERT INTO users (nip, username, password, name, email, status)
		VALUES (?, ?, ?, ?, ?, ?)
	`, params.NIP, params.Username, params.HashedPassword, params.Name, emailVal, params.Status)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	userID, err := res.LastInsertId()
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	if err := insertUserStores(tx, userID, params.StoreIDs); err != nil {
		tx.Rollback()
		return 0, err
	}

	if len(roleIDs) > 0 {
		stmt, err := tx.Prepare(`
			INSERT INTO model_has_roles (role_id, model_type, model_id)
			VALUES (?, ?, ?)
		`)
		if err != nil {
			tx.Rollback()
			return 0, err
		}
		defer stmt.Close()

		for _, roleID := range roleIDs {
			if _, err := stmt.Exec(roleID, userModelType, userID); err != nil {
				tx.Rollback()
				return 0, err
			}
		}
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return 0, err
	}

	return userID, nil
}

// UpdateUserWithRoles memperbarui data user beserta role assignments dalam satu transaksi.
func (r *UserRepository) UpdateUserWithRoles(params UserUpdateParams, roleIDs []int64) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}

	var emailVal interface{}
	if strings.TrimSpace(params.Email) == "" {
		emailVal = nil
	} else {
		emailVal = params.Email
	}

	if params.HashedPassword != "" {
		if _, err := tx.Exec(`
			UPDATE users
			SET nip = ?, username = ?, password = ?, name = ?, email = ?, status = ?
			WHERE id = ?
		`, params.NIP, params.Username, params.HashedPassword, params.Name, emailVal, params.Status, params.ID); err != nil {
			tx.Rollback()
			return err
		}
	} else {
		if _, err := tx.Exec(`
			UPDATE users
			SET nip = ?, username = ?, name = ?, email = ?, status = ?
			WHERE id = ?
		`, params.NIP, params.Username, params.Name, emailVal, params.Status, params.ID); err != nil {
			tx.Rollback()
			return err
		}
	}

	if _, err := tx.Exec(`DELETE FROM user_stores WHERE user_id = ?`, params.ID); err != nil {
		tx.Rollback()
		return err
	}

	if err := insertUserStores(tx, int64(params.ID), params.StoreIDs); err != nil {
		tx.Rollback()
		return err
	}

	if _, err := tx.Exec(`DELETE FROM model_has_roles WHERE model_id = ? AND model_type = ?`, params.ID, userModelType); err != nil {
		tx.Rollback()
		return err
	}

	if len(roleIDs) > 0 {
		stmt, err := tx.Prepare(`
			INSERT INTO model_has_roles (role_id, model_type, model_id)
			VALUES (?, ?, ?)
		`)
		if err != nil {
			tx.Rollback()
			return err
		}
		defer stmt.Close()

		for _, roleID := range roleIDs {
			if _, err := stmt.Exec(roleID, userModelType, params.ID); err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	return tx.Commit()
}

// ExistsByUsername mengecek apakah username sudah digunakan.
func (r *UserRepository) ExistsByUsername(username string) (bool, error) {
	var count int
	err := r.DB.QueryRow(`SELECT COUNT(1) FROM users WHERE username = ?`, username).Scan(&count)
	return count > 0, err
}

// ExistsByUsernameExceptID mengecek apakah username sudah digunakan oleh user lain.
func (r *UserRepository) ExistsByUsernameExceptID(username string, id int) (bool, error) {
	var count int
	err := r.DB.QueryRow(`SELECT COUNT(1) FROM users WHERE username = ? AND id <> ?`, username, id).Scan(&count)
	return count > 0, err
}

// ExistsByNIP mengecek apakah NIP sudah digunakan.
func (r *UserRepository) ExistsByNIP(nip int) (bool, error) {
	var count int
	err := r.DB.QueryRow(`SELECT COUNT(1) FROM users WHERE nip = ?`, nip).Scan(&count)
	return count > 0, err
}

// ExistsByNIPExceptID mengecek apakah NIP sudah digunakan user lain.
func (r *UserRepository) ExistsByNIPExceptID(nip int, id int) (bool, error) {
	var count int
	err := r.DB.QueryRow(`SELECT COUNT(1) FROM users WHERE nip = ? AND id <> ?`, nip, id).Scan(&count)
	return count > 0, err
}

// ExistsByEmail mengecek apakah email sudah digunakan (abaikan jika kosong).
func (r *UserRepository) ExistsByEmail(email string) (bool, error) {
	if strings.TrimSpace(email) == "" {
		return false, nil
	}

	var count int
	err := r.DB.QueryRow(`SELECT COUNT(1) FROM users WHERE email = ?`, email).Scan(&count)
	return count > 0, err
}

// ExistsByEmailExceptID mengecek apakah email sudah digunakan user lain (abaikan jika kosong).
func (r *UserRepository) ExistsByEmailExceptID(email string, id int) (bool, error) {
	if strings.TrimSpace(email) == "" {
		return false, nil
	}

	var count int
	err := r.DB.QueryRow(`SELECT COUNT(1) FROM users WHERE email = ? AND id <> ?`, email, id).Scan(&count)
	return count > 0, err
}

// GetRoleIDsByNames mengambil role_id berdasarkan nama role yang diberikan.
func (r *UserRepository) GetRoleIDsByNames(names []string) (map[string]int64, error) {
	result := make(map[string]int64)

	if len(names) == 0 {
		return result, nil
	}

	placeholders := make([]string, len(names))
	args := make([]interface{}, len(names))

	for i, name := range names {
		placeholders[i] = "?"
		args[i] = name
	}

	query := `
		SELECT id, name
		FROM roles
		WHERE name IN (` + strings.Join(placeholders, ",") + `)
	`

	rows, err := r.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			id   int64
			name string
		)
		if err := rows.Scan(&id, &name); err != nil {
			return nil, err
		}
		result[name] = id
	}

	return result, rows.Err()
}

func splitAndTrimCSV(val string) []string {
	val = strings.TrimSpace(val)
	if val == "" || val == "-" {
		return nil
	}

	parts := strings.Split(val, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

func splitIntCSV(val string) []int {
	val = strings.TrimSpace(val)
	if val == "" {
		return nil
	}

	parts := strings.Split(val, ",")
	result := make([]int, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		id, err := strconv.Atoi(part)
		if err != nil {
			continue
		}
		result = append(result, id)
	}
	return result
}

func insertUserStores(tx *sql.Tx, userID int64, storeIDs []int) error {
	if len(storeIDs) == 0 {
		return nil
	}

	stmt, err := tx.Prepare(`
		INSERT INTO user_stores (user_id, store_id)
		VALUES (?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, storeID := range storeIDs {
		if _, err := stmt.Exec(userID, storeID); err != nil {
			return err
		}
	}

	return nil
}

// DeleteUser removes a user and related role/permission mappings in a single transaction.
func (r *UserRepository) DeleteUser(id int) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}

	if _, err := tx.Exec(`DELETE FROM model_has_roles WHERE model_id = ? AND model_type = ?`, id, userModelType); err != nil {
		tx.Rollback()
		return err
	}

	if _, err := tx.Exec(`DELETE FROM model_has_permissions WHERE model_id = ? AND model_type = ?`, id, userModelType); err != nil {
		tx.Rollback()
		return err
	}

	if _, err := tx.Exec(`DELETE FROM user_stores WHERE user_id = ?`, id); err != nil {
		tx.Rollback()
		return err
	}

	if _, err := tx.Exec(`DELETE FROM users WHERE id = ?`, id); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}
