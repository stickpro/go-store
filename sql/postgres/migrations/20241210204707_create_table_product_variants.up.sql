create table product_variants
(
    id               uuid                    default gen_random_uuid() not null
        primary key,
    product_id       uuid           not null,
    category_id      uuid                    default null,
    name             varchar(255)   not null,
    slug             varchar(255)   not null unique,
    description      text,
    model           varchar(255)   not null unique,
    meta_title       varchar(255),
    meta_h1          varchar(255),
    meta_description varchar(400),
    meta_keyword     varchar(255),
    image            varchar(255)            default null,
    sort_order       int            not null default 0,
    is_enable        boolean        not null default true,
    viewed           bigint         not null default 0,
    created_at       timestamp      not null default current_timestamp,
    updated_at       timestamp      null     default current_timestamp
);

create index idx_product_variants_slug on product_variants (slug);
create index idx_product_variants_product_id on product_variants (product_id);

alter table product_variants
    add constraint fk_product_variants_product
        foreign key (product_id)
            references products (id)
            on delete cascade;

alter table product_variants
    add constraint fk_product_variants_category
        foreign key (category_id)
            references categories (id)
            on delete set null;