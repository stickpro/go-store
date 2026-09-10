DROP INDEX IF EXISTS idx_orders_quick_queue;

ALTER TABLE orders
    DROP COLUMN source;
