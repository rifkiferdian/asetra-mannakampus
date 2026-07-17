-- Tambahkan pelaku posting, proses lewati, dan referensi audit depresiasi.
ALTER TABLE asset_depreciation_schedules
    ADD COLUMN posted_by INT NULL AFTER posted_at,
    ADD COLUMN skipped_at TIMESTAMP NULL DEFAULT NULL AFTER posted_by,
    ADD COLUMN skipped_by INT NULL AFTER skipped_at,
    ADD COLUMN skip_reason TEXT NULL AFTER skipped_by,
    ADD KEY idx_ads_posted_by (posted_by),
    ADD KEY idx_ads_skipped_by (skipped_by),
    ADD CONSTRAINT fk_ads_posted_by FOREIGN KEY (posted_by) REFERENCES users(id) ON DELETE SET NULL,
    ADD CONSTRAINT fk_ads_skipped_by FOREIGN KEY (skipped_by) REFERENCES users(id) ON DELETE SET NULL;

ALTER TABLE audit_logs
    MODIFY COLUMN ref_type ENUM(
        'PR',
        'PO',
        'GR',
        'INVOICE',
        'PAYMENT',
        'APPROVAL',
        'ASSET_DEPRECIATION',
        'DEPRECIATION_PROFILE'
    ) NOT NULL;
