create table attribute_products
(
    product_id   uuid not null,
    attribute_id uuid not null,
    primary key (product_id, attribute_id),
    foreign key (attribute_id) references attributes (id) on delete cascade,
    foreign key (product_id) references products (id) on delete cascade
);
