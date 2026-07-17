-- Menjaga histori posting asli dan membuat draft koreksi sebagai versi baru.
ALTER TABLE asset_depreciation_schedules
    MODIFY COLUMN status ENUM('DRAFT', 'POSTED', 'SKIPPED', 'REVERSED') NOT NULL DEFAULT 'DRAFT',
    ADD COLUMN version_no INT UNSIGNED NOT NULL DEFAULT 1 AFTER period_date,
    ADD COLUMN original_schedule_id BIGINT UNSIGNED NULL AFTER version_no,
    ADD COLUMN correction_reason TEXT NULL AFTER original_schedule_id,
    ADD COLUMN reversed_at TIMESTAMP NULL DEFAULT NULL AFTER skip_reason,
    ADD COLUMN reversed_by INT NULL AFTER reversed_at,
    ADD COLUMN reversal_reason TEXT NULL AFTER reversed_by,
    DROP INDEX uq_asset_depreciation_period,
    ADD UNIQUE KEY uq_asset_depreciation_period_version (asset_id, period_year, period_month, version_no),
    ADD KEY idx_ads_original_schedule (original_schedule_id),
    ADD KEY idx_ads_reversed_by (reversed_by),
    ADD CONSTRAINT fk_ads_original_schedule FOREIGN KEY (original_schedule_id)
        REFERENCES asset_depreciation_schedules(id),
    ADD CONSTRAINT fk_ads_reversed_by FOREIGN KEY (reversed_by)
        REFERENCES users(id) ON DELETE SET NULL;
