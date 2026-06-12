package repositories

import (
	"database/sql"
	"gobase-app/models"
)

type VendorRepository struct {
	DB *sql.DB
}

func (r *VendorRepository) GetAll() ([]models.Vendor, error) {
	rows, err := r.DB.Query(`
		SELECT id, name, COALESCE(phone, ''), COALESCE(email, ''), COALESCE(address, ''), is_active, created_at
		FROM vendors
		ORDER BY id ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var vendors []models.Vendor
	for rows.Next() {
		var (
			v         models.Vendor
			isActive  int
			createdAt sql.NullTime
		)
		if err := rows.Scan(&v.ID, &v.Name, &v.Phone, &v.Email, &v.Address, &isActive, &createdAt); err != nil {
			return nil, err
		}
		v.IsActive = isActive == 1
		if v.IsActive {
			v.IsActiveLabel = "Aktif"
		} else {
			v.IsActiveLabel = "Non Aktif"
		}
		if createdAt.Valid {
			v.CreatedAt = createdAt.Time.Format("2006-01-02 15:04:05")
			v.CreatedAtDisplay = createdAt.Time.Format("02 Jan 2006 15:04:05")
		} else {
			v.CreatedAt = "-"
			v.CreatedAtDisplay = "-"
		}
		vendors = append(vendors, v)
	}

	return vendors, rows.Err()
}

func (r *VendorRepository) ExistsByName(name string) (bool, error) {
	var count int
	err := r.DB.QueryRow(`SELECT COUNT(1) FROM vendors WHERE name = ?`, name).Scan(&count)
	return count > 0, err
}

func (r *VendorRepository) ExistsByNameExceptID(name string, id int64) (bool, error) {
	var count int
	err := r.DB.QueryRow(`SELECT COUNT(1) FROM vendors WHERE name = ? AND id <> ?`, name, id).Scan(&count)
	return count > 0, err
}

func (r *VendorRepository) ExistsByID(id int64) (bool, error) {
	var count int
	err := r.DB.QueryRow(`SELECT COUNT(1) FROM vendors WHERE id = ?`, id).Scan(&count)
	return count > 0, err
}

func (r *VendorRepository) Create(input models.VendorCreateInput) error {
	isActive := 0
	if input.IsActive {
		isActive = 1
	}

	_, err := r.DB.Exec(`
		INSERT INTO vendors (name, phone, email, address, is_active)
		VALUES (?, ?, ?, ?, ?)
	`, input.Name, nullableString(input.Phone), nullableString(input.Email), nullableString(input.Address), isActive)

	return err
}

func (r *VendorRepository) Update(input models.VendorUpdateInput) error {
	isActive := 0
	if input.IsActive {
		isActive = 1
	}

	_, err := r.DB.Exec(`
		UPDATE vendors
		SET name = ?, phone = ?, email = ?, address = ?, is_active = ?
		WHERE id = ?
	`, input.Name, nullableString(input.Phone), nullableString(input.Email), nullableString(input.Address), isActive, input.ID)

	return err
}

func (r *VendorRepository) DeleteByID(id int64) error {
	_, err := r.DB.Exec(`DELETE FROM vendors WHERE id = ?`, id)
	return err
}

func nullableString(value string) interface{} {
	if value == "" {
		return nil
	}
	return value
}
