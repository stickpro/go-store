-- Optimistic-lock counter for admin edits. Bumped by a trigger on every update,
-- so any concurrent change to the order (status, details, refund, items) makes
-- an edit prepared against the old version fail instead of overwriting it.
alter table orders
    add column version bigint default 1 not null;

create function orders_bump_version() returns trigger as
$$
begin
    new.version := old.version + 1;
    return new;
end;
$$ language plpgsql;

create trigger trg_orders_bump_version
    before update
    on orders
    for each row
execute function orders_bump_version();

-- What the customer was charged when the order was paid. grand_total can
-- change afterwards (admin item edits), so this is what refunds and edits of a
-- paid order are measured against: net paid = paid_total - refunded_total.
alter table orders
    add column paid_total decimal(15, 4) default 0 not null;

update orders
set paid_total = grand_total
where payment_status in ('paid', 'partially_refunded', 'refunded');

-- Audit trail of admin item edits: the full line set and totals before and
-- after, plus who did it and why.
create table order_edits
(
    id                 uuid      default gen_random_uuid() not null primary key,
    order_id           uuid                                not null references orders (id) on delete cascade,
    actor              varchar(128)                        not null,
    comment            text,
    lines_before       jsonb                               not null,
    lines_after        jsonb                               not null,
    grand_total_before decimal(15, 4)                      not null,
    grand_total_after  decimal(15, 4)                      not null,
    created_at         timestamp default current_timestamp not null
);
create index idx_order_edits_order on order_edits (order_id, created_at desc);
