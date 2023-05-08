-- auto-generated definition
create table arbuz_prices
(
    product_id    integer not null,
    product_price integer not null,
    is_last       boolean,
    created_at    timestamp with time zone default CURRENT_TIMESTAMP
);

create index arbuz_price_product
    on arbuz_prices (product_id);

create index arbuz_price_product_last
    on arbuz_prices (product_id, is_last);
 