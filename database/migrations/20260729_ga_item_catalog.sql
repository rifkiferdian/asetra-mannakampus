-- Katalog barang/jasa untuk permintaan GA/IT dan sumber item PR.
-- Master vendor tetap menggunakan tabel vendors yang sudah tersedia.

CREATE TABLE catalog_item_types (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    code VARCHAR(30) NOT NULL,
    name VARCHAR(100) NOT NULL,
    description VARCHAR(255) NULL,
    is_active TINYINT(1) NOT NULL DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uq_catalog_item_type_code (code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

INSERT INTO catalog_item_types (code, name, description)
VALUES
    ('GOODS', 'Barang', 'Barang fisik yang dibeli dan diterima melalui proses GR.'),
    ('SERVICE', 'Jasa', 'Pekerjaan atau layanan yang dipesan dari vendor.')
ON DUPLICATE KEY UPDATE
    name = VALUES(name),
    description = VALUES(description),
    is_active = 1;

CREATE TABLE catalog_item_categories (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    code VARCHAR(40) NOT NULL,
    name VARCHAR(120) NOT NULL,
    parent_id BIGINT UNSIGNED NULL,
    owner_division_id INT NULL,
    description VARCHAR(255) NULL,
    is_active TINYINT(1) NOT NULL DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uq_catalog_item_category_code (code),
    KEY idx_catalog_item_category_parent (parent_id),
    KEY idx_catalog_item_category_owner (owner_division_id),
    KEY idx_catalog_item_category_active (is_active),
    CONSTRAINT fk_catalog_item_category_parent
        FOREIGN KEY (parent_id) REFERENCES catalog_item_categories(id) ON DELETE SET NULL,
    CONSTRAINT fk_catalog_item_category_owner
        FOREIGN KEY (owner_division_id) REFERENCES divisions(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE catalog_items (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    item_code VARCHAR(50) NOT NULL,
    category_id BIGINT UNSIGNED NOT NULL,
    item_type_id BIGINT UNSIGNED NOT NULL,
    asset_type_id BIGINT UNSIGNED NULL,
    name VARCHAR(200) NOT NULL,
    uom VARCHAR(30) NOT NULL DEFAULT 'unit',
    description TEXT NULL,
    is_asset_candidate TINYINT(1) NOT NULL DEFAULT 0,
    is_active TINYINT(1) NOT NULL DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uq_catalog_item_code (item_code),
    KEY idx_catalog_item_category (category_id),
    KEY idx_catalog_item_type (item_type_id),
    KEY idx_catalog_item_asset_type (asset_type_id),
    KEY idx_catalog_item_active (is_active),
    CONSTRAINT fk_catalog_item_category
        FOREIGN KEY (category_id) REFERENCES catalog_item_categories(id),
    CONSTRAINT fk_catalog_item_type
        FOREIGN KEY (item_type_id) REFERENCES catalog_item_types(id),
    CONSTRAINT fk_catalog_item_asset_type
        FOREIGN KEY (asset_type_id) REFERENCES asset_types(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE catalog_item_details (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    item_id BIGINT UNSIGNED NOT NULL,
    detail_name VARCHAR(100) NOT NULL,
    detail_value VARCHAR(255) NOT NULL,
    unit VARCHAR(30) NULL,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uq_catalog_item_detail_name (item_id, detail_name),
    KEY idx_catalog_item_detail_sort (item_id, sort_order),
    CONSTRAINT fk_catalog_item_detail_item
        FOREIGN KEY (item_id) REFERENCES catalog_items(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE catalog_packages (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    package_code VARCHAR(50) NOT NULL,
    name VARCHAR(200) NOT NULL,
    category_id BIGINT UNSIGNED NULL,
    owner_division_id INT NULL,
    description TEXT NULL,
    is_active TINYINT(1) NOT NULL DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uq_catalog_package_code (package_code),
    KEY idx_catalog_package_category (category_id),
    KEY idx_catalog_package_owner (owner_division_id),
    KEY idx_catalog_package_active (is_active),
    CONSTRAINT fk_catalog_package_category
        FOREIGN KEY (category_id) REFERENCES catalog_item_categories(id) ON DELETE SET NULL,
    CONSTRAINT fk_catalog_package_owner
        FOREIGN KEY (owner_division_id) REFERENCES divisions(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE catalog_package_items (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    package_id BIGINT UNSIGNED NOT NULL,
    item_id BIGINT UNSIGNED NOT NULL,
    qty DECIMAL(18,2) NOT NULL DEFAULT 1.00,
    is_optional TINYINT(1) NOT NULL DEFAULT 0,
    sort_order INT NOT NULL DEFAULT 0,
    notes VARCHAR(255) NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uq_catalog_package_item (package_id, item_id),
    KEY idx_catalog_package_item_sort (package_id, sort_order),
    KEY idx_catalog_package_item_item (item_id),
    CONSTRAINT fk_catalog_package_item_package
        FOREIGN KEY (package_id) REFERENCES catalog_packages(id) ON DELETE CASCADE,
    CONSTRAINT fk_catalog_package_item_item
        FOREIGN KEY (item_id) REFERENCES catalog_items(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE vendor_item_prices (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    vendor_id BIGINT UNSIGNED NOT NULL,
    item_id BIGINT UNSIGNED NOT NULL,
    unit_price DECIMAL(18,2) NOT NULL DEFAULT 0.00,
    currency_code CHAR(3) NOT NULL DEFAULT 'IDR',
    minimum_qty DECIMAL(18,2) NOT NULL DEFAULT 1.00,
    valid_from DATE NOT NULL,
    valid_until DATE NULL,
    lead_time_days INT NULL,
    quotation_reference VARCHAR(100) NULL,
    is_preferred TINYINT(1) NOT NULL DEFAULT 0,
    is_active TINYINT(1) NOT NULL DEFAULT 1,
    notes VARCHAR(255) NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uq_vendor_item_price_period (vendor_id, item_id, valid_from),
    KEY idx_vendor_item_price_item_period (item_id, valid_from, valid_until),
    KEY idx_vendor_item_price_vendor (vendor_id),
    KEY idx_vendor_item_price_preferred (item_id, is_preferred, is_active),
    CONSTRAINT fk_vendor_item_price_vendor
        FOREIGN KEY (vendor_id) REFERENCES vendors(id),
    CONSTRAINT fk_vendor_item_price_item
        FOREIGN KEY (item_id) REFERENCES catalog_items(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

ALTER TABLE purchase_requests
    MODIFY COLUMN status ENUM(
        'DRAFT',
        'IN_ITEM_REVIEW',
        'REVISION_REQUIRED',
        'SUBMITTED',
        'IN_APPROVAL',
        'REJECTED',
        'APPROVED',
        'CONVERTED_TO_PO',
        'CLOSED'
    ) NOT NULL DEFAULT 'DRAFT',
    ADD COLUMN requested_location_id BIGINT UNSIGNED NULL AFTER store_id,
    ADD KEY idx_pr_requested_location (requested_location_id),
    ADD CONSTRAINT fk_pr_requested_location
        FOREIGN KEY (requested_location_id) REFERENCES asset_locations(id) ON DELETE SET NULL;

ALTER TABLE purchase_request_items
    ADD COLUMN catalog_item_id BIGINT UNSIGNED NULL AFTER pr_id,
    ADD COLUMN item_type_id BIGINT UNSIGNED NULL AFTER catalog_item_id,
    ADD COLUMN item_category_id BIGINT UNSIGNED NULL AFTER item_type_id,
    ADD COLUMN source_package_id BIGINT UNSIGNED NULL AFTER item_category_id,
    ADD COLUMN source_package_item_id BIGINT UNSIGNED NULL AFTER source_package_id,
    ADD COLUMN is_non_catalog TINYINT(1) NOT NULL DEFAULT 0 AFTER source_package_item_id,
    ADD COLUMN assigned_division_id INT NULL AFTER vendor_note,
    ADD COLUMN review_status VARCHAR(30) NOT NULL DEFAULT 'PENDING' AFTER assigned_division_id,
    ADD COLUMN reviewed_by INT NULL AFTER review_status,
    ADD COLUMN reviewed_at TIMESTAMP NULL DEFAULT NULL AFTER reviewed_by,
    ADD COLUMN review_note VARCHAR(500) NULL AFTER reviewed_at,
    ADD COLUMN specification_snapshot TEXT NULL AFTER notes,
    ADD KEY idx_pr_item_catalog_item (catalog_item_id),
    ADD KEY idx_pr_item_type (item_type_id),
    ADD KEY idx_pr_item_category (item_category_id),
    ADD KEY idx_pr_item_source_package (source_package_id),
    ADD KEY idx_pr_item_source_package_item (source_package_item_id),
    ADD KEY idx_pr_item_assigned_division (assigned_division_id),
    ADD KEY idx_pr_item_review_status (review_status),
    ADD KEY idx_pr_item_reviewed_by (reviewed_by),
    ADD CONSTRAINT fk_pr_item_catalog_item
        FOREIGN KEY (catalog_item_id) REFERENCES catalog_items(id) ON DELETE SET NULL,
    ADD CONSTRAINT fk_pr_item_type
        FOREIGN KEY (item_type_id) REFERENCES catalog_item_types(id) ON DELETE SET NULL,
    ADD CONSTRAINT fk_pr_item_category
        FOREIGN KEY (item_category_id) REFERENCES catalog_item_categories(id) ON DELETE SET NULL,
    ADD CONSTRAINT fk_pr_item_source_package
        FOREIGN KEY (source_package_id) REFERENCES catalog_packages(id) ON DELETE SET NULL,
    ADD CONSTRAINT fk_pr_item_source_package_item
        FOREIGN KEY (source_package_item_id) REFERENCES catalog_package_items(id) ON DELETE SET NULL,
    ADD CONSTRAINT fk_pr_item_assigned_division
        FOREIGN KEY (assigned_division_id) REFERENCES divisions(id) ON DELETE SET NULL,
    ADD CONSTRAINT fk_pr_item_reviewed_by
        FOREIGN KEY (reviewed_by) REFERENCES users(id) ON DELETE SET NULL;

-- Data PR lama dianggap sudah selesai ditinjau agar tidak mengganggu alur yang berjalan.
UPDATE purchase_request_items
SET review_status = 'COMPLETED'
WHERE catalog_item_id IS NULL
  AND source_package_id IS NULL
  AND review_status = 'PENDING';

ALTER TABLE purchase_order_items
    ADD COLUMN pr_item_id BIGINT UNSIGNED NULL AFTER po_id,
    ADD KEY idx_po_item_pr_item (pr_item_id),
    ADD CONSTRAINT fk_po_item_pr_item
        FOREIGN KEY (pr_item_id) REFERENCES purchase_request_items(id) ON DELETE SET NULL;
