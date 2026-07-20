-- Snapshot nilai aset saat disposal dan status sebelum posting untuk proses pembatalan.

ALTER TABLE asset_disposals
    ADD COLUMN depreciation_profile_id BIGINT UNSIGNED NULL AFTER asset_id,
    ADD COLUMN acquisition_value DECIMAL(18,2) NOT NULL DEFAULT 0.00 AFTER disposal_value,
    ADD COLUMN accumulated_depreciation DECIMAL(18,2) NOT NULL DEFAULT 0.00 AFTER acquisition_value,
    ADD COLUMN book_value DECIMAL(18,2) NOT NULL DEFAULT 0.00 AFTER accumulated_depreciation,
    ADD COLUMN gain_loss_amount DECIMAL(18,2) NOT NULL DEFAULT 0.00 AFTER book_value,
    ADD COLUMN prior_asset_status VARCHAR(30) NULL AFTER cancellation_reason,
    ADD COLUMN prior_profile_status VARCHAR(30) NULL AFTER prior_asset_status,
    ADD KEY idx_asset_disposal_profile (depreciation_profile_id),
    ADD CONSTRAINT fk_asset_disposal_profile
        FOREIGN KEY (depreciation_profile_id)
        REFERENCES asset_depreciation_profiles(id) ON DELETE SET NULL;
