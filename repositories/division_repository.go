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

func (r *DivisionRepository) ExistsByCode(code string) (bool, error) {
	var count int
	err := r.DB.QueryRow(`SELECT COUNT(1) FROM divisions WHERE division_code = ?`, code).Scan(&count)
	return count > 0, err
}

func (r *DivisionRepository) ExistsByCodeExceptID(code string, id int) (bool, error) {
	var count int
	err := r.DB.QueryRow(`SELECT COUNT(1) FROM divisions WHERE division_code = ? AND id <> ?`, code, id).Scan(&count)
	return count > 0, err
}

func (r *DivisionRepository) Create(input models.DivisionCreateInput) error {
	_, err := r.DB.Exec(`
		INSERT INTO divisions (division_code, division_name)
		VALUES (?, ?)
	`, input.DivisionCode, input.DivisionName)
	return err
}

func (r *DivisionRepository) Update(input models.DivisionUpdateInput) error {
	_, err := r.DB.Exec(`
		UPDATE divisions
		SET division_code = ?, division_name = ?
		WHERE id = ?
	`, input.DivisionCode, input.DivisionName, input.ID)
	return err
}

func (r *DivisionRepository) DeleteByID(id int) error {
	_, err := r.DB.Exec(`DELETE FROM divisions WHERE id = ?`, id)
	return err
}
