package repositories

import (
	"database/sql"
	"errors"
	"gobase-app/models"
)

func (r *AssetRepository) GetAssetTypes() ([]models.AssetType, error) {
	rows, err := r.DB.Query(`
		SELECT
			at.id,
			at.code,
			at.asset_prefix,
			at.name,
			COALESCE(at.description, ''),
			at.is_active,
			COUNT(a.id) AS asset_count,
			at.created_at,
			at.updated_at
		FROM asset_types at
		LEFT JOIN assets a ON a.asset_type_id = at.id
		GROUP BY at.id, at.code, at.asset_prefix, at.name, at.description, at.is_active, at.created_at, at.updated_at
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
		if err := rows.Scan(&item.ID, &item.Code, &item.AssetPrefix, &item.Name, &item.Description, &isActive, &item.AssetCount, &createdAt, &updatedAt); err != nil {
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
		INSERT INTO asset_types (code, asset_prefix, name, description, is_active)
		VALUES (?, ?, ?, ?, ?)
	`, input.Code, input.AssetPrefix, input.Name, nullableString(input.Description), isActive)
	return err
}

func (r *AssetRepository) UpdateAssetType(input models.AssetTypeInput) error {
	isActive := boolToInt(input.IsActive)
	_, err := r.DB.Exec(`
		UPDATE asset_types
		SET code = ?, asset_prefix = ?, name = ?, description = ?, is_active = ?
		WHERE id = ?
	`, input.Code, input.AssetPrefix, input.Name, nullableString(input.Description), isActive, input.ID)
	return err
}

func (r *AssetRepository) DeleteAssetType(id int64) error {
	_, err := r.DB.Exec(`DELETE FROM asset_types WHERE id = ?`, id)
	return err
}

func (r *AssetRepository) GetAssetCategories() ([]models.AssetCategory, error) {
	rows, err := r.DB.Query(`
		SELECT
			ac.id, ac.asset_type_id, at.name, ac.code, ac.category_prefix,
			ac.name, COALESCE(ac.description, ''), ac.is_active,
			COUNT(a.id), ac.created_at, ac.updated_at
		FROM asset_categories ac
		JOIN asset_types at ON at.id = ac.asset_type_id
		LEFT JOIN assets a ON a.asset_category_id = ac.id
		GROUP BY
			ac.id, ac.asset_type_id, at.name, ac.code, ac.category_prefix,
			ac.name, ac.description, ac.is_active, ac.created_at, ac.updated_at
		ORDER BY at.name, ac.name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.AssetCategory
	for rows.Next() {
		var item models.AssetCategory
		var isActive int
		var createdAt, updatedAt sql.NullTime
		if err := rows.Scan(
			&item.ID, &item.AssetTypeID, &item.AssetTypeName, &item.Code,
			&item.CategoryPrefix, &item.Name, &item.Description, &isActive,
			&item.AssetCount, &createdAt, &updatedAt,
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

func (r *AssetRepository) CreateAssetCategory(input models.AssetCategoryInput) error {
	_, err := r.DB.Exec(`
		INSERT INTO asset_categories (
			asset_type_id, code, category_prefix, name, description, is_active
		) VALUES (?, ?, ?, ?, ?, ?)
	`, input.AssetTypeID, input.Code, input.CategoryPrefix, input.Name,
		nullableString(input.Description), boolToInt(input.IsActive))
	return err
}

func (r *AssetRepository) UpdateAssetCategory(input models.AssetCategoryInput) error {
	var currentTypeID int64
	var linkedAssets int
	if err := r.DB.QueryRow(`
		SELECT ac.asset_type_id, COUNT(a.id)
		FROM asset_categories ac
		LEFT JOIN assets a ON a.asset_category_id = ac.id
		WHERE ac.id = ?
		GROUP BY ac.id, ac.asset_type_id
	`, input.ID).Scan(&currentTypeID, &linkedAssets); err != nil {
		return err
	}
	if linkedAssets > 0 && currentTypeID != input.AssetTypeID {
		return errors.New("tipe kategori tidak dapat diubah karena sudah digunakan oleh aset")
	}
	_, err := r.DB.Exec(`
		UPDATE asset_categories
		SET asset_type_id = ?, code = ?, category_prefix = ?, name = ?,
			description = ?, is_active = ?
		WHERE id = ?
	`, input.AssetTypeID, input.Code, input.CategoryPrefix, input.Name,
		nullableString(input.Description), boolToInt(input.IsActive), input.ID)
	return err
}

func (r *AssetRepository) DeleteAssetCategory(id int64) error {
	_, err := r.DB.Exec(`DELETE FROM asset_categories WHERE id = ?`, id)
	return err
}

func (r *AssetRepository) AssetCategoryBelongsToType(categoryID, assetTypeID int64) (bool, error) {
	var count int
	err := r.DB.QueryRow(`
		SELECT COUNT(*)
		FROM asset_categories
		WHERE id = ? AND asset_type_id = ?
	`, categoryID, assetTypeID).Scan(&count)
	return count > 0, err
}

func (r *AssetRepository) LocationBelongsToStore(locationID int64, storeID int) (bool, error) {
	if locationID <= 0 || storeID <= 0 {
		return true, nil
	}
	var count int
	err := r.DB.QueryRow(`
		SELECT COUNT(*)
		FROM asset_locations
		WHERE id = ? AND store_id = ? AND is_active = 1
	`, locationID, storeID).Scan(&count)
	return count > 0, err
}

func (r *AssetRepository) GetLocationStoreID(locationID int64) (int, error) {
	if locationID <= 0 {
		return 0, nil
	}
	var storeID sql.NullInt64
	if err := r.DB.QueryRow(`SELECT store_id FROM asset_locations WHERE id = ? AND is_active = 1`, locationID).Scan(&storeID); err != nil {
		return 0, err
	}
	if !storeID.Valid {
		return 0, nil
	}
	return int(storeID.Int64), nil
}

func (r *AssetRepository) GetComponentTypes() ([]models.ComponentType, error) {
	rows, err := r.DB.Query(`
		SELECT
			ct.id,
			ct.code,
			ct.name,
			COALESCE(ct.description, ''),
			ct.is_active,
			COUNT(ac.id) AS component_count,
			ct.created_at,
			ct.updated_at
		FROM component_types ct
		LEFT JOIN asset_components ac ON ac.component_type_id = ct.id
		GROUP BY ct.id, ct.code, ct.name, ct.description, ct.is_active, ct.created_at, ct.updated_at
		ORDER BY ct.code ASC
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
		var updatedAt sql.NullTime
		if err := rows.Scan(&item.ID, &item.Code, &item.Name, &item.Description, &isActive, &item.ComponentCount, &createdAt, &updatedAt); err != nil {
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
