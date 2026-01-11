create table product_attribute_values
(
    product_id         uuid      not null references products (id) on delete cascade,
    attribute_value_id uuid      not null references attribute_values (id) on delete cascade,
    created_at         timestamp not null default current_timestamp,
    primary key (product_id, attribute_value_id)
);

create index idx_product_attr_vals_product on product_attribute_values (product_id);
create index idx_product_attr_vals_value on product_attribute_values (attribute_value_id);