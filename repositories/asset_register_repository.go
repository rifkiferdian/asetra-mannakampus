package repositories

import (
	"database/sql"
	"errors"
	"gobase-app/models"
)

func (r *AssetRepository) GetAssets() ([]models.Asset, error) {
	return r.getAssetsWithPage(0, 0)
}

func (r *AssetRepository) GetAssetsPage(limit, offset int, scope models.AccessScope) ([]models.Asset, error) {
	if limit <= 0 {
		return nil, errors.New("asset page size tidak valid")
	}
	if offset < 0 {
		offset = 0
	}
	return r.getAssetsWithPage(limit, offset, scope)
}

func (r *AssetRepository) CountAssets(scope models.AccessScope) (int, error) {
	var total int
	where, args := assetStoreScopeSQL("store_id", scope)
	err := r.DB.QueryRow(`SELECT COUNT(*) FROM assets WHERE `+where, args...).Scan(&total)
	return total, err
}

func (r *AssetRepository) GetAssetStatusCounts(scope models.AccessScope) (map[string]int, error) {
	where, args := assetStoreScopeSQL("store_id", scope)
	rows, err := r.DB.Query(`SELECT status, COUNT(*) FROM assets WHERE `+where+` GROUP BY status`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	counts := make(map[string]int)
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, err
		}
		counts[status] = count
	}
	return counts, rows.Err()
}

func (r *AssetRepository) getAssetsWithPage(limit, offset int, scopes ...models.AccessScope) ([]models.Asset, error) {
	query := `
		SELECT
			a.id, a.asset_code, a.asset_name, a.asset_type_id, COALESCE(at.name, ''),
			COALESCE(a.asset_category_id, 0), COALESCE(ac.name, ''), COALESCE(ac.category_prefix, ''),
			COALESCE(a.serial_number, ''), COALESCE(a.store_id, 0), COALESCE(s.store_name, ''),
			COALESCE(a.location_id, 0), COALESCE(al.location_name, ''),
			COALESCE(a.assigned_person_nip, ''), COALESCE(a.assigned_person_name, ''),
			COALESCE(a.assigned_person_department, ''),
			COALESCE(a.source_gr_item_id, 0), a.acquisition_date, a.acquisition_value,
			a.status, COALESCE(a.notes, ''), a.created_at
		FROM assets a
		JOIN asset_types at ON at.id = a.asset_type_id
		LEFT JOIN asset_categories ac ON ac.id = a.asset_category_id
		LEFT JOIN stores s ON s.store_id = a.store_id
		LEFT JOIN asset_locations al ON al.id = a.location_id`
	args := make([]any, 0, 2)
	if len(scopes) > 0 {
		where, scopeArgs := assetStoreScopeSQL("a.store_id", scopes[0])
		query += ` WHERE ` + where
		args = append(args, scopeArgs...)
	}
	query += ` ORDER BY a.id DESC`
	if limit > 0 {
		query += ` LIMIT ? OFFSET ?`
		args = append(args, limit, offset)
	}
	rows, err := r.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.Asset
	for rows.Next() {
		var item models.Asset
		var acquisitionDate sql.NullTime
		var createdAt sql.NullTime
		if err := rows.Scan(
			&item.ID, &item.AssetCode, &item.AssetName, &item.AssetTypeID, &item.AssetTypeName,
			&item.AssetCategoryID, &item.AssetCategoryName, &item.AssetCategoryPrefix,
			&item.SerialNumber, &item.StoreID, &item.StoreName, &item.LocationID, &item.LocationName,
			&item.AssignedPersonNIP, &item.AssignedPersonName, &item.AssignedPersonDepartment,
			&item.SourceGRItemID, &acquisitionDate,
			&item.AcquisitionValue, &item.Status, &item.Notes, &createdAt,
		); err != nil {
			return nil, err
		}
		if acquisitionDate.Valid {
			item.AcquisitionDate = acquisitionDate.Time.Format("2006-01-02")
		}
		item.AcquisitionValueInput = item.AcquisitionValue.DecimalString()
		item.AcquisitionValueDisplay = item.AcquisitionValue.FormatIDR()
		item.StatusLabel = formatAssetStatusLabel(item.Status)
		item.CreatedAtDisplay = formatNullTime(createdAt)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *AssetRepository) GetAssetByID(id int64, scopes ...models.AccessScope) (*models.Asset, error) {
	query := `
		SELECT
			a.id, a.asset_code, a.asset_name, a.asset_type_id, COALESCE(at.name, ''),
			COALESCE(a.asset_category_id, 0), COALESCE(ac.name, ''), COALESCE(ac.category_prefix, ''),
			COALESCE(a.serial_number, ''), COALESCE(a.store_id, 0), COALESCE(s.store_name, ''),
			COALESCE(a.location_id, 0), COALESCE(al.location_name, ''),
			COALESCE(a.assigned_person_nip, ''), COALESCE(a.assigned_person_name, ''),
			COALESCE(a.assigned_person_department, ''),
			COALESCE(a.source_gr_item_id, 0), a.acquisition_date, a.acquisition_value,
			a.status, COALESCE(a.notes, ''), a.created_at
		FROM assets a
		JOIN asset_types at ON at.id = a.asset_type_id
		LEFT JOIN asset_categories ac ON ac.id = a.asset_category_id
		LEFT JOIN stores s ON s.store_id = a.store_id
		LEFT JOIN asset_locations al ON al.id = a.location_id
		WHERE a.id = ?`
	args := []any{id}
	if len(scopes) > 0 {
		where, scopeArgs := assetStoreScopeSQL("a.store_id", scopes[0])
		query += ` AND ` + where
		args = append(args, scopeArgs...)
	}
	row := r.DB.QueryRow(query, args...)

	var item models.Asset
	var acquisitionDate sql.NullTime
	var createdAt sql.NullTime
	if err := row.Scan(
		&item.ID, &item.AssetCode, &item.AssetName, &item.AssetTypeID, &item.AssetTypeName,
		&item.AssetCategoryID, &item.AssetCategoryName, &item.AssetCategoryPrefix,
		&item.SerialNumber, &item.StoreID, &item.StoreName, &item.LocationID, &item.LocationName,
		&item.AssignedPersonNIP, &item.AssignedPersonName, &item.AssignedPersonDepartment,
		&item.SourceGRItemID, &acquisitionDate,
		&item.AcquisitionValue, &item.Status, &item.Notes, &createdAt,
	); err != nil {
		return nil, err
	}
	if acquisitionDate.Valid {
		item.AcquisitionDate = acquisitionDate.Time.Format("2006-01-02")
	}
	item.AcquisitionValueInput = item.AcquisitionValue.DecimalString()
	item.AcquisitionValueDisplay = item.AcquisitionValue.FormatIDR()
	item.StatusLabel = formatAssetStatusLabel(item.Status)
	item.CreatedAtDisplay = formatNullTime(createdAt)
	return &item, nil
}

func (r *AssetRepository) CreateAsset(input models.AssetInput) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err = tx.Exec(
		`CALL sp_generate_asset_code(?, ?, ?, @generated_asset_code)`,
		input.AssetTypeID, input.AssetCategoryID, nullableDate(input.AcquisitionDate),
	); err != nil {
		return err
	}
	if err = tx.QueryRow(`SELECT @generated_asset_code`).Scan(&input.AssetCode); err != nil {
		return err
	}

	result, err := tx.Exec(`
		INSERT INTO assets (
			asset_code, asset_name, asset_type_id, asset_category_id, serial_number, store_id, location_id,
			assigned_person_nip, assigned_person_name, assigned_person_department,
			source_gr_item_id, acquisition_date, acquisition_value, status, notes
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, input.AssetCode, input.AssetName, input.AssetTypeID, input.AssetCategoryID, nullableString(input.SerialNumber),
		nullableInt(input.StoreID), assetNullableInt64(input.LocationID), nullableString(input.AssignedPersonNIP),
		nullableString(input.AssignedPersonName), nullableString(input.AssignedPersonDepartment),
		assetNullableInt64(input.SourceGRItemID), nullableDate(input.AcquisitionDate), input.AcquisitionValue,
		input.Status, nullableString(input.Notes))
	if err != nil {
		return err
	}
	assetID, err := result.LastInsertId()
	if err != nil {
		return err
	}
	if err := insertAuditLogTx(tx, "ASSET", assetID, "CREATE", "Aset "+input.AssetCode+" dibuat", input.AuditContext); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *AssetRepository) UpdateAsset(input models.AssetInput) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var currentCode string
	if err := tx.QueryRow(`SELECT asset_code FROM assets WHERE id = ? FOR UPDATE`, input.ID).Scan(&currentCode); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("aset tidak ditemukan")
		}
		return err
	}
	var postedDisposalCount int
	if err := tx.QueryRow(`SELECT COUNT(*) FROM asset_disposals WHERE asset_id = ? AND status = 'POSTED'`, input.ID).Scan(&postedDisposalCount); err != nil {
		return err
	}
	if postedDisposalCount > 0 {
		return errors.New("aset sudah memiliki disposal yang diposting; batalkan disposal untuk mengubah aset")
	}
	if _, err := tx.Exec(`
		UPDATE assets
		SET asset_name = ?, asset_type_id = ?, asset_category_id = ?, serial_number = ?, store_id = ?,
			location_id = ?, assigned_person_nip = ?, assigned_person_name = ?,
			assigned_person_department = ?, source_gr_item_id = ?, acquisition_date = ?,
			acquisition_value = ?, status = ?, notes = ?
		WHERE id = ?
	`, input.AssetName, input.AssetTypeID, input.AssetCategoryID, nullableString(input.SerialNumber),
		nullableInt(input.StoreID), assetNullableInt64(input.LocationID), nullableString(input.AssignedPersonNIP),
		nullableString(input.AssignedPersonName), nullableString(input.AssignedPersonDepartment),
		assetNullableInt64(input.SourceGRItemID), nullableDate(input.AcquisitionDate), input.AcquisitionValue,
		input.Status, nullableString(input.Notes), input.ID); err != nil {
		return err
	}
	if err := insertAuditLogTx(tx, "ASSET", input.ID, "UPDATE", "Aset "+currentCode+" diperbarui", input.AuditContext); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *AssetRepository) DeleteAsset(id int64) error {
	var dependencies int
	err := r.DB.QueryRow(`
		SELECT
			(SELECT COUNT(*) FROM asset_movements WHERE asset_id = ?) +
			(SELECT COUNT(*) FROM asset_components WHERE parent_asset_id = ?) +
			(SELECT COUNT(*) FROM asset_depreciation_profiles WHERE asset_id = ?) +
			(SELECT COUNT(*) FROM asset_disposals WHERE asset_id = ?)
	`, id, id, id, id).Scan(&dependencies)
	if err != nil {
		return err
	}
	if dependencies > 0 {
		return errors.New("aset memiliki histori atau transaksi terkait dan tidak dapat dihapus")
	}
	result, err := r.DB.Exec(`DELETE FROM assets WHERE id = ?`, id)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return errors.New("aset tidak ditemukan")
	}
	return nil
}
