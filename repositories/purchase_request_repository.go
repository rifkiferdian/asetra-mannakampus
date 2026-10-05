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

type PurchaseRequestRepository struct {
	DB *sql.DB
}

type approvalRuleMatch struct {
	ID   int64
	Name string
}

type approvalRuleStepResolved struct {
	StepOrder      int
	RoleID         int64
	Scope          string
	IsParallel     bool
	IsRequired     bool
	AssignedUserID int
}

func (r *PurchaseRequestRepository) GetAll(scope models.AccessScope) ([]models.PurchaseRequest, error) {
	query := `
		SELECT
			pr.id,
			pr.pr_number,
			pr.requester_user_id,
			COALESCE(u.name, ''),
			pr.store_id,
			COALESCE(s.store_code, ''),
			COALESCE(s.store_name, ''),
			COALESCE(pr.division_id, 0),
			COALESCE(d.division_name, ''),
			pr.gl_account_id,
			COALESCE(ga.gl_name, ''),
			pr.spend_type,
			pr.urgent_level,
			pr.needed_date,
			COALESCE(pr.justification, ''),
			pr.total_amount,
			pr.status,
			pr.created_at,
			COALESCE(cur_step.current_step, '')
		FROM purchase_requests pr
		LEFT JOIN users u ON u.id = pr.requester_user_id
		LEFT JOIN stores s ON s.store_id = pr.store_id
		LEFT JOIN divisions d ON d.id = pr.division_id
		LEFT JOIN gl_accounts ga ON ga.id = pr.gl_account_id
		LEFT JOIN (
			SELECT
				a.ref_id,
				COALESCE(r.name, '') AS current_step
			FROM approvals a
			JOIN approval_tasks at ON at.approval_id = a.id AND at.status = 'WAITING'
			LEFT JOIN roles r ON r.id = at.role_id
			WHERE a.ref_type = 'PR' AND a.status = 'PENDING'
			AND at.id = (
				SELECT at2.id
				FROM approval_tasks at2
				WHERE at2.approval_id = a.id AND at2.status = 'WAITING'
				ORDER BY at2.step_order ASC, at2.id ASC
				LIMIT 1
			)
		) cur_step ON cur_step.ref_id = pr.id
	`
	args := make([]interface{}, 0)
	if !scope.CanViewAll {
		predicate, predicateArgs := purchaseRequestScopePredicate("pr", scope, true)
		query += " WHERE " + predicate
		args = append(args, predicateArgs...)
	}
	query += " ORDER BY pr.id DESC"

	rows, err := r.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.PurchaseRequest
	for rows.Next() {
		var (
			item       models.PurchaseRequest
			divisionID sql.NullInt64
			neededDate sql.NullTime
			createdAt  sql.NullTime
		)
		if err := rows.Scan(
			&item.ID,
			&item.PRNumber,
			&item.RequesterUserID,
			&item.RequesterName,
			&item.StoreID,
			&item.StoreCode,
			&item.StoreName,
			&divisionID,
			&item.DivisionName,
			&item.GLAccountID,
			&item.GLAccountName,
			&item.SpendType,
			&item.UrgentLevel,
			&neededDate,
			&item.Justification,
			&item.TotalAmount,
			&item.Status,
			&createdAt,
			&item.CurrentStep,
		); err != nil {
			return nil, err
		}
		if divisionID.Valid {
			item.DivisionID = int(divisionID.Int64)
		}
		if neededDate.Valid {
			item.NeededDate = neededDate.Time.Format("2006-01-02")
		}
		if createdAt.Valid {
			item.CreatedAtDisplay = createdAt.Time.Format("02 Jan 2006 15:04:05")
		}
		item.TotalAmountDisplay = formatAmountID(item.TotalAmount)
		item.StatusLabel = formatStatusLabel(item.Status)
		if item.CurrentStep == "" {
			item.CurrentStep = defaultCurrentStep(item.Status)
		}
		item.SLALabel, item.SLAState = formatSLALabel(item.Status, item.NeededDate)
		items = append(items, item)
	}

	return items, rows.Err()
}

func (r *PurchaseRequestRepository) GetDetailByID(id int64, scope models.AccessScope) (*models.PurchaseRequestDetail, error) {
	allowed, err := r.canViewPurchaseRequest(id, scope)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, errors.New("purchase request tidak ditemukan atau akses ditolak")
	}

	var (
		item           models.PurchaseRequestDetail
		divisionID     sql.NullInt64
		neededDate     sql.NullTime
		createdAt      sql.NullTime
		isSingleSource int
	)
	err = r.DB.QueryRow(`
		SELECT
			pr.id,
			pr.pr_number,
			pr.requester_user_id,
			COALESCE(u.name, ''),
			pr.store_id,
			COALESCE(s.store_code, ''),
			COALESCE(s.store_name, ''),
			COALESCE(pr.division_id, 0),
			COALESCE(d.division_name, ''),
			pr.gl_account_id,
			COALESCE(ga.gl_name, ''),
			pr.spend_type,
			pr.urgent_level,
			pr.needed_date,
			COALESCE(pr.justification, ''),
			COALESCE(pr.request_title, ''),
			COALESCE(pr.request_category, ''),
			COALESCE(pr.delivery_location, ''),
			COALESCE(pr.impact_if_not_approved, ''),
			COALESCE(pr.urgency_reason, ''),
			COALESCE(pr.recommended_vendor_id, 0),
			COALESCE(v.name, ''),
			COALESCE(pr.vendor_recommendation_reason, ''),
			pr.is_single_source,
			COALESCE(pr.single_source_reason, ''),
			COALESCE(pr.budget_exception_reason, ''),
			COALESCE(pr.asset_request_type, ''),
			COALESCE(pr.existing_asset_code, ''),
			COALESCE(pr.asset_location, ''),
			COALESCE(pr.asset_pic, ''),
			pr.total_amount,
			pr.status,
			pr.created_at,
			COALESCE(cur_step.current_step, '')
		FROM purchase_requests pr
		LEFT JOIN users u ON u.id = pr.requester_user_id
		LEFT JOIN stores s ON s.store_id = pr.store_id
		LEFT JOIN divisions d ON d.id = pr.division_id
		LEFT JOIN gl_accounts ga ON ga.id = pr.gl_account_id
		LEFT JOIN vendors v ON v.id = pr.recommended_vendor_id
		LEFT JOIN (
			SELECT
				a.ref_id,
				COALESCE(r.name, '') AS current_step
			FROM approvals a
			JOIN approval_tasks at ON at.approval_id = a.id AND at.status = 'WAITING'
			LEFT JOIN roles r ON r.id = at.role_id
			WHERE a.ref_type = 'PR' AND a.status = 'PENDING'
			AND at.id = (
				SELECT at2.id
				FROM approval_tasks at2
				WHERE at2.approval_id = a.id AND at2.status = 'WAITING'
				ORDER BY at2.step_order ASC, at2.id ASC
				LIMIT 1
			)
		) cur_step ON cur_step.ref_id = pr.id
		WHERE pr.id = ?
	`, id).Scan(
		&item.ID,
		&item.PRNumber,
		&item.RequesterUserID,
		&item.RequesterName,
		&item.StoreID,
		&item.StoreCode,
		&item.StoreName,
		&divisionID,
		&item.DivisionName,
		&item.GLAccountID,
		&item.GLAccountName,
		&item.SpendType,
		&item.UrgentLevel,
		&neededDate,
		&item.Justification,
		&item.RequestTitle,
		&item.RequestCategory,
		&item.DeliveryLocation,
		&item.ImpactIfNotApproved,
		&item.UrgencyReason,
		&item.RecommendedVendorID,
		&item.RecommendedVendorName,
		&item.VendorRecommendationReason,
		&isSingleSource,
		&item.SingleSourceReason,
		&item.BudgetExceptionReason,
		&item.AssetRequestType,
		&item.ExistingAssetCode,
		&item.AssetLocation,
		&item.AssetPIC,
		&item.TotalAmount,
		&item.Status,
		&createdAt,
		&item.CurrentStep,
	)
	if err != nil {
		return nil, err
	}

	if divisionID.Valid {
		item.DivisionID = int(divisionID.Int64)
	}
	if neededDate.Valid {
		item.NeededDate = neededDate.Time.Format("2006-01-02")
	}
	if createdAt.Valid {
		item.CreatedAtDisplay = createdAt.Time.Format("02 Jan 2006")
	}
	item.IsSingleSource = isSingleSource == 1
	item.TotalAmountDisplay = formatAmountID(item.TotalAmount)
	item.StatusLabel = formatStatusLabel(item.Status)
	if item.CurrentStep == "" {
		item.CurrentStep = defaultCurrentStep(item.Status)
	}
	item.SLALabel, item.SLAState = formatSLALabel(item.Status, item.NeededDate)
	budgetCheck, budgetErr := r.GetBudgetCheck(item.StoreID, item.DivisionID, item.GLAccountID, item.NeededDate, item.TotalAmount)
	item.BudgetImpactLabel = buildBudgetImpactLabel(item.DivisionName, item.GLAccountName)
	if budgetErr == nil {
		item.BudgetUtilizedPct = budgetCheck.UtilizedPct
		item.BudgetMessage = budgetCheck.Message
	} else {
		item.BudgetMessage = "Budget belum dapat dihitung."
	}

	items, err := r.getItemsByPRID(id)
	if err != nil {
		return nil, err
	}
	item.Items = items

	steps, err := r.getApprovalStepsByPRID(id)
	if err != nil {
		return nil, err
	}
	item.ApprovalSteps = steps

	taskID, err := r.getCurrentUserTaskID(id, scope.UserID)
	if err != nil {
		return nil, err
	}
	item.CurrentUserTaskID = taskID

	attachments, err := r.getAttachmentsByRef("PR", id)
	if err != nil {
		return nil, err
	}
	item.Attachments = attachments

	return &item, nil
}

func (r *PurchaseRequestRepository) getItemsByPRID(prID int64) ([]models.PurchaseRequestItem, error) {
	rows, err := r.DB.Query(`
		SELECT id, pr_id, item_name, COALESCE(specification, ''), qty, uom, est_unit_price, est_total, COALESCE(price_source, ''), COALESCE(notes, '')
		FROM purchase_request_items
		WHERE pr_id = ?
		ORDER BY id ASC
	`, prID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.PurchaseRequestItem
	for rows.Next() {
		var item models.PurchaseRequestItem
		if err := rows.Scan(&item.ID, &item.PRID, &item.ItemName, &item.Specification, &item.Qty, &item.UOM, &item.EstUnitPrice, &item.EstTotal, &item.PriceSource, &item.Notes); err != nil {
			return nil, err
		}
		item.QtyDisplay = formatQty(item.Qty)
		item.EstUnitPriceDisplay = formatAmountID(item.EstUnitPrice)
		item.EstTotalDisplay = formatAmountID(item.EstTotal)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *PurchaseRequestRepository) getApprovalStepsByPRID(prID int64) ([]models.PurchaseRequestApprovalStep, error) {
	rows, err := r.DB.Query(`
		SELECT
			at.id,
			at.step_order,
			COALESCE(role.name, ''),
			COALESCE(assigned.name, ''),
			at.status,
			at.acted_at,
			at.created_at
		FROM approvals a
		JOIN approval_tasks at ON at.approval_id = a.id
		LEFT JOIN roles role ON role.id = at.role_id
		LEFT JOIN users assigned ON assigned.id = at.assigned_user_id
		WHERE a.ref_type = 'PR' AND a.ref_id = ?
		ORDER BY at.step_order ASC, at.id ASC
	`, prID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var steps []models.PurchaseRequestApprovalStep
	for rows.Next() {
		var (
			step      models.PurchaseRequestApprovalStep
			actedAt   sql.NullTime
			createdAt sql.NullTime
		)
		if err := rows.Scan(&step.TaskID, &step.StepOrder, &step.RoleName, &step.AssignedUserName, &step.Status, &actedAt, &createdAt); err != nil {
			return nil, err
		}
		step.StatusLabel = formatApprovalTaskStatusLabel(step.Status)
		if actedAt.Valid {
			step.ActedAtDisplay = actedAt.Time.Format("02 Jan, 03:04 PM")
		}
		if createdAt.Valid {
			step.CreatedAtDisplay = createdAt.Time.Format("02 Jan, 03:04 PM")
		}
		steps = append(steps, step)
	}
	return steps, rows.Err()
}

func (r *PurchaseRequestRepository) getCurrentUserTaskID(prID int64, userID int) (int64, error) {
	if userID <= 0 {
		return 0, nil
	}
	var taskID int64
	err := r.DB.QueryRow(`
		SELECT at.id
		FROM approvals a
		JOIN approval_tasks at ON at.approval_id = a.id
		WHERE a.ref_type = 'PR'
		AND a.ref_id = ?
		AND a.status = 'PENDING'
		AND at.assigned_user_id = ?
		AND at.status = 'WAITING'
		AND NOT EXISTS (
			SELECT 1
			FROM approval_tasks prev
			WHERE prev.approval_id = at.approval_id
			AND prev.step_order < at.step_order
			AND prev.status NOT IN ('APPROVED', 'SKIPPED')
		)
		ORDER BY at.step_order ASC, at.id ASC
		LIMIT 1
	`, prID, userID).Scan(&taskID)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return taskID, err
}

func (r *PurchaseRequestRepository) getAttachmentsByRef(refType string, refID int64) ([]models.Attachment, error) {
	rows, err := r.DB.Query(`
		SELECT id, ref_type, ref_id, file_path, file_name, COALESCE(mime_type, ''), COALESCE(file_size, 0), uploaded_by, created_at
		FROM attachments
		WHERE ref_type = ? AND ref_id = ?
		ORDER BY id ASC
	`, refType, refID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var attachments []models.Attachment
	for rows.Next() {
		var (
			item      models.Attachment
			createdAt sql.NullTime
		)
		if err := rows.Scan(&item.ID, &item.RefType, &item.RefID, &item.FilePath, &item.FileName, &item.MimeType, &item.FileSize, &item.UploadedBy, &createdAt); err != nil {
			return nil, err
		}
		if createdAt.Valid {
			item.CreatedAtDisplay = createdAt.Time.Format("02 Jan 2006 15:04")
		}
		attachments = append(attachments, item)
	}
	return attachments, rows.Err()
}

func (r *PurchaseRequestRepository) Create(input models.PurchaseRequestCreateInput, totalAmount models.Money) (int64, error) {
	tx, err := r.DB.Begin()
	if err != nil {
		return 0, err
	}

	prNumber, err := r.nextPRNumberTx(tx, input.StoreID)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	status := "DRAFT"
	if input.Action == "submit" {
		status = "SUBMITTED"
	}

	res, err := tx.Exec(`
		INSERT INTO purchase_requests (
			pr_number, request_title, request_category, requester_user_id, store_id, division_id, gl_account_id, spend_type, urgent_level,
			needed_date, delivery_location, justification, impact_if_not_approved, urgency_reason, recommended_vendor_id,
			vendor_recommendation_reason, is_single_source, single_source_reason, budget_exception_reason,
			asset_request_type, existing_asset_code, asset_location, asset_pic, total_amount, status
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, prNumber, input.RequestTitle, input.RequestCategory, input.RequesterUserID, input.StoreID, nullableInt(input.DivisionID), input.GLAccountID,
		input.SpendType, input.UrgentLevel, nullableDate(input.NeededDate), nullableString(input.DeliveryLocation), nullableString(input.Justification),
		nullableString(input.ImpactIfNotApproved), nullableString(input.UrgencyReason), nullableInt64(input.RecommendedVendorID),
		nullableString(input.VendorRecommendationReason), boolToInt(input.IsSingleSource), nullableString(input.SingleSourceReason),
		nullableString(input.BudgetExceptionReason), nullableString(input.AssetRequestType), nullableString(input.ExistingAssetCode),
		nullableString(input.AssetLocation), nullableString(input.AssetPIC), totalAmount, status)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	prID, err := res.LastInsertId()
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	if err := insertPRItems(tx, prID, input.Items); err != nil {
		tx.Rollback()
		return 0, err
	}

	if err := insertPRAttachments(tx, prID, input.RequesterUserID, input.Attachments); err != nil {
		tx.Rollback()
		return 0, err
	}

	if input.Action == "submit" {
		rule, steps, err := r.buildApprovalFlowTx(tx, prID, input.StoreID, input.SpendType, input.UrgentLevel, totalAmount, input.RequesterUserID)
		if err != nil {
			tx.Rollback()
			return 0, err
		}
		if rule != nil && len(steps) > 0 {
			approvalID, err := insertApprovalFlowTx(tx, prID, *rule, steps, input.RequesterUserID)
			if err != nil {
				tx.Rollback()
				return 0, err
			}
			if _, err := tx.Exec(`UPDATE purchase_requests SET status = 'IN_APPROVAL' WHERE id = ?`, prID); err != nil {
				tx.Rollback()
				return 0, err
			}
			if err := insertAuditLogTx(tx, "PR", prID, "SUBMIT", "PR disubmit dan approval flow dibuat", input.AuditContext); err != nil {
				tx.Rollback()
				return 0, err
			}
			if err := insertAuditLogTx(tx, "APPROVAL", approvalID, "CREATE", "Approval flow generated by system", input.AuditContext); err != nil {
				tx.Rollback()
				return 0, err
			}
		} else {
			if err := insertAuditLogTx(tx, "PR", prID, "SUBMIT", "PR disubmit tanpa approval rule yang cocok", input.AuditContext); err != nil {
				tx.Rollback()
				return 0, err
			}
		}
	} else {
		if err := insertAuditLogTx(tx, "PR", prID, "CREATE_DRAFT", "PR draft dibuat", input.AuditContext); err != nil {
			tx.Rollback()
			return 0, err
		}
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return 0, err
	}

	return prID, nil
}

func (r *PurchaseRequestRepository) RegenerateApprovalFlow(prID int64, scope models.AccessScope, auditCtx models.AuditContext) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}

	var (
		status      string
		storeID     int
		spendType   string
		urgentLevel string
		totalAmount models.Money
		requesterID int
	)
	err = tx.QueryRow(`
		SELECT status, store_id, spend_type, urgent_level, total_amount, requester_user_id
		FROM purchase_requests
		WHERE id = ?
		FOR UPDATE
	`, prID).Scan(&status, &storeID, &spendType, &urgentLevel, &totalAmount, &requesterID)
	if err == sql.ErrNoRows {
		tx.Rollback()
		return fmt.Errorf("purchase request tidak ditemukan")
	}
	if err != nil {
		tx.Rollback()
		return err
	}
	if !scope.AllowsOwnerOrStore(requesterID, storeID) {
		tx.Rollback()
		return errors.New("purchase request tidak ditemukan atau akses ditolak")
	}

	if status != "SUBMITTED" {
		tx.Rollback()
		return fmt.Errorf("approval workflow hanya bisa dibuat ulang untuk PR berstatus SUBMITTED, status saat ini %s", status)
	}

	var existing int
	if err := tx.QueryRow(`SELECT COUNT(1) FROM approvals WHERE ref_type = 'PR' AND ref_id = ?`, prID).Scan(&existing); err != nil {
		tx.Rollback()
		return err
	}
	if existing > 0 {
		tx.Rollback()
		return fmt.Errorf("approval workflow sudah ada untuk PR ini")
	}

	rule, steps, err := r.buildApprovalFlowTx(tx, prID, storeID, spendType, urgentLevel, totalAmount, requesterID)
	if err != nil {
		tx.Rollback()
		return err
	}
	if rule == nil || len(steps) == 0 {
		tx.Rollback()
		return fmt.Errorf("approval workflow belum bisa dibuat karena tidak ada approval rule aktif yang cocok dengan PR ini")
	}

	approvalID, err := insertApprovalFlowTx(tx, prID, *rule, steps, requesterID)
	if err != nil {
		tx.Rollback()
		return err
	}

	if _, err := tx.Exec(`UPDATE purchase_requests SET status = 'IN_APPROVAL' WHERE id = ?`, prID); err != nil {
		tx.Rollback()
		return err
	}

	if err := insertAuditLogTx(tx, "PR", prID, "REGENERATE_APPROVAL", "Approval workflow dibuat ulang dari rule yang aktif", auditCtx); err != nil {
		tx.Rollback()
		return err
	}
	if err := insertAuditLogTx(tx, "APPROVAL", approvalID, "CREATE", "Approval flow regenerated by system", auditCtx); err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

func (r *PurchaseRequestRepository) UpdateEditable(input models.PurchaseRequestUpdateInput, totalAmount models.Money) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}

	var (
		status         string
		requesterID    int
		currentStoreID int
	)
	if err := tx.QueryRow(`SELECT status, requester_user_id, store_id FROM purchase_requests WHERE id = ? FOR UPDATE`, input.ID).Scan(&status, &requesterID, &currentStoreID); err != nil {
		tx.Rollback()
		if err == sql.ErrNoRows {
			return fmt.Errorf("purchase request tidak ditemukan")
		}
		return err
	}
	if !input.AccessScope.AllowsOwnerOrStore(requesterID, currentStoreID) || !input.AccessScope.AllowsStore(input.StoreID) {
		tx.Rollback()
		return errors.New("purchase request tidak ditemukan atau akses ditolak")
	}

	if status != "DRAFT" && status != "SUBMITTED" {
		tx.Rollback()
		return fmt.Errorf("PR berstatus %s tidak bisa diedit", status)
	}

	var approvalCount int
	if err := tx.QueryRow(`SELECT COUNT(1) FROM approvals WHERE ref_type = 'PR' AND ref_id = ?`, input.ID).Scan(&approvalCount); err != nil {
		tx.Rollback()
		return err
	}
	if approvalCount > 0 {
		tx.Rollback()
		return fmt.Errorf("PR tidak bisa diedit karena approval workflow sudah dibuat")
	}

	_, err = tx.Exec(`
		UPDATE purchase_requests
		SET request_title = ?,
			request_category = ?,
			store_id = ?,
			division_id = ?,
			gl_account_id = ?,
			spend_type = ?,
			urgent_level = ?,
			needed_date = ?,
			delivery_location = ?,
			justification = ?,
			impact_if_not_approved = ?,
			urgency_reason = ?,
			recommended_vendor_id = ?,
			vendor_recommendation_reason = ?,
			is_single_source = ?,
			single_source_reason = ?,
			budget_exception_reason = ?,
			asset_request_type = ?,
			existing_asset_code = ?,
			asset_location = ?,
			asset_pic = ?,
			total_amount = ?
		WHERE id = ?
	`, input.RequestTitle, input.RequestCategory, input.StoreID, nullableInt(input.DivisionID), input.GLAccountID, input.SpendType,
		input.UrgentLevel, nullableDate(input.NeededDate), nullableString(input.DeliveryLocation), nullableString(input.Justification),
		nullableString(input.ImpactIfNotApproved), nullableString(input.UrgencyReason), nullableInt64(input.RecommendedVendorID),
		nullableString(input.VendorRecommendationReason), boolToInt(input.IsSingleSource), nullableString(input.SingleSourceReason),
		nullableString(input.BudgetExceptionReason), nullableString(input.AssetRequestType), nullableString(input.ExistingAssetCode),
		nullableString(input.AssetLocation), nullableString(input.AssetPIC), totalAmount, input.ID)
	if err != nil {
		tx.Rollback()
		return err
	}

	if _, err := tx.Exec(`DELETE FROM purchase_request_items WHERE pr_id = ?`, input.ID); err != nil {
		tx.Rollback()
		return err
	}
	if err := insertPRItems(tx, input.ID, input.Items); err != nil {
		tx.Rollback()
		return err
	}

	if err := insertAuditLogTx(tx, "PR", input.ID, "UPDATE", "PR editable fields updated", input.AuditContext); err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

func (r *PurchaseRequestRepository) GetStoreWithCode(storeID int) (*models.Store, error) {
	var item models.Store
	err := r.DB.QueryRow(`SELECT store_id, store_code, store_name FROM stores WHERE store_id = ?`, storeID).Scan(&item.StoreID, &item.StoreCode, &item.StoreName)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *PurchaseRequestRepository) GLAccountSpendType(glAccountID int) (string, error) {
	var spendType string
	err := r.DB.QueryRow(`SELECT spend_type FROM gl_accounts WHERE id = ?`, glAccountID).Scan(&spendType)
	return spendType, err
}

func (r *PurchaseRequestRepository) GLAccountExists(glAccountID int) (bool, error) {
	var count int
	err := r.DB.QueryRow(`SELECT COUNT(1) FROM gl_accounts WHERE id = ?`, glAccountID).Scan(&count)
	return count > 0, err
}

func (r *PurchaseRequestRepository) StoreExists(storeID int) (bool, error) {
	var count int
	err := r.DB.QueryRow(`SELECT COUNT(1) FROM stores WHERE store_id = ?`, storeID).Scan(&count)
	return count > 0, err
}

func (r *PurchaseRequestRepository) UserExists(userID int) (bool, error) {
	var count int
	err := r.DB.QueryRow(`SELECT COUNT(1) FROM users WHERE id = ?`, userID).Scan(&count)
	return count > 0, err
}

func (r *PurchaseRequestRepository) VendorExists(vendorID int64) (bool, error) {
	var count int
	err := r.DB.QueryRow(`SELECT COUNT(1) FROM vendors WHERE id = ? AND is_active = 1`, vendorID).Scan(&count)
	return count > 0, err
}

func (r *PurchaseRequestRepository) canViewPurchaseRequest(id int64, scope models.AccessScope) (bool, error) {
	if id <= 0 || scope.UserID <= 0 {
		return false, nil
	}
	query := "SELECT COUNT(1) FROM purchase_requests pr WHERE pr.id = ?"
	args := []interface{}{id}
	if !scope.CanViewAll {
		predicate, predicateArgs := purchaseRequestScopePredicate("pr", scope, true)
		query += " AND (" + predicate + ")"
		args = append(args, predicateArgs...)
	}
	var count int
	if err := r.DB.QueryRow(query, args...).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

func purchaseRequestScopePredicate(alias string, scope models.AccessScope, includeAssignedApproval bool) (string, []interface{}) {
	clauses := []string{alias + ".requester_user_id = ?"}
	args := []interface{}{scope.UserID}
	if len(scope.StoreIDs) > 0 {
		placeholders := make([]string, 0, len(scope.StoreIDs))
		for _, storeID := range scope.StoreIDs {
			if storeID <= 0 {
				continue
			}
			placeholders = append(placeholders, "?")
			args = append(args, storeID)
		}
		if len(placeholders) > 0 {
			clauses = append(clauses, alias+".store_id IN ("+strings.Join(placeholders, ",")+")")
		}
	}
	if includeAssignedApproval {
		clauses = append(clauses, `EXISTS (
			SELECT 1 FROM approvals scope_approval
			JOIN approval_tasks scope_task ON scope_task.approval_id = scope_approval.id
			WHERE scope_approval.ref_type = 'PR'
			AND scope_approval.ref_id = `+alias+`.id
			AND scope_task.assigned_user_id = ?
		)`)
		args = append(args, scope.UserID)
	}
	return "(" + strings.Join(clauses, " OR ") + ")", args
}

func (r *PurchaseRequestRepository) nextPRNumberTx(tx *sql.Tx, storeID int) (string, error) {
	var storeCode string
	// Lock the store row so sequence allocation is serialized per store.
	if err := tx.QueryRow(`SELECT store_code FROM stores WHERE store_id = ? FOR UPDATE`, storeID).Scan(&storeCode); err != nil {
		return "", err
	}

	year := time.Now().Year()
	prefix := fmt.Sprintf("PR-%s-%d-", storeCode, year)
	var lastNumber sql.NullString
	if err := tx.QueryRow(`
		SELECT pr_number
		FROM purchase_requests
		WHERE pr_number LIKE ?
		ORDER BY id DESC
		LIMIT 1
	`, prefix+"%").Scan(&lastNumber); err != nil && err != sql.ErrNoRows {
		return "", err
	}

	seq := 1
	if lastNumber.Valid {
		var parsed int
		fmt.Sscanf(lastNumber.String, prefix+"%d", &parsed)
		if parsed > 0 {
			seq = parsed + 1
		}
	}

	return fmt.Sprintf("%s%04d", prefix, seq), nil
}

func insertPRItems(tx *sql.Tx, prID int64, items []models.PurchaseRequestItemInput) error {
	stmt, err := tx.Prepare(`
		INSERT INTO purchase_request_items (pr_id, item_name, specification, qty, uom, est_unit_price, est_total, price_source, notes)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, item := range items {
		total, err := models.MultiplyMoneyByQuantity(item.EstUnitPrice, item.Qty)
		if err != nil {
			return err
		}
		if _, err := stmt.Exec(prID, item.ItemName, nullableString(item.Specification), item.Qty, item.UOM, item.EstUnitPrice, total, nullableString(item.PriceSource), nullableString(item.Notes)); err != nil {
			return err
		}
	}
	return nil
}

func insertPRAttachments(tx *sql.Tx, prID int64, uploadedBy int, attachments []models.AttachmentFileInput) error {
	if len(attachments) == 0 {
		return nil
	}
	stmt, err := tx.Prepare(`
		INSERT INTO attachments (ref_type, ref_id, file_path, file_name, mime_type, file_size, uploaded_by)
		VALUES ('PR', ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, attachment := range attachments {
		if _, err := stmt.Exec(prID, attachment.FilePath, attachment.FileName, nullableString(attachment.MimeType), attachment.FileSize, uploadedBy); err != nil {
			return err
		}
	}
	return nil
}

func (r *PurchaseRequestRepository) GetFormCheck(storeID, divisionID, glAccountID int, neededDate string, prAmount models.Money, urgentLevel string) (*models.PurchaseRequestFormCheck, error) {
	spendType, err := r.GLAccountSpendType(glAccountID)
	if err != nil {
		return nil, err
	}
	budget, err := r.GetBudgetCheck(storeID, divisionID, glAccountID, neededDate, prAmount)
	if err != nil {
		return nil, err
	}
	ruleName, steps, err := r.getApprovalPreview(storeID, spendType, urgentLevel, prAmount)
	if err != nil {
		return nil, err
	}
	return &models.PurchaseRequestFormCheck{
		SpendType:     strings.ToUpper(spendType),
		Budget:        *budget,
		ApprovalRule:  ruleName,
		ApprovalSteps: steps,
	}, nil
}

func (r *PurchaseRequestRepository) GetBudgetCheck(storeID, divisionID, glAccountID int, neededDate string, prAmount models.Money) (*models.PurchaseRequestBudgetCheck, error) {
	checkDate := time.Now()
	if strings.TrimSpace(neededDate) != "" {
		parsed, err := time.Parse("2006-01-02", neededDate)
		if err != nil {
			return nil, fmt.Errorf("needed date tidak valid")
		}
		checkDate = parsed
	}
	monthKey := checkDate.Format("2006-01")
	yearKey := checkDate.Format("2006")
	quarterKey := fmt.Sprintf("%d-Q%d", checkDate.Year(), (int(checkDate.Month())-1)/3+1)

	var result models.PurchaseRequestBudgetCheck
	var periodType, periodKey string
	err := r.DB.QueryRow(`
		SELECT b.id, b.period_type, b.period_key, b.amount, COALESCE(SUM(bu.used_amount), 0)
		FROM budgets b
		LEFT JOIN budget_usages bu ON bu.budget_id = b.id
		WHERE b.fiscal_year = ?
		  AND b.gl_account_id = ?
		  AND (b.store_id IS NULL OR b.store_id = ?)
		  AND (b.division_id IS NULL OR b.division_id = ?)
		  AND ((b.period_type = 'MONTHLY' AND b.period_key = ?)
		    OR (b.period_type = 'QUARTERLY' AND b.period_key = ?)
		    OR (b.period_type = 'YEARLY' AND b.period_key = ?))
		GROUP BY b.id, b.period_type, b.period_key, b.amount, b.store_id, b.division_id
		ORDER BY (b.store_id IS NOT NULL) DESC, (b.division_id IS NOT NULL) DESC,
		  FIELD(b.period_type, 'MONTHLY', 'QUARTERLY', 'YEARLY') ASC, b.id DESC
		LIMIT 1
	`, checkDate.Year(), glAccountID, storeID, divisionID, monthKey, quarterKey, yearKey).Scan(
		&result.BudgetID, &periodType, &periodKey, &result.Amount, &result.UsedAmount,
	)
	result.PRAmount = prAmount
	if err == sql.ErrNoRows {
		result.Status = "UNBUDGETED"
		result.Message = "Budget yang sesuai dengan store, divisi, GL, dan periode belum tersedia."
		return &result, nil
	}
	if err != nil {
		return nil, err
	}

	result.PeriodLabel = periodType + " " + periodKey
	result.RemainingAmount = result.Amount.Sub(result.UsedAmount)
	result.AfterPRAmount = result.RemainingAmount.Sub(prAmount)
	if result.Amount > 0 {
		result.UtilizedPct = models.PercentageRounded(result.UsedAmount.Add(prAmount), result.Amount)
	}
	switch {
	case result.AfterPRAmount < 0:
		result.Status = "INSUFFICIENT"
		result.Message = "Estimasi PR melebihi sisa budget. Alasan pengecualian budget wajib diisi sebelum submit."
	case result.UtilizedPct >= 90:
		result.Status = "WARNING"
		result.Message = "Budget tersedia, tetapi utilisasi setelah PR mencapai sedikitnya 90%."
	default:
		result.Status = "AVAILABLE"
		result.Message = "Estimasi PR masih berada dalam budget yang tersedia."
	}
	return &result, nil
}

func (r *PurchaseRequestRepository) getApprovalPreview(storeID int, spendType, urgentLevel string, totalAmount models.Money) (string, []models.PurchaseRequestApprovalPreviewStep, error) {
	var ruleID int64
	var ruleName string
	err := r.DB.QueryRow(`
		SELECT id, name FROM approval_rules
		WHERE is_active = 1
		  AND location_scope IN ('ANY', 'STORE')
		  AND spend_type IN ('ANY', ?)
		  AND urgent_level IN ('ANY', ?)
		  AND min_amount <= ?
		  AND (max_amount IS NULL OR max_amount >= ?)
		ORDER BY min_amount DESC, id ASC LIMIT 1
	`, spendType, urgentLevel, totalAmount, totalAmount).Scan(&ruleID, &ruleName)
	if err == sql.ErrNoRows {
		return "", nil, nil
	}
	if err != nil {
		return "", nil, err
	}
	rows, err := r.DB.Query(`
		SELECT ars.step_order, ars.role_id, ars.scope, COALESCE(role.name, '')
		FROM approval_rule_steps ars
		LEFT JOIN roles role ON role.id = ars.role_id
		WHERE ars.rule_id = ? ORDER BY ars.step_order, ars.id
	`, ruleID)
	if err != nil {
		return "", nil, err
	}
	type previewAssignment struct {
		step   models.PurchaseRequestApprovalPreviewStep
		roleID int64
		scope  string
	}
	var assignments []previewAssignment
	for rows.Next() {
		var assignment previewAssignment
		if err := rows.Scan(&assignment.step.StepOrder, &assignment.roleID, &assignment.scope, &assignment.step.RoleName); err != nil {
			rows.Close()
			return "", nil, err
		}
		assignments = append(assignments, assignment)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return "", nil, err
	}
	if err := rows.Close(); err != nil {
		return "", nil, err
	}
	var steps []models.PurchaseRequestApprovalPreviewStep
	for _, assignment := range assignments {
		step := assignment.step
		if assignment.scope == "STORE" {
			_ = r.DB.QueryRow(`SELECT COALESCE(u.name, '') FROM store_approvers sa JOIN users u ON u.id = sa.user_id WHERE sa.store_id = ? AND sa.role_id = ? AND sa.is_active = 1 ORDER BY sa.id LIMIT 1`, storeID, assignment.roleID).Scan(&step.ApproverName)
		} else {
			_ = r.DB.QueryRow(`SELECT COALESCE(u.name, '') FROM users u JOIN model_has_roles mhr ON mhr.model_id = u.id AND mhr.model_type = 'Models\\User' WHERE mhr.role_id = ? AND u.status = 'active' ORDER BY u.id LIMIT 1`, assignment.roleID).Scan(&step.ApproverName)
		}
		steps = append(steps, step)
	}
	return ruleName, steps, nil
}

func (r *PurchaseRequestRepository) buildApprovalFlowTx(tx *sql.Tx, prID int64, storeID int, spendType, urgentLevel string, totalAmount models.Money, actorUserID int) (*approvalRuleMatch, []approvalRuleStepResolved, error) {
	var rule approvalRuleMatch
	err := tx.QueryRow(`
		SELECT id, name
		FROM approval_rules
		WHERE is_active = 1
		AND location_scope IN ('ANY', 'STORE')
		AND spend_type IN ('ANY', ?)
		AND urgent_level IN ('ANY', ?)
		AND min_amount <= ?
		AND (max_amount IS NULL OR max_amount >= ?)
		ORDER BY min_amount DESC, id ASC
		LIMIT 1
	`, spendType, urgentLevel, totalAmount, totalAmount).Scan(&rule.ID, &rule.Name)
	if err == sql.ErrNoRows {
		return nil, nil, nil
	}
	if err != nil {
		return nil, nil, err
	}

	rows, err := tx.Query(`
		SELECT step_order, role_id, scope, is_parallel, is_required
		FROM approval_rule_steps
		WHERE rule_id = ?
		ORDER BY step_order ASC, id ASC
	`, rule.ID)
	if err != nil {
		return nil, nil, err
	}

	var steps []approvalRuleStepResolved
	for rows.Next() {
		var (
			step       approvalRuleStepResolved
			isParallel int
			isRequired int
		)
		if err := rows.Scan(&step.StepOrder, &step.RoleID, &step.Scope, &isParallel, &isRequired); err != nil {
			return nil, nil, err
		}
		step.IsParallel = isParallel == 1
		step.IsRequired = isRequired == 1
		steps = append(steps, step)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, nil, err
	}
	if err := rows.Close(); err != nil {
		return nil, nil, err
	}

	for i := range steps {
		assignedUserID, err := resolveApproverUserTx(tx, storeID, steps[i].RoleID, steps[i].Scope)
		if err != nil {
			return nil, nil, err
		}
		steps[i].AssignedUserID = assignedUserID
	}

	return &rule, steps, nil
}

func resolveApproverUserTx(tx *sql.Tx, storeID int, roleID int64, scope string) (int, error) {
	var userID int
	if scope == "STORE" {
		err := tx.QueryRow(`
			SELECT user_id
			FROM store_approvers
			WHERE store_id = ? AND role_id = ? AND is_active = 1
			ORDER BY id ASC
			LIMIT 1
		`, storeID, roleID).Scan(&userID)
		if err != nil {
			storeLabel, roleLabel := describeStoreRoleTx(tx, storeID, roleID)
			return 0, fmt.Errorf("store approver tidak ditemukan untuk %s dengan role %s", storeLabel, roleLabel)
		}
		return userID, nil
	}

	err := tx.QueryRow(`
		SELECT u.id
		FROM users u
		JOIN model_has_roles mhr ON mhr.model_id = u.id AND mhr.model_type = 'Models\\User'
		WHERE mhr.role_id = ? AND u.status = 'active'
		ORDER BY u.id ASC
		LIMIT 1
	`, roleID).Scan(&userID)
	if err != nil {
		storeLabel, roleLabel := describeStoreRoleTx(tx, storeID, roleID)
		return 0, fmt.Errorf("approver HO tidak ditemukan untuk role %s pada PR store %s", roleLabel, storeLabel)
	}
	return userID, nil
}

func describeStoreRoleTx(tx *sql.Tx, storeID int, roleID int64) (string, string) {
	storeLabel := fmt.Sprintf("store ID %d", storeID)
	roleLabel := describeRoleTx(tx, roleID)

	var storeCode, storeName string
	if err := tx.QueryRow(`
		SELECT COALESCE(store_code, ''), COALESCE(store_name, '')
		FROM stores
		WHERE store_id = ?
	`, storeID).Scan(&storeCode, &storeName); err == nil {
		switch {
		case strings.TrimSpace(storeCode) != "" && strings.TrimSpace(storeName) != "":
			storeLabel = fmt.Sprintf("%s - %s", storeCode, storeName)
		case strings.TrimSpace(storeCode) != "":
			storeLabel = storeCode
		case strings.TrimSpace(storeName) != "":
			storeLabel = storeName
		}
	}

	return storeLabel, roleLabel
}

func describeRoleTx(tx *sql.Tx, roleID int64) string {
	roleLabel := fmt.Sprintf("role ID %d", roleID)
	var roleName string
	if err := tx.QueryRow(`SELECT COALESCE(name, '') FROM roles WHERE id = ?`, roleID).Scan(&roleName); err == nil && strings.TrimSpace(roleName) != "" {
		roleLabel = fmt.Sprintf("%s (ID %d)", roleName, roleID)
	}
	return roleLabel
}

func insertApprovalFlowTx(tx *sql.Tx, prID int64, rule approvalRuleMatch, steps []approvalRuleStepResolved, actorUserID int) (int64, error) {
	res, err := tx.Exec(`
		INSERT INTO approvals (ref_type, ref_id, rule_id, status)
		VALUES ('PR', ?, ?, 'PENDING')
	`, prID, rule.ID)
	if err != nil {
		return 0, err
	}
	approvalID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	stmt, err := tx.Prepare(`
		INSERT INTO approval_tasks (approval_id, step_order, role_id, scope, assigned_user_id, status)
		VALUES (?, ?, ?, ?, ?, 'WAITING')
	`)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	for _, step := range steps {
		if _, err := stmt.Exec(approvalID, step.StepOrder, step.RoleID, step.Scope, step.AssignedUserID); err != nil {
			return 0, err
		}
	}

	if _, err := tx.Exec(`
		INSERT INTO approval_histories (approval_id, task_id, ref_type, ref_id, action, actor_user_id, actor_role_id, note)
		VALUES (?, NULL, 'PR', ?, 'CREATED', ?, NULL, ?)
	`, approvalID, prID, actorUserID, fmt.Sprintf("Approval flow generated by system from rule %s", rule.Name)); err != nil {
		return 0, err
	}

	return approvalID, nil
}

func insertAuditLogTx(tx *sql.Tx, refType string, refID int64, action, message string, ctx models.AuditContext) error {
	_, err := tx.Exec(`
		INSERT INTO audit_logs (ref_type, ref_id, action, message, actor_user_id, ip_address, user_agent)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, refType, refID, action, nullableString(message), ctx.ActorUserID, nullableString(ctx.IPAddress), nullableString(ctx.UserAgent))
	return err
}

func nullableInt(value int) interface{} {
	if value <= 0 {
		return nil
	}
	return value
}

func nullableInt64(value int64) interface{} {
	if value <= 0 {
		return nil
	}
	return value
}

func nullableDate(value string) interface{} {
	if value == "" {
		return nil
	}
	return value
}

func formatAmountID(value models.Money) string {
	return value.FormatIDR()
}

func formatStatusLabel(status string) string {
	switch status {
	case "DRAFT":
		return "Draft"
	case "SUBMITTED":
		return "Submitted"
	case "IN_APPROVAL":
		return "In Approval"
	case "REJECTED":
		return "Rejected"
	case "APPROVED":
		return "Approved"
	case "CONVERTED_TO_PO":
		return "Converted"
	case "CLOSED":
		return "Closed"
	default:
		return status
	}
}

func defaultCurrentStep(status string) string {
	switch status {
	case "DRAFT":
		return "Drafting"
	case "SUBMITTED":
		return "Internal Check"
	case "IN_APPROVAL":
		return "Manager Approval"
	case "APPROVED":
		return "PO Creation"
	case "CONVERTED_TO_PO":
		return "PO Created"
	case "REJECTED":
		return "Manager Approval"
	case "CLOSED":
		return "Closed"
	default:
		return "-"
	}
}

func formatSLALabel(status, neededDate string) (string, string) {
	if status == "APPROVED" || status == "CONVERTED_TO_PO" || status == "CLOSED" {
		return "Completed", "completed"
	}
	if status == "REJECTED" {
		return "Overdue", "overdue"
	}
	if neededDate == "" {
		return "-", "neutral"
	}

	deadline, err := time.ParseInLocation("2006-01-02", neededDate, time.Local)
	if err != nil {
		return "-", "neutral"
	}

	now := time.Now()
	diff := deadline.Sub(now)
	if diff < 0 {
		return "Overdue", "overdue"
	}

	hours := int(math.Ceil(diff.Hours()))
	if hours <= 24 {
		if hours <= 0 {
			return "Overdue", "overdue"
		}
		return fmt.Sprintf("%dh left", hours), "warning"
	}

	days := hours / 24
	remainingHours := hours % 24
	if remainingHours == 0 {
		return fmt.Sprintf("%dd left", days), "normal"
	}
	return fmt.Sprintf("%dd %dh left", days, remainingHours), "normal"
}

func formatQty(value float64) string {
	if math.Mod(value, 1) == 0 {
		return fmt.Sprintf("%.0f", value)
	}
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.2f", value), "0"), ".")
}

func formatApprovalTaskStatusLabel(status string) string {
	switch status {
	case "WAITING":
		return "Pending Approval"
	case "APPROVED":
		return "Verified"
	case "REJECTED":
		return "Rejected"
	case "SKIPPED":
		return "Skipped"
	default:
		return status
	}
}

func buildBudgetImpactLabel(divisionName, glName string) string {
	if strings.TrimSpace(divisionName) != "" && strings.TrimSpace(glName) != "" {
		return divisionName + " " + glName + " Budget"
	}
	if strings.TrimSpace(glName) != "" {
		return glName + " Budget"
	}
	if strings.TrimSpace(divisionName) != "" {
		return divisionName + " Budget"
	}
	return "Purchase Request Budget"
}
