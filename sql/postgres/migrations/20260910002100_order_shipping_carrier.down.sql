ALTER TABLE orders
    DROP COLUMN ship_provider,
    DROP COLUMN ship_tariff_code,
    DROP COLUMN ship_point_code,
    DROP COLUMN ship_min_days,
    DROP COLUMN ship_max_days;
