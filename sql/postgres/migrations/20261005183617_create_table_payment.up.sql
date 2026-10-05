create table payments
(
    id                    uuid      default gen_random_uuid() not null primary key,
    order_id              uuid                                not null references orders (id) on delete cascade,
    provider              varchar(32)                         not null,
    provider_payment_id   varchar(128),
    status                varchar(32)                         not null,
    amount                decimal(15, 4)                      not null,
    currency              char(3)   default 'RUB'             not null,
    payment_url           varchar(512),
    raw_init_response     jsonb,
    raw_last_notification jsonb,
    created_at            timestamp default current_timestamp not null,
    updated_at            timestamp default current_timestamp,

    unique (provider, provider_payment_id)
);
create index idx_payments_order on payments (order_id);