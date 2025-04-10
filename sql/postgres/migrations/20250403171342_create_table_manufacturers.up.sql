create table manufacturers
(
    id         uuid                  default gen_random_uuid() not null
        primary key,
    name             varchar(255) not null,
    slug             varchar(255) not null unique,
    description      text,
    image_path       text         null,
    meta_title       varchar(255),
    meta_h1          varchar(255),
    meta_description varchar(400),
    meta_keyword     varchar(255),
    is_enable        boolean      not null default true,
    created_at       timestamp    not null default current_timestamp,
    updated_at       timestamp    null     default current_timestamp
)