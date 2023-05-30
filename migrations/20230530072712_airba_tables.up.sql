-- products
create table airba_products
(

);


-- prices
create table airba_prices
(
    product_id      integer not null,
    product_price   integer not null,
    is_last         boolean,
    created_at      timestamp with time zone default CURRENT_TIMESTAMP
);
CREATE INDEX airba_price_product ON airba_prices (product_id);
CREATE INDEX airba_price_product_last ON airba_prices (product_id, is_last);