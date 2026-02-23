create table products
(
    id               uuid                    default gen_random_uuid() not null
        primary key,
    manufacturer_id  uuid                    default null,
    model            varchar(255)   not null unique,
    sku              varchar(64)             default null,
    upc              varchar(12)             default null,
    ean              varchar(14)             default null,
    jan              varchar(13)             default null,
    isbn             varchar(13)             default null,
    mpn              varchar(64)             default null,
    location         varchar(128)            default null,
    quantity         bigint         not null default 0,
    stock_status     varchar(255),
    price            decimal(15, 4) not null default 0.0000,
    weight           decimal(15, 8) not null default 0.00000000,
    length           decimal(15, 8) not null default 0.00000000,
    width            decimal(15, 8) not null default 0.00000000,
    height           decimal(15, 8) not null default 0.00000000,
    subtract         boolean        not null default true,
    minimum          bigint         not null default 1,
    sort_order       int            not null default 0,
    is_enable        boolean        not null default true,
    created_at       timestamp      not null default current_timestamp,
    updated_at       timestamp      null     default current_timestamp
);

create index idx_products_model on products (model);
