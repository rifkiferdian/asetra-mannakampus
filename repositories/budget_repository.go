package repositories

import (
	"database/sql"
	"fmt"
	"gobase-app/models"
	"math"
)

type BudgetRepository struct {
	DB *sql.DB
}

func (r *BudgetRepository) GetAll() ([]models.Budget, error) {
	rows, err := r.DB.Query(`
		SELECT
			b.id, b.fiscal_year, b.period_type, b.period_key,
			COALESCE(b.store_id, 0), COALESCE(s.store_name, ''),
			COALESCE(b.division_id, 0), COALESCE(d.division_name, ''),
			b.gl_account_id, CONCAT(ga.gl_code, ' - ', ga.gl_name),
			b.amount, COALESCE(SUM(bu.used_amount), 0),
			b.created_at, b.updated_at
		FROM budgets b
		LEFT JOIN stores s ON s.store_id = b.store_id
		LEFT JOIN divisions d ON d.id = b.division_id
		JOIN gl_accounts ga ON ga.id = b.gl_account_id
		LEFT JOIN budget_usages bu ON bu.budget_id = b.id
		GROUP BY b.id, b.fiscal_year, b.period_type, b.period_key, b.store_id, s.store_name, b.division_id, d.division_name, b.gl_account_id, ga.gl_code, ga.gl_name, b.amount, b.created_at, b.updated_at
		ORDER BY b.fiscal_year DESC, b.period_key DESC, b.id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var budgets []models.Budget
	for rows.Next() {
		var (
			item      models.Budget
			createdAt sql.NullTime
			updatedAt sql.NullTime
		)
		if err := rows.Scan(
			&item.ID,
			&item.FiscalYear,
			&item.PeriodType,
			&item.PeriodKey,
			&item.StoreID,
			&item.StoreName,
			&item.DivisionID,
			&item.DivisionName,
			&item.GLAccountID,
			&item.GLAccountName,
			&item.Amount,
			&item.UsedAmount,
			&createdAt,
			&updatedAt,
		); err != nil {
			return nil, err
		}
		item.RemainingAmount = item.Amount - item.UsedAmount
		item.AmountDisplay = formatIDR(item.Amount)
		item.UsedAmountDisplay = formatIDR(item.UsedAmount)
		item.RemainingDisplay = formatIDR(item.RemainingAmount)
		if createdAt.Valid {
			item.CreatedAt = createdAt.Time.Format("2006-01-02 15:04:05")
			item.CreatedAtDisplay = createdAt.Time.Format("02 Jan 2006 15:04:05")
		} else {
			item.CreatedAt = "-"
			item.CreatedAtDisplay = "-"
		}
		if updatedAt.Valid {
			item.UpdatedAt = updatedAt.Time.Format("2006-01-02 15:04:05")
			item.UpdatedAtDisplay = updatedAt.Time.Format("02 Jan 2006 15:04:05")
		} else {
			item.UpdatedAt = "-"
			item.UpdatedAtDisplay = "-"
		}
		budgets = append(budgets, item)
	}

	return budgets, rows.Err()
}

func (r *BudgetRepository) ExistsByID(id int64) (bool, error) {
	var count int
	err := r.DB.QueryRow(`SELECT COUNT(1) FROM budgets WHERE id = ?`, id).Scan(&count)
	return count > 0, err
}

func (r *BudgetRepository) ExistsDuplicate(input models.BudgetCreateInput) (bool, error) {
	var count int
	err := r.DB.QueryRow(`
		SELECT COUNT(1)
		FROM budgets
		WHERE fiscal_year = ?
			AND period_type = ?
			AND period_key = ?
			AND ((store_id IS NULL AND ? = 0) OR store_id = ?)
			AND ((division_id IS NULL AND ? = 0) OR division_id = ?)
			AND gl_account_id = ?
	`, input.FiscalYear, input.PeriodType, input.PeriodKey, input.StoreID, input.StoreID, input.DivisionID, input.DivisionID, input.GLAccountID).Scan(&count)
	return count > 0, err
}

func (r *BudgetRepository) ExistsDuplicateExceptID(input models.BudgetUpdateInput) (bool, error) {
	var count int
	err := r.DB.QueryRow(`
		SELECT COUNT(1)
		FROM budgets
		WHERE id <> ?
			AND fiscal_year = ?
			AND period_type = ?
			AND period_key = ?
			AND ((store_id IS NULL AND ? = 0) OR store_id = ?)
			AND ((division_id IS NULL AND ? = 0) OR division_id = ?)
			AND gl_account_id = ?
	`, input.ID, input.FiscalYear, input.PeriodType, input.PeriodKey, input.StoreID, input.StoreID, input.DivisionID, input.DivisionID, input.GLAccountID).Scan(&count)
	return count > 0, err
}

func (r *BudgetRepository) Create(input models.BudgetCreateInput) error {
	_, err := r.DB.Exec(`
		INSERT INTO budgets (fiscal_year, period_type, period_key, store_id, division_id, gl_account_id, amount)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, input.FiscalYear, input.PeriodType, input.PeriodKey, nullableInt(input.StoreID), nullableInt(input.DivisionID), input.GLAccountID, input.Amount)
	return err
}

func (r *BudgetRepository) Update(input models.BudgetUpdateInput) error {
	_, err := r.DB.Exec(`
		UPDATE budgets
		SET fiscal_year = ?, period_type = ?, period_key = ?, store_id = ?, division_id = ?, gl_account_id = ?, amount = ?
		WHERE id = ?
	`, input.FiscalYear, input.PeriodType, input.PeriodKey, nullableInt(input.StoreID), nullableInt(input.DivisionID), input.GLAccountID, input.Amount, input.ID)
	return err
}

func (r *BudgetRepository) DeleteByID(id int64) error {
	_, err := r.DB.Exec(`DELETE FROM budgets WHERE id = ?`, id)
	return err
}

func formatIDR(value float64) string {
	sign := ""
	if value < 0 {
		sign = "-"
		value = math.Abs(value)
	}
	n := int64(math.Round(value))
	raw := fmt.Sprintf("%d", n)
	out := ""
	for i, r := range raw {
		if i > 0 && (len(raw)-i)%3 == 0 {
			out += "."
		}
		out += string(r)
	}
	return sign + "Rp " + out
}
