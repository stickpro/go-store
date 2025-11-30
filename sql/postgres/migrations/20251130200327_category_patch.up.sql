create table category_paths
(
    ancestor_id   uuid not null references categories (id) on delete cascade,
    descendant_id uuid not null references categories (id) on delete cascade,
    depth         int  not null,
    primary key (ancestor_id, descendant_id)
);

create index idx_category_paths_descendant on category_paths (descendant_id);
create index idx_category_paths_depth on category_paths (depth);
