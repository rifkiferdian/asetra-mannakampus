package repositories

import (
	"database/sql"
	"errors"
	"fmt"
	"gobase-app/models"
	"math"
	"strings"
	"time"
)

type AssetDisposalApprovalRepository struct {
	DB *sql.DB
}

func (r *AssetDisposalApprovalRepository) GetRules() ([]models.AssetDisposalApprovalRule, error) {
	rows, err := r.DB.Query(`
		SELECT rule.id, rule.name, COALESCE(rule.disposal_type_id, 0), COALESCE(dtype.name, 'Semua jenis'),
			COALESCE(rule.asset_type_id, 0), COALESCE(atype.name, 'Semua tipe aset'),
			rule.min_book_value, rule.max_book_value, rule.priority, rule.is_active,
			rule.effective_from, rule.effective_until, COUNT(DISTINCT approval.id)
		FROM asset_disposal_approval_rules rule
		LEFT JOIN asset_disposal_types dtype ON dtype.id = rule.disposal_type_id
		LEFT JOIN asset_types atype ON atype.id = rule.asset_type_id
		LEFT JOIN asset_disposal_approvals approval ON approval.rule_id = rule.id
		GROUP BY rule.id, rule.name, rule.disposal_type_id, dtype.name, rule.asset_type_id, atype.name,
			rule.min_book_value, rule.max_book_value, rule.priority, rule.is_active,
			rule.effective_from, rule.effective_until
		ORDER BY rule.priority, rule.id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]models.AssetDisposalApprovalRule, 0)
	for rows.Next() {
		var item models.AssetDisposalApprovalRule
		var max sql.NullFloat64
		var active int
		var from, until sql.NullTime
		if err := rows.Scan(&item.ID, &item.Name, &item.DisposalTypeID, &item.DisposalTypeName,
			&item.AssetTypeID, &item.AssetTypeName, &item.MinBookValue, &max, &item.Priority,
			&active, &from, &until, &item.ApprovalCount); err != nil {
			return nil, err
		}
		item.IsActive = active == 1
		item.MinBookValueInput = formatNumberInput(item.MinBookValue)
		item.MinBookValueDisplay = formatAssetAmountID(item.MinBookValue)
		item.MaxBookValueDisplay = "Tanpa batas"
		if max.Valid {
			value := max.Float64
			item.MaxBookValue = &value
			item.MaxBookValueInput = formatNumberInput(value)
			item.MaxBookValueDisplay = formatAssetAmountID(value)
		}
		if from.Valid {
			item.EffectiveFrom = from.Time.Format("2006-01-02")
		}
		if until.Valid {
			item.EffectiveUntil = until.Time.Format("2006-01-02")
		}
		item.EffectivePeriodLabel = approvalEffectiveLabel(from, until)
		steps, err := r.getRuleSteps(item.ID)
		if err != nil {
			return nil, err
		}
		item.Steps = steps
		parts := make([]string, 0, len(steps))
		for _, step := range steps {
			parts = append(parts, fmt.Sprintf("%d. %s (%s)", step.StepOrder, step.RoleName, step.Scope))
		}
		item.StepSummary = strings.Join(parts, " -> ")
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *AssetDisposalApprovalRepository) getRuleSteps(ruleID int64) ([]models.AssetDisposalApprovalRuleStep, error) {
	rows, err := r.DB.Query(`
		SELECT step.id, step.step_order, step.role_id, role.name, step.scope, step.is_parallel, step.is_required
		FROM asset_disposal_approval_rule_steps step
		JOIN roles role ON role.id = step.role_id
		WHERE step.rule_id = ? ORDER BY step.step_order, step.id
	`, ruleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]models.AssetDisposalApprovalRuleStep, 0)
	for rows.Next() {
		var item models.AssetDisposalApprovalRuleStep
		var parallel, required int
		if err := rows.Scan(&item.ID, &item.StepOrder, &item.RoleID, &item.RoleName, &item.Scope, &parallel, &required); err != nil {
			return nil, err
		}
		item.IsParallel, item.IsRequired = parallel == 1, required == 1
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *AssetDisposalApprovalRepository) SaveRule(input models.AssetDisposalApprovalRuleInput) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if input.ID > 0 {
		var used int
		if err := tx.QueryRow(`SELECT COUNT(*) FROM asset_disposal_approvals WHERE rule_id = ?`, input.ID).Scan(&used); err != nil {
			return err
		}
		if used > 0 {
			return errors.New("aturan sudah digunakan; nonaktifkan lalu buat aturan baru untuk mengubah tahap approval")
		}
	}
	var ruleID int64
	if input.ID == 0 {
		res, err := tx.Exec(`
			INSERT INTO asset_disposal_approval_rules
			(name, disposal_type_id, asset_type_id, min_book_value, max_book_value, priority, is_active, effective_from, effective_until)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, input.Name, nullablePositiveInt64(input.DisposalTypeID), nullablePositiveInt64(input.AssetTypeID),
			input.MinBookValue, input.MaxBookValue, input.Priority, boolToInt(input.IsActive),
			nullableDateString(input.EffectiveFrom), nullableDateString(input.EffectiveUntil))
		if err != nil {
			return err
		}
		ruleID, err = res.LastInsertId()
		if err != nil {
			return err
		}
	} else {
		ruleID = input.ID
		res, err := tx.Exec(`
			UPDATE asset_disposal_approval_rules SET name=?, disposal_type_id=?, asset_type_id=?, min_book_value=?,
				max_book_value=?, priority=?, is_active=?, effective_from=?, effective_until=?, updated_at=CURRENT_TIMESTAMP
			WHERE id=?
		`, input.Name, nullablePositiveInt64(input.DisposalTypeID), nullablePositiveInt64(input.AssetTypeID),
			input.MinBookValue, input.MaxBookValue, input.Priority, boolToInt(input.IsActive),
			nullableDateString(input.EffectiveFrom), nullableDateString(input.EffectiveUntil), ruleID)
		if err != nil {
			return err
		}
		if affected, _ := res.RowsAffected(); affected == 0 {
			return errors.New("aturan approval disposal tidak ditemukan")
		}
		if _, err := tx.Exec(`DELETE FROM asset_disposal_approval_rule_steps WHERE rule_id=?`, ruleID); err != nil {
			return err
		}
	}
	for _, step := range input.Steps {
		if _, err := tx.Exec(`
			INSERT INTO asset_disposal_approval_rule_steps
			(rule_id, step_order, role_id, scope, is_parallel, is_required) VALUES (?, ?, ?, ?, ?, ?)
		`, ruleID, step.StepOrder, step.RoleID, step.Scope, boolToInt(step.IsParallel), boolToInt(step.IsRequired)); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *AssetDisposalApprovalRepository) DeleteRule(id int64) error {
	var used int
	if err := r.DB.QueryRow(`SELECT COUNT(*) FROM asset_disposal_approvals WHERE rule_id=?`, id).Scan(&used); err != nil {
		return err
	}
	if used > 0 {
		return errors.New("aturan sudah digunakan dan tidak dapat dihapus; nonaktifkan aturan tersebut")
	}
	res, err := r.DB.Exec(`DELETE FROM asset_disposal_approval_rules WHERE id=?`, id)
	if err != nil {
		return err
	}
	if affected, _ := res.RowsAffected(); affected == 0 {
		return errors.New("aturan approval disposal tidak ditemukan")
	}
	return nil
}

func (r *AssetDisposalApprovalRepository) GetApprovers() ([]models.AssetDisposalApprover, error) {
	rows, err := r.DB.Query(`
		SELECT mapping.id, mapping.scope, COALESCE(mapping.store_id,0), COALESCE(store.store_name,'Head Office'),
			mapping.role_id, role.name, mapping.user_id, user.name, mapping.is_active, mapping.updated_at
		FROM asset_disposal_approvers mapping
		LEFT JOIN stores store ON store.store_id=mapping.store_id
		JOIN roles role ON role.id=mapping.role_id
		JOIN users user ON user.id=mapping.user_id
		ORDER BY mapping.scope, store.store_name, role.name, user.name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]models.AssetDisposalApprover, 0)
	for rows.Next() {
		var item models.AssetDisposalApprover
		var active int
		var updated time.Time
		if err := rows.Scan(&item.ID, &item.Scope, &item.StoreID, &item.StoreName, &item.RoleID, &item.RoleName,
			&item.UserID, &item.UserName, &active, &updated); err != nil {
			return nil, err
		}
		item.IsActive = active == 1
		item.UpdatedAt = formatDepreciationDateID(updated, true)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *AssetDisposalApprovalRepository) SaveApprover(input models.AssetDisposalApproverInput) error {
	var roleCount int
	if err := r.DB.QueryRow(`SELECT COUNT(*) FROM model_has_roles WHERE model_id=? AND role_id=? AND model_type='Models\\User'`, input.UserID, input.RoleID).Scan(&roleCount); err != nil {
		return err
	}
	if roleCount == 0 {
		return errors.New("user yang dipilih tidak memiliki role tersebut")
	}
	store := nullableDisposalApproverStoreID(input.StoreID)
	if input.ID == 0 {
		_, err := r.DB.Exec(`INSERT INTO asset_disposal_approvers (scope,store_id,role_id,user_id,is_active) VALUES (?,?,?,?,?)`,
			input.Scope, store, input.RoleID, input.UserID, boolToInt(input.IsActive))
		return err
	}
	res, err := r.DB.Exec(`UPDATE asset_disposal_approvers SET scope=?,store_id=?,role_id=?,user_id=?,is_active=? WHERE id=?`,
		input.Scope, store, input.RoleID, input.UserID, boolToInt(input.IsActive), input.ID)
	if err != nil {
		return err
	}
	if affected, _ := res.RowsAffected(); affected == 0 {
		return errors.New("pemetaan approver tidak ditemukan")
	}
	return nil
}

func (r *AssetDisposalApprovalRepository) DeleteApprover(id int64) error {
	var used int
	if err := r.DB.QueryRow(`SELECT COUNT(*) FROM asset_disposal_approval_tasks task JOIN asset_disposal_approvers mapping ON mapping.user_id=task.assigned_user_id AND mapping.role_id=task.role_id WHERE mapping.id=?`, id).Scan(&used); err != nil {
		return err
	}
	if used > 0 {
		return errors.New("pemetaan sudah digunakan; nonaktifkan pemetaan tersebut")
	}
	res, err := r.DB.Exec(`DELETE FROM asset_disposal_approvers WHERE id=?`, id)
	if err != nil {
		return err
	}
	if affected, _ := res.RowsAffected(); affected == 0 {
		return errors.New("pemetaan approver tidak ditemukan")
	}
	return nil
}

func (r *AssetDisposalApprovalRepository) Submit(disposalID int64, auditCtx models.AuditContext) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var number, status string
	var assetID, disposalTypeID, assetTypeID int64
	var storeID sql.NullInt64
	var processedBy int
	if err := tx.QueryRow(`
		SELECT disposal.disposal_number, disposal.status, disposal.asset_id, disposal.disposal_type_id,
			COALESCE(asset.asset_type_id,0), asset.store_id, disposal.processed_by
		FROM asset_disposals disposal JOIN assets asset ON asset.id=disposal.asset_id
		WHERE disposal.id=? FOR UPDATE
	`, disposalID).Scan(&number, &status, &assetID, &disposalTypeID, &assetTypeID, &storeID, &processedBy); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("transaksi disposal tidak ditemukan")
		}
		return err
	}
	if status != "DRAFT" && status != "REJECTED" {
		return errors.New("hanya disposal DRAFT atau REJECTED yang dapat diajukan")
	}
	bookValue, err := disposalBookValueTx(tx, assetID)
	if err != nil {
		return err
	}
	var ruleID int64
	err = tx.QueryRow(`
		SELECT id FROM asset_disposal_approval_rules
		WHERE is_active=1
		  AND (disposal_type_id IS NULL OR disposal_type_id=?)
		  AND (asset_type_id IS NULL OR asset_type_id=?)
		  AND min_book_value<=? AND (max_book_value IS NULL OR max_book_value>=?)
		  AND (effective_from IS NULL OR effective_from<=CURDATE())
		  AND (effective_until IS NULL OR effective_until>=CURDATE())
		ORDER BY (disposal_type_id IS NOT NULL)+(asset_type_id IS NOT NULL) DESC, priority ASC, id ASC LIMIT 1
		FOR UPDATE
	`, disposalTypeID, assetTypeID, bookValue, bookValue).Scan(&ruleID)
	if errors.Is(err, sql.ErrNoRows) {
		return errors.New("tidak ada aturan approval aktif yang sesuai dengan jenis, tipe, dan nilai buku aset")
	}
	if err != nil {
		return err
	}
	steps, err := getApprovalRuleStepsTx(tx, ruleID)
	if err != nil {
		return err
	}
	if len(steps) == 0 {
		return errors.New("aturan approval belum memiliki tahap")
	}
	type resolvedTask struct {
		step   models.AssetDisposalApprovalRuleStep
		userID int
		scope  string
	}
	resolved := make([]resolvedTask, 0, len(steps))
	for _, step := range steps {
		userID, resolvedScope, err := resolveDisposalApproverTx(tx, step, storeID, processedBy)
		if err != nil {
			return err
		}
		resolved = append(resolved, resolvedTask{step: step, userID: userID, scope: resolvedScope})
	}
	var attempt int
	if err := tx.QueryRow(`SELECT COALESCE(MAX(attempt_no),0)+1 FROM asset_disposal_approvals WHERE disposal_id=?`, disposalID).Scan(&attempt); err != nil {
		return err
	}
	firstStep := steps[0].StepOrder
	res, err := tx.Exec(`INSERT INTO asset_disposal_approvals (disposal_id,rule_id,attempt_no,current_step,status,submitted_by) VALUES (?,?,?,?, 'PENDING',?)`,
		disposalID, ruleID, attempt, firstStep, auditCtx.ActorUserID)
	if err != nil {
		return err
	}
	approvalID, err := res.LastInsertId()
	if err != nil {
		return err
	}
	for _, task := range resolved {
		taskStatus := "PENDING"
		if task.step.StepOrder == firstStep {
			taskStatus = "WAITING"
		}
		if _, err := tx.Exec(`INSERT INTO asset_disposal_approval_tasks
			(approval_id,rule_step_id,step_order,role_id,scope,assigned_user_id,status) VALUES (?,?,?,?,?,?,?)`,
			approvalID, task.step.ID, task.step.StepOrder, task.step.RoleID, task.scope, task.userID, taskStatus); err != nil {
			return err
		}
	}
	if _, err := tx.Exec(`UPDATE asset_disposals SET status='IN_APPROVAL', submitted_by=?, submitted_at=CURRENT_TIMESTAMP,
		rejected_by=NULL,rejected_at=NULL,rejection_reason=NULL,approved_by=NULL,approved_at=NULL,updated_at=CURRENT_TIMESTAMP WHERE id=?`,
		auditCtx.ActorUserID, disposalID); err != nil {
		return err
	}
	if err := insertDisposalApprovalHistoryTx(tx, approvalID, nil, disposalID, "SUBMIT", status, "IN_APPROVAL", "Disposal diajukan untuk approval", auditCtx); err != nil {
		return err
	}
	if err := insertAuditLogTx(tx, "ASSET_DISPOSAL", disposalID, "SUBMIT_APPROVAL", fmt.Sprintf("Disposal %s diajukan ke approval percobaan %d", number, attempt), auditCtx); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *AssetDisposalApprovalRepository) GetInbox(userID int, filter models.AssetDisposalApprovalInboxFilter) (models.AssetDisposalApprovalInboxResult, error) {
	result := models.AssetDisposalApprovalInboxResult{}
	where := "task.assigned_user_id=?"
	args := []any{userID}
	if filter.Status != "ALL" {
		where += " AND task.status=?"
		args = append(args, filter.Status)
	}
	if filter.Search != "" {
		where += " AND (disposal.disposal_number LIKE ? OR asset.asset_code LIKE ? OR asset.asset_name LIKE ?)"
		term := "%" + filter.Search + "%"
		args = append(args, term, term, term)
	}
	if err := r.DB.QueryRow(`SELECT COALESCE(SUM(status='WAITING'),0),COALESCE(SUM(status='PENDING'),0),COALESCE(SUM(status='APPROVED'),0),COALESCE(SUM(status='REJECTED'),0) FROM asset_disposal_approval_tasks WHERE assigned_user_id=?`, userID).Scan(&result.Stats.Waiting, &result.Stats.Pending, &result.Stats.Approved, &result.Stats.Rejected); err != nil {
		return result, err
	}
	countQuery := `SELECT COUNT(*) FROM asset_disposal_approval_tasks task JOIN asset_disposal_approvals approval ON approval.id=task.approval_id JOIN asset_disposals disposal ON disposal.id=approval.disposal_id JOIN assets asset ON asset.id=disposal.asset_id WHERE ` + where
	if err := r.DB.QueryRow(countQuery, args...).Scan(&result.TotalRows); err != nil {
		return result, err
	}
	result.TotalPages = 1
	if result.TotalRows > 0 {
		result.TotalPages = (result.TotalRows + filter.PerPage - 1) / filter.PerPage
	}
	if filter.Page > result.TotalPages {
		filter.Page = result.TotalPages
	}
	queryArgs := append(append([]any{}, args...), filter.PerPage, (filter.Page-1)*filter.PerPage)
	rows, err := r.DB.Query(`
		SELECT task.id,approval.id,disposal.id,disposal.disposal_number,asset.id,asset.asset_code,asset.asset_name,
			COALESCE(store.store_name,'Head Office'),dtype.name,disposal.disposal_date,disposal.disposal_value,
			GREATEST(COALESCE(profile.salvage_value,0),COALESCE(profile.depreciable_basis,asset.acquisition_value)-COALESCE(posted.amount,0)),
			rule.name,approval.attempt_no,task.step_order,role.name,task.scope,assignee.name,task.status,
			COALESCE(task.comment,''),submitter.name,approval.submitted_at,task.acted_at
		FROM asset_disposal_approval_tasks task
		JOIN asset_disposal_approvals approval ON approval.id=task.approval_id
		JOIN asset_disposals disposal ON disposal.id=approval.disposal_id
		JOIN assets asset ON asset.id=disposal.asset_id LEFT JOIN stores store ON store.store_id=asset.store_id
		JOIN asset_disposal_types dtype ON dtype.id=disposal.disposal_type_id
		LEFT JOIN asset_depreciation_profiles profile ON profile.asset_id=asset.id
		LEFT JOIN (SELECT profile_id,SUM(depreciation_amount) amount FROM asset_depreciation_schedules WHERE status='POSTED' GROUP BY profile_id) posted ON posted.profile_id=profile.id
		JOIN asset_disposal_approval_rules rule ON rule.id=approval.rule_id JOIN roles role ON role.id=task.role_id
		JOIN users assignee ON assignee.id=task.assigned_user_id JOIN users submitter ON submitter.id=approval.submitted_by
		WHERE `+where+` ORDER BY (task.status='WAITING') DESC,approval.submitted_at DESC,task.id DESC LIMIT ? OFFSET ?`, queryArgs...)
	if err != nil {
		return result, err
	}
	defer rows.Close()
	for rows.Next() {
		var item models.AssetDisposalApprovalTask
		var date time.Time
		var value, book float64
		var submitted time.Time
		var acted sql.NullTime
		if err := rows.Scan(&item.ID, &item.ApprovalID, &item.DisposalID, &item.DisposalNumber, &item.AssetID, &item.AssetCode, &item.AssetName, &item.StoreName, &item.DisposalTypeName, &date, &value, &book, &item.RuleName, &item.AttemptNo, &item.StepOrder, &item.RoleName, &item.Scope, &item.AssignedUserName, &item.Status, &item.Comment, &item.SubmittedByName, &submitted, &acted); err != nil {
			return result, err
		}
		item.DisposalDateDisplay = formatDepreciationDateID(date, false)
		item.DisposalValueDisplay = formatAssetAmountID(value)
		item.BookValueDisplay = formatAssetAmountID(book)
		item.SubmittedAtDisplay = formatDepreciationDateID(submitted, true)
		if acted.Valid {
			item.ActedAtDisplay = formatDepreciationDateID(acted.Time, true)
		}
		result.Items = append(result.Items, item)
	}
	return result, rows.Err()
}

func (r *AssetDisposalApprovalRepository) ApproveTask(taskID int64, comment string, auditCtx models.AuditContext) error {
	return r.actTask(taskID, "APPROVED", comment, auditCtx)
}

func (r *AssetDisposalApprovalRepository) RejectTask(taskID int64, reason string, auditCtx models.AuditContext) error {
	return r.actTask(taskID, "REJECTED", reason, auditCtx)
}

func (r *AssetDisposalApprovalRepository) actTask(taskID int64, action, comment string, auditCtx models.AuditContext) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var approvalID, disposalID int64
	var assignedUser, submittedBy, stepOrder int
	var taskStatus, number string
	if err := tx.QueryRow(`SELECT task.approval_id,approval.disposal_id,task.assigned_user_id,approval.submitted_by,task.step_order,task.status,disposal.disposal_number FROM asset_disposal_approval_tasks task JOIN asset_disposal_approvals approval ON approval.id=task.approval_id JOIN asset_disposals disposal ON disposal.id=approval.disposal_id WHERE task.id=? FOR UPDATE`, taskID).Scan(&approvalID, &disposalID, &assignedUser, &submittedBy, &stepOrder, &taskStatus, &number); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("tugas approval tidak ditemukan")
		}
		return err
	}
	if assignedUser != auditCtx.ActorUserID {
		return errors.New("tugas approval ini bukan milik Anda")
	}
	if submittedBy == auditCtx.ActorUserID {
		return errors.New("pembuat disposal tidak boleh menyetujui pengajuannya sendiri")
	}
	if taskStatus != "WAITING" {
		return errors.New("tugas approval sudah diproses atau belum aktif")
	}
	if _, err := tx.Exec(`UPDATE asset_disposal_approval_tasks SET status=?,comment=?,acted_at=CURRENT_TIMESTAMP,updated_at=CURRENT_TIMESTAMP WHERE id=? AND status='WAITING'`, action, nullableString(comment), taskID); err != nil {
		return err
	}
	if err := insertDisposalApprovalHistoryTx(tx, approvalID, &taskID, disposalID, action, taskStatus, action, comment, auditCtx); err != nil {
		return err
	}
	if action == "REJECTED" {
		if _, err := tx.Exec(`UPDATE asset_disposal_approval_tasks SET status='CANCELLED',updated_at=CURRENT_TIMESTAMP WHERE approval_id=? AND status IN ('PENDING','WAITING')`, approvalID); err != nil {
			return err
		}
		if _, err := tx.Exec(`UPDATE asset_disposal_approvals SET status='REJECTED',completed_at=CURRENT_TIMESTAMP WHERE id=?`, approvalID); err != nil {
			return err
		}
		if _, err := tx.Exec(`UPDATE asset_disposals SET status='REJECTED',rejected_by=?,rejected_at=CURRENT_TIMESTAMP,rejection_reason=?,updated_at=CURRENT_TIMESTAMP WHERE id=?`, auditCtx.ActorUserID, comment, disposalID); err != nil {
			return err
		}
		if err := insertAuditLogTx(tx, "ASSET_DISPOSAL", disposalID, "REJECT_APPROVAL", fmt.Sprintf("Approval disposal %s ditolak: %s", number, comment), auditCtx); err != nil {
			return err
		}
		return tx.Commit()
	}
	var remaining int
	if err := tx.QueryRow(`SELECT COUNT(*) FROM asset_disposal_approval_tasks WHERE approval_id=? AND step_order=? AND status NOT IN ('APPROVED','SKIPPED')`, approvalID, stepOrder).Scan(&remaining); err != nil {
		return err
	}
	if remaining == 0 {
		var next sql.NullInt64
		if err := tx.QueryRow(`SELECT MIN(step_order) FROM asset_disposal_approval_tasks WHERE approval_id=? AND status='PENDING'`, approvalID).Scan(&next); err != nil {
			return err
		}
		if next.Valid {
			if _, err := tx.Exec(`UPDATE asset_disposal_approval_tasks SET status='WAITING',updated_at=CURRENT_TIMESTAMP WHERE approval_id=? AND step_order=? AND status='PENDING'`, approvalID, next.Int64); err != nil {
				return err
			}
			if _, err := tx.Exec(`UPDATE asset_disposal_approvals SET current_step=? WHERE id=?`, next.Int64, approvalID); err != nil {
				return err
			}
		} else {
			if _, err := tx.Exec(`UPDATE asset_disposal_approvals SET status='APPROVED',completed_at=CURRENT_TIMESTAMP WHERE id=?`, approvalID); err != nil {
				return err
			}
			if _, err := tx.Exec(`UPDATE asset_disposals SET status='APPROVED',approved_by=?,approved_at=CURRENT_TIMESTAMP,updated_at=CURRENT_TIMESTAMP WHERE id=? AND status='IN_APPROVAL'`, auditCtx.ActorUserID, disposalID); err != nil {
				return err
			}
			if err := insertDisposalApprovalHistoryTx(tx, approvalID, nil, disposalID, "COMPLETE", "IN_APPROVAL", "APPROVED", "Semua tahap approval selesai", auditCtx); err != nil {
				return err
			}
		}
	}
	if err := insertAuditLogTx(tx, "ASSET_DISPOSAL", disposalID, "APPROVE_TASK", fmt.Sprintf("Tahap %d approval disposal %s disetujui", stepOrder, number), auditCtx); err != nil {
		return err
	}
	return tx.Commit()
}

func getApprovalRuleStepsTx(tx *sql.Tx, ruleID int64) ([]models.AssetDisposalApprovalRuleStep, error) {
	rows, err := tx.Query(`SELECT id,step_order,role_id,scope,is_parallel,is_required FROM asset_disposal_approval_rule_steps WHERE rule_id=? ORDER BY step_order,id FOR UPDATE`, ruleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]models.AssetDisposalApprovalRuleStep, 0)
	for rows.Next() {
		var s models.AssetDisposalApprovalRuleStep
		var p, q int
		if err := rows.Scan(&s.ID, &s.StepOrder, &s.RoleID, &s.Scope, &p, &q); err != nil {
			return nil, err
		}
		s.IsParallel = p == 1
		s.IsRequired = q == 1
		items = append(items, s)
	}
	return items, rows.Err()
}

func resolveDisposalApproverTx(tx *sql.Tx, step models.AssetDisposalApprovalRuleStep, storeID sql.NullInt64, submitter int) (int, string, error) {
	type candidate struct {
		user  int
		scope string
	}
	candidates := make([]candidate, 0)
	add := func(scope string, store any) error {
		rows, err := tx.Query(`SELECT mapping.user_id FROM asset_disposal_approvers mapping JOIN model_has_roles mhr ON mhr.model_id=mapping.user_id AND mhr.role_id=mapping.role_id AND mhr.model_type='Models\\User' JOIN users user ON user.id=mapping.user_id AND user.status='active' WHERE mapping.is_active=1 AND mapping.scope=? AND ((? IS NULL AND mapping.store_id IS NULL) OR mapping.store_id=?) AND mapping.role_id=? AND mapping.user_id<>? ORDER BY mapping.id`, scope, store, store, step.RoleID, submitter)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var id int
			if err := rows.Scan(&id); err != nil {
				return err
			}
			candidates = append(candidates, candidate{id, scope})
		}
		return rows.Err()
	}
	if step.Scope == "STORE" || step.Scope == "ANY" {
		if storeID.Valid {
			if err := add("STORE", storeID.Int64); err != nil {
				return 0, "", err
			}
		}
	}
	if len(candidates) == 0 && (step.Scope == "HO" || step.Scope == "ANY") {
		if err := add("HO", nil); err != nil {
			return 0, "", err
		}
	}
	if len(candidates) == 0 {
		return 0, "", fmt.Errorf("approver tahap %d belum tersedia atau hanya menunjuk pembuat disposal", step.StepOrder)
	}
	return candidates[0].user, candidates[0].scope, nil
}

func disposalBookValueTx(tx *sql.Tx, assetID int64) (float64, error) {
	var acquisition float64
	if err := tx.QueryRow(`SELECT acquisition_value FROM assets WHERE id=?`, assetID).Scan(&acquisition); err != nil {
		return 0, err
	}
	var basis, salvage sql.NullFloat64
	var profileID sql.NullInt64
	err := tx.QueryRow(`SELECT id,depreciable_basis,salvage_value FROM asset_depreciation_profiles WHERE asset_id=?`, assetID).Scan(&profileID, &basis, &salvage)
	if errors.Is(err, sql.ErrNoRows) {
		return acquisition, nil
	}
	if err != nil {
		return 0, err
	}
	var posted float64
	if err := tx.QueryRow(`SELECT COALESCE(SUM(depreciation_amount),0) FROM asset_depreciation_schedules WHERE profile_id=? AND status='POSTED'`, profileID.Int64).Scan(&posted); err != nil {
		return 0, err
	}
	return math.Max(salvage.Float64, basis.Float64-posted), nil
}

func insertDisposalApprovalHistoryTx(tx *sql.Tx, approvalID int64, taskID *int64, disposalID int64, action, oldStatus, newStatus, note string, auditCtx models.AuditContext) error {
	var roleID sql.NullInt64
	_ = tx.QueryRow(`SELECT MIN(role_id) FROM model_has_roles WHERE model_id=? AND model_type='Models\\User'`, auditCtx.ActorUserID).Scan(&roleID)
	_, err := tx.Exec(`INSERT INTO asset_disposal_approval_histories (approval_id,task_id,disposal_id,action,old_status,new_status,actor_user_id,actor_role_id,note,ip_address,user_agent) VALUES (?,?,?,?,?,?,?,?,?,?,?)`, approvalID, nullableTaskID(taskID), disposalID, action, nullableString(oldStatus), nullableString(newStatus), auditCtx.ActorUserID, disposalNullableSQLInt64(roleID), nullableString(note), nullableString(auditCtx.IPAddress), nullableString(auditCtx.UserAgent))
	return err
}

func approvalEffectiveLabel(from, until sql.NullTime) string {
	if from.Valid && until.Valid {
		return from.Time.Format("02 Jan 2006") + " - " + until.Time.Format("02 Jan 2006")
	}
	if from.Valid {
		return "Mulai " + from.Time.Format("02 Jan 2006")
	}
	if until.Valid {
		return "Sampai " + until.Time.Format("02 Jan 2006")
	}
	return "Selalu berlaku"
}
func nullablePositiveInt64(value int64) any {
	if value <= 0 {
		return nil
	}
	return value
}
func nullableDateString(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return value
}
func nullableDisposalApproverStoreID(value int) any {
	if value <= 0 {
		return nil
	}
	return value
}
func nullableTaskID(value *int64) any {
	if value == nil {
		return nil
	}
	return *value
}
