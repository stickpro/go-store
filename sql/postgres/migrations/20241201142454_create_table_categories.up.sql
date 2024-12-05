CREATE TABLE categories
(
    id               UUID                  DEFAULT gen_random_uuid() NOT NULL
        PRIMARY KEY,
    parent_id        UUID REFERENCES categories (id) ON DELETE CASCADE,
    name             VARCHAR(255) NOT NULL,
    slug             VARCHAR(255) NOT NULL UNIQUE,
    description      TEXT,
    meta_title       VARCHAR(255),
    meta_h1          VARCHAR(255),
    meta_description VARCHAR(400),
    meta_keyword     varchar(255),
    is_enable        BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at       TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at       TIMESTAMP    NULL     DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_categories_parent_id ON categories (parent_id);
CREATE INDEX idx_categories_slug ON categories (slug);