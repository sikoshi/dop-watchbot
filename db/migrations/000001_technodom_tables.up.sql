-- products
CREATE TABLE technodom_products
(
    sku                 varchar not null primary key,
    title               varchar not null,
    brand               varchar null,
    uri                 varchar null
);

-- prices
CREATE TABLE technodom_prices
(
    product_sku     varchar not null,
    product_price   integer not null,
    is_last         boolean,
    created_at      timestamp with time zone default CURRENT_TIMESTAMP
);

CREATE INDEX technodom_price_product ON technodom_prices (product_sku);
CREATE INDEX technodom_price_product_last ON technodom_prices (product_sku, is_last);