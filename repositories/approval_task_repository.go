package repositories

import (
	"database/sql"
	"errors"
	"fmt"
	"gobase-app/models"
	"strings"
)

type ApprovalTaskRepository struct {
	DB *sql.DB
}

type lockedApprovalTask struct {
	ID             int64
	ApprovalID     int64
	RefType        string
	RefID          int64
	StepOrder      int
	RoleID         int64
	AssignedUserID int
	Status         string
	ApprovalStatus string
}

func (r *ApprovalTaskRepository) GetInboxByUser(filter models.ApprovalTaskInboxFilter) (*models.ApprovalTaskInboxResult, error) {
	whereSQL, args := approvalTaskInboxWhere(filter)

	var totalRows int
	countArgs := append([]interface{}{}, args...)
	if err := r.DB.QueryRow(`
		SELECT COUNT(1)
		FROM approval_tasks at
		JOIN approvals a ON a.id = at.approval_id
		LEFT JOIN purchase_requests pr ON a.ref_type = 'PR' AND pr.id = a.ref_id
		LEFT JOIN users u ON u.id = pr.requester_user_id
		LEFT JOIN stores s ON s.store_id = pr.store_id
		LEFT JOIN roles r ON r.id = at.role_id
	`+whereSQL, countArgs...).Scan(&totalRows); err != nil {
		return nil, err
	}

	var queueValue float64
	sumArgs := append([]interface{}{}, args...)
	if err := r.DB.QueryRow(`
		SELECT COALESCE(SUM(pr.total_amount), 0)
		FROM approval_tasks at
		JOIN approvals a ON a.id = at.approval_id
		LEFT JOIN purchase_requests pr ON a.ref_type = 'PR' AND pr.id = a.ref_id
		LEFT JOIN users u ON u.id = pr.requester_user_id
		LEFT JOIN stores s ON s.store_id = pr.store_id
		LEFT JOIN roles r ON r.id = at.role_id
	`+whereSQL, sumArgs...).Scan(&queueValue); err != nil {
		return nil, err
	}

	orderSQL := " ORDER BY at.created_at ASC, at.id ASC"
	if filter.NeededDateSort == "asc" {
		orderSQL = " ORDER BY pr.needed_date IS NULL ASC, pr.needed_date ASC, at.created_at ASC, at.id ASC"
	} else if filter.NeededDateSort == "desc" {
		orderSQL = " ORDER BY pr.needed_date IS NULL ASC, pr.needed_date DESC, at.created_at ASC, at.id ASC"
	}

	offset := (filter.Page - 1) * filter.PerPage
	queryArgs := append([]interface{}{}, args...)
	queryArgs = append(queryArgs, filter.PerPage, offset)

	rows, err := r.DB.Query(`
		SELECT
			at.id,
			at.approval_id,
			a.ref_type,
			a.ref_id,
			CASE
				WHEN a.ref_type = 'PR' THEN COALESCE(pr.pr_number, '')
				ELSE CONCAT(a.ref_type, '-', a.ref_id)
			END AS document_number,
			COALESCE(u.name, '') AS requester_name,
			COALESCE(s.store_name, '') AS store_name,
			COALESCE(r.name, '') AS role_name,
			at.step_order,
			at.scope,
			COALESCE(pr.total_amount, 0) AS total_amount,
			COALESCE(pr.spend_type, '') AS spend_type,
			COALESCE(pr.urgent_level, '') AS urgent_level,
			pr.needed_date,
			at.status,
			at.created_at
		FROM approval_tasks at
		JOIN approvals a ON a.id = at.approval_id
		LEFT JOIN purchase_requests pr ON a.ref_type = 'PR' AND pr.id = a.ref_id
		LEFT JOIN users u ON u.id = pr.requester_user_id
		LEFT JOIN stores s ON s.store_id = pr.store_id
		LEFT JOIN roles r ON r.id = at.role_id
	`+whereSQL+orderSQL+`
		LIMIT ? OFFSET ?
	`, queryArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.ApprovalTaskInboxItem
	for rows.Next() {
		var (
			item        models.ApprovalTaskInboxItem
			totalAmount float64
			createdAt   sql.NullTime
			neededDate  sql.NullTime
		)
		if err := rows.Scan(
			&item.ID,
			&item.ApprovalID,
			&item.RefType,
			&item.RefID,
			&item.DocumentNumber,
			&item.RequesterName,
			&item.StoreName,
			&item.RoleName,
			&item.StepOrder,
			&item.Scope,
			&totalAmount,
			&item.SpendType,
			&item.UrgentLevel,
			&neededDate,
			&item.Status,
			&createdAt,
		); err != nil {
			return nil, err
		}
		item.Amount = totalAmount
		item.AmountDisplay = formatAmountID(totalAmount)
		if createdAt.Valid {
			item.CreatedAtDisplay = createdAt.Time.Format("02 Jan 2006 15:04")
		}
		if neededDate.Valid {
			item.NeededDate = neededDate.Time.Format("02 Jan 2006")
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &models.ApprovalTaskInboxResult{
		Items:      items,
		TotalRows:  totalRows,
		QueueValue: queueValue,
	}, nil
}

func approvalTaskInboxWhere(filter models.ApprovalTaskInboxFilter) (string, []interface{}) {
	args := []interface{}{filter.UserID}
	conditions := []string{
		"at.assigned_user_id = ?",
		"at.status = 'WAITING'",
		"a.status = 'PENDING'",
		`NOT EXISTS (
			SELECT 1
			FROM approval_tasks prev
			WHERE prev.approval_id = at.approval_id
			AND prev.step_order < at.step_order
			AND prev.status NOT IN ('APPROVED', 'SKIPPED')
		)`,
	}
	if filter.Urgency != "" {
		conditions = append(conditions, "pr.urgent_level = ?")
		args = append(args, filter.Urgency)
	}
	if filter.SpendType != "" {
		conditions = append(conditions, "pr.spend_type = ?")
		args = append(args, filter.SpendType)
	}
	return " WHERE " + strings.Join(conditions, "\n\t\tAND "), args
}

func (r *ApprovalTaskRepository) GetDetailByID(taskID int64, userID int) (*models.ApprovalTaskDetail, error) {
	var (
		item        models.ApprovalTaskDetail
		totalAmount float64
		createdAt   sql.NullTime
		neededDate  sql.NullTime
	)

	err := r.DB.QueryRow(`
		SELECT
			at.id,
			at.approval_id,
			a.ref_type,
			a.ref_id,
			CASE
				WHEN a.ref_type = 'PR' THEN COALESCE(pr.pr_number, '')
				ELSE CONCAT(a.ref_type, '-', a.ref_id)
			END AS document_number,
			COALESCE(u.name, '') AS requester_name,
			COALESCE(s.store_name, '') AS store_name,
			COALESCE(r.name, '') AS role_name,
			at.step_order,
			at.scope,
			COALESCE(pr.total_amount, 0) AS total_amount,
			COALESCE(pr.urgent_level, '') AS urgent_level,
			at.status,
			at.created_at,
			COALESCE(pr.spend_type, '') AS spend_type,
			pr.needed_date,
			COALESCE(pr.justification, '') AS justification,
			a.status AS approval_status,
			COALESCE(pr.status, '') AS document_status
		FROM approval_tasks at
		JOIN approvals a ON a.id = at.approval_id
		LEFT JOIN purchase_requests pr ON a.ref_type = 'PR' AND pr.id = a.ref_id
		LEFT JOIN users u ON u.id = pr.requester_user_id
		LEFT JOIN stores s ON s.store_id = pr.store_id
		LEFT JOIN roles r ON r.id = at.role_id
		WHERE at.id = ?
		AND at.assigned_user_id = ?
		AND at.status = 'WAITING'
		AND a.status = 'PENDING'
		AND NOT EXISTS (
			SELECT 1
			FROM approval_tasks prev
			WHERE prev.approval_id = at.approval_id
			AND prev.step_order < at.step_order
			AND prev.status NOT IN ('APPROVED', 'SKIPPED')
		)
	`, taskID, userID).Scan(
		&item.ID,
		&item.ApprovalID,
		&item.RefType,
		&item.RefID,
		&item.DocumentNumber,
		&item.RequesterName,
		&item.StoreName,
		&item.RoleName,
		&item.StepOrder,
		&item.Scope,
		&totalAmount,
		&item.UrgentLevel,
		&item.Status,
		&createdAt,
		&item.SpendType,
		&neededDate,
		&item.Justification,
		&item.ApprovalStatus,
		&item.DocumentStatus,
	)
	if err != nil {
		return nil, err
	}

	item.AmountDisplay = formatAmountID(totalAmount)
	item.Amount = totalAmount
	if createdAt.Valid {
		item.CreatedAtDisplay = createdAt.Time.Format("02 Jan 2006 15:04")
	}
	if neededDate.Valid {
		item.NeededDate = neededDate.Time.Format("2006-01-02")
	}
	item.CurrentApprovalStep = fmt.Sprintf("Step %d", item.StepOrder)
	item.CurrentApprovalStatus = "Waiting for your approval"

	return &item, nil
}

func (r *ApprovalTaskRepository) Approve(input models.ApprovalActionInput) error {
	return r.act(input, "APPROVED")
}

func (r *ApprovalTaskRepository) Reject(input models.ApprovalActionInput) error {
	return r.act(input, "REJECTED")
}

func (r *ApprovalTaskRepository) act(input models.ApprovalActionInput, taskStatus string) error {
	if input.TaskID <= 0 {
		return errors.New("task approval tidak valid")
	}
	if input.ActorUserID <= 0 {
		return errors.New("user approval tidak valid")
	}
	if taskStatus != "APPROVED" && taskStatus != "REJECTED" {
		return errors.New("aksi approval tidak valid")
	}

	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	task, err := r.lockTaskTx(tx, input.TaskID)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("task approval tidak ditemukan")
		}
		return err
	}

	if task.AssignedUserID != input.ActorUserID {
		return errors.New("task approval bukan milik user login")
	}
	if task.Status != "WAITING" {
		return fmt.Errorf("task approval sudah berstatus %s", task.Status)
	}
	if task.ApprovalStatus != "PENDING" {
		return fmt.Errorf("approval sudah berstatus %s", task.ApprovalStatus)
	}

	ready, err := r.previousStepsApprovedTx(tx, task.ApprovalID, task.StepOrder)
	if err != nil {
		return err
	}
	if !ready {
		return errors.New("step approval sebelumnya belum selesai")
	}

	comment := strings.TrimSpace(input.Comment)
	if _, err := tx.Exec(`
		UPDATE approval_tasks
		SET status = ?, comment = ?, acted_at = NOW()
		WHERE id = ?
	`, taskStatus, nullableString(comment), task.ID); err != nil {
		return err
	}

	historyAction := "APPROVED"
	auditAction := "APPROVE"
	if taskStatus == "REJECTED" {
		historyAction = "REJECTED"
		auditAction = "REJECT"
	}

	if _, err := tx.Exec(`
		INSERT INTO approval_histories (approval_id, task_id, ref_type, ref_id, action, actor_user_id, actor_role_id, note)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, task.ApprovalID, task.ID, task.RefType, task.RefID, historyAction, input.ActorUserID, task.RoleID, nullableString(comment)); err != nil {
		return err
	}

	message := fmt.Sprintf("%s task approval step %d", strings.Title(strings.ToLower(auditAction)), task.StepOrder)
	if comment != "" {
		message += ": " + comment
	}

	if err := insertAuditLogTx(tx, "APPROVAL", task.ApprovalID, auditAction, message, input.AuditContext); err != nil {
		return err
	}
	if err := insertAuditLogTx(tx, task.RefType, task.RefID, auditAction, message, input.AuditContext); err != nil {
		return err
	}

	if taskStatus == "REJECTED" {
		if err := r.rejectApprovalTx(tx, task); err != nil {
			return err
		}
	} else {
		done, err := r.allApprovalTasksDoneTx(tx, task.ApprovalID)
		if err != nil {
			return err
		}
		if done {
			if err := r.approveApprovalTx(tx, task); err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

func (r *ApprovalTaskRepository) lockTaskTx(tx *sql.Tx, taskID int64) (*lockedApprovalTask, error) {
	var task lockedApprovalTask
	err := tx.QueryRow(`
		SELECT
			at.id,
			at.approval_id,
			a.ref_type,
			a.ref_id,
			at.step_order,
			at.role_id,
			at.assigned_user_id,
			at.status,
			a.status
		FROM approval_tasks at
		JOIN approvals a ON a.id = at.approval_id
		WHERE at.id = ?
		FOR UPDATE
	`, taskID).Scan(
		&task.ID,
		&task.ApprovalID,
		&task.RefType,
		&task.RefID,
		&task.StepOrder,
		&task.RoleID,
		&task.AssignedUserID,
		&task.Status,
		&task.ApprovalStatus,
	)
	if err != nil {
		return nil, err
	}

	var lockedID int64
	if err := tx.QueryRow(`SELECT id FROM approvals WHERE id = ? FOR UPDATE`, task.ApprovalID).Scan(&lockedID); err != nil {
		return nil, err
	}
	if task.RefType == "PR" {
		if err := tx.QueryRow(`SELECT id FROM purchase_requests WHERE id = ? FOR UPDATE`, task.RefID).Scan(&lockedID); err != nil {
			return nil, err
		}
	}

	return &task, nil
}

func (r *ApprovalTaskRepository) previousStepsApprovedTx(tx *sql.Tx, approvalID int64, stepOrder int) (bool, error) {
	var count int
	err := tx.QueryRow(`
		SELECT COUNT(1)
		FROM approval_tasks
		WHERE approval_id = ?
		AND step_order < ?
		AND status NOT IN ('APPROVED', 'SKIPPED')
	`, approvalID, stepOrder).Scan(&count)
	return count == 0, err
}

func (r *ApprovalTaskRepository) allApprovalTasksDoneTx(tx *sql.Tx, approvalID int64) (bool, error) {
	var count int
	err := tx.QueryRow(`
		SELECT COUNT(1)
		FROM approval_tasks
		WHERE approval_id = ?
		AND status NOT IN ('APPROVED', 'SKIPPED')
	`, approvalID).Scan(&count)
	return count == 0, err
}

func (r *ApprovalTaskRepository) approveApprovalTx(tx *sql.Tx, task *lockedApprovalTask) error {
	if _, err := tx.Exec(`UPDATE approvals SET status = 'APPROVED' WHERE id = ?`, task.ApprovalID); err != nil {
		return err
	}
	if task.RefType == "PR" {
		if _, err := tx.Exec(`UPDATE purchase_requests SET status = 'APPROVED' WHERE id = ?`, task.RefID); err != nil {
			return err
		}
	}
	return nil
}

func (r *ApprovalTaskRepository) rejectApprovalTx(tx *sql.Tx, task *lockedApprovalTask) error {
	if _, err := tx.Exec(`
		UPDATE approval_tasks
		SET status = 'SKIPPED'
		WHERE approval_id = ?
		AND id <> ?
		AND status = 'WAITING'
	`, task.ApprovalID, task.ID); err != nil {
		return err
	}
	if _, err := tx.Exec(`UPDATE approvals SET status = 'REJECTED' WHERE id = ?`, task.ApprovalID); err != nil {
		return err
	}
	if task.RefType == "PR" {
		if _, err := tx.Exec(`UPDATE purchase_requests SET status = 'REJECTED' WHERE id = ?`, task.RefID); err != nil {
			return err
		}
	}
	return nil
}
