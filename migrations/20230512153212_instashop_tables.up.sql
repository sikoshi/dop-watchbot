-- stores
CREATE TABLE instashop_stores
(
    store_id    serial primary key,
    slug        varchar,
    title       varchar,
    link        varchar,
    created_at  timestamp with time zone default CURRENT_TIMESTAMP
);
CREATE INDEX idx_instashop_store_slug ON instashop_stores(slug);

-- categories
CREATE TABLE instashop_categories
(
    category_id integer primary key,
    store_id    integer,
    title       varchar,
    link        varchar,
    created_at  timestamp with time zone default CURRENT_TIMESTAMP
);
CREATE INDEX idx_instashop_categories_store ON instashop_categories(store_id);

-- products
CREATE TABLE instashop_products
(
    product_id  integer primary key,
    store_id    integer,
    category_id integer,
    brand       varchar,
    title       varchar,
    link        varchar,
    created_at  timestamp with time zone default CURRENT_TIMESTAMP
);
CREATE INDEX idx_instashop_products_store ON instashop_products(store_id);
CREATE INDEX idx_instashop_products_category ON instashop_products(category_id);

-- prices
create table instashop_prices
(
    store_id        integer,
    product_id      integer not null,
    product_price   integer not null,
    is_last         boolean,
    created_at      timestamp with time zone default CURRENT_TIMESTAMP
);

CREATE INDEX instashop_price_store ON instashop_prices (store_id);
CREATE INDEX instashop_price_product ON instashop_prices (product_id);
CREATE INDEX instashop_price_product_last ON instashop_prices (product_id, is_last);