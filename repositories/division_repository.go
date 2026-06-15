package repositories

import (
	"database/sql"
	"gobase-app/models"
)

type DivisionRepository struct {
	DB *sql.DB
}

func (r *DivisionRepository) GetAll() ([]models.Division, error) {
	rows, err := r.DB.Query(`
		SELECT id, division_code, division_name
		FROM divisions
		ORDER BY division_name ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var divisions []models.Division
	for rows.Next() {
		var item models.Division
		if err := rows.Scan(&item.ID, &item.DivisionCode, &item.DivisionName); err != nil {
			return nil, err
		}
		divisions = append(divisions, item)
	}

	return divisions, rows.Err()
}

func (r *DivisionRepository) ExistsByID(id int) (bool, error) {
	var count int
	err := r.DB.QueryRow(`SELECT COUNT(1) FROM divisions WHERE id = ?`, id).Scan(&count)
	return count > 0, err
}
