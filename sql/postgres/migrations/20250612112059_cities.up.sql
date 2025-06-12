create table if not exists cities
(
    id uuid primary key default gen_random_uuid(),
    address varchar(255) not null,
    postal_code varchar(10) not null,
    country varchar(100) not null,
    federal_district varchar(255) not null,
    region_type varchar(50) not null,
    region varchar(100) not null,
    area_type varchar(50),
    area varchar(100),
    city_type varchar(10) not null,
    city varchar(100) not null,
    settlement_type varchar(10),
    settlement varchar(100),
    kladr_id varchar(20) not null,
    fias_id uuid not null,
    fias_level smallint not null,
    capital_marker smallint not null,
    okato varchar(11) not null,
    oktmo varchar(11) not null,
    tax_office varchar(11) not null,
    timezone varchar(50) not null,
    geo_lat numeric(13, 10) not null,
    geo_lon numeric(13, 10) not null,
    population bigint not null,
    foundation_year smallint not null
);

CREATE INDEX idx_cities_city ON cities(city);
CREATE INDEX idx_cities_country ON cities(country);
CREATE INDEX idx_cities_region ON cities(region);
CREATE INDEX idx_cities_federal_district ON cities(federal_district);
