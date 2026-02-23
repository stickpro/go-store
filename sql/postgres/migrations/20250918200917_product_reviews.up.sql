create table product_reviews
(
    id         uuid default gen_random_uuid() not null
        primary key,
    variant_id uuid                           not null,
    user_id    uuid                           not null,
    order_id   uuid,
    rating     smallint                       not null check (rating between 1 and 5),
    title      varchar(255),
    body       text,
    status     varchar(50)                    not null,
    created_at timestamp                      not null default current_timestamp,
    updated_at timestamp                               default current_timestamp,
    deleted_at timestamp,

    foreign key (variant_id) references product_variants (id) on delete cascade
);

create unique index ux_variant_user on product_reviews (variant_id, user_id) where deleted_at is null;
create index idx_product_reviews_variant_status_created on product_reviews (variant_id, status, created_at desc);
create index idx_product_reviews_rating on product_reviews (variant_id, rating);
create index idx_product_reviews_user on product_reviews (user_id);
