-- Carrier selection snapshot for an order: which carrier + tariff the customer
-- picked at checkout, the pickup point (if any) and the quoted delivery window.
-- The price itself lives in orders.shipping_total (already present).
ALTER TABLE orders
    ADD COLUMN ship_provider    varchar(32),
    ADD COLUMN ship_tariff_code varchar(64),
    ADD COLUMN ship_point_code  varchar(64),
    ADD COLUMN ship_min_days    integer,
    ADD COLUMN ship_max_days    integer;
