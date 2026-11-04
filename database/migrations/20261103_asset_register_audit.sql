-- Audit trail untuk master aset, komponen, dan movement.
ALTER TABLE audit_logs
    MODIFY COLUMN ref_type ENUM(
        'PR', 'PO', 'GR', 'INVOICE', 'PAYMENT', 'APPROVAL',
        'ASSET', 'ASSET_COMPONENT', 'ASSET_MOVEMENT', 'ASSET_COMPONENT_MOVEMENT',
        'ASSET_DEPRECIATION', 'DEPRECIATION_PROFILE', 'DEPRECIATION_PERIOD',
        'ASSET_DISPOSAL'
    ) NOT NULL;
