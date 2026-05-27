-- +goose Up
CREATE TABLE IF NOT EXISTS product_images (
    id         UUID     NOT NULL PRIMARY KEY,
    product_id UUID     NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    url        TEXT     NOT NULL,
    object_key TEXT     NOT NULL,
    position   SMALLINT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX ON product_images(product_id);

-- +goose Down
DROP TABLE product_images;
