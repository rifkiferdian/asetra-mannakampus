package repositories

import (
	"database/sql"
	"errors"
	"gobase-app/models"
	"math"
	"strconv"
	"strings"
)

type AssetRepository struct {
	DB *sql.DB
}

func (r *AssetRepository) GetAssetTypes() ([]models.AssetType, error) {
	rows, err := r.DB.Query(`
		SELECT
			at.id,
			at.code,
			at.name,
			COALESCE(at.description, ''),
			at.is_active,
			COUNT(a.id) AS asset_count,
			at.created_at,
			at.updated_at
		FROM asset_types at
		LEFT JOIN assets a ON a.asset_type_id = at.id
		GROUP BY at.id, at.code, at.name, at.description, at.is_active, at.created_at, at.updated_at
		ORDER BY at.code ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.AssetType
	for rows.Next() {
		var item models.AssetType
		var isActive int
		var createdAt sql.NullTime
		var updatedAt sql.NullTime
		if err := rows.Scan(&item.ID, &item.Code, &item.Name, &item.Description, &isActive, &item.AssetCount, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		item.IsActive = isActive == 1
		item.IsActiveLabel = activeLabel(item.IsActive)
		item.CreatedAtDisplay = formatNullTime(createdAt)
		item.UpdatedAtDisplay = formatNullTime(updatedAt)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *AssetRepository) CreateAssetType(input models.AssetTypeInput) error {
	isActive := boolToInt(input.IsActive)
	_, err := r.DB.Exec(`
		INSERT INTO asset_types (code, name, description, is_active)
		VALUES (?, ?, ?, ?)
	`, input.Code, input.Name, nullableString(input.Description), isActive)
	return err
}

func (r *AssetRepository) UpdateAssetType(input models.AssetTypeInput) error {
	isActive := boolToInt(input.IsActive)
	_, err := r.DB.Exec(`
		UPDATE asset_types
		SET code = ?, name = ?, description = ?, is_active = ?
		WHERE id = ?
	`, input.Code, input.Name, nullableString(input.Description), isActive, input.ID)
	return err
}

func (r *AssetRepository) DeleteAssetType(id int64) error {
	_, err := r.DB.Exec(`DELETE FROM asset_types WHERE id = ?`, id)
	return err
}

func (r *AssetRepository) GetComponentTypes() ([]models.ComponentType, error) {
	rows, err := r.DB.Query(`
		SELECT id, code, name, COALESCE(description, ''), is_active, created_at
		FROM component_types
		ORDER BY code ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.ComponentType
	for rows.Next() {
		var item models.ComponentType
		var isActive int
		var createdAt sql.NullTime
		if err := rows.Scan(&item.ID, &item.Code, &item.Name, &item.Description, &isActive, &createdAt); err != nil {
			return nil, err
		}
		item.IsActive = isActive == 1
		item.IsActiveLabel = activeLabel(item.IsActive)
		item.CreatedAtDisplay = formatNullTime(createdAt)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *AssetRepository) CreateComponentType(input models.ComponentTypeInput) error {
	isActive := boolToInt(input.IsActive)
	_, err := r.DB.Exec(`
		INSERT INTO component_types (code, name, description, is_active)
		VALUES (?, ?, ?, ?)
	`, input.Code, input.Name, nullableString(input.Description), isActive)
	return err
}

func (r *AssetRepository) UpdateComponentType(input models.ComponentTypeInput) error {
	isActive := boolToInt(input.IsActive)
	_, err := r.DB.Exec(`
		UPDATE component_types
		SET code = ?, name = ?, description = ?, is_active = ?
		WHERE id = ?
	`, input.Code, input.Name, nullableString(input.Description), isActive, input.ID)
	return err
}

func (r *AssetRepository) DeleteComponentType(id int64) error {
	_, err := r.DB.Exec(`DELETE FROM component_types WHERE id = ?`, id)
	return err
}

func (r *AssetRepository) GetLocations() ([]models.AssetLocation, error) {
	rows, err := r.DB.Query(`
		SELECT
			al.id, al.location_code, al.location_name, COALESCE(al.store_id, 0),
			COALESCE(s.store_name, ''), COALESCE(al.parent_id, 0), COALESCE(parent.location_name, ''),
			al.location_type, al.is_active, al.created_at, al.updated_at
		FROM asset_locations al
		LEFT JOIN stores s ON s.store_id = al.store_id
		LEFT JOIN asset_locations parent ON parent.id = al.parent_id
		ORDER BY al.location_name ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.AssetLocation
	for rows.Next() {
		var item models.AssetLocation
		var isActive int
		var createdAt sql.NullTime
		var updatedAt sql.NullTime
		if err := rows.Scan(
			&item.ID, &item.LocationCode, &item.LocationName, &item.StoreID, &item.StoreName,
			&item.ParentID, &item.ParentName, &item.LocationType, &isActive, &createdAt, &updatedAt,
		); err != nil {
			return nil, err
		}
		item.IsActive = isActive == 1
		item.IsActiveLabel = activeLabel(item.IsActive)
		item.CreatedAtDisplay = formatNullTime(createdAt)
		item.UpdatedAtDisplay = formatNullTime(updatedAt)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *AssetRepository) CreateLocation(input models.AssetLocationInput) error {
	_, err := r.DB.Exec(`
		INSERT INTO asset_locations (location_code, location_name, store_id, parent_id, location_type, is_active)
		VALUES (?, ?, ?, ?, ?, ?)
	`, input.LocationCode, input.LocationName, nullableInt(input.StoreID), assetNullableInt64(input.ParentID), input.LocationType, boolToInt(input.IsActive))
	return err
}

func (r *AssetRepository) UpdateLocation(input models.AssetLocationInput) error {
	_, err := r.DB.Exec(`
		UPDATE asset_locations
		SET location_code = ?, location_name = ?, store_id = ?, parent_id = ?, location_type = ?, is_active = ?
		WHERE id = ?
	`, input.LocationCode, input.LocationName, nullableInt(input.StoreID), assetNullableInt64(input.ParentID), input.LocationType, boolToInt(input.IsActive), input.ID)
	return err
}

func (r *AssetRepository) DeleteLocation(id int64) error {
	_, err := r.DB.Exec(`DELETE FROM asset_locations WHERE id = ?`, id)
	return err
}

func (r *AssetRepository) GetAssets() ([]models.Asset, error) {
	rows, err := r.DB.Query(`
		SELECT
			a.id, a.asset_code, a.asset_name, a.asset_type_id, COALESCE(at.name, ''),
			COALESCE(a.serial_number, ''), COALESCE(a.store_id, 0), COALESCE(s.store_name, ''),
			COALESCE(a.location_id, 0), COALESCE(al.location_name, ''),
			COALESCE(a.assigned_person_nip, ''), COALESCE(a.assigned_person_name, ''),
			COALESCE(a.assigned_person_department, ''),
			COALESCE(a.source_gr_item_id, 0), a.acquisition_date, a.acquisition_value,
			a.status, COALESCE(a.notes, ''), a.created_at
		FROM assets a
		JOIN asset_types at ON at.id = a.asset_type_id
		LEFT JOIN stores s ON s.store_id = a.store_id
		LEFT JOIN asset_locations al ON al.id = a.location_id
		ORDER BY a.id DESC
	`)
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
		item.AcquisitionValueDisplay = formatAssetAmountID(item.AcquisitionValue)
		item.StatusLabel = formatAssetStatusLabel(item.Status)
		item.CreatedAtDisplay = formatNullTime(createdAt)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *AssetRepository) GetAssetByID(id int64) (*models.Asset, error) {
	row := r.DB.QueryRow(`
		SELECT
			a.id, a.asset_code, a.asset_name, a.asset_type_id, COALESCE(at.name, ''),
			COALESCE(a.serial_number, ''), COALESCE(a.store_id, 0), COALESCE(s.store_name, ''),
			COALESCE(a.location_id, 0), COALESCE(al.location_name, ''),
			COALESCE(a.assigned_person_nip, ''), COALESCE(a.assigned_person_name, ''),
			COALESCE(a.assigned_person_department, ''),
			COALESCE(a.source_gr_item_id, 0), a.acquisition_date, a.acquisition_value,
			a.status, COALESCE(a.notes, ''), a.created_at
		FROM assets a
		JOIN asset_types at ON at.id = a.asset_type_id
		LEFT JOIN stores s ON s.store_id = a.store_id
		LEFT JOIN asset_locations al ON al.id = a.location_id
		WHERE a.id = ?
	`, id)

	var item models.Asset
	var acquisitionDate sql.NullTime
	var createdAt sql.NullTime
	if err := row.Scan(
		&item.ID, &item.AssetCode, &item.AssetName, &item.AssetTypeID, &item.AssetTypeName,
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
	item.AcquisitionValueDisplay = formatAssetAmountID(item.AcquisitionValue)
	item.StatusLabel = formatAssetStatusLabel(item.Status)
	item.CreatedAtDisplay = formatNullTime(createdAt)
	return &item, nil
}

func (r *AssetRepository) CreateAsset(input models.AssetInput) error {
	_, err := r.DB.Exec(`
		INSERT INTO assets (
			asset_code, asset_name, asset_type_id, serial_number, store_id, location_id,
			assigned_person_nip, assigned_person_name, assigned_person_department,
			source_gr_item_id, acquisition_date, acquisition_value, status, notes
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, input.AssetCode, input.AssetName, input.AssetTypeID, nullableString(input.SerialNumber),
		nullableInt(input.StoreID), assetNullableInt64(input.LocationID), nullableString(input.AssignedPersonNIP),
		nullableString(input.AssignedPersonName), nullableString(input.AssignedPersonDepartment),
		assetNullableInt64(input.SourceGRItemID), nullableDate(input.AcquisitionDate), input.AcquisitionValue,
		input.Status, nullableString(input.Notes))
	return err
}

func (r *AssetRepository) UpdateAsset(input models.AssetInput) error {
	_, err := r.DB.Exec(`
		UPDATE assets
		SET asset_code = ?, asset_name = ?, asset_type_id = ?, serial_number = ?, store_id = ?,
			location_id = ?, assigned_person_nip = ?, assigned_person_name = ?,
			assigned_person_department = ?, source_gr_item_id = ?, acquisition_date = ?,
			acquisition_value = ?, status = ?, notes = ?
		WHERE id = ?
	`, input.AssetCode, input.AssetName, input.AssetTypeID, nullableString(input.SerialNumber),
		nullableInt(input.StoreID), assetNullableInt64(input.LocationID), nullableString(input.AssignedPersonNIP),
		nullableString(input.AssignedPersonName), nullableString(input.AssignedPersonDepartment),
		assetNullableInt64(input.SourceGRItemID), nullableDate(input.AcquisitionDate), input.AcquisitionValue,
		input.Status, nullableString(input.Notes), input.ID)
	return err
}

func (r *AssetRepository) DeleteAsset(id int64) error {
	_, err := r.DB.Exec(`DELETE FROM assets WHERE id = ?`, id)
	return err
}

func (r *AssetRepository) GetComponents() ([]models.AssetComponent, error) {
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
		ORDER BY c.id DESC
	`)
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
		item.CreatedAtDisplay = formatNullTime(createdAt)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *AssetRepository) CreateComponent(input models.AssetComponentInput) error {
	_, err := r.DB.Exec(`
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
	return err
}

func (r *AssetRepository) UpdateComponent(input models.AssetComponentInput) error {
	_, err := r.DB.Exec(`
		UPDATE asset_components
		SET component_code = ?, component_name = ?, component_type_id = ?, brand = ?, model = ?,
			specification = ?, serial_number = ?, parent_asset_id = ?, location_id = ?,
			source_gr_item_id = ?, acquisition_date = ?, acquisition_value = ?, status = ?, notes = ?
		WHERE id = ?
	`, input.ComponentCode, input.ComponentName, input.ComponentTypeID, nullableString(input.Brand),
		nullableString(input.Model), nullableString(input.Specification), nullableString(input.SerialNumber),
		assetNullableInt64(input.ParentAssetID), assetNullableInt64(input.LocationID),
		assetNullableInt64(input.SourceGRItemID), nullableDate(input.AcquisitionDate),
		input.AcquisitionValue, input.Status, nullableString(input.Notes), input.ID)
	return err
}

func (r *AssetRepository) DeleteComponent(id int64) error {
	_, err := r.DB.Exec(`DELETE FROM asset_components WHERE id = ?`, id)
	return err
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
		item.CreatedAtDisplay = formatNullTime(createdAt)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *AssetRepository) GetAssetMovements() ([]models.AssetMovement, error) {
	rows, err := r.DB.Query(`
		SELECT
			m.id, m.asset_id, COALESCE(a.asset_code, ''), m.movement_type,
			COALESCE(fs.store_name, ''), COALESCE(ts.store_name, ''),
			COALESCE(fl.location_name, ''), COALESCE(tl.location_name, ''),
			COALESCE(fu.name, ''), COALESCE(tu.name, ''), COALESCE(au.name, ''),
			m.movement_date, COALESCE(m.notes, '')
		FROM asset_movements m
		JOIN assets a ON a.id = m.asset_id
		LEFT JOIN stores fs ON fs.store_id = m.from_store_id
		LEFT JOIN stores ts ON ts.store_id = m.to_store_id
		LEFT JOIN asset_locations fl ON fl.id = m.from_location_id
		LEFT JOIN asset_locations tl ON tl.id = m.to_location_id
		LEFT JOIN users fu ON fu.id = m.from_user_id
		LEFT JOIN users tu ON tu.id = m.to_user_id
		LEFT JOIN users au ON au.id = m.acted_by
		ORDER BY m.movement_date DESC, m.id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.AssetMovement
	for rows.Next() {
		var item models.AssetMovement
		var movementDate sql.NullTime
		if err := rows.Scan(
			&item.ID, &item.AssetID, &item.AssetCode, &item.MovementType,
			&item.FromStoreName, &item.ToStoreName, &item.FromLocationName, &item.ToLocationName,
			&item.FromUserName, &item.ToUserName, &item.ActedByName, &movementDate, &item.Notes,
		); err != nil {
			return nil, err
		}
		item.MovementDate = formatNullTime(movementDate)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *AssetRepository) GetAssetMovementsByAssetID(assetID int64, limit int) ([]models.AssetMovement, error) {
	if limit <= 0 {
		limit = 5
	}
	rows, err := r.DB.Query(`
		SELECT
			m.id, m.asset_id, COALESCE(a.asset_code, ''), m.movement_type,
			COALESCE(fs.store_name, ''), COALESCE(ts.store_name, ''),
			COALESCE(fl.location_name, ''), COALESCE(tl.location_name, ''),
			COALESCE(m.from_person_name, fu.name, ''), COALESCE(m.to_person_name, tu.name, ''),
			COALESCE(au.name, ''),
			m.movement_date, COALESCE(m.notes, '')
		FROM asset_movements m
		JOIN assets a ON a.id = m.asset_id
		LEFT JOIN stores fs ON fs.store_id = m.from_store_id
		LEFT JOIN stores ts ON ts.store_id = m.to_store_id
		LEFT JOIN asset_locations fl ON fl.id = m.from_location_id
		LEFT JOIN asset_locations tl ON tl.id = m.to_location_id
		LEFT JOIN users fu ON fu.id = m.from_user_id
		LEFT JOIN users tu ON tu.id = m.to_user_id
		LEFT JOIN users au ON au.id = m.acted_by
		WHERE m.asset_id = ?
		ORDER BY m.movement_date DESC, m.id DESC
		LIMIT ?
	`, assetID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.AssetMovement
	for rows.Next() {
		var item models.AssetMovement
		var movementDate sql.NullTime
		if err := rows.Scan(
			&item.ID, &item.AssetID, &item.AssetCode, &item.MovementType,
			&item.FromStoreName, &item.ToStoreName, &item.FromLocationName, &item.ToLocationName,
			&item.FromUserName, &item.ToUserName, &item.ActedByName, &movementDate, &item.Notes,
		); err != nil {
			return nil, err
		}
		item.MovementDate = formatNullTime(movementDate)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *AssetRepository) CreateAssetMovement(input models.AssetMovementInput) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}

	var fromStoreID, fromLocationID, fromUserID sql.NullInt64
	if err := tx.QueryRow(`
		SELECT store_id, location_id, assigned_user_id
		FROM assets
		WHERE id = ?
		FOR UPDATE
	`, input.AssetID).Scan(&fromStoreID, &fromLocationID, &fromUserID); err != nil {
		tx.Rollback()
		return err
	}

	if _, err := tx.Exec(`
		INSERT INTO asset_movements (
			asset_id, movement_type, from_store_id, to_store_id, from_location_id, to_location_id,
			from_user_id, to_user_id, acted_by, notes
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, input.AssetID, input.MovementType, nullInt64Value(fromStoreID), nullableInt(input.ToStoreID),
		nullInt64Value(fromLocationID), assetNullableInt64(input.ToLocationID), nullInt64Value(fromUserID),
		nullableInt(input.ToUserID), nullableInt(input.ActedBy), nullableString(input.Notes)); err != nil {
		tx.Rollback()
		return err
	}

	status := nextAssetStatus(input.MovementType, input.ToUserID)
	if _, err := tx.Exec(`
		UPDATE assets
		SET store_id = ?, location_id = ?, assigned_user_id = ?, status = ?
		WHERE id = ?
	`, nullableInt(input.ToStoreID), assetNullableInt64(input.ToLocationID), nullableInt(input.ToUserID), status, input.AssetID); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (r *AssetRepository) GetComponentMovements() ([]models.AssetComponentMovement, error) {
	rows, err := r.DB.Query(`
		SELECT
			m.id, m.component_id, COALESCE(c.component_code, ''), m.movement_type,
			COALESCE(fa.asset_code, ''), COALESCE(ta.asset_code, ''),
			COALESCE(fl.location_name, ''), COALESCE(tl.location_name, ''),
			COALESCE(au.name, ''), m.movement_date, COALESCE(m.notes, '')
		FROM asset_component_movements m
		JOIN asset_components c ON c.id = m.component_id
		LEFT JOIN assets fa ON fa.id = m.from_asset_id
		LEFT JOIN assets ta ON ta.id = m.to_asset_id
		LEFT JOIN asset_locations fl ON fl.id = m.from_location_id
		LEFT JOIN asset_locations tl ON tl.id = m.to_location_id
		LEFT JOIN users au ON au.id = m.acted_by
		ORDER BY m.movement_date DESC, m.id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.AssetComponentMovement
	for rows.Next() {
		var item models.AssetComponentMovement
		var movementDate sql.NullTime
		if err := rows.Scan(
			&item.ID, &item.ComponentID, &item.ComponentCode, &item.MovementType,
			&item.FromAssetCode, &item.ToAssetCode, &item.FromLocationName, &item.ToLocationName,
			&item.ActedByName, &movementDate, &item.Notes,
		); err != nil {
			return nil, err
		}
		item.MovementDate = formatNullTime(movementDate)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *AssetRepository) CreateComponentMovement(input models.AssetComponentMovementInput) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}

	var fromAssetID, fromLocationID sql.NullInt64
	if err := tx.QueryRow(`
		SELECT parent_asset_id, location_id
		FROM asset_components
		WHERE id = ?
		FOR UPDATE
	`, input.ComponentID).Scan(&fromAssetID, &fromLocationID); err != nil {
		tx.Rollback()
		return err
	}

	if _, err := tx.Exec(`
		INSERT INTO asset_component_movements (
			component_id, movement_type, from_asset_id, to_asset_id, from_location_id,
			to_location_id, acted_by, notes
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, input.ComponentID, input.MovementType, nullInt64Value(fromAssetID), assetNullableInt64(input.ToAssetID),
		nullInt64Value(fromLocationID), assetNullableInt64(input.ToLocationID), nullableInt(input.ActedBy),
		nullableString(input.Notes)); err != nil {
		tx.Rollback()
		return err
	}

	nextAssetID := input.ToAssetID
	if input.MovementType == "UNINSTALL" || input.MovementType == "RETURN_TO_STORAGE" || input.MovementType == "TRANSFER" {
		nextAssetID = 0
	}
	status := nextComponentStatus(input.MovementType, nextAssetID)

	if _, err := tx.Exec(`
		UPDATE asset_components
		SET parent_asset_id = ?, location_id = ?, status = ?
		WHERE id = ?
	`, assetNullableInt64(nextAssetID), assetNullableInt64(input.ToLocationID), status, input.ComponentID); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (r *AssetRepository) GetStoreOptions() ([]models.Store, error) {
	rows, err := r.DB.Query(`
		SELECT store_id, store_code, store_name
		FROM stores
		ORDER BY store_name ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.Store
	for rows.Next() {
		var item models.Store
		if err := rows.Scan(&item.StoreID, &item.StoreCode, &item.StoreName); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *AssetRepository) GetUserOptions() ([]models.UserSelectOption, error) {
	rows, err := r.DB.Query(`
		SELECT id, name
		FROM users
		WHERE status = 'active'
		ORDER BY name ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.UserSelectOption
	for rows.Next() {
		var item models.UserSelectOption
		if err := rows.Scan(&item.ID, &item.Name); err != nil {
			return nil, err
		}
		item.Label = item.Name
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *AssetRepository) GetAssetOptions() ([]models.AssetSelectOption, error) {
	rows, err := r.DB.Query(`
		SELECT id, asset_code, asset_name
		FROM assets
		WHERE status <> 'DISPOSED'
		ORDER BY asset_code ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.AssetSelectOption
	for rows.Next() {
		var item models.AssetSelectOption
		if err := rows.Scan(&item.ID, &item.Code, &item.Name); err != nil {
			return nil, err
		}
		item.Label = strings.TrimSpace(item.Code + " - " + item.Name)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *AssetRepository) GetComponentOptions() ([]models.AssetSelectOption, error) {
	rows, err := r.DB.Query(`
		SELECT id, component_code, component_name
		FROM asset_components
		WHERE status <> 'DISPOSED'
		ORDER BY component_code ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.AssetSelectOption
	for rows.Next() {
		var item models.AssetSelectOption
		if err := rows.Scan(&item.ID, &item.Code, &item.Name); err != nil {
			return nil, err
		}
		item.Label = strings.TrimSpace(item.Code + " - " + item.Name)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *AssetRepository) ExistsByID(table string, id int64) (bool, error) {
	if !assetAllowedTable(table) {
		return false, errors.New("invalid table")
	}
	var count int
	err := r.DB.QueryRow(`SELECT COUNT(1) FROM `+table+` WHERE id = ?`, id).Scan(&count)
	return count > 0, err
}

func assetAllowedTable(table string) bool {
	switch table {
	case "asset_types", "component_types", "asset_locations", "assets", "asset_components":
		return true
	default:
		return false
	}
}

func assetNullableInt64(value int64) interface{} {
	if value <= 0 {
		return nil
	}
	return value
}

func nullInt64Value(value sql.NullInt64) interface{} {
	if !value.Valid {
		return nil
	}
	return value.Int64
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func activeLabel(value bool) string {
	if value {
		return "Aktif"
	}
	return "Non Aktif"
}

func formatNullTime(value sql.NullTime) string {
	if value.Valid {
		return value.Time.Format("02 Jan 2006 15:04")
	}
	return "-"
}

func formatAssetAmountID(value float64) string {
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

func formatAssetStatusLabel(status string) string {
	switch status {
	case "AVAILABLE":
		return "Available"
	case "IN_USE":
		return "In Use"
	case "MAINTENANCE":
		return "Maintenance"
	case "BROKEN":
		return "Broken"
	case "DISPOSED":
		return "Disposed"
	default:
		return status
	}
}

func nextAssetStatus(movementType string, toUserID int) string {
	switch movementType {
	case "MAINTENANCE":
		return "MAINTENANCE"
	case "BROKEN":
		return "BROKEN"
	case "DISPOSE":
		return "DISPOSED"
	case "RETURN":
		return "AVAILABLE"
	default:
		if toUserID > 0 || movementType == "ASSIGN" || movementType == "TRANSFER" || movementType == "RECEIVE" {
			return "IN_USE"
		}
		return "AVAILABLE"
	}
}

func nextComponentStatus(movementType string, toAssetID int64) string {
	switch movementType {
	case "INSTALL":
		return "INSTALLED"
	case "MAINTENANCE":
		return "MAINTENANCE"
	case "BROKEN":
		return "BROKEN"
	case "DISPOSE":
		return "DISPOSED"
	default:
		if toAssetID > 0 {
			return "INSTALLED"
		}
		return "IN_STORAGE"
	}
}
