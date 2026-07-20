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

type AssetDisposalRepository struct {
	DB *sql.DB
}

func (r *AssetDisposalRepository) GetPostedDisposalByAssetID(assetID int64) (*models.AssetDisposal, error) {
	var item models.AssetDisposal
	var disposalDate time.Time
	var postedAt sql.NullTime
	var disposalValue, acquisitionValue, accumulated, bookValue, gainLoss float64
	err := r.DB.QueryRow(`
		SELECT disposal.id, disposal.disposal_number, disposal.asset_id,
			asset.asset_code, asset.asset_name, COALESCE(asset_type.name, ''),
			disposal.disposal_type_id, type.code, type.name, disposal.disposal_date,
			disposal.disposal_value, COALESCE(disposal.buyer_name, ''),
			COALESCE(disposal.document_reference, ''), disposal.reason, disposal.status,
			processor.name, COALESCE(approver.name, ''), disposal.posted_at,
			COALESCE(disposal.notes, ''), disposal.acquisition_value,
			disposal.accumulated_depreciation, disposal.book_value, disposal.gain_loss_amount
		FROM asset_disposals disposal
		JOIN assets asset ON asset.id = disposal.asset_id
		LEFT JOIN asset_types asset_type ON asset_type.id = asset.asset_type_id
		JOIN asset_disposal_types type ON type.id = disposal.disposal_type_id
		JOIN users processor ON processor.id = disposal.processed_by
		LEFT JOIN users approver ON approver.id = disposal.approved_by
		WHERE disposal.asset_id = ? AND disposal.status = 'POSTED'
		ORDER BY disposal.posted_at DESC, disposal.id DESC
		LIMIT 1
	`, assetID).Scan(
		&item.ID, &item.DisposalNumber, &item.AssetID, &item.AssetCode, &item.AssetName, &item.AssetTypeName,
		&item.DisposalTypeID, &item.DisposalTypeCode, &item.DisposalTypeName, &disposalDate,
		&disposalValue, &item.BuyerName, &item.DocumentReference, &item.Reason, &item.Status,
		&item.ProcessedByName, &item.ApprovedByName, &postedAt, &item.Notes,
		&acquisitionValue, &accumulated, &bookValue, &gainLoss,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	item.DisposalDate = disposalDate.Format("2006-01-02")
	item.DisposalDateDisplay = formatDepreciationDateID(disposalDate, false)
	item.DisposalValueInput = formatNumberInput(disposalValue)
	item.DisposalValueDisplay = formatAssetAmountID(disposalValue)
	item.AcquisitionValueDisplay = formatAssetAmountID(acquisitionValue)
	item.AccumulatedDepreciationDisplay = formatAssetAmountID(accumulated)
	item.BookValueDisplay = formatAssetAmountID(bookValue)
	item.GainLossAmountDisplay = formatAssetAmountID(math.Abs(gainLoss))
	if gainLoss > 0.005 {
		item.GainLossLabel = "Laba"
	} else if gainLoss < -0.005 {
		item.GainLossLabel = "Rugi"
	} else {
		item.GainLossLabel = "Impas"
	}
	if postedAt.Valid {
		item.PostedAtDisplay = formatDepreciationDateID(postedAt.Time, true)
	}
	return &item, nil
}

func (r *AssetDisposalRepository) GetDisposalTypes() ([]models.AssetDisposalType, error) {
	rows, err := r.DB.Query(`
		SELECT type.id, type.code, type.name, COALESCE(type.description, ''), type.is_active,
			COUNT(disposal.id), type.created_at, type.updated_at
		FROM asset_disposal_types type
		LEFT JOIN asset_disposals disposal ON disposal.disposal_type_id = type.id
		GROUP BY type.id, type.code, type.name, type.description, type.is_active, type.created_at, type.updated_at
		ORDER BY type.code
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]models.AssetDisposalType, 0)
	for rows.Next() {
		var item models.AssetDisposalType
		var active int
		var createdAt, updatedAt sql.NullTime
		if err := rows.Scan(&item.ID, &item.Code, &item.Name, &item.Description, &active,
			&item.DisposalCount, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		item.IsActive = active == 1
		item.IsActiveLabel = activeLabel(item.IsActive)
		item.CreatedAtDisplay = formatNullTime(createdAt)
		item.UpdatedAtDisplay = formatNullTime(updatedAt)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *AssetDisposalRepository) SaveDisposalType(input models.AssetDisposalTypeInput) error {
	if input.ID <= 0 {
		_, err := r.DB.Exec(`
			INSERT INTO asset_disposal_types (code, name, description, is_active)
			VALUES (?, ?, ?, ?)
		`, input.Code, input.Name, nullableString(input.Description), boolToInt(input.IsActive))
		return err
	}
	_, err := r.DB.Exec(`
		UPDATE asset_disposal_types
		SET code = ?, name = ?, description = ?, is_active = ?
		WHERE id = ?
	`, input.Code, input.Name, nullableString(input.Description), boolToInt(input.IsActive), input.ID)
	return err
}

func (r *AssetDisposalRepository) DeleteDisposalType(id int64) error {
	var used int
	if err := r.DB.QueryRow(`SELECT COUNT(*) FROM asset_disposals WHERE disposal_type_id = ?`, id).Scan(&used); err != nil {
		return err
	}
	if used > 0 {
		return errors.New("jenis disposal sudah digunakan dan tidak dapat dihapus; nonaktifkan jenis tersebut")
	}
	result, err := r.DB.Exec(`DELETE FROM asset_disposal_types WHERE id = ?`, id)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return errors.New("jenis disposal tidak ditemukan")
	}
	return nil
}

func (r *AssetDisposalRepository) GetDisposals(filter models.AssetDisposalFilter) (models.AssetDisposalResult, error) {
	result := models.AssetDisposalResult{}
	where, args := assetDisposalWhere(filter)
	var ignoredTotalValue float64
	if err := r.DB.QueryRow(`
		SELECT COUNT(*)
		FROM asset_disposals disposal
		JOIN assets asset ON asset.id = disposal.asset_id
		JOIN asset_disposal_types type ON type.id = disposal.disposal_type_id
		WHERE `+where, args...).Scan(&result.TotalRows); err != nil {
		return result, err
	}
	result.TotalPages = 1
	if result.TotalRows > 0 {
		result.TotalPages = (result.TotalRows + filter.PerPage - 1) / filter.PerPage
	}
	if filter.Page > result.TotalPages {
		filter.Page = result.TotalPages
	}
	offset := (filter.Page - 1) * filter.PerPage

	if err := r.DB.QueryRow(`
		SELECT COUNT(*),
			COALESCE(SUM(status = 'DRAFT'), 0),
			COALESCE(SUM(status = 'POSTED'), 0),
			COALESCE(SUM(status = 'CANCELLED'), 0),
			COALESCE(SUM(CASE WHEN status = 'POSTED' THEN disposal_value ELSE 0 END), 0)
		FROM asset_disposals
	`).Scan(&result.Stats.Total, &result.Stats.Draft, &result.Stats.Posted, &result.Stats.Cancelled, &ignoredTotalValue); err != nil {
		return result, err
	}
	var totalValue float64
	if err := r.DB.QueryRow(`SELECT COALESCE(SUM(disposal_value), 0) FROM asset_disposals WHERE status = 'POSTED'`).Scan(&totalValue); err != nil {
		return result, err
	}
	result.Stats.TotalValueDisplay = formatAssetAmountID(totalValue)

	queryArgs := append(append([]any{}, args...), filter.PerPage, offset)
	rows, err := r.DB.Query(`
		SELECT
			disposal.id, disposal.disposal_number, disposal.asset_id,
			asset.asset_code, asset.asset_name, COALESCE(asset_type.name, ''),
			disposal.disposal_type_id, type.code, type.name, disposal.disposal_date,
			disposal.disposal_value, COALESCE(disposal.buyer_name, ''),
			COALESCE(disposal.document_reference, ''), disposal.reason, disposal.status,
			processor.name, COALESCE(approver.name, ''), disposal.posted_at,
			disposal.cancelled_at, COALESCE(canceller.name, ''),
			COALESCE(disposal.cancellation_reason, ''), COALESCE(disposal.notes, ''),
			disposal.acquisition_value, disposal.accumulated_depreciation,
			disposal.book_value, disposal.gain_loss_amount, disposal.created_at
		FROM asset_disposals disposal
		JOIN assets asset ON asset.id = disposal.asset_id
		LEFT JOIN asset_types asset_type ON asset_type.id = asset.asset_type_id
		JOIN asset_disposal_types type ON type.id = disposal.disposal_type_id
		JOIN users processor ON processor.id = disposal.processed_by
		LEFT JOIN users approver ON approver.id = disposal.approved_by
		LEFT JOIN users canceller ON canceller.id = disposal.cancelled_by
		WHERE `+where+`
		ORDER BY disposal.id DESC
		LIMIT ? OFFSET ?
	`, queryArgs...)
	if err != nil {
		return result, err
	}
	defer rows.Close()
	for rows.Next() {
		var item models.AssetDisposal
		var disposalDate time.Time
		var postedAt, cancelledAt sql.NullTime
		var disposalValue, acquisitionValue, accumulated, bookValue, gainLoss float64
		var createdAt sql.NullTime
		if err := rows.Scan(
			&item.ID, &item.DisposalNumber, &item.AssetID,
			&item.AssetCode, &item.AssetName, &item.AssetTypeName,
			&item.DisposalTypeID, &item.DisposalTypeCode, &item.DisposalTypeName, &disposalDate,
			&disposalValue, &item.BuyerName, &item.DocumentReference, &item.Reason, &item.Status,
			&item.ProcessedByName, &item.ApprovedByName, &postedAt, &cancelledAt,
			&item.CancelledByName, &item.CancellationReason, &item.Notes,
			&acquisitionValue, &accumulated, &bookValue, &gainLoss, &createdAt,
		); err != nil {
			return result, err
		}
		item.DisposalDate = disposalDate.Format("2006-01-02")
		item.DisposalDateDisplay = formatDepreciationDateID(disposalDate, false)
		item.DisposalValueInput = formatNumberInput(disposalValue)
		item.DisposalValueDisplay = formatAssetAmountID(disposalValue)
		item.AcquisitionValueDisplay = formatAssetAmountID(acquisitionValue)
		item.AccumulatedDepreciationDisplay = formatAssetAmountID(accumulated)
		item.BookValueDisplay = formatAssetAmountID(bookValue)
		item.GainLossAmountDisplay = formatAssetAmountID(math.Abs(gainLoss))
		if gainLoss > 0.005 {
			item.GainLossLabel = "Laba"
		} else if gainLoss < -0.005 {
			item.GainLossLabel = "Rugi"
		} else {
			item.GainLossLabel = "Impas"
		}
		if postedAt.Valid {
			item.PostedAtDisplay = formatDepreciationDateID(postedAt.Time, true)
		}
		if cancelledAt.Valid {
			item.CancelledAtDisplay = formatDepreciationDateID(cancelledAt.Time, true)
		}
		item.CreatedAtDisplay = formatNullTime(createdAt)
		result.Items = append(result.Items, item)
	}
	return result, rows.Err()
}

func assetDisposalWhere(filter models.AssetDisposalFilter) (string, []any) {
	clauses := []string{"1 = 1"}
	args := make([]any, 0)
	if filter.Status != "" && filter.Status != "ALL" {
		clauses = append(clauses, "disposal.status = ?")
		args = append(args, filter.Status)
	}
	if filter.Search != "" {
		term := "%" + filter.Search + "%"
		clauses = append(clauses, "(disposal.disposal_number LIKE ? OR asset.asset_code LIKE ? OR asset.asset_name LIKE ? OR type.name LIKE ?)")
		args = append(args, term, term, term, term)
	}
	return strings.Join(clauses, " AND "), args
}

func (r *AssetDisposalRepository) GetDisposalAssetOptions() ([]models.AssetDisposalAssetOption, error) {
	rows, err := r.DB.Query(`
		SELECT asset.id, asset.asset_code, asset.asset_name, COALESCE(asset_type.name, ''),
			asset.status, asset.acquisition_date, asset.acquisition_value,
			COALESCE(profile.status, ''),
			GREATEST(COALESCE(profile.salvage_value, 0),
				COALESCE(profile.depreciable_basis, asset.acquisition_value) - COALESCE(posted.amount, 0))
		FROM assets asset
		LEFT JOIN asset_types asset_type ON asset_type.id = asset.asset_type_id
		LEFT JOIN asset_depreciation_profiles profile ON profile.asset_id = asset.id
		LEFT JOIN (
			SELECT profile_id, SUM(depreciation_amount) amount
			FROM asset_depreciation_schedules
			WHERE status = 'POSTED'
			GROUP BY profile_id
		) posted ON posted.profile_id = profile.id
		WHERE asset.status <> 'DISPOSED'
		  AND NOT EXISTS (
			SELECT 1 FROM asset_disposals disposal
			WHERE disposal.asset_id = asset.id AND disposal.status = 'POSTED'
		  )
		ORDER BY asset.asset_code
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]models.AssetDisposalAssetOption, 0)
	for rows.Next() {
		var item models.AssetDisposalAssetOption
		var acquisitionDate sql.NullTime
		var acquisitionValue, bookValue float64
		if err := rows.Scan(&item.ID, &item.AssetCode, &item.AssetName, &item.AssetTypeName,
			&item.AssetStatus, &acquisitionDate, &acquisitionValue, &item.ProfileStatus, &bookValue); err != nil {
			return nil, err
		}
		if acquisitionDate.Valid {
			item.AcquisitionDate = acquisitionDate.Time.Format("2006-01-02")
		}
		item.AcquisitionValueInput = formatNumberInput(acquisitionValue)
		item.CurrentBookValueInput = formatNumberInput(bookValue)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *AssetDisposalRepository) SaveDisposal(input models.AssetDisposalInput) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var typeCode string
	if err := tx.QueryRow(`SELECT code FROM asset_disposal_types WHERE id = ? AND is_active = 1`, input.DisposalTypeID).Scan(&typeCode); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("jenis disposal tidak ditemukan atau tidak aktif")
		}
		return err
	}
	if typeCode == "SOLD" && input.BuyerName == "" {
		return errors.New("nama pembeli wajib diisi untuk disposal penjualan")
	}
	var assetCode, assetStatus string
	var acquisitionDate sql.NullTime
	if err := tx.QueryRow(`SELECT asset_code, status, acquisition_date FROM assets WHERE id = ? FOR UPDATE`, input.AssetID).Scan(&assetCode, &assetStatus, &acquisitionDate); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("aset tidak ditemukan")
		}
		return err
	}
	if assetStatus == "DISPOSED" {
		return errors.New("aset sudah berstatus DISPOSED")
	}
	if acquisitionDate.Valid && input.DisposalDate < acquisitionDate.Time.Format("2006-01-02") {
		return errors.New("tanggal disposal tidak boleh sebelum tanggal perolehan aset")
	}
	var duplicateCount int
	duplicateQuery := `SELECT COUNT(*) FROM asset_disposals WHERE asset_id = ? AND status IN ('DRAFT', 'POSTED')`
	duplicateArgs := []any{input.AssetID}
	if input.ID > 0 {
		duplicateQuery += " AND id <> ?"
		duplicateArgs = append(duplicateArgs, input.ID)
	}
	if err := tx.QueryRow(duplicateQuery, duplicateArgs...).Scan(&duplicateCount); err != nil {
		return err
	}
	if duplicateCount > 0 {
		return errors.New("aset sudah memiliki transaksi disposal aktif")
	}

	if input.ID <= 0 {
		temporaryNumber := "TMP-" + strings.ToUpper(strings.ReplaceAll(fmt.Sprintf("%d-%d", time.Now().UnixNano(), input.AssetID), "-", ""))
		result, err := tx.Exec(`
			INSERT INTO asset_disposals (
				disposal_number, asset_id, disposal_type_id, disposal_date, disposal_value,
				buyer_name, document_reference, reason, status, processed_by, notes
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, 'DRAFT', ?, ?)
		`, temporaryNumber, input.AssetID, input.DisposalTypeID, input.DisposalDate, input.DisposalValue,
			nullableString(input.BuyerName), nullableString(input.DocumentReference), input.Reason,
			input.AuditContext.ActorUserID, nullableString(input.Notes))
		if err != nil {
			return err
		}
		disposalID, err := result.LastInsertId()
		if err != nil {
			return err
		}
		disposalNumber := fmt.Sprintf("DSP-%s-%06d", input.DisposalDate[:4], disposalID)
		if _, err := tx.Exec(`UPDATE asset_disposals SET disposal_number = ? WHERE id = ?`, disposalNumber, disposalID); err != nil {
			return err
		}
		if err := insertAuditLogTx(tx, "ASSET_DISPOSAL", disposalID, "CREATE_DRAFT",
			fmt.Sprintf("Draft disposal %s dibuat untuk aset %s", disposalNumber, assetCode), input.AuditContext); err != nil {
			return err
		}
		return tx.Commit()
	}

	var status string
	if err := tx.QueryRow(`SELECT status FROM asset_disposals WHERE id = ? AND asset_id = ? FOR UPDATE`, input.ID, input.AssetID).Scan(&status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("transaksi disposal tidak ditemukan")
		}
		return err
	}
	if status != "DRAFT" {
		return errors.New("hanya disposal DRAFT yang dapat diubah")
	}
	if _, err := tx.Exec(`
		UPDATE asset_disposals
		SET disposal_type_id = ?, disposal_date = ?, disposal_value = ?, buyer_name = ?,
			document_reference = ?, reason = ?, notes = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND status = 'DRAFT'
	`, input.DisposalTypeID, input.DisposalDate, input.DisposalValue, nullableString(input.BuyerName),
		nullableString(input.DocumentReference), input.Reason, nullableString(input.Notes), input.ID); err != nil {
		return err
	}
	if err := insertAuditLogTx(tx, "ASSET_DISPOSAL", input.ID, "UPDATE_DRAFT",
		fmt.Sprintf("Draft disposal aset %s diperbarui", assetCode), input.AuditContext); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *AssetDisposalRepository) PostDisposal(id int64, auditCtx models.AuditContext) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var assetID int64
	var disposalNumber, disposalStatus, disposalReason, typeCode string
	var disposalDate time.Time
	var disposalValue float64
	if err := tx.QueryRow(`
		SELECT disposal.asset_id, disposal.disposal_number, disposal.status, disposal.reason,
			disposal.disposal_date, disposal.disposal_value, type.code
		FROM asset_disposals disposal
		JOIN asset_disposal_types type ON type.id = disposal.disposal_type_id
		WHERE disposal.id = ?
		FOR UPDATE
	`, id).Scan(&assetID, &disposalNumber, &disposalStatus, &disposalReason, &disposalDate, &disposalValue, &typeCode); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("transaksi disposal tidak ditemukan")
		}
		return err
	}
	if disposalStatus != "DRAFT" {
		return errors.New("hanya disposal DRAFT yang dapat diposting")
	}
	if disposalDate.After(time.Now()) {
		return errors.New("disposal dengan tanggal masa depan belum dapat diposting")
	}

	var assetCode, assetStatus string
	var acquisitionValue float64
	if err := tx.QueryRow(`SELECT asset_code, status, acquisition_value FROM assets WHERE id = ? FOR UPDATE`, assetID).Scan(&assetCode, &assetStatus, &acquisitionValue); err != nil {
		return err
	}
	if assetStatus == "DISPOSED" {
		return errors.New("aset sudah berstatus DISPOSED")
	}

	var profileID sql.NullInt64
	var profileStatus, lastPolicyCode sql.NullString
	var basis, salvage sql.NullFloat64
	err = tx.QueryRow(`
		SELECT profile.id, profile.status, profile.depreciable_basis, profile.salvage_value, policy.code
		FROM asset_depreciation_profiles profile
		JOIN asset_depreciation_last_month_policies policy ON policy.id = profile.last_month_policy_id
		WHERE profile.asset_id = ?
		FOR UPDATE
	`, assetID).Scan(&profileID, &profileStatus, &basis, &salvage, &lastPolicyCode)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	postedAmount := 0.0
	bookValue := acquisitionValue
	if profileID.Valid {
		if err := tx.QueryRow(`
			SELECT COALESCE(SUM(depreciation_amount), 0)
			FROM asset_depreciation_schedules
			WHERE profile_id = ? AND status = 'POSTED'
		`, profileID.Int64).Scan(&postedAmount); err != nil {
			return err
		}
		bookValue = math.Max(salvage.Float64, basis.Float64-postedAmount)
		if err := validateAndFinalizeDisposalDepreciationTx(tx, id, profileID.Int64, assetID, disposalDate,
			lastPolicyCode.String, bookValue, salvage.Float64, auditCtx); err != nil {
			return err
		}
	}

	if _, err := tx.Exec(`UPDATE assets SET status = 'DISPOSED', updated_at = CURRENT_TIMESTAMP WHERE id = ?`, assetID); err != nil {
		return err
	}
	if profileID.Valid && profileStatus.String != "FINISHED" {
		terminationReason := fmt.Sprintf("Disposal %s (%s): %s", disposalNumber, typeCode, disposalReason)
		if _, err := tx.Exec(`
			UPDATE asset_depreciation_profiles
			SET status = 'TERMINATED', terminated_at = CURRENT_TIMESTAMP, terminated_by = ?,
				termination_reason = ?, updated_at = CURRENT_TIMESTAMP
			WHERE id = ?
		`, auditCtx.ActorUserID, terminationReason, profileID.Int64); err != nil {
			return err
		}
		if err := insertAuditLogTx(tx, "DEPRECIATION_PROFILE", profileID.Int64, "TERMINATE_PROFILE", terminationReason, auditCtx); err != nil {
			return err
		}
	}
	gainLoss := disposalValue - bookValue
	if _, err := tx.Exec(`
		UPDATE asset_disposals
		SET status = 'POSTED', depreciation_profile_id = ?, acquisition_value = ?,
			accumulated_depreciation = ?, book_value = ?, gain_loss_amount = ?,
			prior_asset_status = ?, prior_profile_status = ?, approved_by = ?, posted_at = CURRENT_TIMESTAMP,
			cancelled_at = NULL, cancelled_by = NULL, cancellation_reason = NULL,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND status = 'DRAFT'
	`, disposalNullableSQLInt64(profileID), acquisitionValue, postedAmount, bookValue, gainLoss,
		assetStatus, disposalNullableSQLString(profileStatus), auditCtx.ActorUserID, id); err != nil {
		return err
	}
	if err := insertAuditLogTx(tx, "ASSET_DISPOSAL", id, "POST_DISPOSAL",
		fmt.Sprintf("Disposal %s untuk aset %s diposting; nilai buku %s, nilai disposal %s",
			disposalNumber, assetCode, formatAssetAmountID(bookValue), formatAssetAmountID(disposalValue)), auditCtx); err != nil {
		return err
	}
	return tx.Commit()
}

func validateAndFinalizeDisposalDepreciationTx(tx *sql.Tx, disposalID, profileID, assetID int64, disposalDate time.Time,
	lastPolicyCode string, bookValue, salvage float64, auditCtx models.AuditContext) error {
	periodDate := time.Date(disposalDate.Year(), disposalDate.Month(), 1, 0, 0, 0, 0, time.Local)
	var laterPosted int
	if err := tx.QueryRow(`
		SELECT COUNT(*) FROM asset_depreciation_schedules
		WHERE profile_id = ? AND status = 'POSTED' AND period_date > ?
	`, profileID, periodDate).Scan(&laterPosted); err != nil {
		return err
	}
	if laterPosted > 0 {
		return errors.New("terdapat depresiasi yang diposting setelah bulan disposal; lakukan reversal terlebih dahulu")
	}

	var currentStatus sql.NullString
	err := tx.QueryRow(`
		SELECT status
		FROM asset_depreciation_schedules
		WHERE asset_id = ? AND period_year = ? AND period_month = ?
		ORDER BY version_no DESC
		LIMIT 1
		FOR UPDATE
	`, assetID, disposalDate.Year(), int(disposalDate.Month())).Scan(&currentStatus)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	if lastPolicyCode == "NO_DEPRECIATION" {
		if currentStatus.Valid && currentStatus.String == "POSTED" {
			return errors.New("kebijakan bulan terakhir tidak menghitung bulan disposal; lakukan reversal posting bulan tersebut")
		}
	} else if bookValue > salvage+0.005 && (!currentStatus.Valid || currentStatus.String != "POSTED") {
		return errors.New("depresiasi bulan disposal harus dibuat dan diposting sebelum disposal diposting")
	}

	rows, err := tx.Query(`
		SELECT id, period_date
		FROM asset_depreciation_schedules
		WHERE profile_id = ? AND status = 'DRAFT' AND period_date >= ?
		ORDER BY period_date
		FOR UPDATE
	`, profileID, periodDate)
	if err != nil {
		return err
	}
	records := make([]depreciationScheduleActionRecord, 0)
	for rows.Next() {
		var record depreciationScheduleActionRecord
		record.ProfileID = profileID
		if err := rows.Scan(&record.ID, &record.PeriodDate); err != nil {
			rows.Close()
			return err
		}
		records = append(records, record)
	}
	if err := rows.Close(); err != nil {
		return err
	}
	for _, record := range records {
		if _, err := assertDepreciationPeriodOpenTx(tx, record.PeriodDate.Year(), int(record.PeriodDate.Month())); err != nil {
			return err
		}
		if _, err := tx.Exec(`
			UPDATE asset_depreciation_schedules
			SET status = 'SKIPPED', accumulated_depreciation = GREATEST(0, accumulated_depreciation - depreciation_amount),
				closing_book_value = opening_book_value, depreciation_amount = 0,
				skipped_at = CURRENT_TIMESTAMP, skipped_by = ?, skip_reason = ?, updated_at = CURRENT_TIMESTAMP
			WHERE id = ? AND status = 'DRAFT'
		`, auditCtx.ActorUserID, fmt.Sprintf("Dilewati otomatis karena disposal ID %d", disposalID), record.ID); err != nil {
			return err
		}
		if err := insertAuditLogTx(tx, "ASSET_DEPRECIATION", record.ID, "SKIP_FOR_DISPOSAL",
			"Draft dilewati otomatis saat disposal aset diposting", auditCtx); err != nil {
			return err
		}
	}
	return syncDepreciationActionPeriodsTx(tx, records, auditCtx.ActorUserID)
}

func (r *AssetDisposalRepository) CancelDisposal(id int64, reason string, auditCtx models.AuditContext) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var assetID int64
	var profileID sql.NullInt64
	var number, status string
	var priorAssetStatus, priorProfileStatus sql.NullString
	if err := tx.QueryRow(`
		SELECT asset_id, depreciation_profile_id, disposal_number, status,
			prior_asset_status, prior_profile_status
		FROM asset_disposals WHERE id = ? FOR UPDATE
	`, id).Scan(&assetID, &profileID, &number, &status, &priorAssetStatus, &priorProfileStatus); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("transaksi disposal tidak ditemukan")
		}
		return err
	}
	if status == "CANCELLED" {
		return errors.New("transaksi disposal sudah dibatalkan")
	}
	if status == "POSTED" {
		if !priorAssetStatus.Valid || priorAssetStatus.String == "" {
			return errors.New("status aset sebelum disposal tidak tersedia")
		}
		if _, err := tx.Exec(`UPDATE assets SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, priorAssetStatus.String, assetID); err != nil {
			return err
		}
		if profileID.Valid && priorProfileStatus.Valid && priorProfileStatus.String != "" {
			if _, err := tx.Exec(`
				UPDATE asset_depreciation_profiles
				SET status = ?, terminated_at = NULL, terminated_by = NULL, termination_reason = NULL,
					updated_at = CURRENT_TIMESTAMP
				WHERE id = ?
			`, priorProfileStatus.String, profileID.Int64); err != nil {
				return err
			}
			if err := insertAuditLogTx(tx, "DEPRECIATION_PROFILE", profileID.Int64, "RESTORE_AFTER_DISPOSAL_CANCEL",
				fmt.Sprintf("Status profil dipulihkan menjadi %s karena disposal %s dibatalkan", priorProfileStatus.String, number), auditCtx); err != nil {
				return err
			}
		}
		if profileID.Valid {
			rows, err := tx.Query(`
			SELECT id, period_date
			FROM asset_depreciation_schedules
			WHERE profile_id = ? AND status = 'SKIPPED' AND skip_reason = ?
			FOR UPDATE
			`, profileID.Int64, fmt.Sprintf("Dilewati otomatis karena disposal ID %d", id))
			if err != nil {
				return err
			}
			records := make([]depreciationScheduleActionRecord, 0)
			for rows.Next() {
				var record depreciationScheduleActionRecord
				if err := rows.Scan(&record.ID, &record.PeriodDate); err != nil {
					rows.Close()
					return err
				}
				record.ProfileID = profileID.Int64
				records = append(records, record)
			}
			if err := rows.Close(); err != nil {
				return err
			}
			for _, record := range records {
				if _, err := assertDepreciationPeriodOpenTx(tx, record.PeriodDate.Year(), int(record.PeriodDate.Month())); err != nil {
					return errors.New("disposal tidak dapat dibatalkan karena periode depresiasi terkait sudah ditutup")
				}
				if _, err := tx.Exec(`
				UPDATE asset_depreciation_schedules
				SET status = 'DRAFT', skipped_at = NULL, skipped_by = NULL, skip_reason = NULL,
					updated_at = CURRENT_TIMESTAMP
				WHERE id = ? AND status = 'SKIPPED'
				`, record.ID); err != nil {
					return err
				}
			}
			if err := syncDepreciationActionPeriodsTx(tx, records, auditCtx.ActorUserID); err != nil {
				return err
			}
		}
	}
	if _, err := tx.Exec(`
		UPDATE asset_disposals
		SET status = 'CANCELLED', cancelled_at = CURRENT_TIMESTAMP, cancelled_by = ?,
			cancellation_reason = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND status IN ('DRAFT', 'POSTED')
	`, auditCtx.ActorUserID, reason, id); err != nil {
		return err
	}
	if err := insertAuditLogTx(tx, "ASSET_DISPOSAL", id, "CANCEL_DISPOSAL",
		fmt.Sprintf("Disposal %s dibatalkan. Alasan: %s", number, reason), auditCtx); err != nil {
		return err
	}
	return tx.Commit()
}

func disposalNullableSQLInt64(value sql.NullInt64) any {
	if !value.Valid {
		return nil
	}
	return value.Int64
}

func disposalNullableSQLString(value sql.NullString) any {
	if !value.Valid || value.String == "" {
		return nil
	}
	return value.String
}
