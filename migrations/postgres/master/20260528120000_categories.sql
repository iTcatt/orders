-- +goose Up

CREATE TABLE categories (
    id   SERIAL       NOT NULL PRIMARY KEY,
    slug VARCHAR(100) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL
);

INSERT INTO categories (slug, name) VALUES
    ('face-care',      'Уход за лицом'),
    ('body-care',      'Уход за телом'),
    ('hair-care',      'Уход за волосами'),
    ('makeup',         'Декоративная косметика'),
    ('perfumery',      'Парфюмерия'),
    ('hand-nail-care', 'Уход за руками и ногтями'),
    ('kids-care',      'Детская косметика'),
    ('natural-care',   'Натуральная/органическая косметика'),
    ('sun-care',       'Солнцезащитные средства');

ALTER TABLE products
    ADD COLUMN category_id INT NOT NULL REFERENCES categories(id) ON DELETE RESTRICT;

CREATE INDEX ON products(category_id);

-- +goose Down
ALTER TABLE products DROP COLUMN category_id;
DROP TABLE categories;
