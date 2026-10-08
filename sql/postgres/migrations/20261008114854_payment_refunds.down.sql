alter table orders drop column if exists refunded_total;
drop table if exists payment_refunds;
alter table payments drop column if exists refunded_amount;
