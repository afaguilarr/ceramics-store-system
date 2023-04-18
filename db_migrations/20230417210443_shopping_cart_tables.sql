-- +goose Up
-- +goose StatementBegin
CREATE TABLE shopping_carts (
  id SERIAL PRIMARY KEY,
  user_id INT NOT NULL,
  ip_address VARCHAR(255) NOT NULL
);

CREATE TABLE shopping_cart_items (
  id SERIAL PRIMARY KEY,
  shopping_cart_id INT NOT NULL REFERENCES shopping_carts(id) ON DELETE CASCADE,
  product_id INT NOT NULL REFERENCES products(id),
  number_of_products INT NOT NULL
);

INSERT INTO shopping_carts (user_id, ip_address) VALUES
(1, '192.168.1.100'),
(2, '192.168.1.101'),
(3, '192.168.1.102');

INSERT INTO shopping_cart_items (shopping_cart_id, product_id, number_of_products) VALUES
(1, 1, 2),
(1, 2, 1),
(2, 3, 3),
(3, 1, 1),
(3, 2, 2);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS shopping_cart_items;
DROP TABLE IF EXISTS shopping_carts;
-- +goose StatementEnd
