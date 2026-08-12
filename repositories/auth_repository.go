package repositories

import (
	"database/sql"
	"errors"
	"strings"

	"gobase-app/models"
)

var ErrAuthAccountNotFound = errors.New("auth account not found")

type AuthRepository struct {
	DB *sql.DB
}

func (r *AuthRepository) FindActiveByUsername(username string) (models.AuthAccount, error) {
	var account models.AuthAccount

	err := r.DB.QueryRow(`
		SELECT
			u.id,
			u.username,
			u.name,
			u.password,
			COALESCE(u.nip, '') AS nip,
			COALESCE(GROUP_CONCAT(DISTINCT r.name ORDER BY r.name SEPARATOR ', '), '') AS role,
			COALESCE(GROUP_CONCAT(DISTINCT us.store_id ORDER BY us.store_id SEPARATOR ','), '') AS store_id
		FROM users u
		LEFT JOIN model_has_roles mhr ON mhr.model_id = u.id AND mhr.model_type = ?
		LEFT JOIN roles r ON r.id = mhr.role_id
		LEFT JOIN user_stores us ON us.user_id = u.id
		WHERE u.username = ? AND u.status = 'active'
		GROUP BY u.id, u.username, u.name, u.password, u.nip
	`, userModelType, strings.TrimSpace(username)).Scan(
		&account.UserID,
		&account.Username,
		&account.Name,
		&account.PasswordHash,
		&account.NIP,
		&account.Role,
		&account.StoreID,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return models.AuthAccount{}, ErrAuthAccountNotFound
	}
	if err != nil {
		return models.AuthAccount{}, err
	}

	return account, nil
}
