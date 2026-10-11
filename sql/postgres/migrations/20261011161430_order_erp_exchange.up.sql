-- Order exchange with the ERP (see docs/order-exchange.md).

-- What the ERP reports back about an order: its own document number / GUID
-- and the carrier tracking number. external_id is deliberately not unique —
-- ERP document numbering may restart (e.g. yearly).
ALTER TABLE orders
    ADD COLUMN external_id       varchar(128),
    ADD COLUMN tracking_number   varchar(128),
    ADD COLUMN tracking_provider varchar(32),
    ADD COLUMN tracking_url      varchar(512);

-- Snapshot of products.external_id at checkout — the key the ERP matches its
-- nomenclature on. Kept on the line like every other order_item column, so it
-- survives catalogue deletes.
ALTER TABLE order_items
    ADD COLUMN external_id varchar(128);

-- Inbound ERP order events already applied. Delivery is at-least-once, so the
-- consumer registers event_id here (in the same transaction as the change)
-- and skips an event it has seen before.
CREATE TABLE order_inbound_events
(
    event_id   uuid                                not null primary key,
    event_type varchar(64)                         not null,
    order_id   uuid                                not null references orders (id) on delete cascade,
    created_at timestamp default current_timestamp not null
);

CREATE INDEX idx_order_inbound_events_order ON order_inbound_events (order_id, created_at);
