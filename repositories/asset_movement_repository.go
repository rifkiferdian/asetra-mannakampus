package repositories

import (
	"database/sql"
	"gobase-app/models"
)

func (r *AssetRepository) GetAssetMovements(scopes ...models.AccessScope) ([]models.AssetMovement, error) {
	query := `
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
		LEFT JOIN users au ON au.id = m.acted_by`
	args := make([]any, 0)
	if len(scopes) > 0 {
		where, scopeArgs := assetStoreScopeSQL("a.store_id", scopes[0])
		query += ` WHERE ` + where
		args = append(args, scopeArgs...)
	}
	query += ` ORDER BY m.movement_date DESC, m.id DESC`
	rows, err := r.DB.Query(query, args...)
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
	defer tx.Rollback()

	var fromStoreID, fromLocationID sql.NullInt64
	var fromPersonNIP, fromPersonName, fromPersonDepartment sql.NullString
	if err := tx.QueryRow(`
		SELECT store_id, location_id, assigned_person_nip, assigned_person_name, assigned_person_department
		FROM assets
		WHERE id = ?
		FOR UPDATE
	`, input.AssetID).Scan(&fromStoreID, &fromLocationID, &fromPersonNIP, &fromPersonName, &fromPersonDepartment); err != nil {
		tx.Rollback()
		return err
	}

	toPersonNIP := fromPersonNIP
	toPersonName := fromPersonName
	toPersonDepartment := fromPersonDepartment
	if input.ToUserID > 0 {
		toPersonDepartment = sql.NullString{}
		if err := tx.QueryRow(`
			SELECT CAST(nip AS CHAR), name
			FROM users
			WHERE id = ?
		`, input.ToUserID).Scan(&toPersonNIP, &toPersonName); err != nil {
			tx.Rollback()
			return err
		}
	}
	if input.MovementType == "RETURN" || input.MovementType == "DISPOSE" {
		toPersonNIP = sql.NullString{}
		toPersonName = sql.NullString{}
		toPersonDepartment = sql.NullString{}
	}

	if _, err := tx.Exec(`
		INSERT INTO asset_movements (
			asset_id, movement_type, from_store_id, to_store_id, from_location_id, to_location_id,
			from_person_nip, from_person_name, from_person_department,
			to_user_id, to_person_nip, to_person_name, to_person_department,
			acted_by, notes
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, input.AssetID, input.MovementType, nullInt64Value(fromStoreID), nullableInt(input.ToStoreID),
		nullInt64Value(fromLocationID), assetNullableInt64(input.ToLocationID), nullStringValue(fromPersonNIP),
		nullStringValue(fromPersonName), nullStringValue(fromPersonDepartment), nullableInt(input.ToUserID),
		nullStringValue(toPersonNIP), nullStringValue(toPersonName), nullStringValue(toPersonDepartment),
		nullableInt(input.ActedBy), nullableString(input.Notes)); err != nil {
		tx.Rollback()
		return err
	}
	if err := insertAuditLogTx(tx, "ASSET_MOVEMENT", input.AssetID, input.MovementType,
		"Movement aset dicatat", input.AuditContext); err != nil {
		return err
	}

	if _, err := tx.Exec(`
		UPDATE assets
		SET store_id = ?, location_id = ?, assigned_person_nip = ?, assigned_person_name = ?,
			assigned_person_department = ?, status = ?
		WHERE id = ?
	`, nullableInt(input.ToStoreID), assetNullableInt64(input.ToLocationID), nullStringValue(toPersonNIP),
		nullStringValue(toPersonName), nullStringValue(toPersonDepartment), input.ResultStatus, input.AssetID); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

func (r *AssetRepository) GetComponentMovements(scopes ...models.AccessScope) ([]models.AssetComponentMovement, error) {
	query := `
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
		LEFT JOIN users au ON au.id = m.acted_by`
	args := make([]any, 0)
	if len(scopes) > 0 {
		where, scopeArgs := assetStoreScopeSQL("COALESCE(ta.store_id, fa.store_id, tl.store_id, fl.store_id)", scopes[0])
		query += ` WHERE ` + where
		args = append(args, scopeArgs...)
	}
	query += ` ORDER BY m.movement_date DESC, m.id DESC`
	rows, err := r.DB.Query(query, args...)
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

func (r *AssetRepository) GetComponentMovementsByComponentID(componentID int64) ([]models.AssetComponentMovement, error) {
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
		WHERE m.component_id = ?
		ORDER BY m.movement_date DESC, m.id DESC
	`, componentID)
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
	defer tx.Rollback()

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

	if _, err := tx.Exec(`
		UPDATE asset_components
		SET parent_asset_id = ?, location_id = ?, status = ?
		WHERE id = ?
	`, assetNullableInt64(input.ResultAssetID), assetNullableInt64(input.ToLocationID), input.ResultStatus, input.ComponentID); err != nil {
		tx.Rollback()
		return err
	}
	if err := insertAuditLogTx(tx, "ASSET_COMPONENT_MOVEMENT", input.ComponentID, input.MovementType,
		"Movement komponen aset dicatat", input.AuditContext); err != nil {
		return err
	}

	return tx.Commit()
}
