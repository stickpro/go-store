create table attribute_groups
(
    id          uuid primary key      default gen_random_uuid(),
    name        varchar(255) not null,
    slug        varchar(255) not null,
    description text         null,
    created_at  timestamp    not null default current_timestamp,
    updated_at  timestamp    null     default current_timestamp
);

create unique index idx_attribute_groups_slug on attribute_groups(slug);

insert into attribute_groups (id, name, slug, description)
values ('00000000-0000-0000-0000-000000000001', 'Основные', 'default', 'Группа по умолчанию для атрибутов из внешних систем')
on conflict (slug) do nothing;
