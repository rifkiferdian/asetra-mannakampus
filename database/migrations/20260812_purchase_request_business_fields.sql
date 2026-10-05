-- Informasi bisnis tambahan untuk form Purchase Request SPV ke atas.
-- Jalankan setelah skema dasar gobase_app.sql.

ALTER TABLE purchase_requests
    ADD COLUMN request_title VARCHAR(200) NOT NULL DEFAULT '' AFTER pr_number,
    ADD COLUMN request_category ENUM('GOODS','SERVICE','MAINTENANCE','RENTAL','PROJECT','ASSET_REPLACEMENT') NOT NULL DEFAULT 'GOODS' AFTER request_title,
    ADD COLUMN delivery_location VARCHAR(200) NULL AFTER needed_date,
    ADD COLUMN impact_if_not_approved TEXT NULL AFTER justification,
    ADD COLUMN urgency_reason TEXT NULL AFTER impact_if_not_approved,
    ADD COLUMN recommended_vendor_id BIGINT UNSIGNED NULL AFTER urgency_reason,
    ADD COLUMN vendor_recommendation_reason VARCHAR(500) NULL AFTER recommended_vendor_id,
    ADD COLUMN is_single_source TINYINT(1) NOT NULL DEFAULT 0 AFTER vendor_recommendation_reason,
    ADD COLUMN single_source_reason VARCHAR(500) NULL AFTER is_single_source,
    ADD COLUMN budget_exception_reason VARCHAR(500) NULL AFTER single_source_reason,
    ADD COLUMN asset_request_type ENUM('NEW','REPLACEMENT') NULL AFTER budget_exception_reason,
    ADD COLUMN existing_asset_code VARCHAR(100) NULL AFTER asset_request_type,
    ADD COLUMN asset_location VARCHAR(200) NULL AFTER existing_asset_code,
    ADD COLUMN asset_pic VARCHAR(150) NULL AFTER asset_location,
    ADD KEY idx_pr_recommended_vendor (recommended_vendor_id),
    ADD CONSTRAINT fk_pr_recommended_vendor FOREIGN KEY (recommended_vendor_id) REFERENCES vendors(id) ON DELETE SET NULL;

ALTER TABLE purchase_request_items
    ADD COLUMN specification TEXT NULL AFTER item_name,
    ADD COLUMN price_source VARCHAR(100) NULL AFTER est_total;

UPDATE purchase_requests
SET request_title = CONCAT('Purchase Request ', pr_number)
WHERE request_title = '';
