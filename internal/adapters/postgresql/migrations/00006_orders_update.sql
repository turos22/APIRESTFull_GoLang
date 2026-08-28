-- +goose Up
   ALTER TABLE orders add COLUMN IF NOT EXISTS status VARCHAR(30) NULL;
   ALTER TABLE orders add COLUMN IF NOT EXISTS total_cents INTEGER NULL;
   ALTER TABLE orders add COLUMN IF NOT EXISTS customer_id BIGINT NULL;

   alter table orders add constraint FK_orders_users  FOREIGN KEY (customer_id) references users (id);
   ALTER TABLE ORDER_items ADD CONSTRAINT FK_ORDERS_items_PRODUCTS FOREIGN KEY (product_id) REFERENCES PRODUCTS (ID);

    CREATE INDEX idx_orders_customer_id ON orders (customer_id);
    create index idx_orders_product_id on ORDER_items (product_id);
    create index idx_orders_order_id on order_items (order_id);
-- +goose Down
    ALTER TABLE orders DROP COLUMN IF EXISTS status;
    ALTER TABLE orders DROP COLUMN IF EXISTS total_cents;
    ALTER TABLE orders DROP COLUMN IF EXISTS customer_id;

    drop index idx_orders_customer_id;
    drop index idx_orders_product_id;
    drop index idx_orders_order_id;

    alter table orders drop constraint IF EXISTS FK_orders_users;
    alter table ORDER_items drop constraint IF EXISTS FK_ORDERS_PRODUCTS;
