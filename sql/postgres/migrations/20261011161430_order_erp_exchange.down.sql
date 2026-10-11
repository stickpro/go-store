DROP TABLE IF EXISTS order_inbound_events;

ALTER TABLE order_items
    DROP COLUMN IF EXISTS external_id;

ALTER TABLE orders
    DROP COLUMN IF EXISTS external_id,
    DROP COLUMN IF EXISTS tracking_number,
    DROP COLUMN IF EXISTS tracking_provider,
    DROP COLUMN IF EXISTS tracking_url;
