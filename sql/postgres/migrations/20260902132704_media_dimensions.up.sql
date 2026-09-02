alter table media
    add column if not exists width  int not null default 0,
    add column if not exists height int not null default 0;
