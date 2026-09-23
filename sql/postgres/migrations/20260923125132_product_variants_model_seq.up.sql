-- product_variants.model is always generated from a sequence (same scheme as
-- orders_number_seq) and can no longer be set through the API or the import.
-- Non-transactional: a failed insert burns its number, so gaps are expected.
create sequence product_variants_model_seq start 100000 owned by product_variants.model;

-- Renumber existing variants in creation order. Two passes: a single pass could
-- hit the unique constraint if an old model already equals a new number.
update product_variants set model = 'tmp-' || id;

with numbered as (
    select id, 99999 + row_number() over (order by created_at, id) as n
    from product_variants
)
update product_variants pv
set model = numbered.n::text
from numbered
where pv.id = numbered.id;

select setval('product_variants_model_seq', max(model::bigint))
from product_variants
having count(*) > 0;

alter table product_variants
    alter column model set default nextval('product_variants_model_seq')::text;
