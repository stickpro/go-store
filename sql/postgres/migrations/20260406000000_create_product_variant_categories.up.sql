create table product_variant_categories
(
    product_variant_id uuid      not null references product_variants (id) on delete cascade,
    category_id        uuid      not null references categories (id) on delete cascade,
    created_at         timestamp not null default current_timestamp,
    primary key (product_variant_id, category_id)
);

create index idx_pvc_variant on product_variant_categories (product_variant_id);
create index idx_pvc_category on product_variant_categories (category_id);
