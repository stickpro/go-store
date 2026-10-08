-- Running total of what has been refunded against a payment. Kept on the row
-- (not derived on the fly) so a refund can check the remaining balance under
-- a single row lock.
alter table payments
    add column refunded_amount decimal(15, 4) default 0 not null;

-- One row per refund attempt. A row is written as 'pending' before the
-- provider is called, so a retry/timeout can never refund the same money
-- twice: pending rows count against the remaining balance until resolved.
create table payment_refunds
(
    id              uuid      default gen_random_uuid() not null primary key,
    payment_id      uuid                                not null references payments (id) on delete cascade,
    amount          decimal(15, 4)                      not null check (amount > 0),
    status          varchar(16)                         not null,
    idempotency_key varchar(64)                         not null unique,
    reason          text,
    actor           varchar(128)                        not null,
    raw_response    jsonb,
    created_at      timestamp default current_timestamp not null,
    updated_at      timestamp
);
create index idx_payment_refunds_payment on payment_refunds (payment_id);

-- Money already returned to the customer for this order, so revenue and the
-- admin panel can show the net amount of a partially refunded order.
alter table orders
    add column refunded_total decimal(15, 4) default 0 not null;
