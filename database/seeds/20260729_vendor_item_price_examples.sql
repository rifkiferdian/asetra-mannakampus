-- Dua contoh harga vendor untuk item katalog.
-- Aman dijalankan ulang karena kombinasi vendor, item, dan tanggal mulai bersifat unik.

INSERT INTO vendor_item_prices (
    vendor_id,
    variant_id,
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
    variant.id,
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
        'VAR-RAM-KINGSTON-16GB' AS variant_code,
        700000.00 AS unit_price,
        1.00 AS minimum_qty,
        3 AS lead_time_days,
        'QT-SIT/RAM/2026-010' AS quotation_reference,
        'Harga RAM untuk konfigurasi PC Kasir.' AS notes
    UNION ALL
    SELECT
        'VAR-MON-SAMSUNG-24-IPS',
        2150000.00,
        1.00,
        4,
        'QT-SIT/MON/2026-002',
        'Harga monitor 24 inch termasuk kabel display dan garansi.'
) example
JOIN vendors vendor
    ON vendor.name = 'PT Sumber IT'
   AND vendor.is_active = 1
JOIN catalog_item_variants variant
    ON variant.variant_code = example.variant_code
   AND variant.is_active = 1
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
