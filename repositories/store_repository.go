package repositories

import (
	"database/sql"
	"gobase-app/models"
	"strings"
)

type StoreRepository struct {
	DB *sql.DB
}

// GetAll mengambil seluruh data store.
func (r *StoreRepository) GetAll() ([]models.Store, error) {
	rows, err := r.DB.Query(`
		SELECT store_id, store_code, store_name, store_address, is_active, created_at, updated_at
		FROM stores
		ORDER BY store_id asc
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stores []models.Store
	for rows.Next() {
		var (
			s         models.Store
			isActive  int
			createdAt sql.NullTime
			updatedAt sql.NullTime
		)
		if err := rows.Scan(
			&s.StoreID,
			&s.StoreCode,
			&s.StoreName,
			&s.StoreAddress,
			&isActive,
			&createdAt,
			&updatedAt,
		); err != nil {
			return nil, err
		}
		s.IsActive = isActive == 1
		if s.IsActive {
			s.IsActiveLabel = "Aktif"
		} else {
			s.IsActiveLabel = "Non Aktif"
		}
		if createdAt.Valid {
			s.CreatedAt = createdAt.Time.Format("2006-01-02 15:04:05")
			s.CreatedAtDisplay = createdAt.Time.Format("02 Jan 2006 15:04:05")
		} else {
			s.CreatedAt = "-"
			s.CreatedAtDisplay = "-"
		}
		if updatedAt.Valid {
			s.UpdatedAt = updatedAt.Time.Format("2006-01-02 15:04:05")
			s.UpdatedAtDisplay = updatedAt.Time.Format("02 Jan 2006 15:04:05")
		} else {
			s.UpdatedAt = "-"
			s.UpdatedAtDisplay = "-"
		}
		stores = append(stores, s)
	}

	return stores, rows.Err()
}

// ExistsByID mengecek apakah store_id sudah tersedia.
func (r *StoreRepository) ExistsByID(id int) (bool, error) {
	var count int
	err := r.DB.QueryRow(`SELECT COUNT(1) FROM stores WHERE store_id = ?`, id).Scan(&count)
	return count > 0, err
}

// Create menyimpan store baru.
func (r *StoreRepository) Create(input models.StoreCreateInput) error {
	isActive := 0
	if input.IsActive {
		isActive = 1
	}

	_, err := r.DB.Exec(`
		INSERT INTO stores (store_id, store_code, store_name, store_address, is_active)
		VALUES (?, ?, ?, ?, ?)
	`, input.StoreID, input.StoreCode, input.StoreName, input.StoreAddress, isActive)

	return err
}

// Update memperbarui data store.
func (r *StoreRepository) Update(input models.StoreUpdateInput) error {
	isActive := 0
	if input.IsActive {
		isActive = 1
	}

	_, err := r.DB.Exec(`
		UPDATE stores
		SET store_code = ?, store_name = ?, store_address = ?, is_active = ?
		WHERE store_id = ?
	`, input.StoreCode, input.StoreName, input.StoreAddress, isActive, input.StoreID)

	return err
}

// DeleteByID menghapus store berdasarkan ID.
func (r *StoreRepository) DeleteByID(id int) error {
	_, err := r.DB.Exec(`DELETE FROM stores WHERE store_id = ?`, id)
	return err
}

// GetByIDs mengambil daftar store berdasarkan id yang diberikan.
func (r *StoreRepository) GetByIDs(ids []int) ([]models.Store, error) {
	if len(ids) == 0 {
		return []models.Store{}, nil
	}

	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}

	query := `
		SELECT store_id, store_name
		FROM stores
		WHERE store_id IN (` + strings.Join(placeholders, ",") + `)
		ORDER BY store_name
	`

	rows, err := r.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stores []models.Store
	for rows.Next() {
		var s models.Store
		if err := rows.Scan(&s.StoreID, &s.StoreName); err != nil {
			return nil, err
		}
		stores = append(stores, s)
	}

	return stores, nil
}
