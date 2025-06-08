create table attribute_groups
(
    id          uuid primary key      default gen_random_uuid(),
    name        varchar(255) not null,
    description text         null,
    created_at  timestamp    not null default current_timestamp,
    updated_at  timestamp    null     default current_timestamp
);