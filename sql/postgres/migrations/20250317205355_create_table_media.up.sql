create table media
(
    id         uuid                  default gen_random_uuid() not null
        primary key,
    name       varchar(255) not null,
    path       text         not null,
    file_name  varchar(255) not null,
    mime_type  varchar(64)  not null,
    disk_type  varchar(64)  not null,
    size       bigint                default 0,
    created_at timestamp    not null default current_timestamp
)