create table attributes (
    id uuid primary key default gen_random_uuid(),
    attribute_group_id uuid references attribute_groups(id) on delete cascade,
    name varchar(255) not null,
    type varchar(50) not null,
    is_filterable boolean default false,
    is_visible boolean default true,
    sort_order integer default 0,
    created_at timestamp not null default current_timestamp,
    updated_at timestamp null default current_timestamp
);

-- create indexes
create index idx_attributes_group_id on attributes(attribute_group_id);
create index idx_attributes_name on attributes(name);
