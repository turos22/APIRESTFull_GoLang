-- +goose Up
ALTER TABLE products ADD COLUMN IF NOT EXISTS description varchar(150);
ALTER TABLE products ADD COLUMN IF NOT EXISTS image_url varchar(150);
ALTER TABLE products ADD COLUMN IF NOT EXISTS category_id BIGINT NULL;
ALTER TABLE products ADD COLUMN IF NOT EXISTS active BOOLEAN DEFAULT true;
ALTER TABLE products ADD COLUMN IF NOT EXISTS seller_id BIGINT NULL;

ALTER TABLE products ADD CONSTRAINT FK_products_users FOREIGN KEY (seller_id) REFERENCES users(id);

CREATE INDEX IF NOT EXISTS idx_products_seller_id ON products (seller_id);
CREATE INDEX IF NOT EXISTS idx_products_category ON products (category_id);

-- +goose Down
DROP INDEX IF EXISTS idx_products_seller_id;
DROP INDEX IF EXISTS idx_products_category;

ALTER TABLE products DROP CONSTRAINT IF EXISTS FK_products_users;

ALTER TABLE products DROP COLUMN IF EXISTS description;
ALTER TABLE products DROP COLUMN IF EXISTS image_url;
ALTER TABLE products DROP COLUMN IF EXISTS category_id;
ALTER TABLE products DROP COLUMN IF EXISTS active;
ALTER TABLE products DROP COLUMN IF EXISTS seller_id;