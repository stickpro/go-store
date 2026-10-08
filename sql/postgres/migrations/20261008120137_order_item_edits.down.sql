drop table if exists order_edits;
alter table orders drop column if exists paid_total;
drop trigger if exists trg_orders_bump_version on orders;
drop function if exists orders_bump_version();
alter table orders drop column if exists version;
