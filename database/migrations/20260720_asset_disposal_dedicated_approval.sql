-- Approval khusus Asset Disposal. Tidak menggunakan approvals/approval_tasks milik PR dan PO.

ALTER TABLE asset_disposals
    MODIFY COLUMN status ENUM(
        'DRAFT',
        'IN_APPROVAL',
        'REJECTED',
        'APPROVED',
        'POSTED',
        'CANCELLED',
        'REVERSED'
    ) NOT NULL DEFAULT 'DRAFT',
    ADD COLUMN submitted_by INT NULL AFTER status,
    ADD COLUMN submitted_at TIMESTAMP NULL DEFAULT NULL AFTER submitted_by,
    ADD COLUMN rejected_by INT NULL AFTER submitted_at,
    ADD COLUMN rejected_at TIMESTAMP NULL DEFAULT NULL AFTER rejected_by,
    ADD COLUMN rejection_reason TEXT NULL AFTER rejected_at,
    ADD COLUMN approved_at TIMESTAMP NULL DEFAULT NULL AFTER approved_by,
    ADD COLUMN posted_by INT NULL AFTER approved_at,
    ADD COLUMN reversed_by INT NULL AFTER cancellation_reason,
    ADD COLUMN reversed_at TIMESTAMP NULL DEFAULT NULL AFTER reversed_by,
    ADD COLUMN reversal_reason TEXT NULL AFTER reversed_at,
    ADD KEY idx_asset_disposal_submitted_by (submitted_by),
    ADD KEY idx_asset_disposal_rejected_by (rejected_by),
    ADD KEY idx_asset_disposal_posted_by (posted_by),
    ADD KEY idx_asset_disposal_reversed_by (reversed_by),
    ADD CONSTRAINT fk_asset_disposal_submitted_by
        FOREIGN KEY (submitted_by) REFERENCES users(id) ON DELETE SET NULL,
    ADD CONSTRAINT fk_asset_disposal_rejected_by
        FOREIGN KEY (rejected_by) REFERENCES users(id) ON DELETE SET NULL,
    ADD CONSTRAINT fk_asset_disposal_posted_by
        FOREIGN KEY (posted_by) REFERENCES users(id) ON DELETE SET NULL,
    ADD CONSTRAINT fk_asset_disposal_reversed_by
        FOREIGN KEY (reversed_by) REFERENCES users(id) ON DELETE SET NULL;

-- Menjaga data lama: pengguna yang sebelumnya tercatat sebagai approved_by adalah pelaku posting lama.
UPDATE asset_disposals
SET posted_by = approved_by,
    approved_at = posted_at
WHERE status = 'POSTED'
  AND posted_by IS NULL;

CREATE TABLE asset_disposal_approval_rules (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    name VARCHAR(150) NOT NULL,
    disposal_type_id BIGINT UNSIGNED NULL,
    asset_type_id BIGINT UNSIGNED NULL,
    min_book_value DECIMAL(18,2) NOT NULL DEFAULT 0.00,
    max_book_value DECIMAL(18,2) NULL,
    priority INT UNSIGNED NOT NULL DEFAULT 100,
    is_active TINYINT(1) NOT NULL DEFAULT 1,
    effective_from DATE NULL,
    effective_until DATE NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uq_asset_disposal_approval_rule_name (name),
    KEY idx_adar_disposal_type (disposal_type_id),
    KEY idx_adar_asset_type (asset_type_id),
    KEY idx_adar_active_priority (is_active, priority),
    CONSTRAINT fk_adar_disposal_type
        FOREIGN KEY (disposal_type_id) REFERENCES asset_disposal_types(id) ON DELETE SET NULL,
    CONSTRAINT fk_adar_asset_type
        FOREIGN KEY (asset_type_id) REFERENCES asset_types(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE asset_disposal_approval_rule_steps (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    rule_id BIGINT UNSIGNED NOT NULL,
    step_order INT UNSIGNED NOT NULL,
    role_id BIGINT UNSIGNED NOT NULL,
    scope ENUM('STORE', 'HO', 'ANY') NOT NULL DEFAULT 'ANY',
    is_parallel TINYINT(1) NOT NULL DEFAULT 0,
    is_required TINYINT(1) NOT NULL DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uq_adars_rule_step_role (rule_id, step_order, role_id),
    KEY idx_adars_role (role_id),
    CONSTRAINT fk_adars_rule
        FOREIGN KEY (rule_id) REFERENCES asset_disposal_approval_rules(id) ON DELETE CASCADE,
    CONSTRAINT fk_adars_role
        FOREIGN KEY (role_id) REFERENCES roles(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE asset_disposal_approvers (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    scope ENUM('STORE', 'HO') NOT NULL,
    store_id INT NULL,
    role_id BIGINT UNSIGNED NOT NULL,
    user_id INT NOT NULL,
    is_active TINYINT(1) NOT NULL DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uq_asset_disposal_approver (scope, store_id, role_id, user_id),
    KEY idx_ada_store (store_id),
    KEY idx_ada_role (role_id),
    KEY idx_ada_user (user_id),
    KEY idx_ada_active_scope (is_active, scope),
    CONSTRAINT fk_ada_store
        FOREIGN KEY (store_id) REFERENCES stores(store_id) ON DELETE CASCADE,
    CONSTRAINT fk_ada_role
        FOREIGN KEY (role_id) REFERENCES roles(id),
    CONSTRAINT fk_ada_user
        FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE asset_disposal_approvals (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    disposal_id BIGINT UNSIGNED NOT NULL,
    rule_id BIGINT UNSIGNED NOT NULL,
    attempt_no INT UNSIGNED NOT NULL DEFAULT 1,
    current_step INT UNSIGNED NOT NULL DEFAULT 1,
    status ENUM('PENDING', 'APPROVED', 'REJECTED', 'CANCELLED') NOT NULL DEFAULT 'PENDING',
    submitted_by INT NOT NULL,
    submitted_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP NULL DEFAULT NULL,
    cancelled_by INT NULL,
    cancelled_at TIMESTAMP NULL DEFAULT NULL,
    cancellation_reason TEXT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uq_asset_disposal_approval_attempt (disposal_id, attempt_no),
    KEY idx_ada_header_rule (rule_id),
    KEY idx_ada_header_status (status),
    KEY idx_ada_header_submitted_by (submitted_by),
    KEY idx_ada_header_cancelled_by (cancelled_by),
    CONSTRAINT fk_ada_header_disposal
        FOREIGN KEY (disposal_id) REFERENCES asset_disposals(id),
    CONSTRAINT fk_ada_header_rule
        FOREIGN KEY (rule_id) REFERENCES asset_disposal_approval_rules(id),
    CONSTRAINT fk_ada_header_submitted_by
        FOREIGN KEY (submitted_by) REFERENCES users(id),
    CONSTRAINT fk_ada_header_cancelled_by
        FOREIGN KEY (cancelled_by) REFERENCES users(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE asset_disposal_approval_tasks (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    approval_id BIGINT UNSIGNED NOT NULL,
    rule_step_id BIGINT UNSIGNED NOT NULL,
    step_order INT UNSIGNED NOT NULL,
    role_id BIGINT UNSIGNED NOT NULL,
    scope ENUM('STORE', 'HO') NOT NULL,
    assigned_user_id INT NOT NULL,
    status ENUM('PENDING', 'WAITING', 'APPROVED', 'REJECTED', 'SKIPPED', 'CANCELLED') NOT NULL DEFAULT 'PENDING',
    comment TEXT NULL,
    acted_at TIMESTAMP NULL DEFAULT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uq_adat_approval_rule_step_user (approval_id, rule_step_id, assigned_user_id),
    KEY idx_adat_rule_step (rule_step_id),
    KEY idx_adat_role (role_id),
    KEY idx_adat_assigned_status (assigned_user_id, status),
    CONSTRAINT fk_adat_approval
        FOREIGN KEY (approval_id) REFERENCES asset_disposal_approvals(id) ON DELETE CASCADE,
    CONSTRAINT fk_adat_rule_step
        FOREIGN KEY (rule_step_id) REFERENCES asset_disposal_approval_rule_steps(id),
    CONSTRAINT fk_adat_role
        FOREIGN KEY (role_id) REFERENCES roles(id),
    CONSTRAINT fk_adat_assigned_user
        FOREIGN KEY (assigned_user_id) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE asset_disposal_approval_histories (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    approval_id BIGINT UNSIGNED NOT NULL,
    task_id BIGINT UNSIGNED NULL,
    disposal_id BIGINT UNSIGNED NOT NULL,
    action VARCHAR(50) NOT NULL,
    old_status VARCHAR(30) NULL,
    new_status VARCHAR(30) NULL,
    actor_user_id INT NOT NULL,
    actor_role_id BIGINT UNSIGNED NULL,
    note TEXT NULL,
    ip_address VARCHAR(45) NULL,
    user_agent VARCHAR(255) NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_adah_approval (approval_id),
    KEY idx_adah_task (task_id),
    KEY idx_adah_disposal (disposal_id),
    KEY idx_adah_actor (actor_user_id),
    KEY idx_adah_action (action),
    CONSTRAINT fk_adah_approval
        FOREIGN KEY (approval_id) REFERENCES asset_disposal_approvals(id) ON DELETE CASCADE,
    CONSTRAINT fk_adah_task
        FOREIGN KEY (task_id) REFERENCES asset_disposal_approval_tasks(id) ON DELETE SET NULL,
    CONSTRAINT fk_adah_disposal
        FOREIGN KEY (disposal_id) REFERENCES asset_disposals(id),
    CONSTRAINT fk_adah_actor_user
        FOREIGN KEY (actor_user_id) REFERENCES users(id),
    CONSTRAINT fk_adah_actor_role
        FOREIGN KEY (actor_role_id) REFERENCES roles(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
