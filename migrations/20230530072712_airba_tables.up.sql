-- products
create table airba_fresh_products
(
    sku                 varchar not null primary key,
    title               varchar not null,
    brand               varchar null,
    uri                 varchar null,
    merchant_code       varchar null,
    merchant_name       varchar null,
    measurement_code    varchar null,
    measurement_name    varchar null,
    measurement_step    varchar null
);

-- prices
create table airba_fresh_prices
(
    product_sku     varchar not null,
    product_price   integer not null,
    is_last         boolean,
    created_at      timestamp with time zone default CURRENT_TIMESTAMP
);

CREATE INDEX airba_fresh_price_product ON airba_fresh_prices (product_sku);
CREATE INDEX airba_fresh_price_product_last ON airba_fresh_prices (product_sku, is_last);