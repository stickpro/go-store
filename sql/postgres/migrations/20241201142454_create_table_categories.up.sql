create table categories
(
    id               uuid                  default gen_random_uuid() not null
        primary key,
    parent_id        uuid references categories (id) on delete cascade,
    name             varchar(255) not null,
    slug             varchar(255) not null unique,
    description      text,
    meta_title       varchar(255),
    meta_h1          varchar(255),
    meta_description varchar(400),
    meta_keyword     varchar(255),
    is_enable        boolean      not null default true,
    created_at       timestamp    not null default current_timestamp,
    updated_at       timestamp    null     default current_timestamp
);

create index idx_categories_parent_id on categories (parent_id);
create index idx_categories_slug on categories (slug);