-- name: ListProducts :many
SELECT * FROM products;

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
SELECT * FROM users WHERE email = $1 and password_hash = $2;

-- name: Me :one
SELECT * FROM users WHERE id = $1;