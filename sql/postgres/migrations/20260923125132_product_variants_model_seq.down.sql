-- The original (hand-entered) models are not restored: the up migration
-- overwrote them. Generated numbers stay in place.
alter table product_variants
    alter column model drop default;

drop sequence if exists product_variants_model_seq;
