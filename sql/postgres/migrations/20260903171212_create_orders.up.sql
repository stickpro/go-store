-- Human-facing sequential order number, independent of the uuid primary key.
-- Non-transactional: a rolled-back order burns its number, so gaps are expected.
create sequence orders_number_seq start 100000;

create table orders
(
    id              uuid          default gen_random_uuid()          not null primary key,
    order_number    bigint        default nextval('orders_number_seq') not null unique,
    user_id         uuid,                                             -- null = guest checkout
    status          varchar(32)   default 'pending'                  not null,
    payment_status  varchar(32)   default 'unpaid'                   not null,
    payment_method  varchar(32),
    currency        char(3)       default 'RUB'                      not null,

    -- contact
    email           varchar(255)                                     not null,
    phone           varchar(32),

    -- shipping address (denormalised snapshot; ship_city_id is a soft pointer only)
    ship_city_id    uuid,
    ship_city_name  varchar(255)                                     not null,
    ship_address    varchar(512)                                     not null,
    ship_postcode   varchar(16),
    ship_recipient  varchar(255)                                     not null,
    shipping_method varchar(64),

    -- money, all decimal(15,4) to match products.price_*
    subtotal        decimal(15, 4)                                   not null,
    discount_total  decimal(15, 4) default 0                         not null,
    shipping_total  decimal(15, 4) default 0                         not null,
    tax_total       decimal(15, 4) default 0                         not null,
    grand_total     decimal(15, 4)                                   not null,

    comment         text,
    idempotency_key varchar(80),

    created_at      timestamp     default current_timestamp          not null,
    updated_at      timestamp     default current_timestamp,
    paid_at         timestamp,
    cancelled_at    timestamp,

    foreign key (user_id) references users (id) on delete set null,
    foreign key (ship_city_id) references cities (id) on delete set null
);

create unique index ux_orders_idempotency_key on orders (idempotency_key) where idempotency_key is not null;
create index idx_orders_user_created on orders (user_id, created_at desc);
create index idx_orders_open on orders (status) where status in ('pending', 'paid');

-- Full snapshot of each line at checkout time. Deliberately NO foreign key to
-- products / product_variants: historical orders must survive catalogue deletes.
create table order_items
(
    id         uuid           default gen_random_uuid() not null primary key,
    order_id   uuid                                     not null,
    product_id uuid,
    variant_id uuid,
    sku        varchar(255),
    name       varchar(512)                             not null,
    slug       varchar(255),
    image_path varchar(255),
    unit_price decimal(15, 4)                           not null,
    quantity   bigint                                   not null,
    line_total decimal(15, 4)                           not null,

    foreign key (order_id) references orders (id) on delete cascade
);

create index idx_order_items_order on order_items (order_id);

create table order_status_history
(
    id          uuid        default gen_random_uuid() not null primary key,
    order_id    uuid                                  not null,
    from_status varchar(32),
    to_status   varchar(32)                           not null,
    actor       varchar(64)                           not null, -- 'customer' | 'admin:<id>' | 'system' | 'payment:<provider>'
    comment     text,
    created_at  timestamp   default current_timestamp not null,

    foreign key (order_id) references orders (id) on delete cascade
);

create index idx_order_status_history_order on order_status_history (order_id, created_at);
