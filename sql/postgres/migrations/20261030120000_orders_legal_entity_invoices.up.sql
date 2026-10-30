-- Buyer type and the legal-entity requisites captured at checkout. They are a
-- snapshot on the order (not a link to a profile), so issued invoices never
-- change when the customer edits their details later.
alter table orders
    add column customer_type       varchar(16) default 'individual' not null,
    add column legal_name          varchar(255),
    add column legal_inn           varchar(12),
    add column legal_kpp           varchar(9),
    add column legal_ogrn          varchar(15),
    add column legal_address       varchar(512),
    add constraint orders_customer_type_check check (customer_type in ('individual', 'legal')),
    add constraint orders_legal_requisites_check check (
        customer_type <> 'legal'
            or (legal_name is not null and legal_inn is not null and legal_address is not null)
        );

create sequence invoice_number_seq;

-- Invoices issued for "pay by invoice" orders. A row is immutable: data holds
-- the full snapshot (seller, buyer, lines, VAT) the PDF is rendered from. When
-- the order total changes, the active invoice is voided and a new one issued.
create table invoices
(
    id         uuid      default gen_random_uuid()           not null primary key,
    order_id   uuid                                          not null references orders (id) on delete cascade,
    number     bigint    default nextval('invoice_number_seq') not null unique,
    status     varchar(16)                                   not null default 'active',
    amount     decimal(15, 4)                                not null,
    currency   varchar(3)                                    not null,
    data       jsonb                                         not null,
    issued_at  timestamp default current_timestamp           not null,
    due_at     timestamp                                     not null,
    voided_at  timestamp,
    constraint invoices_status_check check (status in ('active', 'void'))
);

alter sequence invoice_number_seq owned by invoices.number;

-- At most one active invoice per order.
create unique index idx_invoices_active_order on invoices (order_id) where status = 'active';
create index idx_invoices_order on invoices (order_id);
