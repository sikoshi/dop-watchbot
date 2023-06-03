-- products
CREATE TABLE arbuz_products
(
    id  integer not null constraint arbuz_products_pkey primary key,
    catalog_id        integer,
    name              varchar,
    producer_country  varchar,
    brand_name        varchar,
    description       text,
    uri               varchar,
    image             varchar,
    measure           varchar,
    is_weighted       boolean,
    weight_avg        double precision,
    weight_min        double precision,
    weight_max        double precision,
    piece_weight_max  double precision,
    quantity_min_step double precision,
    barcode           varchar,
    is_available      boolean,
    is_local          boolean,
    created_at        timestamp with time zone default CURRENT_TIMESTAMP
);

-- prices
CREATE TABLE arbuz_prices
(
    product_id    integer not null,
    product_price integer not null,
    is_last       boolean,
    created_at    timestamp with time zone default CURRENT_TIMESTAMP
);

CREATE INDEX arbuz_price_product on arbuz_prices (product_id);
CREATE INDEX arbuz_price_product_last on arbuz_prices (product_id, is_last);