-- +goose Up
CREATE TABLE IF NOT EXISTS products(
    id          UUID          NOT NULL PRIMARY KEY,
    title       VARCHAR(255)  NOT NULL,
    price       BIGINT        NOT NULL,
    description TEXT          NOT NULL,
    created_at  TIMESTAMP     NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP     NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE products;
