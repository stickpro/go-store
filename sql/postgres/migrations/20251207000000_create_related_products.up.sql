create table related_products
(
    variant_id         uuid not null,
    related_variant_id uuid not null,
    primary key (variant_id, related_variant_id),
    foreign key (variant_id) references product_variants (id) on delete cascade,
    foreign key (related_variant_id) references product_variants (id) on delete cascade,
    constraint chk_not_self_related check (variant_id != related_variant_id)
);

create index idx_related_products_variant_id on related_products (variant_id);
create index idx_related_products_related_variant_id on related_products (related_variant_id);
