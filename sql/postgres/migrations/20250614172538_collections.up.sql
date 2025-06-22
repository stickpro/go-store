create table if not exists collections
(
    id          uuid                     not null primary key default gen_random_uuid(),
    name        varchar(255)             not null,
    description text                     null,
    slug        varchar(255)             not null unique,
    created_at  timestamp with time zone not null default (timezone('utc', now())),
    updated_at  timestamp with time zone
);