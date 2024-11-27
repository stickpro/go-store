create table if not exists users
(
    id                uuid       default gen_random_uuid() not null
        primary key,
    email             varchar(255)                         not null
        constraint users_email_unique
            unique,
    email_verified_at timestamp,
    password          varchar(255)                         not null,
    remember_token    varchar(100),
    location          varchar(50),
    language          varchar(2) default 'en'::character varying      not null,
    created_at        timestamp,
    updated_at        timestamp,
    deleted_at        timestamp,
    banned            boolean    default false
);

create table if not exists personal_access_tokens
(
    id             uuid default gen_random_uuid() not null
        primary key,
    tokenable_type varchar(255)                   not null,
    tokenable_id   uuid                           not null,
    name           varchar(255)                   not null,
    token          varchar(64)                    not null,
    last_used_at   timestamp(0),
    expires_at     timestamp(0),
    created_at     timestamp(0),
    updated_at     timestamp(0)
);

create index if not exists personal_access_tokens_tokenable_type_tokenable_id_index
    on personal_access_tokens (tokenable_type, tokenable_id);

create unique index if not exists personal_access_tokens_token_unique
    on personal_access_tokens (token);