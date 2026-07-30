-- Migrasi katalog generik menjadi katalog varian dan BOM/rakitan.
-- Jalankan setelah 20260729_ga_item_catalog.sql.
-- Migrasi ini menghapus catalog_item_details dan vendor_item_prices.item_id
-- setelah seluruh data yang masih relevan dipindahkan.

INSERT INTO catalog_item_types (code, name, description, is_active)
VALUES (
    'COMPOSITE',
    'Rakitan',
    'Item parent yang nilainya dihitung dari komponen BOM dan tidak dibeli langsung.',
    1
)
ON DUPLICATE KEY UPDATE
    name = VALUES(name),
    description = VALUES(description),
    is_active = 1;

ALTER TABLE catalog_items
    ADD COLUMN component_type_id BIGINT UNSIGNED NULL AFTER asset_type_id,
    ADD COLUMN is_purchasable TINYINT(1) NOT NULL DEFAULT 1 AFTER is_asset_candidate,
    ADD KEY idx_catalog_item_component_type (component_type_id),
    ADD KEY idx_catalog_item_purchasable (is_purchasable, is_active),
    ADD CONSTRAINT fk_catalog_item_component_type
        FOREIGN KEY (component_type_id) REFERENCES component_types(id) ON DELETE SET NULL;

UPDATE catalog_items item
JOIN catalog_item_types item_type ON item_type.code = 'COMPOSITE'
SET
    item.item_type_id = item_type.id,
    item.is_purchasable = 0,
    item.is_asset_candidate = 1
WHERE item.item_code = 'IT-PC-KASIR';

CREATE TABLE catalog_brands (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    code VARCHAR(50) NOT NULL,
    name VARCHAR(150) NOT NULL,
    description VARCHAR(255) NULL,
    is_active TINYINT(1) NOT NULL DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uq_catalog_brand_code (code),
    UNIQUE KEY uq_catalog_brand_name (name),
    KEY idx_catalog_brand_active (is_active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

INSERT INTO catalog_brands (code, name, description, is_active)
VALUES
    ('GENERIC', 'Generic', 'Merek sementara untuk data katalog lama.', 1),
    ('KINGSTON', 'Kingston', 'Merek perangkat penyimpanan dan memori.', 1),
    ('INTEL', 'Intel', 'Merek processor dan perangkat komputasi.', 1),
    ('ASUS', 'ASUS', 'Merek perangkat dan komponen komputer.', 1),
    ('SAMSUNG', 'Samsung', 'Merek monitor dan perangkat penyimpanan.', 1),
    ('COOLER-MASTER', 'Cooler Master', 'Merek power supply dan perangkat komputer.', 1)
ON DUPLICATE KEY UPDATE
    name = VALUES(name),
    description = VALUES(description),
    is_active = 1;

CREATE TABLE catalog_item_variants (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    item_id BIGINT UNSIGNED NOT NULL,
    brand_id BIGINT UNSIGNED NULL,
    variant_code VARCHAR(80) NOT NULL,
    model_name VARCHAR(150) NOT NULL,
    manufacturer_sku VARCHAR(100) NULL,
    specification TEXT NULL,
    is_active TINYINT(1) NOT NULL DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uq_catalog_item_variant_code (variant_code),
    KEY idx_catalog_item_variant_item (item_id, is_active),
    KEY idx_catalog_item_variant_brand (brand_id),
    CONSTRAINT fk_catalog_item_variant_item
        FOREIGN KEY (item_id) REFERENCES catalog_items(id),
    CONSTRAINT fk_catalog_item_variant_brand
        FOREIGN KEY (brand_id) REFERENCES catalog_brands(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

INSERT INTO catalog_item_variants (
    item_id,
    brand_id,
    variant_code,
    model_name,
    specification,
    is_active
)
SELECT
    item.id,
    brand.id,
    CONCAT(item.item_code, '-DEFAULT'),
    CONCAT(item.name, ' Standar'),
    COALESCE(
        GROUP_CONCAT(
            CONCAT(
                detail.detail_name,
                ': ',
                detail.detail_value,
                IF(detail.unit IS NULL OR detail.unit = '', '', CONCAT(' ', detail.unit))
            )
            ORDER BY detail.sort_order, detail.id
            SEPARATOR '; '
        ),
        item.description
    ),
    item.is_active
FROM catalog_items item
JOIN catalog_brands brand ON brand.code = 'GENERIC'
LEFT JOIN catalog_item_details detail ON detail.item_id = item.id
WHERE item.is_purchasable = 1
GROUP BY
    item.id,
    brand.id,
    item.item_code,
    item.name,
    item.description,
    item.is_active
ON DUPLICATE KEY UPDATE
    model_name = VALUES(model_name),
    specification = VALUES(specification),
    is_active = VALUES(is_active);

INSERT INTO catalog_items (
    item_code,
    category_id,
    item_type_id,
    asset_type_id,
    component_type_id,
    name,
    uom,
    description,
    is_asset_candidate,
    is_purchasable,
    is_active
)
SELECT
    source.item_code,
    category.id,
    item_type.id,
    NULL,
    component_type.id,
    source.item_name,
    'unit',
    source.description,
    0,
    1,
    1
FROM (
    SELECT 'IT-RAM-16GB' item_code, 'RAM' component_code, 'RAM 16 GB' item_name,
           'Komponen memori standar untuk PC Kasir.' description
    UNION ALL
    SELECT 'IT-PROCESSOR-I5', 'CPU', 'Processor Intel Core i5',
           'Processor standar untuk PC Kasir.'
    UNION ALL
    SELECT 'IT-MAINBOARD-H610', 'MOTHERBOARD', 'Mainboard H610',
           'Mainboard standar untuk PC Kasir.'
    UNION ALL
    SELECT 'IT-SSD-512GB', 'SSD', 'SSD 512 GB',
           'Media penyimpanan SSD standar untuk PC Kasir.'
    UNION ALL
    SELECT 'IT-PSU-500W', 'PSU', 'PSU 500 Watt',
           'Power supply standar untuk PC Kasir.'
) source
JOIN catalog_item_categories category ON category.code = 'IT-EQUIPMENT'
JOIN catalog_item_types item_type ON item_type.code = 'GOODS'
JOIN component_types component_type ON component_type.code = source.component_code
ON DUPLICATE KEY UPDATE
    category_id = VALUES(category_id),
    item_type_id = VALUES(item_type_id),
    asset_type_id = VALUES(asset_type_id),
    component_type_id = VALUES(component_type_id),
    name = VALUES(name),
    uom = VALUES(uom),
    description = VALUES(description),
    is_asset_candidate = VALUES(is_asset_candidate),
    is_purchasable = VALUES(is_purchasable),
    is_active = 1;

INSERT INTO catalog_item_variants (
    item_id,
    brand_id,
    variant_code,
    model_name,
    manufacturer_sku,
    specification,
    is_active
)
SELECT
    item.id,
    brand.id,
    source.variant_code,
    source.model_name,
    source.manufacturer_sku,
    source.specification,
    1
FROM (
    SELECT 'IT-RAM-16GB' item_code, 'KINGSTON' brand_code,
           'VAR-RAM-KINGSTON-16GB' variant_code, 'Fury Beast 16 GB' model_name,
           NULL manufacturer_sku, 'DDR4 16 GB' specification
    UNION ALL
    SELECT 'IT-PROCESSOR-I5', 'INTEL',
           'VAR-CPU-INTEL-I5', 'Core i5-12400',
           'BX8071512400', '6 core, 12 thread'
    UNION ALL
    SELECT 'IT-MAINBOARD-H610', 'ASUS',
           'VAR-MB-ASUS-H610', 'PRIME H610M-K',
           NULL, 'Micro-ATX, socket LGA1700'
    UNION ALL
    SELECT 'IT-SSD-512GB', 'SAMSUNG',
           'VAR-SSD-SAMSUNG-512GB', 'SSD NVMe 512 GB',
           NULL, 'NVMe, kapasitas 512 GB'
    UNION ALL
    SELECT 'IT-PSU-500W', 'COOLER-MASTER',
           'VAR-PSU-CM-500W', 'MWE 500',
           NULL, '500 Watt, 80 Plus'
    UNION ALL
    SELECT 'IT-MONITOR-24', 'SAMSUNG',
           'VAR-MON-SAMSUNG-24-IPS', 'Samsung 24 Inch IPS',
           NULL, 'Panel IPS, resolusi 1920 x 1080'
) source
JOIN catalog_items item ON item.item_code = source.item_code
JOIN catalog_brands brand ON brand.code = source.brand_code
ON DUPLICATE KEY UPDATE
    item_id = VALUES(item_id),
    brand_id = VALUES(brand_id),
    model_name = VALUES(model_name),
    manufacturer_sku = VALUES(manufacturer_sku),
    specification = VALUES(specification),
    is_active = 1;

CREATE TABLE catalog_item_boms (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    parent_item_id BIGINT UNSIGNED NOT NULL,
    bom_code VARCHAR(80) NOT NULL,
    name VARCHAR(180) NOT NULL,
    version_no INT NOT NULL DEFAULT 1,
    description VARCHAR(255) NULL,
    is_active TINYINT(1) NOT NULL DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uq_catalog_item_bom_code (bom_code),
    UNIQUE KEY uq_catalog_item_bom_version (parent_item_id, version_no),
    KEY idx_catalog_item_bom_active (is_active),
    CONSTRAINT fk_catalog_item_bom_parent
        FOREIGN KEY (parent_item_id) REFERENCES catalog_items(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE catalog_item_bom_items (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    bom_id BIGINT UNSIGNED NOT NULL,
    component_item_id BIGINT UNSIGNED NOT NULL,
    variant_id BIGINT UNSIGNED NULL,
    preferred_vendor_id BIGINT UNSIGNED NULL,
    qty DECIMAL(18,2) NOT NULL DEFAULT 1.00,
    is_required TINYINT(1) NOT NULL DEFAULT 1,
    sort_order INT NOT NULL DEFAULT 0,
    notes VARCHAR(255) NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uq_catalog_item_bom_component (bom_id, component_item_id),
    KEY idx_catalog_item_bom_item_sort (bom_id, sort_order),
    KEY idx_catalog_item_bom_item_component (component_item_id),
    KEY idx_catalog_item_bom_item_variant (variant_id),
    KEY idx_catalog_item_bom_item_vendor (preferred_vendor_id),
    CONSTRAINT fk_catalog_item_bom_item_bom
        FOREIGN KEY (bom_id) REFERENCES catalog_item_boms(id) ON DELETE CASCADE,
    CONSTRAINT fk_catalog_item_bom_item_component
        FOREIGN KEY (component_item_id) REFERENCES catalog_items(id),
    CONSTRAINT fk_catalog_item_bom_item_variant
        FOREIGN KEY (variant_id) REFERENCES catalog_item_variants(id) ON DELETE SET NULL,
    CONSTRAINT fk_catalog_item_bom_item_vendor
        FOREIGN KEY (preferred_vendor_id) REFERENCES vendors(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

INSERT INTO catalog_item_boms (
    parent_item_id,
    bom_code,
    name,
    version_no,
    description,
    is_active
)
SELECT
    item.id,
    'BOM-PC-KASIR-STD',
    'PC Kasir Standar',
    1,
    'Konfigurasi komponen standar untuk satu unit PC Kasir.',
    1
FROM catalog_items item
WHERE item.item_code = 'IT-PC-KASIR'
ON DUPLICATE KEY UPDATE
    name = VALUES(name),
    description = VALUES(description),
    is_active = 1;

INSERT INTO catalog_item_bom_items (
    bom_id,
    component_item_id,
    variant_id,
    preferred_vendor_id,
    qty,
    is_required,
    sort_order,
    notes
)
SELECT
    bom.id,
    item.id,
    variant.id,
    vendor.id,
    1.00,
    1,
    source.sort_order,
    source.notes
FROM (
    SELECT 'IT-RAM-16GB' item_code, 'VAR-RAM-KINGSTON-16GB' variant_code,
           1 sort_order, 'RAM standar PC Kasir' notes
    UNION ALL
    SELECT 'IT-PROCESSOR-I5', 'VAR-CPU-INTEL-I5', 2, 'Processor standar PC Kasir'
    UNION ALL
    SELECT 'IT-MAINBOARD-H610', 'VAR-MB-ASUS-H610', 3, 'Mainboard standar PC Kasir'
    UNION ALL
    SELECT 'IT-SSD-512GB', 'VAR-SSD-SAMSUNG-512GB', 4, 'Penyimpanan utama PC Kasir'
    UNION ALL
    SELECT 'IT-PSU-500W', 'VAR-PSU-CM-500W', 5, 'Power supply standar PC Kasir'
) source
JOIN catalog_item_boms bom ON bom.bom_code = 'BOM-PC-KASIR-STD'
JOIN catalog_items item ON item.item_code = source.item_code
JOIN catalog_item_variants variant ON variant.variant_code = source.variant_code
LEFT JOIN vendors vendor ON vendor.name = 'PT Sumber IT'
ON DUPLICATE KEY UPDATE
    variant_id = VALUES(variant_id),
    preferred_vendor_id = VALUES(preferred_vendor_id),
    qty = VALUES(qty),
    is_required = VALUES(is_required),
    sort_order = VALUES(sort_order),
    notes = VALUES(notes);

ALTER TABLE catalog_package_items
    ADD COLUMN bom_id BIGINT UNSIGNED NULL AFTER item_id,
    ADD COLUMN variant_id BIGINT UNSIGNED NULL AFTER bom_id,
    ADD COLUMN preferred_vendor_id BIGINT UNSIGNED NULL AFTER variant_id,
    ADD KEY idx_catalog_package_item_bom (bom_id),
    ADD KEY idx_catalog_package_item_variant (variant_id),
    ADD KEY idx_catalog_package_item_vendor (preferred_vendor_id),
    ADD CONSTRAINT fk_catalog_package_item_bom
        FOREIGN KEY (bom_id) REFERENCES catalog_item_boms(id) ON DELETE SET NULL,
    ADD CONSTRAINT fk_catalog_package_item_variant
        FOREIGN KEY (variant_id) REFERENCES catalog_item_variants(id) ON DELETE SET NULL,
    ADD CONSTRAINT fk_catalog_package_item_vendor
        FOREIGN KEY (preferred_vendor_id) REFERENCES vendors(id) ON DELETE SET NULL;

UPDATE catalog_package_items package_item
JOIN catalog_items item ON item.id = package_item.item_id
JOIN catalog_item_variants variant
    ON variant.item_id = item.id
   AND variant.variant_code = CONCAT(item.item_code, '-DEFAULT')
SET package_item.variant_id = variant.id
WHERE item.is_purchasable = 1;

UPDATE catalog_package_items package_item
JOIN catalog_items item ON item.id = package_item.item_id
JOIN catalog_item_boms bom ON bom.parent_item_id = item.id AND bom.bom_code = 'BOM-PC-KASIR-STD'
SET
    package_item.bom_id = bom.id,
    package_item.variant_id = NULL,
    package_item.preferred_vendor_id = NULL
WHERE item.item_code = 'IT-PC-KASIR';

UPDATE catalog_package_items package_item
JOIN catalog_items item ON item.id = package_item.item_id
JOIN catalog_item_variants variant ON variant.variant_code = 'VAR-MON-SAMSUNG-24-IPS'
LEFT JOIN vendors vendor ON vendor.name = 'PT Sumber IT'
SET
    package_item.variant_id = variant.id,
    package_item.preferred_vendor_id = vendor.id
WHERE item.item_code = 'IT-MONITOR-24';

ALTER TABLE vendor_item_prices
    ADD COLUMN variant_id BIGINT UNSIGNED NULL AFTER item_id;

UPDATE vendor_item_prices price
JOIN catalog_items item ON item.id = price.item_id
JOIN catalog_item_variants variant
    ON variant.item_id = item.id
   AND variant.variant_code = CONCAT(item.item_code, '-DEFAULT')
SET price.variant_id = variant.id;

UPDATE vendor_item_prices price
JOIN catalog_items item ON item.id = price.item_id
JOIN catalog_item_variants variant ON variant.variant_code = 'VAR-MON-SAMSUNG-24-IPS'
SET price.variant_id = variant.id
WHERE item.item_code = 'IT-MONITOR-24';

DELETE price
FROM vendor_item_prices price
JOIN catalog_items item ON item.id = price.item_id
WHERE item.is_purchasable = 0;

DELETE FROM vendor_item_prices
WHERE variant_id IS NULL;

ALTER TABLE vendor_item_prices
    DROP FOREIGN KEY fk_vendor_item_price_item,
    DROP INDEX uq_vendor_item_price_period,
    DROP INDEX idx_vendor_item_price_item_period,
    DROP INDEX idx_vendor_item_price_preferred,
    DROP COLUMN item_id,
    MODIFY COLUMN variant_id BIGINT UNSIGNED NOT NULL,
    ADD UNIQUE KEY uq_vendor_variant_price_period (vendor_id, variant_id, valid_from),
    ADD KEY idx_vendor_variant_price_period (variant_id, valid_from, valid_until),
    ADD KEY idx_vendor_variant_price_preferred (variant_id, is_preferred, is_active),
    ADD CONSTRAINT fk_vendor_item_price_variant
        FOREIGN KEY (variant_id) REFERENCES catalog_item_variants(id);

INSERT INTO vendor_item_prices (
    vendor_id,
    variant_id,
    unit_price,
    currency_code,
    minimum_qty,
    valid_from,
    valid_until,
    lead_time_days,
    quotation_reference,
    is_preferred,
    is_active,
    notes
)
SELECT
    vendor.id,
    variant.id,
    source.unit_price,
    'IDR',
    1.00,
    '2026-01-01',
    '2026-12-31',
    source.lead_time_days,
    source.quotation_reference,
    1,
    1,
    source.notes
FROM (
    SELECT 'VAR-RAM-KINGSTON-16GB' variant_code, 700000.00 unit_price,
           3 lead_time_days, 'QT-SIT/RAM/2026-010' quotation_reference,
           'Harga RAM untuk konfigurasi PC Kasir.' notes
    UNION ALL
    SELECT 'VAR-CPU-INTEL-I5', 2500000.00, 5,
           'QT-SIT/CPU/2026-011', 'Harga processor untuk konfigurasi PC Kasir.'
    UNION ALL
    SELECT 'VAR-MB-ASUS-H610', 1500000.00, 5,
           'QT-SIT/MB/2026-012', 'Harga mainboard untuk konfigurasi PC Kasir.'
    UNION ALL
    SELECT 'VAR-SSD-SAMSUNG-512GB', 800000.00, 3,
           'QT-SIT/SSD/2026-013', 'Harga SSD untuk konfigurasi PC Kasir.'
    UNION ALL
    SELECT 'VAR-PSU-CM-500W', 700000.00, 4,
           'QT-SIT/PSU/2026-014', 'Harga PSU untuk konfigurasi PC Kasir.'
) source
JOIN catalog_item_variants variant ON variant.variant_code = source.variant_code
JOIN vendors vendor ON vendor.name = 'PT Sumber IT' AND vendor.is_active = 1
ON DUPLICATE KEY UPDATE
    unit_price = VALUES(unit_price),
    currency_code = VALUES(currency_code),
    minimum_qty = VALUES(minimum_qty),
    valid_until = VALUES(valid_until),
    lead_time_days = VALUES(lead_time_days),
    quotation_reference = VALUES(quotation_reference),
    is_preferred = VALUES(is_preferred),
    is_active = VALUES(is_active),
    notes = VALUES(notes);

ALTER TABLE purchase_request_items
    ADD COLUMN catalog_item_variant_id BIGINT UNSIGNED NULL AFTER catalog_item_id,
    ADD COLUMN source_bom_id BIGINT UNSIGNED NULL AFTER source_package_item_id,
    ADD COLUMN source_bom_item_id BIGINT UNSIGNED NULL AFTER source_bom_id,
    ADD COLUMN parent_pr_item_id BIGINT UNSIGNED NULL AFTER source_bom_item_id,
    ADD KEY idx_pr_item_variant (catalog_item_variant_id),
    ADD KEY idx_pr_item_source_bom (source_bom_id),
    ADD KEY idx_pr_item_source_bom_item (source_bom_item_id),
    ADD KEY idx_pr_item_parent (parent_pr_item_id),
    ADD CONSTRAINT fk_pr_item_variant
        FOREIGN KEY (catalog_item_variant_id) REFERENCES catalog_item_variants(id) ON DELETE SET NULL,
    ADD CONSTRAINT fk_pr_item_source_bom
        FOREIGN KEY (source_bom_id) REFERENCES catalog_item_boms(id) ON DELETE SET NULL,
    ADD CONSTRAINT fk_pr_item_source_bom_item
        FOREIGN KEY (source_bom_item_id) REFERENCES catalog_item_bom_items(id) ON DELETE SET NULL,
    ADD CONSTRAINT fk_pr_item_parent
        FOREIGN KEY (parent_pr_item_id) REFERENCES purchase_request_items(id) ON DELETE SET NULL;

ALTER TABLE purchase_order_items
    ADD COLUMN catalog_item_variant_id BIGINT UNSIGNED NULL AFTER pr_item_id,
    ADD COLUMN source_bom_item_id BIGINT UNSIGNED NULL AFTER catalog_item_variant_id,
    ADD KEY idx_po_item_variant (catalog_item_variant_id),
    ADD KEY idx_po_item_source_bom_item (source_bom_item_id),
    ADD CONSTRAINT fk_po_item_variant
        FOREIGN KEY (catalog_item_variant_id) REFERENCES catalog_item_variants(id) ON DELETE SET NULL,
    ADD CONSTRAINT fk_po_item_source_bom_item
        FOREIGN KEY (source_bom_item_id) REFERENCES catalog_item_bom_items(id) ON DELETE SET NULL;

DROP TABLE catalog_item_details;
