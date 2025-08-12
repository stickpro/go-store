create table collection_products
(
    collection_id uuid not null,
    product_id    uuid not null,
    primary key (collection_id, product_id),
    foreign key (collection_id) references collections (id) on delete cascade,
    foreign key (product_id) references products (id) on delete cascade
);
