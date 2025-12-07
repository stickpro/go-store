create table related_products
(
    product_id         uuid not null,
    related_product_id uuid not null,
    primary key (product_id, related_product_id),
    foreign key (product_id) references products (id) on delete cascade,
    foreign key (related_product_id) references products (id) on delete cascade,
    constraint chk_not_self_related check (product_id != related_product_id)
);

create index idx_related_products_product_id on related_products (product_id);
create index idx_related_products_related_product_id on related_products (related_product_id);