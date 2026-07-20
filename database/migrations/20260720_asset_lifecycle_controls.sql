-- Kontrol siklus hidup depresiasi dan pelepasan aset.
-- FINISHED/TERMINATED disimpan pada profil depresiasi, bukan assets.status.

CREATE TABLE asset_depreciation_first_month_policies (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    code VARCHAR(40) NOT NULL,
    name VARCHAR(100) NOT NULL,
    description VARCHAR(255) NULL,
    is_active TINYINT(1) NOT NULL DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uq_ad_first_month_policy_code (code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

INSERT INTO asset_depreciation_first_month_policies (id, code, name, description)
VALUES
    (1, 'FULL_MONTH', 'Satu Bulan Penuh', 'Depresiasi dimulai penuh pada bulan aset siap digunakan.'),
    (2, 'NEXT_MONTH', 'Mulai Bulan Berikutnya', 'Depresiasi dimulai pada bulan setelah aset siap digunakan.'),
    (3, 'PRORATE_DAILY', 'Prorata Harian', 'Depresiasi bulan pertama dihitung berdasarkan jumlah hari penggunaan.')
ON DUPLICATE KEY UPDATE
    name = VALUES(name),
    description = VALUES(description),
    is_active = 1;

CREATE TABLE asset_depreciation_last_month_policies (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    code VARCHAR(40) NOT NULL,
    name VARCHAR(100) NOT NULL,
    description VARCHAR(255) NULL,
    is_active TINYINT(1) NOT NULL DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uq_ad_last_month_policy_code (code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

INSERT INTO asset_depreciation_last_month_policies (id, code, name, description)
VALUES
    (1, 'NO_DEPRECIATION', 'Tidak Dihitung', 'Tidak ada depresiasi pada bulan aset dilepas.'),
    (2, 'FULL_MONTH', 'Satu Bulan Penuh', 'Depresiasi bulan pelepasan tetap dihitung penuh.'),
    (3, 'PRORATE_DAILY', 'Prorata Harian', 'Depresiasi dihitung sampai tanggal pelepasan aset.')
ON DUPLICATE KEY UPDATE
    name = VALUES(name),
    description = VALUES(description),
    is_active = 1;

ALTER TABLE asset_depreciation_profiles
    MODIFY COLUMN status ENUM('ACTIVE', 'PAUSED', 'FINISHED', 'TERMINATED') NOT NULL DEFAULT 'ACTIVE',
    ADD COLUMN first_month_policy_id BIGINT UNSIGNED NULL AFTER start_date,
    ADD COLUMN last_month_policy_id BIGINT UNSIGNED NULL AFTER first_month_policy_id,
    ADD COLUMN paused_at TIMESTAMP NULL DEFAULT NULL AFTER status,
    ADD COLUMN paused_by INT NULL AFTER paused_at,
    ADD COLUMN pause_reason TEXT NULL AFTER paused_by,
    ADD COLUMN resumed_at TIMESTAMP NULL DEFAULT NULL AFTER pause_reason,
    ADD COLUMN resumed_by INT NULL AFTER resumed_at,
    ADD COLUMN finished_at TIMESTAMP NULL DEFAULT NULL AFTER resumed_by,
    ADD COLUMN terminated_at TIMESTAMP NULL DEFAULT NULL AFTER finished_at,
    ADD COLUMN terminated_by INT NULL AFTER terminated_at,
    ADD COLUMN termination_reason TEXT NULL AFTER terminated_by;

-- Kebijakan standar awal: mulai bulan berikutnya dan tidak menghitung bulan disposal.
UPDATE asset_depreciation_profiles profile
JOIN asset_depreciation_first_month_policies first_policy
    ON first_policy.code = 'NEXT_MONTH'
JOIN asset_depreciation_last_month_policies last_policy
    ON last_policy.code = 'NO_DEPRECIATION'
SET
    profile.first_month_policy_id = first_policy.id,
    profile.last_month_policy_id = last_policy.id
WHERE profile.first_month_policy_id IS NULL
   OR profile.last_month_policy_id IS NULL;

-- Profil lama yang sudah FINISHED diberi tanggal selesai berdasarkan posting terakhir.
UPDATE asset_depreciation_profiles profile
SET profile.finished_at = COALESCE(
    (
        SELECT MAX(schedule.posted_at)
        FROM asset_depreciation_schedules schedule
        WHERE schedule.profile_id = profile.id
          AND schedule.status = 'POSTED'
    ),
    profile.updated_at
)
WHERE profile.status = 'FINISHED'
  AND profile.finished_at IS NULL;

ALTER TABLE asset_depreciation_profiles
    MODIFY COLUMN first_month_policy_id BIGINT UNSIGNED NOT NULL DEFAULT 2,
    MODIFY COLUMN last_month_policy_id BIGINT UNSIGNED NOT NULL DEFAULT 1,
    ADD KEY idx_adp_first_month_policy (first_month_policy_id),
    ADD KEY idx_adp_last_month_policy (last_month_policy_id),
    ADD KEY idx_adp_paused_by (paused_by),
    ADD KEY idx_adp_resumed_by (resumed_by),
    ADD KEY idx_adp_terminated_by (terminated_by),
    ADD CONSTRAINT fk_adp_first_month_policy
        FOREIGN KEY (first_month_policy_id)
        REFERENCES asset_depreciation_first_month_policies(id),
    ADD CONSTRAINT fk_adp_last_month_policy
        FOREIGN KEY (last_month_policy_id)
        REFERENCES asset_depreciation_last_month_policies(id),
    ADD CONSTRAINT fk_adp_paused_by
        FOREIGN KEY (paused_by) REFERENCES users(id) ON DELETE SET NULL,
    ADD CONSTRAINT fk_adp_resumed_by
        FOREIGN KEY (resumed_by) REFERENCES users(id) ON DELETE SET NULL,
    ADD CONSTRAINT fk_adp_terminated_by
        FOREIGN KEY (terminated_by) REFERENCES users(id) ON DELETE SET NULL;

CREATE TABLE asset_disposal_types (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    code VARCHAR(40) NOT NULL,
    name VARCHAR(100) NOT NULL,
    description VARCHAR(255) NULL,
    is_active TINYINT(1) NOT NULL DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uq_asset_disposal_type_code (code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

INSERT INTO asset_disposal_types (code, name, description)
VALUES
    ('SOLD', 'Dijual', 'Aset dilepas melalui proses penjualan.'),
    ('DESTROYED', 'Dimusnahkan', 'Aset dimusnahkan karena tidak dapat digunakan kembali.'),
    ('LOST', 'Hilang', 'Aset tidak ditemukan atau dinyatakan hilang.'),
    ('DONATED', 'Dihibahkan', 'Aset diberikan atau dihibahkan kepada pihak lain.'),
    ('WRITE_OFF', 'Dihapus Buku', 'Aset dihentikan dan dihapus dari pencatatan perusahaan.')
ON DUPLICATE KEY UPDATE
    name = VALUES(name),
    description = VALUES(description),
    is_active = 1;

CREATE TABLE asset_disposals (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    disposal_number VARCHAR(50) NOT NULL,
    asset_id BIGINT UNSIGNED NOT NULL,
    disposal_type_id BIGINT UNSIGNED NOT NULL,
    disposal_date DATE NOT NULL,
    disposal_value DECIMAL(18,2) NOT NULL DEFAULT 0.00,
    buyer_name VARCHAR(150) NULL,
    document_reference VARCHAR(100) NULL,
    reason TEXT NOT NULL,
    status ENUM('DRAFT', 'POSTED', 'CANCELLED') NOT NULL DEFAULT 'DRAFT',
    processed_by INT NOT NULL,
    approved_by INT NULL,
    posted_at TIMESTAMP NULL DEFAULT NULL,
    cancelled_at TIMESTAMP NULL DEFAULT NULL,
    cancelled_by INT NULL,
    cancellation_reason TEXT NULL,
    notes TEXT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uq_asset_disposal_number (disposal_number),
    KEY idx_asset_disposal_asset_status (asset_id, status),
    KEY idx_asset_disposal_type (disposal_type_id),
    KEY idx_asset_disposal_date (disposal_date),
    KEY idx_asset_disposal_processed_by (processed_by),
    KEY idx_asset_disposal_approved_by (approved_by),
    KEY idx_asset_disposal_cancelled_by (cancelled_by),
    CONSTRAINT fk_asset_disposal_asset
        FOREIGN KEY (asset_id) REFERENCES assets(id),
    CONSTRAINT fk_asset_disposal_type
        FOREIGN KEY (disposal_type_id) REFERENCES asset_disposal_types(id),
    CONSTRAINT fk_asset_disposal_processed_by
        FOREIGN KEY (processed_by) REFERENCES users(id),
    CONSTRAINT fk_asset_disposal_approved_by
        FOREIGN KEY (approved_by) REFERENCES users(id) ON DELETE SET NULL,
    CONSTRAINT fk_asset_disposal_cancelled_by
        FOREIGN KEY (cancelled_by) REFERENCES users(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

ALTER TABLE audit_logs
    MODIFY COLUMN ref_type ENUM(
        'PR',
        'PO',
        'GR',
        'INVOICE',
        'PAYMENT',
        'APPROVAL',
        'ASSET_DEPRECIATION',
        'DEPRECIATION_PROFILE',
        'DEPRECIATION_PERIOD',
        'ASSET_DISPOSAL'
    ) NOT NULL;
