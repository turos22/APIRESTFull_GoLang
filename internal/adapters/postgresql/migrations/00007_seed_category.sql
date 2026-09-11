-- +goose Up
ALTER TABLE category ADD CONSTRAINT uq_category_name UNIQUE (name);

INSERT INTO category (name) VALUES
    ('Eletronicos'),
    ('Roupas'),
    ('Casa e Cozinha'),
    ('Livros'),
    ('Esportes')
ON CONFLICT (name) DO NOTHING;

-- +goose Down
DELETE FROM category WHERE name IN
    ('Eletronicos', 'Roupas', 'Casa e Cozinha', 'Livros', 'Esportes');

ALTER TABLE category DROP CONSTRAINT IF EXISTS uq_category_name;
