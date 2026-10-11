-- cities now holds every settlement from the GAR (ФИАС) registry, not only the
-- ~1100 hflabs cities. GAR has no coordinates, population, timezone or
-- foundation year, and not every settlement has a postal code / KLADR / OKATO,
-- so those columns become optional. Rows are upserted by fias_id so re-imports
-- keep ids stable for orders.ship_city_id.
alter table cities
    alter column postal_code drop not null,
    alter column city_type drop not null,
    alter column city drop not null,
    alter column kladr_id drop not null,
    alter column okato drop not null,
    alter column oktmo drop not null,
    alter column tax_office drop not null,
    alter column timezone drop not null,
    alter column geo_lat drop not null,
    alter column geo_lon drop not null,
    alter column foundation_year drop not null,
    alter column population set default 0,
    alter column capital_marker set default 0,
    alter column address type varchar(512),
    alter column area type varchar(255),
    alter column area_type type varchar(50),
    alter column city type varchar(255),
    alter column city_type type varchar(50),
    alter column settlement type varchar(255),
    alter column settlement_type type varchar(50);

create unique index uq_cities_fias_id on cities (fias_id);
create index idx_cities_settlement on cities (settlement);
