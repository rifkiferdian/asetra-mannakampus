-- Contoh minimal katalog sebelum migrasi varian/BOM:
-- 1 paket "Paket Kassa PC" dengan 2 item utama.
-- Migrasi 20260730_catalog_variants_bom.sql akan mengubah PC Kasir
-- menjadi COMPOSITE dan menambahkan komponen BOM-nya.

INSERT INTO catalog_item_categories (
    code, name, parent_id, owner_division_id, description, is_active
)
SELECT
    'IT-EQUIPMENT',
    'Peralatan IT',
    NULL,
    division.id,
    'Komputer dan perangkat pendukung operasional toko.',
    1
FROM divisions division
WHERE division.division_code = 'IT'
ON DUPLICATE KEY UPDATE
    name = VALUES(name),
    owner_division_id = VALUES(owner_division_id),
    description = VALUES(description),
    is_active = 1;

INSERT INTO catalog_items (
    item_code, category_id, item_type_id, asset_type_id,
    name, uom, description, is_asset_candidate, is_active
)
SELECT
    source.item_code,
    category.id,
    item_type.id,
    asset_type.id,
    source.item_name,
    'unit',
    source.description,
    1,
    1
FROM (
    SELECT
        'IT-PC-KASIR' AS item_code,
        'PC Kasir' AS item_name,
        'Komputer utama aplikasi kasir; setelah migrasi nilainya dihitung dari komponen BOM.' AS description
    UNION ALL
    SELECT
        'IT-MONITOR-24',
        'Monitor 24 Inch',
        'Monitor utama untuk satu titik kasir.'
) source
JOIN catalog_item_categories category ON category.code = 'IT-EQUIPMENT'
JOIN catalog_item_types item_type ON item_type.code = 'GOODS'
JOIN asset_types asset_type ON asset_type.code = 'IT'
ON DUPLICATE KEY UPDATE
    category_id = VALUES(category_id),
    item_type_id = VALUES(item_type_id),
    asset_type_id = VALUES(asset_type_id),
    name = VALUES(name),
    uom = VALUES(uom),
    description = VALUES(description),
    is_asset_candidate = VALUES(is_asset_candidate),
    is_active = 1;

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

INSERT INTO catalog_package_items (
    package_id, item_id, qty, is_optional, sort_order, notes
)
SELECT
    package.id,
    item.id,
    1.00,
    0,
    source.sort_order,
    source.notes
FROM (
    SELECT
        'IT-PC-KASIR' AS item_code,
        1 AS sort_order,
        'PC utama aplikasi kasir' AS notes
    UNION ALL
    SELECT
        'IT-MONITOR-24',
        2,
        'Monitor utama kasir'
) source
JOIN catalog_packages package ON package.package_code = 'PKG-KASSA-PC'
JOIN catalog_items item ON item.item_code = source.item_code
ON DUPLICATE KEY UPDATE
    qty = VALUES(qty),
    is_optional = VALUES(is_optional),
    sort_order = VALUES(sort_order),
    notes = VALUES(notes);
