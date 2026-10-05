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

type PurchaseOrderRepository struct {
	DB *sql.DB
}

func (r *PurchaseOrderRepository) GetAll(scope models.AccessScope) ([]models.PurchaseOrder, error) {
	query := `
		SELECT
			po.id,
			po.po_number,
			COALESCE(po.pr_id, 0),
			COALESCE(pr.pr_number, ''),
			po.vendor_id,
			COALESCE(v.name, ''),
			COALESCE(po.store_id, 0),
			COALESCE(s.store_code, ''),
			COALESCE(s.store_name, ''),
			COALESCE(po.division_id, 0),
			COALESCE(d.division_name, ''),
			po.total_amount,
			po.status,
			po.created_at
		FROM purchase_orders po
		LEFT JOIN purchase_requests pr ON pr.id = po.pr_id
		LEFT JOIN vendors v ON v.id = po.vendor_id
		LEFT JOIN stores s ON s.store_id = po.store_id
		LEFT JOIN divisions d ON d.id = po.division_id
	`
	args := make([]interface{}, 0)
	if !scope.CanViewAll {
		predicate, predicateArgs := purchaseOrderScopePredicate("po", scope)
		query += " WHERE " + predicate
		args = append(args, predicateArgs...)
	}
	query += " ORDER BY po.id DESC"
	rows, err := r.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []models.PurchaseOrder
	for rows.Next() {
		var (
			po        models.PurchaseOrder
			createdAt sql.NullTime
		)
		if err := rows.Scan(
			&po.ID,
			&po.PONumber,
			&po.PRID,
			&po.PRNumber,
			&po.VendorID,
			&po.VendorName,
			&po.StoreID,
			&po.StoreCode,
			&po.StoreName,
			&po.DivisionID,
			&po.DivisionName,
			&po.TotalAmount,
			&po.Status,
			&createdAt,
		); err != nil {
			return nil, err
		}
		if createdAt.Valid {
			po.CreatedAtDisplay = createdAt.Time.Format("02 Jan 2006 15:04")
		}
		po.TotalAmountDisplay = formatPOMoney(po.TotalAmount)
		po.StatusLabel = formatPOStatusLabel(po.Status)
		orders = append(orders, po)
	}

	return orders, rows.Err()
}

func (r *PurchaseOrderRepository) GetApprovedPRReadyForPO(scope models.AccessScope) ([]models.ApprovedPRForPO, error) {
	query := `
		SELECT
			pr.id,
			pr.pr_number,
			COALESCE(u.name, ''),
			COALESCE(s.store_name, ''),
			COALESCE(d.division_name, ''),
			COALESCE(ga.gl_name, ''),
			pr.spend_type,
			pr.total_amount,
			pr.updated_at
		FROM purchase_requests pr
		LEFT JOIN users u ON u.id = pr.requester_user_id
		LEFT JOIN stores s ON s.store_id = pr.store_id
		LEFT JOIN divisions d ON d.id = pr.division_id
		LEFT JOIN gl_accounts ga ON ga.id = pr.gl_account_id
		WHERE pr.status = 'APPROVED'
		AND NOT EXISTS (
			SELECT 1 FROM purchase_orders po WHERE po.pr_id = pr.id
		)
	`
	args := make([]interface{}, 0)
	if !scope.CanViewAll {
		predicate, predicateArgs := purchaseRequestScopePredicate("pr", scope, false)
		query += " AND " + predicate
		args = append(args, predicateArgs...)
	}
	query += " ORDER BY pr.updated_at DESC, pr.id DESC"
	rows, err := r.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.ApprovedPRForPO
	for rows.Next() {
		var (
			item      models.ApprovedPRForPO
			updatedAt sql.NullTime
		)
		if err := rows.Scan(
			&item.ID,
			&item.PRNumber,
			&item.RequesterName,
			&item.StoreName,
			&item.DivisionName,
			&item.GLAccountName,
			&item.SpendType,
			&item.TotalAmount,
			&updatedAt,
		); err != nil {
			return nil, err
		}
		if updatedAt.Valid {
			item.ApprovedAtDisplay = updatedAt.Time.Format("02 Jan 2006 15:04")
		}
		item.TotalAmountDisplay = formatPOMoney(item.TotalAmount)
		items = append(items, item)
	}

	return items, rows.Err()
}

func (r *PurchaseOrderRepository) GetCreateFormByPRID(prID int64, scope models.AccessScope) (*models.PurchaseOrderCreateForm, error) {
	prRepo := &PurchaseRequestRepository{DB: r.DB}
	pr, err := prRepo.GetDetailByID(prID, scope)
	if err != nil {
		return nil, err
	}
	if pr.Status != "APPROVED" {
		return nil, fmt.Errorf("PR %s belum approved", pr.PRNumber)
	}

	var existing int
	if err := r.DB.QueryRow(`SELECT COUNT(1) FROM purchase_orders WHERE pr_id = ?`, prID).Scan(&existing); err != nil {
		return nil, err
	}
	if existing > 0 {
		return nil, fmt.Errorf("PR %s sudah dibuatkan PO", pr.PRNumber)
	}

	vendorRepo := &VendorRepository{DB: r.DB}
	vendors, err := vendorRepo.GetAll()
	if err != nil {
		return nil, err
	}

	activeVendors := make([]models.Vendor, 0, len(vendors))
	for _, vendor := range vendors {
		if vendor.IsActive {
			activeVendors = append(activeVendors, vendor)
		}
	}

	return &models.PurchaseOrderCreateForm{
		PR:      *pr,
		Vendors: activeVendors,
	}, nil
}

func (r *PurchaseOrderRepository) GetDetailByID(id int64, scope models.AccessScope) (*models.PurchaseOrderDetail, error) {
	var (
		po        models.PurchaseOrderDetail
		createdAt sql.NullTime
	)
	query := `
		SELECT
			po.id,
			po.po_number,
			COALESCE(po.pr_id, 0),
			COALESCE(pr.pr_number, ''),
			po.vendor_id,
			COALESCE(v.name, ''),
			COALESCE(po.store_id, 0),
			COALESCE(s.store_code, ''),
			COALESCE(s.store_name, ''),
			COALESCE(po.division_id, 0),
			COALESCE(d.division_name, ''),
			po.total_amount,
			po.status,
			po.created_at
		FROM purchase_orders po
		LEFT JOIN purchase_requests pr ON pr.id = po.pr_id
		LEFT JOIN vendors v ON v.id = po.vendor_id
		LEFT JOIN stores s ON s.store_id = po.store_id
		LEFT JOIN divisions d ON d.id = po.division_id
		WHERE po.id = ?
	`
	args := []interface{}{id}
	if !scope.CanViewAll {
		predicate, predicateArgs := purchaseOrderScopePredicate("po", scope)
		query += " AND " + predicate
		args = append(args, predicateArgs...)
	}
	err := r.DB.QueryRow(query, args...).Scan(
		&po.ID,
		&po.PONumber,
		&po.PRID,
		&po.PRNumber,
		&po.VendorID,
		&po.VendorName,
		&po.StoreID,
		&po.StoreCode,
		&po.StoreName,
		&po.DivisionID,
		&po.DivisionName,
		&po.TotalAmount,
		&po.Status,
		&createdAt,
	)
	if err != nil {
		return nil, err
	}
	if createdAt.Valid {
		po.CreatedAtDisplay = createdAt.Time.Format("02 Jan 2006 15:04")
	}
	po.TotalAmountDisplay = formatPOMoney(po.TotalAmount)
	po.StatusLabel = formatPOStatusLabel(po.Status)

	items, err := r.getItemsByPOID(id)
	if err != nil {
		return nil, err
	}
	po.Items = items

	return &po, nil
}

func (r *PurchaseOrderRepository) CreateFromPR(input models.PurchaseOrderCreateInput, totalAmount models.Money) (int64, error) {
	tx, err := r.DB.Begin()
	if err != nil {
		return 0, err
	}

	var (
		status          string
		storeID         int
		divisionID      sql.NullInt64
		prNumber        string
		requesterID     int
		glAccountID     int
		neededDate      sql.NullTime
		prAmount        models.Money
		budgetException sql.NullString
	)
	err = tx.QueryRow(`
		SELECT pr.status, pr.store_id, pr.division_id, pr.pr_number, pr.requester_user_id,
		       pr.gl_account_id, pr.needed_date, pr.total_amount, pr.budget_exception_reason
		FROM purchase_requests pr
		WHERE pr.id = ?
		FOR UPDATE
	`, input.PRID).Scan(&status, &storeID, &divisionID, &prNumber, &requesterID,
		&glAccountID, &neededDate, &prAmount, &budgetException)
	if err == sql.ErrNoRows {
		tx.Rollback()
		return 0, fmt.Errorf("purchase request tidak ditemukan")
	}
	if err != nil {
		tx.Rollback()
		return 0, err
	}
	if status != "APPROVED" {
		tx.Rollback()
		return 0, fmt.Errorf("PR %s harus APPROVED sebelum dibuat PO", prNumber)
	}
	if !input.AccessScope.AllowsOwnerOrStore(requesterID, storeID) {
		tx.Rollback()
		return 0, fmt.Errorf("purchase request tidak ditemukan atau akses ditolak")
	}
	if models.ExceedsVariance(totalAmount, prAmount, input.VarianceToleranceBPS) {
		tx.Rollback()
		return 0, fmt.Errorf("nilai PO %s melebihi toleransi terhadap nilai PR %s; lakukan approval ulang", totalAmount.FormatIDR(), prAmount.FormatIDR())
	}

	var existing int
	if err := tx.QueryRow(`SELECT COUNT(1) FROM purchase_orders WHERE pr_id = ?`, input.PRID).Scan(&existing); err != nil {
		tx.Rollback()
		return 0, err
	}
	if existing > 0 {
		tx.Rollback()
		return 0, fmt.Errorf("PR %s sudah memiliki PO", prNumber)
	}

	var vendorActive int
	if err := tx.QueryRow(`SELECT is_active FROM vendors WHERE id = ?`, input.VendorID).Scan(&vendorActive); err != nil {
		tx.Rollback()
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("vendor tidak ditemukan")
		}
		return 0, err
	}
	if vendorActive != 1 {
		tx.Rollback()
		return 0, fmt.Errorf("vendor tidak aktif")
	}

	budgetID, budgetExceptionUsed, err := lockAndValidatePOBudgetTx(
		tx, storeID, divisionID, glAccountID, neededDate, totalAmount, budgetException.String,
	)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	poNumber, err := r.nextPONumberTx(tx, storeID)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	res, err := tx.Exec(`
		INSERT INTO purchase_orders (po_number, pr_id, vendor_id, store_id, division_id, total_amount, status)
		VALUES (?, ?, ?, ?, ?, ?, 'APPROVED')
	`, poNumber, input.PRID, input.VendorID, storeID, nullableSQLInt64(divisionID), totalAmount)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	poID, err := res.LastInsertId()
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	if err := insertPOItemsTx(tx, poID, input.Items); err != nil {
		tx.Rollback()
		return 0, err
	}
	if budgetID > 0 {
		if _, err := tx.Exec(`
			INSERT INTO budget_usages (budget_id, ref_type, ref_id, used_amount, note)
			VALUES (?, 'PO', ?, ?, ?)
		`, budgetID, poID, totalAmount, fmt.Sprintf("Pemakaian budget untuk %s", poNumber)); err != nil {
			tx.Rollback()
			return 0, err
		}
	}

	if _, err := tx.Exec(`UPDATE purchase_requests SET status = 'CONVERTED_TO_PO' WHERE id = ?`, input.PRID); err != nil {
		tx.Rollback()
		return 0, err
	}

	if err := insertAuditLogTx(tx, "PO", poID, "CREATE_FROM_PR", fmt.Sprintf("PO %s dibuat dari PR %s", poNumber, prNumber), input.AuditContext); err != nil {
		tx.Rollback()
		return 0, err
	}
	approvalDetail := fmt.Sprintf("PO %s disetujui", poNumber)
	if budgetID > 0 {
		approvalDetail += fmt.Sprintf(" dan budget dikomit sebesar %s", totalAmount.FormatIDR())
	} else {
		approvalDetail += " tanpa budget menggunakan pengecualian PR"
	}
	if err := insertAuditLogTx(tx, "PO", poID, "APPROVE", approvalDetail, input.AuditContext); err != nil {
		tx.Rollback()
		return 0, err
	}
	if err := insertAuditLogTx(tx, "PR", input.PRID, "CONVERT_TO_PO", fmt.Sprintf("PR dikonversi menjadi PO %s", poNumber), input.AuditContext); err != nil {
		tx.Rollback()
		return 0, err
	}
	if budgetExceptionUsed {
		if err := insertAuditLogTx(tx, "PO", poID, "BUDGET_EXCEPTION", "PO dibuat menggunakan alasan pengecualian budget dari PR", input.AuditContext); err != nil {
			tx.Rollback()
			return 0, err
		}
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return 0, err
	}

	return poID, nil
}

func lockAndValidatePOBudgetTx(tx *sql.Tx, storeID int, divisionID sql.NullInt64, glAccountID int, neededDate sql.NullTime, poAmount models.Money, exceptionReason string) (int64, bool, error) {
	checkDate := time.Now()
	if neededDate.Valid {
		checkDate = neededDate.Time
	}
	monthKey := checkDate.Format("2006-01")
	yearKey := checkDate.Format("2006")
	quarterKey := fmt.Sprintf("%d-Q%d", checkDate.Year(), (int(checkDate.Month())-1)/3+1)

	var budgetID int64
	var budgetAmount models.Money
	err := tx.QueryRow(`
		SELECT b.id, b.amount
		FROM budgets b
		WHERE b.fiscal_year = ?
		  AND b.gl_account_id = ?
		  AND (b.store_id IS NULL OR b.store_id = ?)
		  AND (b.division_id IS NULL OR b.division_id = ?)
		  AND ((b.period_type = 'MONTHLY' AND b.period_key = ?)
		    OR (b.period_type = 'QUARTERLY' AND b.period_key = ?)
		    OR (b.period_type = 'YEARLY' AND b.period_key = ?))
		ORDER BY (b.store_id IS NOT NULL) DESC, (b.division_id IS NOT NULL) DESC,
		  FIELD(b.period_type, 'MONTHLY', 'QUARTERLY', 'YEARLY') ASC, b.id DESC
		LIMIT 1
		FOR UPDATE
	`, checkDate.Year(), glAccountID, storeID, divisionID.Int64, monthKey, quarterKey, yearKey).Scan(&budgetID, &budgetAmount)
	if err == sql.ErrNoRows {
		if strings.TrimSpace(exceptionReason) == "" {
			return 0, false, errors.New("budget final tidak tersedia; PO tidak dapat disetujui")
		}
		return 0, true, nil
	}
	if err != nil {
		return 0, false, err
	}

	var usedAmount models.Money
	if err := tx.QueryRow(`SELECT COALESCE(SUM(used_amount), 0) FROM budget_usages WHERE budget_id = ?`, budgetID).Scan(&usedAmount); err != nil {
		return 0, false, err
	}
	remaining := budgetAmount.Sub(usedAmount)
	if poAmount > remaining {
		if strings.TrimSpace(exceptionReason) == "" {
			return 0, false, fmt.Errorf("budget final tidak mencukupi: sisa %s, nilai PO %s", remaining.FormatIDR(), poAmount.FormatIDR())
		}
		return budgetID, true, nil
	}
	return budgetID, false, nil
}

func (r *PurchaseOrderRepository) getItemsByPOID(poID int64) ([]models.PurchaseOrderItem, error) {
	rows, err := r.DB.Query(`
		SELECT id, po_id, item_name, qty, uom, unit_price, total
		FROM purchase_order_items
		WHERE po_id = ?
		ORDER BY id ASC
	`, poID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.PurchaseOrderItem
	for rows.Next() {
		var item models.PurchaseOrderItem
		if err := rows.Scan(&item.ID, &item.POID, &item.ItemName, &item.Qty, &item.UOM, &item.UnitPrice, &item.Total); err != nil {
			return nil, err
		}
		item.QtyDisplay = formatQtyLocal(item.Qty)
		item.UnitPriceDisplay = formatPOMoney(item.UnitPrice)
		item.TotalDisplay = formatPOMoney(item.Total)
		items = append(items, item)
	}

	return items, rows.Err()
}

func (r *PurchaseOrderRepository) nextPONumberTx(tx *sql.Tx, storeID int) (string, error) {
	var storeCode string
	// Lock the store row so sequence allocation is serialized per store.
	if err := tx.QueryRow(`SELECT store_code FROM stores WHERE store_id = ? FOR UPDATE`, storeID).Scan(&storeCode); err != nil {
		return "", err
	}

	year := time.Now().Year()
	prefix := fmt.Sprintf("PO-%s-%d-", storeCode, year)
	var lastNumber sql.NullString
	if err := tx.QueryRow(`
		SELECT po_number
		FROM purchase_orders
		WHERE po_number LIKE ?
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

func insertPOItemsTx(tx *sql.Tx, poID int64, items []models.PurchaseOrderItemInput) error {
	stmt, err := tx.Prepare(`
		INSERT INTO purchase_order_items (po_id, item_name, qty, uom, unit_price, total)
		VALUES (?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, item := range items {
		total, err := models.MultiplyMoneyByQuantity(item.UnitPrice, item.Qty)
		if err != nil {
			return err
		}
		if _, err := stmt.Exec(poID, item.ItemName, item.Qty, item.UOM, item.UnitPrice, total); err != nil {
			return err
		}
	}
	return nil
}

func nullableSQLInt64(value sql.NullInt64) interface{} {
	if !value.Valid {
		return nil
	}
	return value.Int64
}

func purchaseOrderScopePredicate(alias string, scope models.AccessScope) (string, []interface{}) {
	if len(scope.StoreIDs) == 0 {
		return "1 = 0", nil
	}
	placeholders := make([]string, 0, len(scope.StoreIDs))
	args := make([]interface{}, 0, len(scope.StoreIDs))
	for _, storeID := range scope.StoreIDs {
		if storeID <= 0 {
			continue
		}
		placeholders = append(placeholders, "?")
		args = append(args, storeID)
	}
	if len(placeholders) == 0 {
		return "1 = 0", nil
	}
	return alias + ".store_id IN (" + strings.Join(placeholders, ",") + ")", args
}

func formatPOMoney(value models.Money) string {
	return value.FormatIDR()
}

func formatQtyLocal(value float64) string {
	if math.Mod(value, 1) == 0 {
		return fmt.Sprintf("%.0f", value)
	}
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.2f", value), "0"), ".")
}

func formatPOStatusLabel(status string) string {
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
	case "RECEIVING":
		return "Receiving"
	case "CLOSED":
		return "Closed"
	default:
		return status
	}
}
