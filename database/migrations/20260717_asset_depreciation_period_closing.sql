CREATE TABLE asset_depreciation_periods (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    period_year INT NOT NULL,
    period_month TINYINT UNSIGNED NOT NULL,
    status ENUM('OPEN', 'GENERATED', 'POSTED', 'CLOSED') NOT NULL DEFAULT 'OPEN',
    generated_at TIMESTAMP NULL DEFAULT NULL,
    generated_by INT NULL,
    posted_at TIMESTAMP NULL DEFAULT NULL,
    posted_by INT NULL,
    closed_at TIMESTAMP NULL DEFAULT NULL,
    closed_by INT NULL,
    closing_notes TEXT NULL,
    reopened_at TIMESTAMP NULL DEFAULT NULL,
    reopened_by INT NULL,
    reopen_reason TEXT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uq_depreciation_period (period_year, period_month),
    KEY idx_adp_status (status),
    KEY idx_adp_generated_by (generated_by),
    KEY idx_adp_posted_by (posted_by),
    KEY idx_adp_closed_by (closed_by),
    KEY idx_adp_reopened_by (reopened_by),
    CONSTRAINT fk_adp_generated_by FOREIGN KEY (generated_by) REFERENCES users(id) ON DELETE SET NULL,
    CONSTRAINT fk_adp_posted_by FOREIGN KEY (posted_by) REFERENCES users(id) ON DELETE SET NULL,
    CONSTRAINT fk_adp_closed_by FOREIGN KEY (closed_by) REFERENCES users(id) ON DELETE SET NULL,
    CONSTRAINT fk_adp_reopened_by FOREIGN KEY (reopened_by) REFERENCES users(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

INSERT INTO asset_depreciation_periods (
    period_year,
    period_month,
    status,
    generated_at,
    posted_at
)
SELECT
    period_year,
    period_month,
    CASE
        WHEN SUM(status = 'DRAFT') > 0 THEN 'GENERATED'
        WHEN COUNT(*) > 0 THEN 'POSTED'
        ELSE 'OPEN'
    END,
    MIN(created_at),
    CASE WHEN SUM(status = 'DRAFT') = 0 THEN MAX(COALESCE(posted_at, skipped_at)) ELSE NULL END
FROM asset_depreciation_schedules
GROUP BY period_year, period_month;

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
        'DEPRECIATION_PERIOD'
    ) NOT NULL;
