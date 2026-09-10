-- Order acquisition channel.
--   'checkout' — full self-service checkout: the customer picked address, delivery
--                and payment, the order starts in status 'pending'.
--   'quick'    — one-click "quick order": the customer left only a name + phone,
--                a manager calls back to collect address / delivery / payment, so
--                the order starts in status 'new' with shipping_total = 0 and no
--                payment_method yet.
ALTER TABLE orders
    ADD COLUMN source varchar(16) NOT NULL DEFAULT 'checkout';

-- Manager work queue: unconfirmed quick orders, oldest first.
CREATE INDEX idx_orders_quick_queue ON orders (created_at) WHERE status = 'new';
