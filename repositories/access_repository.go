package repositories

import (
	"database/sql"
	"errors"
	"strings"

	"gobase-app/models"
)

var ErrAccessUserInactive = errors.New("user session is no longer active")

type AccessRepository struct {
	DB *sql.DB
}

// GetByUserID reloads role and store assignments on every request so revoked
// access does not remain valid until the session cookie expires.
func (r *AccessRepository) GetByUserID(userID int) (models.AccessScope, error) {
	scope := models.AccessScope{UserID: userID}
	if userID <= 0 {
		return scope, ErrAccessUserInactive
	}

	var active int
	if err := r.DB.QueryRow(`SELECT COUNT(1) FROM users WHERE id = ? AND status = 'active'`, userID).Scan(&active); err != nil {
		return scope, err
	}
	if active == 0 {
		return scope, ErrAccessUserInactive
	}

	roleRows, err := r.DB.Query(`
		SELECT DISTINCT LOWER(role.name)
		FROM model_has_roles mapping
		JOIN roles role ON role.id = mapping.role_id
		WHERE mapping.model_id = ? AND mapping.model_type = ?
	`, userID, userModelType)
	if err != nil {
		return scope, err
	}
	defer roleRows.Close()
	for roleRows.Next() {
		var role string
		if err := roleRows.Scan(&role); err != nil {
			return scope, err
		}
		switch strings.TrimSpace(role) {
		case "super-admin", "admin", "procurement", "finance-manager", "gm":
			scope.CanViewAll = true
		}
	}
	if err := roleRows.Err(); err != nil {
		return scope, err
	}

	storeRows, err := r.DB.Query(`SELECT DISTINCT store_id FROM user_stores WHERE user_id = ? ORDER BY store_id`, userID)
	if err != nil {
		return scope, err
	}
	defer storeRows.Close()
	for storeRows.Next() {
		var storeID int
		if err := storeRows.Scan(&storeID); err != nil {
			return scope, err
		}
		if storeID > 0 {
			scope.StoreIDs = append(scope.StoreIDs, storeID)
		}
	}
	return scope, storeRows.Err()
}
