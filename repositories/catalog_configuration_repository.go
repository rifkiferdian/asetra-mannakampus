package repositories

import (
	"database/sql"
	"errors"
	"fmt"
	"gobase-app/models"
)

func (r *CatalogRepository) GetBrands() ([]models.CatalogBrand, error) {
	rows, err := r.DB.Query(`
		SELECT brand.id, brand.code, brand.name, COALESCE(brand.description, ''),
			brand.is_active, COUNT(variant.id), brand.updated_at
		FROM catalog_brands brand
		LEFT JOIN catalog_item_variants variant ON variant.brand_id = brand.id
		GROUP BY brand.id, brand.code, brand.name, brand.description, brand.is_active, brand.updated_at
		ORDER BY brand.name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.CatalogBrand
	for rows.Next() {
		var item models.CatalogBrand
		var isActive int
		var updatedAt sql.NullTime
		if err := rows.Scan(
			&item.ID, &item.Code, &item.Name, &item.Description,
			&isActive, &item.VariantCount, &updatedAt,
		); err != nil {
			return nil, err
		}
		item.IsActive = isActive == 1
		item.UpdatedAtDisplay = formatNullTime(updatedAt)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *CatalogRepository) CreateBrand(input models.CatalogBrandInput) error {
	exists, err := r.brandExists(input.Code, input.Name, 0)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("kode atau nama merek sudah digunakan")
	}
	_, err = r.DB.Exec(`
		INSERT INTO catalog_brands (code, name, description, is_active)
		VALUES (?, ?, ?, ?)
	`, input.Code, input.Name, nullableString(input.Description), boolToInt(input.IsActive))
	return err
}

func (r *CatalogRepository) UpdateBrand(input models.CatalogBrandInput) error {
	exists, err := r.brandExists(input.Code, input.Name, input.ID)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("kode atau nama merek sudah digunakan")
	}
	_, err = r.DB.Exec(`
		UPDATE catalog_brands
		SET code = ?, name = ?, description = ?, is_active = ?
		WHERE id = ?
	`, input.Code, input.Name, nullableString(input.Description), boolToInt(input.IsActive), input.ID)
	return err
}

func (r *CatalogRepository) DeleteBrand(id int64) error {
	var count int
	if err := r.DB.QueryRow(`SELECT COUNT(*) FROM catalog_item_variants WHERE brand_id = ?`, id).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("merek masih digunakan oleh %d varian", count)
	}
	result, err := r.DB.Exec(`DELETE FROM catalog_brands WHERE id = ?`, id)
	if err != nil {
		return err
	}
	return ensureAffected(result, "merek tidak ditemukan")
}

func (r *CatalogRepository) GetItemVariants() ([]models.CatalogItemVariant, error) {
	rows, err := r.DB.Query(`
		SELECT
			variant.id, variant.item_id, item.item_code, item.name,
			COALESCE(variant.brand_id, 0), COALESCE(brand.name, ''),
			variant.variant_code, variant.model_name,
			COALESCE(variant.manufacturer_sku, ''), COALESCE(variant.specification, ''),
			variant.is_active,
			COUNT(DISTINCT price.id), COUNT(DISTINCT bom_item.id),
			COUNT(DISTINCT package_item.id), variant.updated_at
		FROM catalog_item_variants variant
		JOIN catalog_items item ON item.id = variant.item_id
		LEFT JOIN catalog_brands brand ON brand.id = variant.brand_id
		LEFT JOIN vendor_item_prices price ON price.variant_id = variant.id
		LEFT JOIN catalog_item_bom_items bom_item ON bom_item.variant_id = variant.id
		LEFT JOIN catalog_package_items package_item ON package_item.variant_id = variant.id
		GROUP BY
			variant.id, variant.item_id, item.item_code, item.name,
			variant.brand_id, brand.name, variant.variant_code, variant.model_name,
			variant.manufacturer_sku, variant.specification, variant.is_active, variant.updated_at
		ORDER BY item.name, brand.name, variant.model_name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.CatalogItemVariant
	for rows.Next() {
		var item models.CatalogItemVariant
		var isActive int
		var updatedAt sql.NullTime
		if err := rows.Scan(
			&item.ID, &item.ItemID, &item.ItemCode, &item.ItemName,
			&item.BrandID, &item.BrandName, &item.VariantCode, &item.ModelName,
			&item.ManufacturerSKU, &item.Specification, &isActive,
			&item.PriceCount, &item.BOMUsageCount, &item.PackageCount, &updatedAt,
		); err != nil {
			return nil, err
		}
		item.IsActive = isActive == 1
		item.UpdatedAtDisplay = formatNullTime(updatedAt)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *CatalogRepository) CreateItemVariant(input models.CatalogItemVariantInput) error {
	exists, err := r.variantCodeExists(input.VariantCode, 0)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("kode varian sudah digunakan")
	}
	_, err = r.DB.Exec(`
		INSERT INTO catalog_item_variants (
			item_id, brand_id, variant_code, model_name,
			manufacturer_sku, specification, is_active
		) VALUES (?, ?, ?, ?, ?, ?, ?)
	`, input.ItemID, nullablePositiveInt64(input.BrandID), input.VariantCode,
		input.ModelName, nullableString(input.ManufacturerSKU),
		nullableString(input.Specification), boolToInt(input.IsActive))
	return err
}

func (r *CatalogRepository) UpdateItemVariant(input models.CatalogItemVariantInput) error {
	exists, err := r.variantCodeExists(input.VariantCode, input.ID)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("kode varian sudah digunakan")
	}
	_, err = r.DB.Exec(`
		UPDATE catalog_item_variants
		SET item_id = ?, brand_id = ?, variant_code = ?, model_name = ?,
			manufacturer_sku = ?, specification = ?, is_active = ?
		WHERE id = ?
	`, input.ItemID, nullablePositiveInt64(input.BrandID), input.VariantCode,
		input.ModelName, nullableString(input.ManufacturerSKU),
		nullableString(input.Specification), boolToInt(input.IsActive), input.ID)
	return err
}

func (r *CatalogRepository) DeleteItemVariant(id int64) error {
	var priceCount, bomCount, packageCount, prCount, poCount int
	if err := r.DB.QueryRow(`
		SELECT
			(SELECT COUNT(*) FROM vendor_item_prices WHERE variant_id = ?),
			(SELECT COUNT(*) FROM catalog_item_bom_items WHERE variant_id = ?),
			(SELECT COUNT(*) FROM catalog_package_items WHERE variant_id = ?),
			(SELECT COUNT(*) FROM purchase_request_items WHERE catalog_item_variant_id = ?),
			(SELECT COUNT(*) FROM purchase_order_items WHERE catalog_item_variant_id = ?)
	`, id, id, id, id, id).Scan(&priceCount, &bomCount, &packageCount, &prCount, &poCount); err != nil {
		return err
	}
	if priceCount+bomCount+packageCount+prCount+poCount > 0 {
		return fmt.Errorf(
			"varian masih digunakan oleh %d harga, %d BOM, %d paket, %d item PR, atau %d item PO",
			priceCount, bomCount, packageCount, prCount, poCount,
		)
	}
	result, err := r.DB.Exec(`DELETE FROM catalog_item_variants WHERE id = ?`, id)
	if err != nil {
		return err
	}
	return ensureAffected(result, "varian tidak ditemukan")
}

func (r *CatalogRepository) GetItemBOMs() ([]models.CatalogItemBOM, error) {
	rows, err := r.DB.Query(`
		SELECT
			bom.id, bom.parent_item_id, parent.item_code, parent.name,
			bom.bom_code, bom.name, bom.version_no, COALESCE(bom.description, ''),
			bom.is_active, COUNT(DISTINCT bom_item.id),
			COALESCE(SUM(
				CASE
					WHEN bom_item.id IS NOT NULL AND selected_price.id IS NULL THEN 1
					ELSE 0
				END
			), 0),
			COALESCE(SUM(bom_item.qty * selected_price.unit_price), 0),
			COUNT(DISTINCT package_item.id), bom.updated_at
		FROM catalog_item_boms bom
		JOIN catalog_items parent ON parent.id = bom.parent_item_id
		LEFT JOIN catalog_item_bom_items bom_item ON bom_item.bom_id = bom.id
		LEFT JOIN vendor_item_prices selected_price
			ON selected_price.id = (
				SELECT price_lookup.id
				FROM vendor_item_prices price_lookup
				WHERE price_lookup.variant_id = bom_item.variant_id
				  AND price_lookup.vendor_id = bom_item.preferred_vendor_id
				  AND price_lookup.is_active = 1
				  AND price_lookup.valid_from <= CURDATE()
				  AND (price_lookup.valid_until IS NULL OR price_lookup.valid_until >= CURDATE())
				ORDER BY price_lookup.is_preferred DESC, price_lookup.valid_from DESC, price_lookup.id DESC
				LIMIT 1
			)
		LEFT JOIN catalog_package_items package_item ON package_item.bom_id = bom.id
		GROUP BY
			bom.id, bom.parent_item_id, parent.item_code, parent.name,
			bom.bom_code, bom.name, bom.version_no, bom.description,
			bom.is_active, bom.updated_at
		ORDER BY parent.name, bom.version_no
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.CatalogItemBOM
	for rows.Next() {
		var item models.CatalogItemBOM
		var isActive int
		var updatedAt sql.NullTime
		if err := rows.Scan(
			&item.ID, &item.ParentItemID, &item.ParentItemCode, &item.ParentItemName,
			&item.BOMCode, &item.Name, &item.VersionNo, &item.Description,
			&isActive, &item.ComponentCount, &item.MissingPriceCount,
			&item.EstimatedTotal, &item.PackageCount, &updatedAt,
		); err != nil {
			return nil, err
		}
		item.IsActive = isActive == 1
		item.EstimatedTotalDisplay = formatCatalogAmountLocal(item.EstimatedTotal)
		item.UpdatedAtDisplay = formatNullTime(updatedAt)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *CatalogRepository) CreateItemBOM(input models.CatalogItemBOMInput) error {
	exists, err := r.bomExists(input.BOMCode, input.ParentItemID, input.VersionNo, 0)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("kode BOM atau versi parent tersebut sudah digunakan")
	}
	_, err = r.DB.Exec(`
		INSERT INTO catalog_item_boms (
			parent_item_id, bom_code, name, version_no, description, is_active
		) VALUES (?, ?, ?, ?, ?, ?)
	`, input.ParentItemID, input.BOMCode, input.Name, input.VersionNo,
		nullableString(input.Description), boolToInt(input.IsActive))
	return err
}

func (r *CatalogRepository) UpdateItemBOM(input models.CatalogItemBOMInput) error {
	exists, err := r.bomExists(input.BOMCode, input.ParentItemID, input.VersionNo, input.ID)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("kode BOM atau versi parent tersebut sudah digunakan")
	}
	_, err = r.DB.Exec(`
		UPDATE catalog_item_boms
		SET parent_item_id = ?, bom_code = ?, name = ?, version_no = ?,
			description = ?, is_active = ?
		WHERE id = ?
	`, input.ParentItemID, input.BOMCode, input.Name, input.VersionNo,
		nullableString(input.Description), boolToInt(input.IsActive), input.ID)
	return err
}

func (r *CatalogRepository) DeleteItemBOM(id int64) error {
	var packageCount, prCount int
	if err := r.DB.QueryRow(`
		SELECT
			(SELECT COUNT(*) FROM catalog_package_items WHERE bom_id = ?),
			(SELECT COUNT(*) FROM purchase_request_items WHERE source_bom_id = ?)
	`, id, id).Scan(&packageCount, &prCount); err != nil {
		return err
	}
	if packageCount > 0 || prCount > 0 {
		return fmt.Errorf("BOM masih digunakan oleh %d paket atau %d item PR", packageCount, prCount)
	}
	result, err := r.DB.Exec(`DELETE FROM catalog_item_boms WHERE id = ?`, id)
	if err != nil {
		return err
	}
	return ensureAffected(result, "BOM tidak ditemukan")
}

func (r *CatalogRepository) GetItemBOMItems() ([]models.CatalogItemBOMItem, error) {
	rows, err := r.DB.Query(`
		SELECT
			bom_item.id, bom_item.bom_id, bom.bom_code, bom.name,
			parent.item_code, parent.name,
			bom_item.component_item_id, component.item_code, component.name,
			COALESCE(bom_item.variant_id, 0), COALESCE(variant.variant_code, ''),
			COALESCE(variant.model_name, ''), COALESCE(brand.name, ''),
			COALESCE(bom_item.preferred_vendor_id, 0), COALESCE(vendor.name, ''),
			bom_item.qty, bom_item.is_required, bom_item.sort_order,
			COALESCE(bom_item.notes, ''),
			COALESCE(selected_price.unit_price, 0),
			selected_price.id IS NOT NULL,
			bom_item.updated_at
		FROM catalog_item_bom_items bom_item
		JOIN catalog_item_boms bom ON bom.id = bom_item.bom_id
		JOIN catalog_items parent ON parent.id = bom.parent_item_id
		JOIN catalog_items component ON component.id = bom_item.component_item_id
		LEFT JOIN catalog_item_variants variant ON variant.id = bom_item.variant_id
		LEFT JOIN catalog_brands brand ON brand.id = variant.brand_id
		LEFT JOIN vendors vendor ON vendor.id = bom_item.preferred_vendor_id
		LEFT JOIN vendor_item_prices selected_price
			ON selected_price.id = (
				SELECT price_lookup.id
				FROM vendor_item_prices price_lookup
				WHERE price_lookup.variant_id = bom_item.variant_id
				  AND price_lookup.vendor_id = bom_item.preferred_vendor_id
				  AND price_lookup.is_active = 1
				  AND price_lookup.valid_from <= CURDATE()
				  AND (price_lookup.valid_until IS NULL OR price_lookup.valid_until >= CURDATE())
				ORDER BY price_lookup.is_preferred DESC, price_lookup.valid_from DESC, price_lookup.id DESC
				LIMIT 1
			)
		ORDER BY parent.name, bom.version_no, bom_item.sort_order, component.name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.CatalogItemBOMItem
	for rows.Next() {
		var item models.CatalogItemBOMItem
		var isRequired, hasCurrentPrice int
		var updatedAt sql.NullTime
		if err := rows.Scan(
			&item.ID, &item.BOMID, &item.BOMCode, &item.BOMName,
			&item.ParentItemCode, &item.ParentItemName,
			&item.ComponentItemID, &item.ComponentItemCode, &item.ComponentItemName,
			&item.VariantID, &item.VariantCode, &item.VariantName, &item.BrandName,
			&item.PreferredVendorID, &item.PreferredVendorName,
			&item.Qty, &isRequired, &item.SortOrder, &item.Notes,
			&item.UnitPrice, &hasCurrentPrice, &updatedAt,
		); err != nil {
			return nil, err
		}
		item.IsRequired = isRequired == 1
		item.HasCurrentPrice = hasCurrentPrice == 1
		item.LineTotal = item.Qty * item.UnitPrice
		item.QtyDisplay = formatQtyLocal(item.Qty)
		item.UnitPriceDisplay = formatCatalogAmountLocal(item.UnitPrice)
		item.LineTotalDisplay = formatCatalogAmountLocal(item.LineTotal)
		item.UpdatedAtDisplay = formatNullTime(updatedAt)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *CatalogRepository) CreateItemBOMItem(input models.CatalogItemBOMItemInput) error {
	exists, err := r.bomComponentExists(input.BOMID, input.ComponentItemID, 0)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("komponen tersebut sudah ada pada BOM")
	}
	_, err = r.DB.Exec(`
		INSERT INTO catalog_item_bom_items (
			bom_id, component_item_id, variant_id, preferred_vendor_id,
			qty, is_required, sort_order, notes
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, input.BOMID, input.ComponentItemID, nullablePositiveInt64(input.VariantID),
		nullablePositiveInt64(input.PreferredVendorID), input.Qty,
		boolToInt(input.IsRequired), input.SortOrder, nullableString(input.Notes))
	return err
}

func (r *CatalogRepository) UpdateItemBOMItem(input models.CatalogItemBOMItemInput) error {
	exists, err := r.bomComponentExists(input.BOMID, input.ComponentItemID, input.ID)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("komponen tersebut sudah ada pada BOM")
	}
	_, err = r.DB.Exec(`
		UPDATE catalog_item_bom_items
		SET bom_id = ?, component_item_id = ?, variant_id = ?, preferred_vendor_id = ?,
			qty = ?, is_required = ?, sort_order = ?, notes = ?
		WHERE id = ?
	`, input.BOMID, input.ComponentItemID, nullablePositiveInt64(input.VariantID),
		nullablePositiveInt64(input.PreferredVendorID), input.Qty,
		boolToInt(input.IsRequired), input.SortOrder, nullableString(input.Notes), input.ID)
	return err
}

func (r *CatalogRepository) DeleteItemBOMItem(id int64) error {
	var prCount, poCount int
	if err := r.DB.QueryRow(`
		SELECT
			(SELECT COUNT(*) FROM purchase_request_items WHERE source_bom_item_id = ?),
			(SELECT COUNT(*) FROM purchase_order_items WHERE source_bom_item_id = ?)
	`, id, id).Scan(&prCount, &poCount); err != nil {
		return err
	}
	if prCount > 0 || poCount > 0 {
		return fmt.Errorf("komponen BOM sudah digunakan oleh %d item PR atau %d item PO", prCount, poCount)
	}
	result, err := r.DB.Exec(`DELETE FROM catalog_item_bom_items WHERE id = ?`, id)
	if err != nil {
		return err
	}
	return ensureAffected(result, "komponen BOM tidak ditemukan")
}

func (r *CatalogRepository) GetItemTypeCode(itemID int64) (string, error) {
	var code string
	err := r.DB.QueryRow(`
		SELECT item_type.code
		FROM catalog_items item
		JOIN catalog_item_types item_type ON item_type.id = item.item_type_id
		WHERE item.id = ?
	`, itemID).Scan(&code)
	if err == sql.ErrNoRows {
		return "", errors.New("item tidak ditemukan")
	}
	return code, err
}

func (r *CatalogRepository) GetItemTypeCodeByID(itemTypeID int64) (string, error) {
	var code string
	err := r.DB.QueryRow(`SELECT code FROM catalog_item_types WHERE id = ?`, itemTypeID).Scan(&code)
	if err == sql.ErrNoRows {
		return "", errors.New("jenis item tidak ditemukan")
	}
	return code, err
}

func (r *CatalogRepository) VariantBelongsToItem(variantID, itemID int64) (bool, error) {
	var count int
	err := r.DB.QueryRow(`
		SELECT COUNT(*) FROM catalog_item_variants
		WHERE id = ? AND item_id = ?
	`, variantID, itemID).Scan(&count)
	return count > 0, err
}

func (r *CatalogRepository) BOMBelongsToItem(bomID, itemID int64) (bool, error) {
	var count int
	err := r.DB.QueryRow(`
		SELECT COUNT(*) FROM catalog_item_boms
		WHERE id = ? AND parent_item_id = ?
	`, bomID, itemID).Scan(&count)
	return count > 0, err
}

func (r *CatalogRepository) ValidateBOMItemSelection(input models.CatalogItemBOMItemInput) error {
	var parentItemID int64
	if err := r.DB.QueryRow(`
		SELECT parent_item_id FROM catalog_item_boms WHERE id = ?
	`, input.BOMID).Scan(&parentItemID); err != nil {
		if err == sql.ErrNoRows {
			return errors.New("BOM tidak ditemukan")
		}
		return err
	}
	if parentItemID == input.ComponentItemID {
		return errors.New("item parent tidak boleh menjadi komponennya sendiri")
	}
	ok, err := r.VariantBelongsToItem(input.VariantID, input.ComponentItemID)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("varian tidak sesuai dengan item komponen")
	}
	return nil
}

func (r *CatalogRepository) brandExists(code, name string, exceptID int64) (bool, error) {
	var count int
	err := r.DB.QueryRow(`
		SELECT COUNT(*) FROM catalog_brands
		WHERE (code = ? OR name = ?) AND (? = 0 OR id <> ?)
	`, code, name, exceptID, exceptID).Scan(&count)
	return count > 0, err
}

func (r *CatalogRepository) variantCodeExists(code string, exceptID int64) (bool, error) {
	var count int
	err := r.DB.QueryRow(`
		SELECT COUNT(*) FROM catalog_item_variants
		WHERE variant_code = ? AND (? = 0 OR id <> ?)
	`, code, exceptID, exceptID).Scan(&count)
	return count > 0, err
}

func (r *CatalogRepository) bomExists(code string, parentID int64, version int, exceptID int64) (bool, error) {
	var count int
	err := r.DB.QueryRow(`
		SELECT COUNT(*) FROM catalog_item_boms
		WHERE (bom_code = ? OR (parent_item_id = ? AND version_no = ?))
		  AND (? = 0 OR id <> ?)
	`, code, parentID, version, exceptID, exceptID).Scan(&count)
	return count > 0, err
}

func (r *CatalogRepository) bomComponentExists(bomID, componentItemID, exceptID int64) (bool, error) {
	var count int
	err := r.DB.QueryRow(`
		SELECT COUNT(*) FROM catalog_item_bom_items
		WHERE bom_id = ? AND component_item_id = ? AND (? = 0 OR id <> ?)
	`, bomID, componentItemID, exceptID, exceptID).Scan(&count)
	return count > 0, err
}
