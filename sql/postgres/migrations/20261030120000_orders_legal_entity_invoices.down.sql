drop table if exists invoices;
drop sequence if exists invoice_number_seq;

alter table orders
    drop constraint if exists orders_legal_requisites_check,
    drop constraint if exists orders_customer_type_check,
    drop column if exists legal_address,
    drop column if exists legal_ogrn,
    drop column if exists legal_kpp,
    drop column if exists legal_inn,
    drop column if exists legal_name,
    drop column if exists customer_type;
