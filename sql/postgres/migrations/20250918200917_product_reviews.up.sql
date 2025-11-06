create table product_reviews
(
    id         uuid default gen_random_uuid() not null
        primary key,
    product_id uuid                           not null,
    user_id    uuid                           not null,
    order_id   uuid,
    rating     smallint                       NOT NULL CHECK (rating BETWEEN 1 AND 5),
    title      varchar(255),
    body       text,
    status     varchar(50)                    not null,
    created_at       timestamp      not null default current_timestamp,
    updated_at       timestamp      null     default current_timestamp,
    deleted_at       timestamp      null,

    foreign key (product_id) references products (id) on delete cascade

);

create unique index ux_product_user on product_reviews(product_id, user_id) where deleted_at is null;
create index idx_product_reviews_product_status_created on product_reviews(product_id, status, created_at desc);
create index idx_product_reviews_rating on product_reviews(product_id, rating);
create index idx_product_reviews_user on product_reviews(user_id);
