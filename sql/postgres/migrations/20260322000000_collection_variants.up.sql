CREATE TABLE collection_variants
(
    collection_id uuid NOT NULL,
    variant_id    uuid NOT NULL,
    PRIMARY KEY (collection_id, variant_id),
    FOREIGN KEY (collection_id) REFERENCES collections (id) ON DELETE CASCADE,
    FOREIGN KEY (variant_id) REFERENCES product_variants (id) ON DELETE CASCADE
);
