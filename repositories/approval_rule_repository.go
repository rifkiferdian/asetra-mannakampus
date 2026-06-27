package repositories

import (
	"database/sql"
	"fmt"
	"gobase-app/models"
)

type ApprovalRuleRepository struct {
	DB *sql.DB
}

func (r *ApprovalRuleRepository) GetAll() ([]models.ApprovalRule, error) {
	rows, err := r.DB.Query(`
		SELECT
			ar.id,
			ar.name,
			ar.is_active,
			ar.min_amount,
			ar.max_amount,
			ar.location_scope,
			ar.spend_type,
			ar.urgent_level,
			ar.created_at,
			COUNT(ars.id) AS step_count
		FROM approval_rules ar
		LEFT JOIN approval_rule_steps ars ON ars.rule_id = ar.id
		GROUP BY ar.id, ar.name, ar.is_active, ar.min_amount, ar.max_amount, ar.location_scope, ar.spend_type, ar.urgent_level, ar.created_at
		ORDER BY ar.id ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []models.ApprovalRule
	for rows.Next() {
		var (
			item      models.ApprovalRule
			isActive  int
			maxAmount sql.NullFloat64
			createdAt sql.NullTime
		)
		if err := rows.Scan(
			&item.ID,
			&item.Name,
			&isActive,
			&item.MinAmount,
			&maxAmount,
			&item.LocationScope,
			&item.SpendType,
			&item.UrgentLevel,
			&createdAt,
			&item.StepCount,
		); err != nil {
			return nil, err
		}

		item.IsActive = isActive == 1
		if item.IsActive {
			item.IsActiveLabel = "Aktif"
		} else {
			item.IsActiveLabel = "Non Aktif"
		}
		if maxAmount.Valid {
			value := maxAmount.Float64
			item.MaxAmount = &value
			item.MaxAmountLabel = formatMoney(value)
		} else {
			item.MaxAmountLabel = "Tanpa batas"
		}
		if createdAt.Valid {
			item.CreatedAt = createdAt.Time.Format("2006-01-02 15:04:05")
			item.CreatedAtDisplay = createdAt.Time.Format("02 Jan 2006 15:04:05")
		} else {
			item.CreatedAt = "-"
			item.CreatedAtDisplay = "-"
		}
		rules = append(rules, item)
	}

	return rules, rows.Err()
}

func (r *ApprovalRuleRepository) GetByID(id int64) (*models.ApprovalRuleDetail, error) {
	var (
		detail    models.ApprovalRuleDetail
		isActive  int
		maxAmount sql.NullFloat64
	)

	err := r.DB.QueryRow(`
		SELECT id, name, is_active, min_amount, max_amount, location_scope, spend_type, urgent_level
		FROM approval_rules
		WHERE id = ?
	`, id).Scan(
		&detail.ID,
		&detail.Name,
		&isActive,
		&detail.MinAmount,
		&maxAmount,
		&detail.LocationScope,
		&detail.SpendType,
		&detail.UrgentLevel,
	)
	if err != nil {
		return nil, err
	}
	detail.IsActive = isActive == 1
	if maxAmount.Valid {
		value := maxAmount.Float64
		detail.MaxAmount = &value
	}

	rows, err := r.DB.Query(`
		SELECT ars.id, ars.rule_id, ars.step_order, ars.role_id, COALESCE(r.name, ''), ars.scope, ars.is_parallel, ars.is_required
		FROM approval_rule_steps ars
		LEFT JOIN roles r ON r.id = ars.role_id
		WHERE ars.rule_id = ?
		ORDER BY ars.step_order ASC, ars.id ASC
	`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			step       models.ApprovalRuleStep
			isParallel int
			isRequired int
		)
		if err := rows.Scan(&step.ID, &step.RuleID, &step.StepOrder, &step.RoleID, &step.RoleName, &step.Scope, &isParallel, &isRequired); err != nil {
			return nil, err
		}
		step.IsParallel = isParallel == 1
		step.IsRequired = isRequired == 1
		detail.Steps = append(detail.Steps, step)
	}

	return &detail, rows.Err()
}

func (r *ApprovalRuleRepository) ExistsByID(id int64) (bool, error) {
	var count int
	err := r.DB.QueryRow(`SELECT COUNT(1) FROM approval_rules WHERE id = ?`, id).Scan(&count)
	return count > 0, err
}

func (r *ApprovalRuleRepository) ExistsByName(name string) (bool, error) {
	var count int
	err := r.DB.QueryRow(`SELECT COUNT(1) FROM approval_rules WHERE name = ?`, name).Scan(&count)
	return count > 0, err
}

func (r *ApprovalRuleRepository) ExistsByNameExceptID(name string, id int64) (bool, error) {
	var count int
	err := r.DB.QueryRow(`SELECT COUNT(1) FROM approval_rules WHERE name = ? AND id <> ?`, name, id).Scan(&count)
	return count > 0, err
}

func (r *ApprovalRuleRepository) RoleExists(roleID int64) (bool, error) {
	var count int
	err := r.DB.QueryRow(`SELECT COUNT(1) FROM roles WHERE id = ?`, roleID).Scan(&count)
	return count > 0, err
}

func (r *ApprovalRuleRepository) CountApprovalsByRuleID(ruleID int64) (int, error) {
	var count int
	err := r.DB.QueryRow(`SELECT COUNT(1) FROM approvals WHERE rule_id = ?`, ruleID).Scan(&count)
	return count, err
}

func (r *ApprovalRuleRepository) Create(input models.ApprovalRuleCreateInput) (int64, error) {
	tx, err := r.DB.Begin()
	if err != nil {
		return 0, err
	}

	isActive := 0
	if input.IsActive {
		isActive = 1
	}

	res, err := tx.Exec(`
		INSERT INTO approval_rules (name, is_active, min_amount, max_amount, location_scope, spend_type, urgent_level)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, input.Name, isActive, input.MinAmount, input.MaxAmount, input.LocationScope, input.SpendType, input.UrgentLevel)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	ruleID, err := res.LastInsertId()
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	if err := insertApprovalRuleSteps(tx, ruleID, input.Steps); err != nil {
		tx.Rollback()
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return 0, err
	}

	return ruleID, nil
}

func (r *ApprovalRuleRepository) Update(input models.ApprovalRuleUpdateInput) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}

	isActive := 0
	if input.IsActive {
		isActive = 1
	}

	if _, err := tx.Exec(`
		UPDATE approval_rules
		SET name = ?, is_active = ?, min_amount = ?, max_amount = ?, location_scope = ?, spend_type = ?, urgent_level = ?
		WHERE id = ?
	`, input.Name, isActive, input.MinAmount, input.MaxAmount, input.LocationScope, input.SpendType, input.UrgentLevel, input.ID); err != nil {
		tx.Rollback()
		return err
	}

	if _, err := tx.Exec(`DELETE FROM approval_rule_steps WHERE rule_id = ?`, input.ID); err != nil {
		tx.Rollback()
		return err
	}

	if err := insertApprovalRuleSteps(tx, input.ID, input.Steps); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (r *ApprovalRuleRepository) DeleteByID(id int64) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}

	if _, err := tx.Exec(`DELETE FROM approval_rule_steps WHERE rule_id = ?`, id); err != nil {
		tx.Rollback()
		return err
	}
	if _, err := tx.Exec(`DELETE FROM approval_rules WHERE id = ?`, id); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func insertApprovalRuleSteps(tx *sql.Tx, ruleID int64, steps []models.ApprovalRuleStepInput) error {
	if len(steps) == 0 {
		return nil
	}

	stmt, err := tx.Prepare(`
		INSERT INTO approval_rule_steps (rule_id, step_order, role_id, scope, is_parallel, is_required)
		VALUES (?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, step := range steps {
		isParallel := 0
		if step.IsParallel {
			isParallel = 1
		}
		isRequired := 0
		if step.IsRequired {
			isRequired = 1
		}
		if _, err := stmt.Exec(ruleID, step.StepOrder, step.RoleID, step.Scope, isParallel, isRequired); err != nil {
			return err
		}
	}

	return nil
}

func formatMoney(value float64) string {
	return fmt.Sprintf("%.0f", value)
}
