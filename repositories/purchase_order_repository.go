package repositories

import (
	"database/sql"
	"fmt"
	"gobase-app/models"
	"math"
	"strconv"
	"strings"
	"time"
)

type PurchaseOrderRepository struct {
	DB *sql.DB
}

func (r *PurchaseOrderRepository) GetAll() ([]models.PurchaseOrder, error) {
	rows, err := r.DB.Query(`
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
		ORDER BY po.id DESC
	`)
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
		po.TotalAmountDisplay = formatAmountIDLocal(po.TotalAmount)
		po.StatusLabel = formatPOStatusLabel(po.Status)
		orders = append(orders, po)
	}

	return orders, rows.Err()
}

func (r *PurchaseOrderRepository) GetApprovedPRReadyForPO() ([]models.ApprovedPRForPO, error) {
	rows, err := r.DB.Query(`
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
		ORDER BY pr.updated_at DESC, pr.id DESC
	`)
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
		item.TotalAmountDisplay = formatAmountIDLocal(item.TotalAmount)
		items = append(items, item)
	}

	return items, rows.Err()
}

func (r *PurchaseOrderRepository) GetCreateFormByPRID(prID int64, userID int) (*models.PurchaseOrderCreateForm, error) {
	prRepo := &PurchaseRequestRepository{DB: r.DB}
	pr, err := prRepo.GetDetailByID(prID, userID)
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

func (r *PurchaseOrderRepository) GetDetailByID(id int64) (*models.PurchaseOrderDetail, error) {
	var (
		po        models.PurchaseOrderDetail
		createdAt sql.NullTime
	)
	err := r.DB.QueryRow(`
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
	`, id).Scan(
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
	po.TotalAmountDisplay = formatAmountIDLocal(po.TotalAmount)
	po.StatusLabel = formatPOStatusLabel(po.Status)

	items, err := r.getItemsByPOID(id)
	if err != nil {
		return nil, err
	}
	po.Items = items

	return &po, nil
}

func (r *PurchaseOrderRepository) CreateFromPR(input models.PurchaseOrderCreateInput, totalAmount float64) (int64, error) {
	tx, err := r.DB.Begin()
	if err != nil {
		return 0, err
	}

	var (
		status     string
		storeID    int
		divisionID sql.NullInt64
		prNumber   string
	)
	err = tx.QueryRow(`
		SELECT status, store_id, division_id, pr_number
		FROM purchase_requests
		WHERE id = ?
		FOR UPDATE
	`, input.PRID).Scan(&status, &storeID, &divisionID, &prNumber)
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

	if _, err := tx.Exec(`UPDATE purchase_requests SET status = 'CONVERTED_TO_PO' WHERE id = ?`, input.PRID); err != nil {
		tx.Rollback()
		return 0, err
	}

	if err := insertAuditLogTx(tx, "PO", poID, "CREATE_FROM_PR", fmt.Sprintf("PO %s dibuat dari PR %s", poNumber, prNumber), input.AuditContext); err != nil {
		tx.Rollback()
		return 0, err
	}
	if err := insertAuditLogTx(tx, "PR", input.PRID, "CONVERT_TO_PO", fmt.Sprintf("PR dikonversi menjadi PO %s", poNumber), input.AuditContext); err != nil {
		tx.Rollback()
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return 0, err
	}

	return poID, nil
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
		item.UnitPriceDisplay = formatAmountIDLocal(item.UnitPrice)
		item.TotalDisplay = formatAmountIDLocal(item.Total)
		items = append(items, item)
	}

	return items, rows.Err()
}

func (r *PurchaseOrderRepository) nextPONumberTx(tx *sql.Tx, storeID int) (string, error) {
	var storeCode string
	if err := tx.QueryRow(`SELECT store_code FROM stores WHERE store_id = ?`, storeID).Scan(&storeCode); err != nil {
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
		total := item.Qty * item.UnitPrice
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

func formatAmountIDLocal(value float64) string {
	rounded := int64(math.Round(value))
	raw := strconv.FormatInt(rounded, 10)
	var parts []string
	for len(raw) > 3 {
		parts = append([]string{raw[len(raw)-3:]}, parts...)
		raw = raw[:len(raw)-3]
	}
	if raw != "" {
		parts = append([]string{raw}, parts...)
	}
	return "IDR " + strings.Join(parts, ",")
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
