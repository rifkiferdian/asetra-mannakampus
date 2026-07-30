-- Hapus permission modul Detail Item yang sudah digantikan oleh
-- Merek Item, Varian Item, dan Rakitan/BOM.

START TRANSACTION;

DELETE role_permission
FROM role_has_permissions role_permission
JOIN permissions permission ON permission.id = role_permission.permission_id
WHERE permission.name IN (
    'catalog_item_detail_management_access',
    'catalog_item_detail_view',
    'catalog_item_detail_create',
    'catalog_item_detail_edit',
    'catalog_item_detail_delete'
);

DELETE FROM permissions
WHERE name IN (
    'catalog_item_detail_management_access',
    'catalog_item_detail_view',
    'catalog_item_detail_create',
    'catalog_item_detail_edit',
    'catalog_item_detail_delete'
);

COMMIT;
