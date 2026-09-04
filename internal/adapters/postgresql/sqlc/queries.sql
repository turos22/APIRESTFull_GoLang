-- name: ListProducts :many
-- name: ListProducts :many
-- name: ListProducts :many
SELECT id, name, price_in_cents, quantity, description, image_url, category_id, active, seller_id
FROM products
WHERE
    active = true
    AND (sqlc.narg('search')::text IS NULL OR name ILIKE '%' || sqlc.narg('search')::text || '%')
    AND (sqlc.narg('category_id')::bigint IS NULL OR category_id = sqlc.narg('category_id')::bigint)
    AND (sqlc.narg('min_price')::int IS NULL OR price_in_cents >= sqlc.narg('min_price')::int)
    AND (sqlc.narg('max_price')::int IS NULL OR price_in_cents <= sqlc.narg('max_price')::int)
ORDER BY id
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: FindProductByID :one
SELECT * FROM products WHERE id = $1;

-- name: CreateOrder :one
INSERT INTO orders(
    customer_id
) VALUES ($1) RETURNING *;

-- name: CreateOrderItem :one
INSERT INTO order_items(order_id, product_id, quantity, price_cents) 
VALUES ($1, $2, $3, $4) RETURNING *;

-- name: UpdateStock :one
UPDATE products SET quantity = $1 WHERE id = $2 RETURNING id, name, price_in_cents, quantity, created_at;

-- name: Register :one
INSERT INTO users(email, password_hash, name, role) VALUES ($1, $2, $3, $4) RETURNING *;

-- name: FindUserByEmailPassword :one
SELECT * FROM users WHERE email = $1;

-- name: Me :one
SELECT * FROM users WHERE id = $1;

-- name: CreateProduto :one
INSERT INTO products(name, price_in_cents, quantity, description, image_url, category_id, active, seller_id) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING *;

-- name: UpdateProduct :one
UPDATE products SET name = $1, price_in_cents = $2, quantity = $3, description = $4, image_url = $5, category_id = $6, active = $7, seller_id = $8 WHERE id = $9 RETURNING *;

-- name: DeleteProduct :exec
DELETE FROM products WHERE id = $1;

-- name: Meproducts :many
SELECT * FROM products WHERE seller_id = $1;

-- name: OrdersId :one
select * from orders where id = $1;

-- name: OrderItemsOrderId :many
select * from order_items where order_id = $1;

-- name: OrdersMe :many
select * from orders where customer_id = $1;

-- name: OrderMeId :one
select * from orders where id = $1 and customer_id = $2;

