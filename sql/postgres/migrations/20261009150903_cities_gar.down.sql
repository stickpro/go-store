drop index if exists idx_cities_settlement;
drop index if exists uq_cities_fias_id;

-- GAR-only rows can't satisfy the old NOT NULL constraints.
delete from cities
where postal_code is null
   or city_type is null
   or city is null
   or kladr_id is null
   or okato is null
   or oktmo is null
   or tax_office is null
   or timezone is null
   or geo_lat is null
   or geo_lon is null
   or foundation_year is null
   or length(address) > 255
   or length(area) > 100
   or length(area_type) > 50
   or length(city) > 100
   or length(city_type) > 10
   or length(settlement) > 100
   or length(settlement_type) > 10;

alter table cities
    alter column address type varchar(255),
    alter column area type varchar(100),
    alter column city type varchar(100),
    alter column city_type type varchar(10),
    alter column settlement type varchar(100),
    alter column settlement_type type varchar(10),
    alter column population drop default,
    alter column capital_marker drop default,
    alter column postal_code set not null,
    alter column city_type set not null,
    alter column city set not null,
    alter column kladr_id set not null,
    alter column okato set not null,
    alter column oktmo set not null,
    alter column tax_office set not null,
    alter column timezone set not null,
    alter column geo_lat set not null,
    alter column geo_lon set not null,
    alter column foundation_year set not null;
