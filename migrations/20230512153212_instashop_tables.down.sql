-- stores
DROP TABLE IF EXISTS instashop_stores;
DROP INDEX IF EXISTS idx_instashop_store_slug;

-- categories
DROP TABLE IF EXISTS instashop_categories;
DROP INDEX IF EXISTS idx_instashop_categories_slug;
DROP INDEX IF EXISTS idx_instashop_categories_store;

-- products
DROP TABLE IF EXISTS instashop_products;
DROP INDEX IF EXISTS idx_instashop_products_slug;
DROP INDEX IF EXISTS idx_instashop_products_store;
DROP INDEX IF EXISTS idx_instashop_products_category;

-- prices
DROP TABLE IF EXISTS instashop_prices;
DROP INDEX IF EXISTS instashop_price_store;
DROP INDEX IF EXISTS instashop_price_product;
DROP INDEX IF EXISTS instashop_price_product_last;