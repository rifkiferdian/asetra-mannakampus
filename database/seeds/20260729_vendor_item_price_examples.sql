-- Dua contoh harga vendor untuk item katalog.
-- Aman dijalankan ulang karena kombinasi vendor, item, dan tanggal mulai bersifat unik.

INSERT INTO vendor_item_prices (
    vendor_id,
    item_id,
    unit_price,
    currency_code,
    minimum_qty,
    valid_from,
    valid_until,
    lead_time_days,
    quotation_reference,
    is_preferred,
    is_active,
    notes
)
SELECT
    vendor.id,
    item.id,
    example.unit_price,
    'IDR',
    example.minimum_qty,
    '2026-01-01',
    '2026-12-31',
    example.lead_time_days,
    example.quotation_reference,
    1,
    1,
    example.notes
FROM (
    SELECT
        'IT-PC-KASIR' AS item_code,
        7850000.00 AS unit_price,
        1.00 AS minimum_qty,
        7 AS lead_time_days,
        'QT-SIT/PC/2026-001' AS quotation_reference,
        'Harga paket unit PC kasir sesuai spesifikasi katalog.' AS notes
    UNION ALL
    SELECT
        'IT-MONITOR-24',
        2150000.00,
        1.00,
        4,
        'QT-SIT/MON/2026-002',
        'Harga monitor 24 inch termasuk kabel display dan garansi.'
) example
JOIN vendors vendor
    ON vendor.name = 'PT Sumber IT'
   AND vendor.is_active = 1
JOIN catalog_items item
    ON item.item_code = example.item_code
   AND item.is_active = 1
ON DUPLICATE KEY UPDATE
    unit_price = VALUES(unit_price),
    currency_code = VALUES(currency_code),
    minimum_qty = VALUES(minimum_qty),
    valid_until = VALUES(valid_until),
    lead_time_days = VALUES(lead_time_days),
    quotation_reference = VALUES(quotation_reference),
    is_preferred = VALUES(is_preferred),
    is_active = VALUES(is_active),
    notes = VALUES(notes);
