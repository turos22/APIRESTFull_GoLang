-- +goose Up
    CREATE TABLE IF NOT EXISTS category(
        id BIGSERIAL PRIMARY KEY,
        name varchar(150) NOT NULL
    );

-- +goose Down
drop table IF EXISTS category;
