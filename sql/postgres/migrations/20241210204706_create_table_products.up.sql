CREATE TABLE products
(
    id               UUID                    DEFAULT gen_random_uuid() NOT NULL
        PRIMARY KEY,
    name             VARCHAR(255)   NOT NULL,
    model            VARCHAR(255)   NOT NULL UNIQUE,
    slug             VARCHAR(255)   NOT NULL UNIQUE,
    description      TEXT,
    meta_title       VARCHAR(255),
    meta_h1          VARCHAR(255),
    meta_description VARCHAR(400),
    meta_keyword     varchar(255),
    sku              VARCHAR(64)             DEFAULT NULL,
    upc              VARCHAR(12)             DEFAULT NULL,
    ean              VARCHAR(14)             DEFAULT NULL,
    jan              VARCHAR(13)             DEFAULT NULL,
    isbn             VARCHAR(13)             DEFAULT NULL,
    mpn              VARCHAR(64)             DEFAULT NULL,
    location         VARCHAR(128)            DEFAULT NULL,
    quantity         BIGINT         NOT NULL DEFAULT 0,
    stock_status     VARCHAR(255),
    image            VARCHAR(255)            DEFAULT NULL,
    manufacturer_id  UUID                    DEFAULT NULL,
    price            DECIMAL(15, 4) NOT NULL DEFAULT 0.0000,
    weight           DECIMAL(15, 8) NOT NULL DEFAULT 0.00000000,
    length           DECIMAL(15, 8) NOT NULL DEFAULT 0.00000000,
    width            DECIMAL(15, 8) NOT NULL DEFAULT 0.00000000,
    height           DECIMAL(15, 8) NOT NULL DEFAULT 0.00000000,
    subtract         BOOLEAN        NOT NULL DEFAULT TRUE,
    minimum          BIGINT         NOT NULL DEFAULT 1,
    sort_order       INT            NOT NULL DEFAULT 0,
    is_enable        BOOLEAN        NOT NULL DEFAULT TRUE,
    viewed           BIGINT         NOT NULL DEFAULT 0,
    created_at       TIMESTAMP      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at       TIMESTAMP      NULL     DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_products_slug ON products (slug);
CREATE INDEX idx_products_model ON products (model);