package repositories

import (
	"database/sql"
	"errors"
	"gobase-app/models"
)

func (r *AssetRepository) GetComponents(scopes ...models.AccessScope) ([]models.AssetComponent, error) {
	query := `
		SELECT
			c.id, c.component_code, c.component_name, c.component_type_id, COALESCE(ct.name, ''),
			COALESCE(c.brand, ''), COALESCE(c.model, ''), COALESCE(c.specification, ''),
			COALESCE(c.serial_number, ''), COALESCE(c.parent_asset_id, 0), COALESCE(a.asset_code, ''),
			COALESCE(c.location_id, 0), COALESCE(al.location_name, ''),
			COALESCE(c.source_gr_item_id, 0), c.acquisition_date, c.acquisition_value,
			c.status, COALESCE(c.notes, ''), c.created_at
		FROM asset_components c
		JOIN component_types ct ON ct.id = c.component_type_id
		LEFT JOIN assets a ON a.id = c.parent_asset_id
		LEFT JOIN asset_locations al ON al.id = c.location_id`
	args := make([]any, 0)
	if len(scopes) > 0 {
		where, scopeArgs := assetStoreScopeSQL("COALESCE(a.store_id, al.store_id)", scopes[0])
		query += ` WHERE ` + where
		args = append(args, scopeArgs...)
	}
	query += ` ORDER BY c.id DESC`
	rows, err := r.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.AssetComponent
	for rows.Next() {
		var item models.AssetComponent
		var acquisitionDate sql.NullTime
		var createdAt sql.NullTime
		if err := rows.Scan(
			&item.ID, &item.ComponentCode, &item.ComponentName, &item.ComponentTypeID, &item.ComponentTypeName,
			&item.Brand, &item.Model, &item.Specification, &item.SerialNumber, &item.ParentAssetID,
			&item.ParentAssetCode, &item.LocationID, &item.LocationName, &item.SourceGRItemID,
			&acquisitionDate, &item.AcquisitionValue, &item.Status, &item.Notes, &createdAt,
		); err != nil {
			return nil, err
		}
		if acquisitionDate.Valid {
			item.AcquisitionDate = acquisitionDate.Time.Format("2006-01-02")
		}
		item.AcquisitionValueInput = item.AcquisitionValue.DecimalString()
		item.AcquisitionValueDisplay = item.AcquisitionValue.FormatIDR()
		item.StatusLabel = formatComponentStatusLabel(item.Status)
		item.CreatedAtDisplay = formatNullTime(createdAt)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *AssetRepository) GetComponentByID(id int64, scopes ...models.AccessScope) (*models.AssetComponent, error) {
	query := `
		SELECT
			c.id, c.component_code, c.component_name, c.component_type_id, COALESCE(ct.name, ''),
			COALESCE(c.brand, ''), COALESCE(c.model, ''), COALESCE(c.specification, ''),
			COALESCE(c.serial_number, ''), COALESCE(c.parent_asset_id, 0), COALESCE(a.asset_code, ''),
			COALESCE(c.location_id, 0), COALESCE(al.location_name, ''),
			COALESCE(c.source_gr_item_id, 0), c.acquisition_date, c.acquisition_value,
			c.status, COALESCE(c.notes, ''), c.created_at
		FROM asset_components c
		JOIN component_types ct ON ct.id = c.component_type_id
		LEFT JOIN assets a ON a.id = c.parent_asset_id
		LEFT JOIN asset_locations al ON al.id = c.location_id
		WHERE c.id = ?`
	args := []any{id}
	if len(scopes) > 0 {
		where, scopeArgs := assetStoreScopeSQL("COALESCE(a.store_id, al.store_id)", scopes[0])
		query += ` AND ` + where
		args = append(args, scopeArgs...)
	}
	row := r.DB.QueryRow(query, args...)

	var item models.AssetComponent
	var acquisitionDate sql.NullTime
	var createdAt sql.NullTime
	if err := row.Scan(
		&item.ID, &item.ComponentCode, &item.ComponentName, &item.ComponentTypeID, &item.ComponentTypeName,
		&item.Brand, &item.Model, &item.Specification, &item.SerialNumber, &item.ParentAssetID,
		&item.ParentAssetCode, &item.LocationID, &item.LocationName, &item.SourceGRItemID,
		&acquisitionDate, &item.AcquisitionValue, &item.Status, &item.Notes, &createdAt,
	); err != nil {
		return nil, err
	}
	if acquisitionDate.Valid {
		item.AcquisitionDate = acquisitionDate.Time.Format("2006-01-02")
	}
	item.AcquisitionValueInput = item.AcquisitionValue.DecimalString()
	item.AcquisitionValueDisplay = item.AcquisitionValue.FormatIDR()
	item.StatusLabel = formatComponentStatusLabel(item.Status)
	item.CreatedAtDisplay = formatNullTime(createdAt)
	return &item, nil
}

func (r *AssetRepository) CreateComponent(input models.AssetComponentInput) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	result, err := tx.Exec(`
		INSERT INTO asset_components (
			component_code, component_name, component_type_id, brand, model, specification,
			serial_number, parent_asset_id, location_id, source_gr_item_id, acquisition_date,
			acquisition_value, status, notes
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, input.ComponentCode, input.ComponentName, input.ComponentTypeID, nullableString(input.Brand),
		nullableString(input.Model), nullableString(input.Specification), nullableString(input.SerialNumber),
		assetNullableInt64(input.ParentAssetID), assetNullableInt64(input.LocationID),
		assetNullableInt64(input.SourceGRItemID), nullableDate(input.AcquisitionDate),
		input.AcquisitionValue, input.Status, nullableString(input.Notes))
	if err != nil {
		return err
	}
	componentID, err := result.LastInsertId()
	if err != nil {
		return err
	}
	if err := insertAuditLogTx(tx, "ASSET_COMPONENT", componentID, "CREATE",
		"Komponen "+input.ComponentCode+" dibuat", input.AuditContext); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *AssetRepository) UpdateComponent(input models.AssetComponentInput) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var currentCode string
	if err := tx.QueryRow(`SELECT component_code FROM asset_components WHERE id = ? FOR UPDATE`, input.ID).Scan(&currentCode); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("komponen tidak ditemukan")
		}
		return err
	}
	if _, err := tx.Exec(`
		UPDATE asset_components
		SET component_code = ?, component_name = ?, component_type_id = ?, brand = ?, model = ?,
			specification = ?, serial_number = ?, parent_asset_id = ?, location_id = ?,
			source_gr_item_id = ?, acquisition_date = ?, acquisition_value = ?, status = ?, notes = ?
		WHERE id = ?
	`, input.ComponentCode, input.ComponentName, input.ComponentTypeID, nullableString(input.Brand),
		nullableString(input.Model), nullableString(input.Specification), nullableString(input.SerialNumber),
		assetNullableInt64(input.ParentAssetID), assetNullableInt64(input.LocationID),
		assetNullableInt64(input.SourceGRItemID), nullableDate(input.AcquisitionDate),
		input.AcquisitionValue, input.Status, nullableString(input.Notes), input.ID); err != nil {
		return err
	}
	if err := insertAuditLogTx(tx, "ASSET_COMPONENT", input.ID, "UPDATE",
		"Komponen "+currentCode+" diperbarui", input.AuditContext); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *AssetRepository) DeleteComponent(id int64) error {
	var movements int
	if err := r.DB.QueryRow(`SELECT COUNT(*) FROM asset_component_movements WHERE component_id = ?`, id).Scan(&movements); err != nil {
		return err
	}
	if movements > 0 {
		return errors.New("komponen memiliki histori movement dan tidak dapat dihapus")
	}
	result, err := r.DB.Exec(`DELETE FROM asset_components WHERE id = ?`, id)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return errors.New("komponen tidak ditemukan")
	}
	return nil
}

func (r *AssetRepository) GetComponentsByAssetID(assetID int64) ([]models.AssetComponent, error) {
	rows, err := r.DB.Query(`
		SELECT
			c.id, c.component_code, c.component_name, c.component_type_id, COALESCE(ct.name, ''),
			COALESCE(c.brand, ''), COALESCE(c.model, ''), COALESCE(c.specification, ''),
			COALESCE(c.serial_number, ''), COALESCE(c.parent_asset_id, 0), COALESCE(a.asset_code, ''),
			COALESCE(c.location_id, 0), COALESCE(al.location_name, ''),
			COALESCE(c.source_gr_item_id, 0), c.acquisition_date, c.acquisition_value,
			c.status, COALESCE(c.notes, ''), c.created_at
		FROM asset_components c
		JOIN component_types ct ON ct.id = c.component_type_id
		LEFT JOIN assets a ON a.id = c.parent_asset_id
		LEFT JOIN asset_locations al ON al.id = c.location_id
		WHERE c.parent_asset_id = ?
		ORDER BY c.component_name ASC, c.id ASC
	`, assetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.AssetComponent
	for rows.Next() {
		var item models.AssetComponent
		var acquisitionDate sql.NullTime
		var createdAt sql.NullTime
		if err := rows.Scan(
			&item.ID, &item.ComponentCode, &item.ComponentName, &item.ComponentTypeID, &item.ComponentTypeName,
			&item.Brand, &item.Model, &item.Specification, &item.SerialNumber, &item.ParentAssetID,
			&item.ParentAssetCode, &item.LocationID, &item.LocationName, &item.SourceGRItemID,
			&acquisitionDate, &item.AcquisitionValue, &item.Status, &item.Notes, &createdAt,
		); err != nil {
			return nil, err
		}
		if acquisitionDate.Valid {
			item.AcquisitionDate = acquisitionDate.Time.Format("2006-01-02")
		}
		item.AcquisitionValueInput = item.AcquisitionValue.DecimalString()
		item.AcquisitionValueDisplay = item.AcquisitionValue.FormatIDR()
		item.StatusLabel = formatComponentStatusLabel(item.Status)
		item.CreatedAtDisplay = formatNullTime(createdAt)
		items = append(items, item)
	}
	return items, rows.Err()
}
