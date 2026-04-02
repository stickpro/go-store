drop index if exists idx_media_source_url;

alter table media drop column if exists source_url;
