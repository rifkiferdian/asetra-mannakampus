package repositories

import (
	"errors"
	"gobase-app/models"
	"strings"
)

func (r *AssetRepository) GetStoreOptions(scopes ...models.AccessScope) ([]models.Store, error) {
	query := `
		SELECT store_id, store_code, store_name
		FROM stores`
	args := make([]any, 0)
	if len(scopes) > 0 {
		where, scopeArgs := assetStoreScopeSQL("store_id", scopes[0])
		query += ` WHERE ` + where
		args = append(args, scopeArgs...)
	}
	query += ` ORDER BY store_name ASC`
	rows, err := r.DB.Query(query, args...)
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
		SELECT id, name,
			CONCAT(COALESCE(name, ''), CASE WHEN nip IS NULL OR CAST(nip AS CHAR) = '' THEN '' ELSE CONCAT(' - ', CAST(nip AS CHAR)) END)
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
		if err := rows.Scan(&item.ID, &item.Name, &item.Label); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *AssetRepository) GetAssetOptions(scopes ...models.AccessScope) ([]models.AssetSelectOption, error) {
	query := `
		SELECT id, asset_code, asset_name
		FROM assets
		WHERE status <> 'DISPOSED'`
	args := make([]any, 0)
	if len(scopes) > 0 {
		where, scopeArgs := assetStoreScopeSQL("store_id", scopes[0])
		query += ` AND ` + where
		args = append(args, scopeArgs...)
	}
	query += ` ORDER BY asset_code ASC`
	rows, err := r.DB.Query(query, args...)
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

func assetStoreScopeSQL(column string, scope models.AccessScope) (string, []any) {
	if scope.CanViewAll {
		return "1=1", nil
	}
	if len(scope.StoreIDs) == 0 {
		return "1=0", nil
	}
	placeholders := make([]string, 0, len(scope.StoreIDs))
	args := make([]any, 0, len(scope.StoreIDs))
	for _, storeID := range scope.StoreIDs {
		placeholders = append(placeholders, "?")
		args = append(args, storeID)
	}
	return column + " IN (" + strings.Join(placeholders, ",") + ")", args
}

func (r *AssetRepository) GetComponentOptions(scopes ...models.AccessScope) ([]models.AssetSelectOption, error) {
	query := `SELECT c.id, c.component_code, c.component_name
		FROM asset_components c
		LEFT JOIN assets a ON a.id = c.parent_asset_id
		LEFT JOIN asset_locations al ON al.id = c.location_id
		WHERE c.status <> 'DISPOSED'`
	args := make([]any, 0)
	if len(scopes) > 0 {
		where, scopeArgs := assetStoreScopeSQL("COALESCE(a.store_id, al.store_id)", scopes[0])
		query += ` AND ` + where
		args = append(args, scopeArgs...)
	}
	query += ` ORDER BY c.component_code ASC`
	rows, err := r.DB.Query(query, args...)
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
