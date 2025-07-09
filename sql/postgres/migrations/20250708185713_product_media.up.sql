create table if not exists product_media
(
    product_id  uuid                     not null,
    media_id    uuid                     not null,
    sort_order   integer                  not null default 0,
    foreign key (product_id) references products (id) on delete cascade,
    foreign key (media_id) references media (id) on delete cascade
);