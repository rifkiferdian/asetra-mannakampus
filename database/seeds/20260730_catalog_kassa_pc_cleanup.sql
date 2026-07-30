-- Rapikan data contoh setelah migrasi varian/BOM.
-- Hasil akhir:
--   Paket Kassa PC
--   - PC Kasir (BOM PC Kasir Standar)
--   - Monitor 24 Inch (Samsung 24 Inch IPS, PT Sumber IT)

START TRANSACTION;

DELETE package_item
FROM catalog_package_items package_item
JOIN catalog_packages package ON package.id = package_item.package_id
WHERE package.package_code IN (
    'PKG-PC-KASIR-STD',
    'PKG-MEJA-ADMIN',
    'PKG-KASSA-PC'
);

DELETE FROM catalog_packages
WHERE package_code IN ('PKG-PC-KASIR-STD', 'PKG-MEJA-ADMIN');

INSERT INTO catalog_packages (
    package_code, name, category_id, owner_division_id,
    description, is_active
)
SELECT
    'PKG-KASSA-PC',
    'Paket Kassa PC',
    category.id,
    division.id,
    'Paket satu titik kasir yang terdiri dari PC Kasir rakitan dan monitor.',
    1
FROM catalog_item_categories category
JOIN divisions division ON division.division_code = 'IT'
WHERE category.code = 'IT-EQUIPMENT'
ON DUPLICATE KEY UPDATE
    name = VALUES(name),
    category_id = VALUES(category_id),
    owner_division_id = VALUES(owner_division_id),
    description = VALUES(description),
    is_active = 1;

DELETE price
FROM vendor_item_prices price
JOIN catalog_item_variants variant ON variant.id = price.variant_id
JOIN catalog_items item ON item.id = variant.item_id
LEFT JOIN catalog_brands brand ON brand.id = variant.brand_id
WHERE item.item_code IN (
    'GA-MEJA-ADMIN',
    'GA-KURSI-ADMIN',
    'GA-LEMARI-DOKUMEN',
    'GA-TELEPON-MEJA',
    'GA-TEMPAT-SAMPAH',
    'IT-PRINTER-THERMAL',
    'IT-BARCODE-SCANNER',
    'IT-UPS-650VA'
)
OR brand.code = 'GENERIC';

DELETE variant
FROM catalog_item_variants variant
JOIN catalog_items item ON item.id = variant.item_id
LEFT JOIN catalog_brands brand ON brand.id = variant.brand_id
WHERE item.item_code IN (
    'GA-MEJA-ADMIN',
    'GA-KURSI-ADMIN',
    'GA-LEMARI-DOKUMEN',
    'GA-TELEPON-MEJA',
    'GA-TEMPAT-SAMPAH',
    'IT-PRINTER-THERMAL',
    'IT-BARCODE-SCANNER',
    'IT-UPS-650VA'
)
OR brand.code = 'GENERIC';

DELETE FROM catalog_items
WHERE item_code IN (
    'GA-MEJA-ADMIN',
    'GA-KURSI-ADMIN',
    'GA-LEMARI-DOKUMEN',
    'GA-TELEPON-MEJA',
    'GA-TEMPAT-SAMPAH',
    'IT-PRINTER-THERMAL',
    'IT-BARCODE-SCANNER',
    'IT-UPS-650VA'
);

DELETE category
FROM catalog_item_categories category
LEFT JOIN catalog_items item ON item.category_id = category.id
LEFT JOIN catalog_packages package ON package.category_id = category.id
WHERE category.code = 'OFFICE-FURNITURE'
  AND item.id IS NULL
  AND package.id IS NULL;

DELETE brand
FROM catalog_brands brand
LEFT JOIN catalog_item_variants variant ON variant.brand_id = brand.id
WHERE brand.code = 'GENERIC'
  AND variant.id IS NULL;

UPDATE catalog_items item
JOIN catalog_item_types item_type ON item_type.code = 'COMPOSITE'
SET
    item.item_type_id = item_type.id,
    item.is_purchasable = 0,
    item.is_asset_candidate = 1,
    item.is_active = 1
WHERE item.item_code = 'IT-PC-KASIR';

INSERT INTO catalog_package_items (
    package_id,
    item_id,
    bom_id,
    variant_id,
    preferred_vendor_id,
    qty,
    is_optional,
    sort_order,
    notes
)
SELECT
    package.id,
    item.id,
    bom.id,
    NULL,
    NULL,
    1.00,
    0,
    1,
    'PC utama aplikasi kasir'
FROM catalog_packages package
JOIN catalog_items item ON item.item_code = 'IT-PC-KASIR'
JOIN catalog_item_boms bom
    ON bom.parent_item_id = item.id
   AND bom.bom_code = 'BOM-PC-KASIR-STD'
WHERE package.package_code = 'PKG-KASSA-PC'
ON DUPLICATE KEY UPDATE
    bom_id = VALUES(bom_id),
    variant_id = NULL,
    preferred_vendor_id = NULL,
    qty = VALUES(qty),
    is_optional = VALUES(is_optional),
    sort_order = VALUES(sort_order),
    notes = VALUES(notes);

INSERT INTO catalog_package_items (
    package_id,
    item_id,
    bom_id,
    variant_id,
    preferred_vendor_id,
    qty,
    is_optional,
    sort_order,
    notes
)
SELECT
    package.id,
    item.id,
    NULL,
    variant.id,
    vendor.id,
    1.00,
    0,
    2,
    'Monitor utama kasir'
FROM catalog_packages package
JOIN catalog_items item ON item.item_code = 'IT-MONITOR-24'
JOIN catalog_item_variants variant
    ON variant.item_id = item.id
   AND variant.variant_code = 'VAR-MON-SAMSUNG-24-IPS'
JOIN vendors vendor
    ON vendor.name = 'PT Sumber IT'
   AND vendor.is_active = 1
WHERE package.package_code = 'PKG-KASSA-PC'
ON DUPLICATE KEY UPDATE
    bom_id = NULL,
    variant_id = VALUES(variant_id),
    preferred_vendor_id = VALUES(preferred_vendor_id),
    qty = VALUES(qty),
    is_optional = VALUES(is_optional),
    sort_order = VALUES(sort_order),
    notes = VALUES(notes);

COMMIT;
