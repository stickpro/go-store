create table attributes
(
    id                 uuid primary key      default gen_random_uuid(),
    attribute_group_id uuid references attribute_groups (id) on delete cascade,
    name               varchar(255) not null,
    slug               varchar(255) not null unique,
    type               varchar(50)  not null check (type in ('select', 'number', 'boolean', 'text')),
    unit               varchar(50)  null,
    is_filterable      boolean               default true,
    is_visible         boolean               default true,
    is_required        boolean               default false,
    sort_order         integer               default 0,
    created_at         timestamp    not null default current_timestamp,
    updated_at         timestamp    null     default current_timestamp
);
-- create indexes
create index idx_attributes_group_id on attributes (attribute_group_id);
create index idx_attributes_slug on attributes (slug);
create unique index idx_attributes_group_name on attributes (attribute_group_id, name);


create table attribute_values
(
    id               uuid primary key        default gen_random_uuid(),
    attribute_id     uuid           not null references attributes (id) on delete cascade,
    value            text           not null,
    value_normalized text,
    value_numeric    decimal(15, 4) null,
    display_order    integer                 default 0,
    is_active        boolean                 default true,
    created_at       timestamp      not null default current_timestamp,
    updated_at       timestamp      null     default current_timestamp
);

create index idx_attribute_values_attribute_id on attribute_values(attribute_id);
create index idx_attribute_values_numeric on attribute_values(value_numeric) where value_numeric is not null;
create unique index idx_attribute_values_unique on attribute_values(attribute_id, value);