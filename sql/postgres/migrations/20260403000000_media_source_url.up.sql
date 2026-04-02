alter table media add column source_url text null;

create unique index idx_media_source_url on media (source_url) where source_url is not null;
