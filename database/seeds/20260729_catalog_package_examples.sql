-- Dua contoh paket katalog:
-- 1. Paket PC Kasir Standar
-- 2. Paket Meja Admin Toko

INSERT INTO catalog_item_categories (
    code, name, parent_id, owner_division_id, description, is_active
)
SELECT
    'IT-EQUIPMENT',
    'Peralatan IT',
    NULL,
    division.id,
    'Komputer, monitor, printer, scanner, dan perangkat pendukung operasional.',
    1
FROM divisions division
WHERE division.division_code = 'IT'
ON DUPLICATE KEY UPDATE
    name = VALUES(name),
    owner_division_id = VALUES(owner_division_id),
    description = VALUES(description),
    is_active = 1;

INSERT INTO catalog_item_categories (
    code, name, parent_id, owner_division_id, description, is_active
)
SELECT
    'OFFICE-FURNITURE',
    'Furnitur dan Peralatan Kantor',
    NULL,
    division.id,
    'Furnitur dan perlengkapan untuk area administrasi toko.',
    1
FROM divisions division
WHERE division.division_code = 'GA'
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
    source.uom,
    source.description,
    source.is_asset_candidate,
    1
FROM (
    SELECT 'IT-PC-KASIR' item_code, 'PC Kasir' item_name, 'unit' uom,
           'Komputer utama untuk aplikasi kasir dan perangkat POS.' description, 1 is_asset_candidate, 'IT' asset_type_code
    UNION ALL
    SELECT 'IT-MONITOR-24', 'Monitor 24 Inch', 'unit',
           'Monitor kerja kasir ukuran 24 inch.', 1, 'IT'
    UNION ALL
    SELECT 'IT-PRINTER-THERMAL', 'Printer Thermal 80 mm', 'unit',
           'Printer struk thermal untuk transaksi kasir.', 1, 'IT'
    UNION ALL
    SELECT 'IT-BARCODE-SCANNER', 'Barcode Scanner', 'unit',
           'Scanner barcode 1D/2D untuk meja kasir.', 1, 'IT'
    UNION ALL
    SELECT 'IT-UPS-650VA', 'UPS 650 VA', 'unit',
           'Cadangan daya untuk perangkat utama kasir.', 1, 'ELECTRONIC'
) source
JOIN catalog_item_categories category ON category.code = 'IT-EQUIPMENT'
JOIN catalog_item_types item_type ON item_type.code = 'GOODS'
JOIN asset_types asset_type ON asset_type.code = source.asset_type_code
ON DUPLICATE KEY UPDATE
    category_id = VALUES(category_id),
    item_type_id = VALUES(item_type_id),
    asset_type_id = VALUES(asset_type_id),
    name = VALUES(name),
    uom = VALUES(uom),
    description = VALUES(description),
    is_asset_candidate = VALUES(is_asset_candidate),
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
    source.uom,
    source.description,
    source.is_asset_candidate,
    1
FROM (
    SELECT 'GA-MEJA-ADMIN' item_code, 'Meja Admin Toko' item_name, 'unit' uom,
           'Meja kerja administrasi dengan laci penyimpanan.' description, 1 is_asset_candidate, 'FURNITURE' asset_type_code
    UNION ALL
    SELECT 'GA-KURSI-ADMIN', 'Kursi Admin Ergonomis', 'unit',
           'Kursi kerja admin dengan sandaran dan pengaturan tinggi.', 1, 'FURNITURE'
    UNION ALL
    SELECT 'GA-LEMARI-DOKUMEN', 'Lemari Dokumen', 'unit',
           'Lemari berkunci untuk penyimpanan dokumen operasional.', 1, 'FURNITURE'
    UNION ALL
    SELECT 'GA-TELEPON-MEJA', 'Telepon Meja', 'unit',
           'Telepon meja untuk komunikasi operasional toko.', 1, 'OFFICE'
    UNION ALL
    SELECT 'GA-TEMPAT-SAMPAH', 'Tempat Sampah Meja', 'unit',
           'Tempat sampah kecil untuk area meja administrasi.', 0, 'OFFICE'
) source
JOIN catalog_item_categories category ON category.code = 'OFFICE-FURNITURE'
JOIN catalog_item_types item_type ON item_type.code = 'GOODS'
JOIN asset_types asset_type ON asset_type.code = source.asset_type_code
ON DUPLICATE KEY UPDATE
    category_id = VALUES(category_id),
    item_type_id = VALUES(item_type_id),
    asset_type_id = VALUES(asset_type_id),
    name = VALUES(name),
    uom = VALUES(uom),
    description = VALUES(description),
    is_asset_candidate = VALUES(is_asset_candidate),
    is_active = 1;

INSERT INTO catalog_item_details (
    item_id, detail_name, detail_value, unit, sort_order
)
SELECT item.id, detail.detail_name, detail.detail_value, detail.unit, detail.sort_order
FROM (
    SELECT 'IT-PC-KASIR' item_code, 'Processor' detail_name, 'Intel Core i5' detail_value, NULL unit, 1 sort_order
    UNION ALL SELECT 'IT-PC-KASIR', 'RAM', '16', 'GB', 2
    UNION ALL SELECT 'IT-PC-KASIR', 'Storage', '512', 'GB SSD', 3
    UNION ALL SELECT 'IT-PC-KASIR', 'Operating System', 'Windows 11 Pro', NULL, 4
    UNION ALL SELECT 'IT-MONITOR-24', 'Ukuran Layar', '24', 'inch', 1
    UNION ALL SELECT 'IT-MONITOR-24', 'Resolusi', '1920 x 1080', 'pixel', 2
    UNION ALL SELECT 'IT-PRINTER-THERMAL', 'Lebar Kertas', '80', 'mm', 1
    UNION ALL SELECT 'IT-PRINTER-THERMAL', 'Interface', 'USB dan LAN', NULL, 2
    UNION ALL SELECT 'IT-BARCODE-SCANNER', 'Tipe Barcode', '1D dan 2D', NULL, 1
    UNION ALL SELECT 'IT-BARCODE-SCANNER', 'Interface', 'USB', NULL, 2
    UNION ALL SELECT 'IT-UPS-650VA', 'Kapasitas', '650', 'VA', 1
    UNION ALL SELECT 'IT-UPS-650VA', 'Jumlah Outlet', '4', 'port', 2
    UNION ALL SELECT 'GA-MEJA-ADMIN', 'Material', 'Kayu olahan', NULL, 1
    UNION ALL SELECT 'GA-MEJA-ADMIN', 'Ukuran', '120 x 60 x 75', 'cm', 2
    UNION ALL SELECT 'GA-KURSI-ADMIN', 'Sandaran', 'Ergonomis', NULL, 1
    UNION ALL SELECT 'GA-KURSI-ADMIN', 'Pengaturan Tinggi', 'Tersedia', NULL, 2
    UNION ALL SELECT 'GA-LEMARI-DOKUMEN', 'Jumlah Pintu', '2', 'pintu', 1
    UNION ALL SELECT 'GA-LEMARI-DOKUMEN', 'Kunci', 'Tersedia', NULL, 2
    UNION ALL SELECT 'GA-TELEPON-MEJA', 'Caller ID', 'Tersedia', NULL, 1
    UNION ALL SELECT 'GA-TEMPAT-SAMPAH', 'Kapasitas', '5', 'liter', 1
) detail
JOIN catalog_items item ON item.item_code = detail.item_code
ON DUPLICATE KEY UPDATE
    detail_value = VALUES(detail_value),
    unit = VALUES(unit),
    sort_order = VALUES(sort_order);

INSERT INTO catalog_packages (
    package_code, name, category_id, owner_division_id,
    description, is_active
)
SELECT
    'PKG-PC-KASIR-STD',
    'Paket PC Kasir Standar',
    category.id,
    division.id,
    'Paket perangkat utama untuk satu titik kasir standar.',
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

INSERT INTO catalog_packages (
    package_code, name, category_id, owner_division_id,
    description, is_active
)
SELECT
    'PKG-MEJA-ADMIN',
    'Paket Meja Admin Toko',
    category.id,
    division.id,
    'Paket furnitur dan perlengkapan dasar untuk satu meja admin toko.',
    1
FROM catalog_item_categories category
JOIN divisions division ON division.division_code = 'GA'
WHERE category.code = 'OFFICE-FURNITURE'
ON DUPLICATE KEY UPDATE
    name = VALUES(name),
    category_id = VALUES(category_id),
    owner_division_id = VALUES(owner_division_id),
    description = VALUES(description),
    is_active = 1;

INSERT INTO catalog_package_items (
    package_id, item_id, qty, is_optional, sort_order, notes
)
SELECT package.id, item.id, source.qty, source.is_optional, source.sort_order, source.notes
FROM (
    SELECT 'PKG-PC-KASIR-STD' package_code, 'IT-PC-KASIR' item_code, 1.00 qty, 0 is_optional, 1 sort_order, 'Perangkat utama kasir' notes
    UNION ALL SELECT 'PKG-PC-KASIR-STD', 'IT-MONITOR-24', 1.00, 0, 2, NULL
    UNION ALL SELECT 'PKG-PC-KASIR-STD', 'IT-PRINTER-THERMAL', 1.00, 0, 3, NULL
    UNION ALL SELECT 'PKG-PC-KASIR-STD', 'IT-BARCODE-SCANNER', 1.00, 0, 4, NULL
    UNION ALL SELECT 'PKG-PC-KASIR-STD', 'IT-UPS-650VA', 1.00, 1, 5, 'Opsional jika toko sudah memiliki UPS'
    UNION ALL SELECT 'PKG-MEJA-ADMIN', 'GA-MEJA-ADMIN', 1.00, 0, 1, NULL
    UNION ALL SELECT 'PKG-MEJA-ADMIN', 'GA-KURSI-ADMIN', 1.00, 0, 2, NULL
    UNION ALL SELECT 'PKG-MEJA-ADMIN', 'GA-LEMARI-DOKUMEN', 1.00, 1, 3, 'Opsional sesuai kapasitas ruangan'
    UNION ALL SELECT 'PKG-MEJA-ADMIN', 'GA-TELEPON-MEJA', 1.00, 0, 4, NULL
    UNION ALL SELECT 'PKG-MEJA-ADMIN', 'GA-TEMPAT-SAMPAH', 1.00, 0, 5, NULL
) source
JOIN catalog_packages package ON package.package_code = source.package_code
JOIN catalog_items item ON item.item_code = source.item_code
ON DUPLICATE KEY UPDATE
    qty = VALUES(qty),
    is_optional = VALUES(is_optional),
    sort_order = VALUES(sort_order),
    notes = VALUES(notes);
