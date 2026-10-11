-- Transactional outbox for order events going to the ERP (see
-- docs/order-exchange.md). A row is inserted in the same transaction as the
-- order change it describes; the relay worker publishes pending rows to Kafka
-- in id order and stamps published_at.
--
-- Deliberately no foreign key to orders: an event that is still pending must
-- not disappear with its order.
CREATE TABLE order_outbox
(
    id             bigserial                           not null primary key,
    event_id       uuid                                not null unique,
    event_type     varchar(64)                         not null,
    schema_version integer                             not null,
    order_id       uuid                                not null,
    payload        jsonb                               not null,
    occurred_at    timestamptz default now()           not null,
    published_at   timestamptz,
    attempts       integer     default 0               not null,
    last_error     text
);

-- Relay queue: pending events, oldest first.
CREATE INDEX idx_order_outbox_pending ON order_outbox (id) WHERE published_at IS NULL;
-- Retention sweep of already published events.
CREATE INDEX idx_order_outbox_published ON order_outbox (published_at) WHERE published_at IS NOT NULL;
