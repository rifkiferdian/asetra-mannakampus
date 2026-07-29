package repositories

import (
	"database/sql"
	"errors"
	"fmt"
	"gobase-app/models"
	"strconv"
)

type CatalogRepository struct {
	DB *sql.DB
}

func (r *CatalogRepository) GetItemTypes() ([]models.CatalogItemType, error) {
	rows, err := r.DB.Query(`
		SELECT
			item_type.id, item_type.code, item_type.name,
			COALESCE(item_type.description, ''), item_type.is_active,
			COUNT(item.id), item_type.updated_at
		FROM catalog_item_types item_type
		LEFT JOIN catalog_items item ON item.item_type_id = item_type.id
		GROUP BY
			item_type.id, item_type.code, item_type.name, item_type.description,
			item_type.is_active, item_type.updated_at
		ORDER BY item_type.name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.CatalogItemType
	for rows.Next() {
		var item models.CatalogItemType
		var isActive int
		var updatedAt sql.NullTime
		if err := rows.Scan(
			&item.ID, &item.Code, &item.Name, &item.Description,
			&isActive, &item.ItemCount, &updatedAt,
		); err != nil {
			return nil, err
		}
		item.IsActive = isActive == 1
		item.UpdatedAtDisplay = formatNullTime(updatedAt)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *CatalogRepository) CreateItemType(input models.CatalogItemTypeInput) error {
	exists, err := r.itemTypeCodeExists(input.Code, 0)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("kode jenis item sudah digunakan")
	}
	_, err = r.DB.Exec(`
		INSERT INTO catalog_item_types (code, name, description, is_active)
		VALUES (?, ?, ?, ?)
	`, input.Code, input.Name, nullableString(input.Description), boolToInt(input.IsActive))
	return err
}

func (r *CatalogRepository) UpdateItemType(input models.CatalogItemTypeInput) error {
	exists, err := r.itemTypeCodeExists(input.Code, input.ID)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("kode jenis item sudah digunakan")
	}
	_, err = r.DB.Exec(`
		UPDATE catalog_item_types
		SET code = ?, name = ?, description = ?, is_active = ?
		WHERE id = ?
	`, input.Code, input.Name, nullableString(input.Description), boolToInt(input.IsActive), input.ID)
	return err
}

func (r *CatalogRepository) DeleteItemType(id int64) error {
	var itemCount int
	if err := r.DB.QueryRow(`SELECT COUNT(*) FROM catalog_items WHERE item_type_id = ?`, id).Scan(&itemCount); err != nil {
		return err
	}
	if itemCount > 0 {
		return fmt.Errorf("jenis item masih digunakan oleh %d item", itemCount)
	}
	result, err := r.DB.Exec(`DELETE FROM catalog_item_types WHERE id = ?`, id)
	if err != nil {
		return err
	}
	return ensureAffected(result, "jenis item tidak ditemukan")
}

func (r *CatalogRepository) GetCategories() ([]models.CatalogItemCategory, error) {
	rows, err := r.DB.Query(`
		SELECT
			category.id, category.code, category.name,
			COALESCE(category.parent_id, 0), COALESCE(parent.name, ''),
			COALESCE(category.owner_division_id, 0), COALESCE(division.division_name, ''),
			COALESCE(category.description, ''), category.is_active,
			COUNT(DISTINCT item.id), COUNT(DISTINCT child.id), category.updated_at
		FROM catalog_item_categories category
		LEFT JOIN catalog_item_categories parent ON parent.id = category.parent_id
		LEFT JOIN divisions division ON division.id = category.owner_division_id
		LEFT JOIN catalog_items item ON item.category_id = category.id
		LEFT JOIN catalog_item_categories child ON child.parent_id = category.id
		GROUP BY
			category.id, category.code, category.name, category.parent_id, parent.name,
			category.owner_division_id, division.division_name, category.description,
			category.is_active, category.updated_at
		ORDER BY category.name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.CatalogItemCategory
	for rows.Next() {
		var item models.CatalogItemCategory
		var isActive int
		var updatedAt sql.NullTime
		if err := rows.Scan(
			&item.ID, &item.Code, &item.Name, &item.ParentID, &item.ParentName,
			&item.OwnerDivisionID, &item.OwnerDivisionName, &item.Description,
			&isActive, &item.ItemCount, &item.ChildCount, &updatedAt,
		); err != nil {
			return nil, err
		}
		item.IsActive = isActive == 1
		item.UpdatedAtDisplay = formatNullTime(updatedAt)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *CatalogRepository) CreateCategory(input models.CatalogItemCategoryInput) error {
	exists, err := r.categoryCodeExists(input.Code, 0)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("kode kategori sudah digunakan")
	}
	_, err = r.DB.Exec(`
		INSERT INTO catalog_item_categories (
			code, name, parent_id, owner_division_id, description, is_active
		) VALUES (?, ?, ?, ?, ?, ?)
	`, input.Code, input.Name, nullablePositiveInt64(input.ParentID), nullableInt(input.OwnerDivisionID),
		nullableString(input.Description), boolToInt(input.IsActive))
	return err
}

func (r *CatalogRepository) UpdateCategory(input models.CatalogItemCategoryInput) error {
	exists, err := r.categoryCodeExists(input.Code, input.ID)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("kode kategori sudah digunakan")
	}
	cycle, err := r.categoryParentCreatesCycle(input.ID, input.ParentID)
	if err != nil {
		return err
	}
	if cycle {
		return errors.New("parent kategori akan membuat relasi berulang")
	}
	_, err = r.DB.Exec(`
		UPDATE catalog_item_categories
		SET code = ?, name = ?, parent_id = ?, owner_division_id = ?,
			description = ?, is_active = ?
		WHERE id = ?
	`, input.Code, input.Name, nullablePositiveInt64(input.ParentID), nullableInt(input.OwnerDivisionID),
		nullableString(input.Description), boolToInt(input.IsActive), input.ID)
	return err
}

func (r *CatalogRepository) DeleteCategory(id int64) error {
	var itemCount, childCount, packageCount int
	if err := r.DB.QueryRow(`
		SELECT
			(SELECT COUNT(*) FROM catalog_items WHERE category_id = ?),
			(SELECT COUNT(*) FROM catalog_item_categories WHERE parent_id = ?),
			(SELECT COUNT(*) FROM catalog_packages WHERE category_id = ?)
	`, id, id, id).Scan(&itemCount, &childCount, &packageCount); err != nil {
		return err
	}
	if itemCount > 0 || childCount > 0 || packageCount > 0 {
		return fmt.Errorf("kategori masih digunakan oleh %d item, %d subkategori, atau %d paket", itemCount, childCount, packageCount)
	}
	result, err := r.DB.Exec(`DELETE FROM catalog_item_categories WHERE id = ?`, id)
	if err != nil {
		return err
	}
	return ensureAffected(result, "kategori tidak ditemukan")
}

func (r *CatalogRepository) GetItems() ([]models.CatalogItem, error) {
	rows, err := r.DB.Query(`
		SELECT
			item.id, item.item_code, item.category_id, category.name,
			item.item_type_id, item_type.code, item_type.name,
			COALESCE(item.asset_type_id, 0), COALESCE(asset_type.name, ''),
			item.name, item.uom, COALESCE(item.description, ''),
			item.is_asset_candidate, item.is_active,
			COUNT(DISTINCT detail.id), COUNT(DISTINCT package_item.package_id),
			COUNT(DISTINCT price.id), item.updated_at
		FROM catalog_items item
		JOIN catalog_item_categories category ON category.id = item.category_id
		JOIN catalog_item_types item_type ON item_type.id = item.item_type_id
		LEFT JOIN asset_types asset_type ON asset_type.id = item.asset_type_id
		LEFT JOIN catalog_item_details detail ON detail.item_id = item.id
		LEFT JOIN catalog_package_items package_item ON package_item.item_id = item.id
		LEFT JOIN vendor_item_prices price ON price.item_id = item.id
		GROUP BY
			item.id, item.item_code, item.category_id, category.name,
			item.item_type_id, item_type.code, item_type.name,
			item.asset_type_id, asset_type.name, item.name, item.uom,
			item.description, item.is_asset_candidate, item.is_active, item.updated_at
		ORDER BY item.name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.CatalogItem
	for rows.Next() {
		var item models.CatalogItem
		var isAssetCandidate, isActive int
		var updatedAt sql.NullTime
		if err := rows.Scan(
			&item.ID, &item.ItemCode, &item.CategoryID, &item.CategoryName,
			&item.ItemTypeID, &item.ItemTypeCode, &item.ItemTypeName,
			&item.AssetTypeID, &item.AssetTypeName, &item.Name, &item.UOM,
			&item.Description, &isAssetCandidate, &isActive, &item.DetailCount,
			&item.PackageCount, &item.PriceCount, &updatedAt,
		); err != nil {
			return nil, err
		}
		item.IsAssetCandidate = isAssetCandidate == 1
		item.IsActive = isActive == 1
		item.UpdatedAtDisplay = formatNullTime(updatedAt)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *CatalogRepository) CreateItem(input models.CatalogItemInput) error {
	exists, err := r.itemCodeExists(input.ItemCode, 0)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("kode item sudah digunakan")
	}
	_, err = r.DB.Exec(`
		INSERT INTO catalog_items (
			item_code, category_id, item_type_id, asset_type_id, name, uom,
			description, is_asset_candidate, is_active
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, input.ItemCode, input.CategoryID, input.ItemTypeID, nullablePositiveInt64(input.AssetTypeID),
		input.Name, input.UOM, nullableString(input.Description),
		boolToInt(input.IsAssetCandidate), boolToInt(input.IsActive))
	return err
}

func (r *CatalogRepository) UpdateItem(input models.CatalogItemInput) error {
	exists, err := r.itemCodeExists(input.ItemCode, input.ID)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("kode item sudah digunakan")
	}
	_, err = r.DB.Exec(`
		UPDATE catalog_items
		SET item_code = ?, category_id = ?, item_type_id = ?, asset_type_id = ?,
			name = ?, uom = ?, description = ?, is_asset_candidate = ?, is_active = ?
		WHERE id = ?
	`, input.ItemCode, input.CategoryID, input.ItemTypeID, nullablePositiveInt64(input.AssetTypeID),
		input.Name, input.UOM, nullableString(input.Description),
		boolToInt(input.IsAssetCandidate), boolToInt(input.IsActive), input.ID)
	return err
}

func (r *CatalogRepository) DeleteItem(id int64) error {
	var packageCount, prCount, priceCount int
	if err := r.DB.QueryRow(`
		SELECT
			(SELECT COUNT(*) FROM catalog_package_items WHERE item_id = ?),
			(SELECT COUNT(*) FROM purchase_request_items WHERE catalog_item_id = ?),
			(SELECT COUNT(*) FROM vendor_item_prices WHERE item_id = ?)
	`, id, id, id).Scan(&packageCount, &prCount, &priceCount); err != nil {
		return err
	}
	if packageCount > 0 || prCount > 0 || priceCount > 0 {
		return fmt.Errorf("item masih digunakan oleh %d paket, %d item PR, atau %d harga vendor", packageCount, prCount, priceCount)
	}
	result, err := r.DB.Exec(`DELETE FROM catalog_items WHERE id = ?`, id)
	if err != nil {
		return err
	}
	return ensureAffected(result, "item tidak ditemukan")
}

func (r *CatalogRepository) GetItemDetails() ([]models.CatalogItemDetail, error) {
	rows, err := r.DB.Query(`
		SELECT
			detail.id, detail.item_id, item.item_code, item.name, category.name,
			detail.detail_name, detail.detail_value, COALESCE(detail.unit, ''),
			detail.sort_order, detail.updated_at
		FROM catalog_item_details detail
		JOIN catalog_items item ON item.id = detail.item_id
		JOIN catalog_item_categories category ON category.id = item.category_id
		ORDER BY item.name, detail.sort_order, detail.detail_name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.CatalogItemDetail
	for rows.Next() {
		var item models.CatalogItemDetail
		var updatedAt sql.NullTime
		if err := rows.Scan(
			&item.ID, &item.ItemID, &item.ItemCode, &item.ItemName, &item.CategoryName,
			&item.DetailName, &item.DetailValue, &item.Unit, &item.SortOrder, &updatedAt,
		); err != nil {
			return nil, err
		}
		item.UpdatedAtDisplay = formatNullTime(updatedAt)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *CatalogRepository) CreateItemDetail(input models.CatalogItemDetailInput) error {
	exists, err := r.detailNameExists(input.ItemID, input.DetailName, 0)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("nama detail sudah digunakan pada item tersebut")
	}
	_, err = r.DB.Exec(`
		INSERT INTO catalog_item_details (
			item_id, detail_name, detail_value, unit, sort_order
		) VALUES (?, ?, ?, ?, ?)
	`, input.ItemID, input.DetailName, input.DetailValue, nullableString(input.Unit), input.SortOrder)
	return err
}

func (r *CatalogRepository) UpdateItemDetail(input models.CatalogItemDetailInput) error {
	exists, err := r.detailNameExists(input.ItemID, input.DetailName, input.ID)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("nama detail sudah digunakan pada item tersebut")
	}
	_, err = r.DB.Exec(`
		UPDATE catalog_item_details
		SET item_id = ?, detail_name = ?, detail_value = ?, unit = ?, sort_order = ?
		WHERE id = ?
	`, input.ItemID, input.DetailName, input.DetailValue, nullableString(input.Unit), input.SortOrder, input.ID)
	return err
}

func (r *CatalogRepository) DeleteItemDetail(id int64) error {
	result, err := r.DB.Exec(`DELETE FROM catalog_item_details WHERE id = ?`, id)
	if err != nil {
		return err
	}
	return ensureAffected(result, "detail item tidak ditemukan")
}

func (r *CatalogRepository) GetDivisions() ([]models.Division, error) {
	return (&DivisionRepository{DB: r.DB}).GetAll()
}

func (r *CatalogRepository) GetAssetTypes() ([]models.AssetType, error) {
	return (&AssetRepository{DB: r.DB}).GetAssetTypes()
}

func (r *CatalogRepository) GetPackages() ([]models.CatalogPackage, error) {
	rows, err := r.DB.Query(`
		SELECT
			package.id, package.package_code, package.name,
			COALESCE(package.category_id, 0), COALESCE(category.name, ''),
			COALESCE(package.owner_division_id, 0), COALESCE(division.division_name, ''),
			COALESCE(package.description, ''), package.is_active,
			COUNT(package_item.id),
			COALESCE(SUM(package_item.is_optional = 0), 0),
			COALESCE(SUM(package_item.is_optional = 1), 0),
			package.updated_at
		FROM catalog_packages package
		LEFT JOIN catalog_item_categories category ON category.id = package.category_id
		LEFT JOIN divisions division ON division.id = package.owner_division_id
		LEFT JOIN catalog_package_items package_item ON package_item.package_id = package.id
		GROUP BY
			package.id, package.package_code, package.name, package.category_id,
			category.name, package.owner_division_id, division.division_name,
			package.description, package.is_active, package.updated_at
		ORDER BY package.name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.CatalogPackage
	for rows.Next() {
		var item models.CatalogPackage
		var isActive int
		var updatedAt sql.NullTime
		if err := rows.Scan(
			&item.ID, &item.PackageCode, &item.Name, &item.CategoryID, &item.CategoryName,
			&item.OwnerDivisionID, &item.OwnerDivisionName, &item.Description,
			&isActive, &item.ItemCount, &item.RequiredCount, &item.OptionalCount, &updatedAt,
		); err != nil {
			return nil, err
		}
		item.IsActive = isActive == 1
		item.UpdatedAtDisplay = formatNullTime(updatedAt)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *CatalogRepository) CreatePackage(input models.CatalogPackageInput) error {
	exists, err := r.packageCodeExists(input.PackageCode, 0)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("kode paket sudah digunakan")
	}
	_, err = r.DB.Exec(`
		INSERT INTO catalog_packages (
			package_code, name, category_id, owner_division_id, description, is_active
		) VALUES (?, ?, ?, ?, ?, ?)
	`, input.PackageCode, input.Name, nullablePositiveInt64(input.CategoryID),
		nullableInt(input.OwnerDivisionID), nullableString(input.Description), boolToInt(input.IsActive))
	return err
}

func (r *CatalogRepository) UpdatePackage(input models.CatalogPackageInput) error {
	exists, err := r.packageCodeExists(input.PackageCode, input.ID)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("kode paket sudah digunakan")
	}
	_, err = r.DB.Exec(`
		UPDATE catalog_packages
		SET package_code = ?, name = ?, category_id = ?, owner_division_id = ?,
			description = ?, is_active = ?
		WHERE id = ?
	`, input.PackageCode, input.Name, nullablePositiveInt64(input.CategoryID),
		nullableInt(input.OwnerDivisionID), nullableString(input.Description),
		boolToInt(input.IsActive), input.ID)
	return err
}

func (r *CatalogRepository) DeletePackage(id int64) error {
	var prCount int
	if err := r.DB.QueryRow(`SELECT COUNT(*) FROM purchase_request_items WHERE source_package_id = ?`, id).Scan(&prCount); err != nil {
		return err
	}
	if prCount > 0 {
		return fmt.Errorf("paket sudah digunakan oleh %d item PR dan tidak dapat dihapus", prCount)
	}
	result, err := r.DB.Exec(`DELETE FROM catalog_packages WHERE id = ?`, id)
	if err != nil {
		return err
	}
	return ensureAffected(result, "paket tidak ditemukan")
}

func (r *CatalogRepository) GetPackageItems() ([]models.CatalogPackageItem, error) {
	rows, err := r.DB.Query(`
		SELECT
			package_item.id, package_item.package_id, package.package_code, package.name,
			package_item.item_id, item.item_code, item.name, item_type.code,
			category.name, package_item.qty, item.uom, package_item.is_optional,
			package_item.sort_order, COALESCE(package_item.notes, ''), package_item.updated_at
		FROM catalog_package_items package_item
		JOIN catalog_packages package ON package.id = package_item.package_id
		JOIN catalog_items item ON item.id = package_item.item_id
		JOIN catalog_item_types item_type ON item_type.id = item.item_type_id
		JOIN catalog_item_categories category ON category.id = item.category_id
		ORDER BY package.name, package_item.sort_order, item.name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.CatalogPackageItem
	for rows.Next() {
		var item models.CatalogPackageItem
		var isOptional int
		var updatedAt sql.NullTime
		if err := rows.Scan(
			&item.ID, &item.PackageID, &item.PackageCode, &item.PackageName,
			&item.ItemID, &item.ItemCode, &item.ItemName, &item.ItemTypeCode,
			&item.CategoryName, &item.Qty, &item.UOM, &isOptional,
			&item.SortOrder, &item.Notes, &updatedAt,
		); err != nil {
			return nil, err
		}
		item.IsOptional = isOptional == 1
		item.QtyDisplay = formatQtyLocal(item.Qty)
		item.UpdatedAtDisplay = formatNullTime(updatedAt)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *CatalogRepository) CreatePackageItem(input models.CatalogPackageItemInput) error {
	exists, err := r.packageItemExists(input.PackageID, input.ItemID, 0)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("item tersebut sudah ada di dalam paket")
	}
	_, err = r.DB.Exec(`
		INSERT INTO catalog_package_items (
			package_id, item_id, qty, is_optional, sort_order, notes
		) VALUES (?, ?, ?, ?, ?, ?)
	`, input.PackageID, input.ItemID, input.Qty, boolToInt(input.IsOptional),
		input.SortOrder, nullableString(input.Notes))
	return err
}

func (r *CatalogRepository) UpdatePackageItem(input models.CatalogPackageItemInput) error {
	exists, err := r.packageItemExists(input.PackageID, input.ItemID, input.ID)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("item tersebut sudah ada di dalam paket")
	}
	_, err = r.DB.Exec(`
		UPDATE catalog_package_items
		SET package_id = ?, item_id = ?, qty = ?, is_optional = ?, sort_order = ?, notes = ?
		WHERE id = ?
	`, input.PackageID, input.ItemID, input.Qty, boolToInt(input.IsOptional),
		input.SortOrder, nullableString(input.Notes), input.ID)
	return err
}

func (r *CatalogRepository) DeletePackageItem(id int64) error {
	var prCount int
	if err := r.DB.QueryRow(`SELECT COUNT(*) FROM purchase_request_items WHERE source_package_item_id = ?`, id).Scan(&prCount); err != nil {
		return err
	}
	if prCount > 0 {
		return fmt.Errorf("isi paket sudah digunakan oleh %d item PR dan tidak dapat dihapus", prCount)
	}
	result, err := r.DB.Exec(`DELETE FROM catalog_package_items WHERE id = ?`, id)
	if err != nil {
		return err
	}
	return ensureAffected(result, "isi paket tidak ditemukan")
}

func (r *CatalogRepository) GetVendorItemPrices() ([]models.VendorItemPrice, error) {
	rows, err := r.DB.Query(`
		SELECT
			price.id, price.vendor_id, vendor.name,
			price.item_id, item.item_code, item.name, category.name, item.uom,
			price.unit_price, price.currency_code, price.minimum_qty,
			price.valid_from, price.valid_until, COALESCE(price.lead_time_days, 0),
			COALESCE(price.quotation_reference, ''),
			price.is_preferred, price.is_active,
			(price.is_active = 1 AND price.valid_from <= CURDATE()
				AND (price.valid_until IS NULL OR price.valid_until >= CURDATE())),
			COALESCE(price.notes, ''), price.updated_at
		FROM vendor_item_prices price
		JOIN vendors vendor ON vendor.id = price.vendor_id
		JOIN catalog_items item ON item.id = price.item_id
		JOIN catalog_item_categories category ON category.id = item.category_id
		ORDER BY item.name, price.is_preferred DESC, price.unit_price, vendor.name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prices []models.VendorItemPrice
	for rows.Next() {
		var price models.VendorItemPrice
		var validFrom, updatedAt sql.NullTime
		var validUntil sql.NullTime
		var isPreferred, isActive, isCurrentlyValid int
		if err := rows.Scan(
			&price.ID, &price.VendorID, &price.VendorName,
			&price.ItemID, &price.ItemCode, &price.ItemName, &price.CategoryName, &price.UOM,
			&price.UnitPrice, &price.CurrencyCode, &price.MinimumQty,
			&validFrom, &validUntil, &price.LeadTimeDays, &price.QuotationReference,
			&isPreferred, &isActive, &isCurrentlyValid, &price.Notes, &updatedAt,
		); err != nil {
			return nil, err
		}
		price.UnitPriceInput = strconv.FormatFloat(price.UnitPrice, 'f', -1, 64)
		price.UnitPriceDisplay = formatAmountIDLocal(price.UnitPrice)
		price.MinimumQtyDisplay = formatQtyLocal(price.MinimumQty)
		if validFrom.Valid {
			price.ValidFrom = validFrom.Time.Format("2006-01-02")
			price.ValidFromDisplay = validFrom.Time.Format("02 Jan 2006")
		}
		if validUntil.Valid {
			price.ValidUntil = validUntil.Time.Format("2006-01-02")
			price.ValidUntilDisplay = validUntil.Time.Format("02 Jan 2006")
		} else {
			price.ValidUntilDisplay = "Tanpa batas"
		}
		price.IsPreferred = isPreferred == 1
		price.IsActive = isActive == 1
		price.IsCurrentlyValid = isCurrentlyValid == 1
		price.UpdatedAtDisplay = formatNullTime(updatedAt)
		prices = append(prices, price)
	}
	return prices, rows.Err()
}

func (r *CatalogRepository) GetVendors() ([]models.Vendor, error) {
	return (&VendorRepository{DB: r.DB}).GetAll()
}

func (r *CatalogRepository) CreateVendorItemPrice(input models.VendorItemPriceInput) error {
	exists, err := r.vendorItemPriceExists(input.VendorID, input.ItemID, input.ValidFrom, 0)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("harga vendor untuk item dan tanggal mulai tersebut sudah tersedia")
	}
	return r.saveVendorItemPrice(input, false)
}

func (r *CatalogRepository) UpdateVendorItemPrice(input models.VendorItemPriceInput) error {
	exists, err := r.vendorItemPriceExists(input.VendorID, input.ItemID, input.ValidFrom, input.ID)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("harga vendor untuk item dan tanggal mulai tersebut sudah tersedia")
	}
	return r.saveVendorItemPrice(input, true)
}

func (r *CatalogRepository) saveVendorItemPrice(input models.VendorItemPriceInput, update bool) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if input.IsPreferred {
		if _, err := tx.Exec(`
			UPDATE vendor_item_prices
			SET is_preferred = 0
			WHERE item_id = ? AND id <> ?
		`, input.ItemID, input.ID); err != nil {
			return err
		}
	}

	if update {
		_, err = tx.Exec(`
			UPDATE vendor_item_prices
			SET vendor_id = ?, item_id = ?, unit_price = ?, currency_code = ?,
				minimum_qty = ?, valid_from = ?, valid_until = ?, lead_time_days = ?,
				quotation_reference = ?, is_preferred = ?, is_active = ?, notes = ?
			WHERE id = ?
		`, input.VendorID, input.ItemID, input.UnitPrice, input.CurrencyCode,
			input.MinimumQty, input.ValidFrom, nullableString(input.ValidUntil),
			nullableInt(input.LeadTimeDays), nullableString(input.QuotationReference),
			boolToInt(input.IsPreferred), boolToInt(input.IsActive),
			nullableString(input.Notes), input.ID)
	} else {
		_, err = tx.Exec(`
			INSERT INTO vendor_item_prices (
				vendor_id, item_id, unit_price, currency_code, minimum_qty,
				valid_from, valid_until, lead_time_days, quotation_reference,
				is_preferred, is_active, notes
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, input.VendorID, input.ItemID, input.UnitPrice, input.CurrencyCode,
			input.MinimumQty, input.ValidFrom, nullableString(input.ValidUntil),
			nullableInt(input.LeadTimeDays), nullableString(input.QuotationReference),
			boolToInt(input.IsPreferred), boolToInt(input.IsActive), nullableString(input.Notes))
	}
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (r *CatalogRepository) DeleteVendorItemPrice(id int64) error {
	result, err := r.DB.Exec(`DELETE FROM vendor_item_prices WHERE id = ?`, id)
	if err != nil {
		return err
	}
	return ensureAffected(result, "harga vendor tidak ditemukan")
}

func (r *CatalogRepository) vendorItemPriceExists(vendorID, itemID int64, validFrom string, exceptID int64) (bool, error) {
	var count int
	err := r.DB.QueryRow(`
		SELECT COUNT(*) FROM vendor_item_prices
		WHERE vendor_id = ? AND item_id = ? AND valid_from = ? AND id <> ?
	`, vendorID, itemID, validFrom, exceptID).Scan(&count)
	return count > 0, err
}

func (r *CatalogRepository) categoryCodeExists(code string, exceptID int64) (bool, error) {
	var count int
	err := r.DB.QueryRow(`
		SELECT COUNT(*) FROM catalog_item_categories
		WHERE code = ? AND (? = 0 OR id <> ?)
	`, code, exceptID, exceptID).Scan(&count)
	return count > 0, err
}

func (r *CatalogRepository) itemTypeCodeExists(code string, exceptID int64) (bool, error) {
	var count int
	err := r.DB.QueryRow(`
		SELECT COUNT(*) FROM catalog_item_types
		WHERE code = ? AND (? = 0 OR id <> ?)
	`, code, exceptID, exceptID).Scan(&count)
	return count > 0, err
}

func (r *CatalogRepository) itemCodeExists(code string, exceptID int64) (bool, error) {
	var count int
	err := r.DB.QueryRow(`
		SELECT COUNT(*) FROM catalog_items
		WHERE item_code = ? AND (? = 0 OR id <> ?)
	`, code, exceptID, exceptID).Scan(&count)
	return count > 0, err
}

func (r *CatalogRepository) detailNameExists(itemID int64, name string, exceptID int64) (bool, error) {
	var count int
	err := r.DB.QueryRow(`
		SELECT COUNT(*) FROM catalog_item_details
		WHERE item_id = ? AND detail_name = ? AND (? = 0 OR id <> ?)
	`, itemID, name, exceptID, exceptID).Scan(&count)
	return count > 0, err
}

func (r *CatalogRepository) packageCodeExists(code string, exceptID int64) (bool, error) {
	var count int
	err := r.DB.QueryRow(`
		SELECT COUNT(*) FROM catalog_packages
		WHERE package_code = ? AND (? = 0 OR id <> ?)
	`, code, exceptID, exceptID).Scan(&count)
	return count > 0, err
}

func (r *CatalogRepository) packageItemExists(packageID, itemID, exceptID int64) (bool, error) {
	var count int
	err := r.DB.QueryRow(`
		SELECT COUNT(*) FROM catalog_package_items
		WHERE package_id = ? AND item_id = ? AND (? = 0 OR id <> ?)
	`, packageID, itemID, exceptID, exceptID).Scan(&count)
	return count > 0, err
}

func (r *CatalogRepository) categoryParentCreatesCycle(id, parentID int64) (bool, error) {
	if parentID <= 0 {
		return false, nil
	}
	current := parentID
	for current > 0 {
		if current == id {
			return true, nil
		}
		var parent sql.NullInt64
		err := r.DB.QueryRow(`SELECT parent_id FROM catalog_item_categories WHERE id = ?`, current).Scan(&parent)
		if err == sql.ErrNoRows {
			return false, errors.New("parent kategori tidak ditemukan")
		}
		if err != nil {
			return false, err
		}
		if !parent.Valid {
			return false, nil
		}
		current = parent.Int64
	}
	return false, nil
}

func ensureAffected(result sql.Result, notFoundMessage string) error {
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return errors.New(notFoundMessage)
	}
	return nil
}
