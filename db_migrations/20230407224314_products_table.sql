-- +goose Up
-- +goose StatementBegin
CREATE TABLE products (
  id SERIAL PRIMARY KEY,
  name TEXT NOT NULL,
  price DECIMAL(10,2) NOT NULL,
  description TEXT NOT NULL,
  categories TEXT[],
  images TEXT[],
  referenced_name TEXT,
  date_added TIMESTAMP WITH TIME ZONE NOT NULL
);

INSERT INTO products (name, price, description, categories, images, referenced_name, date_added)
VALUES
  ('Product 1', 9.99, 'This is product 1', ARRAY['Category A', 'Category B'], ARRAY['image1.jpg', 'image2.jpg'], 'Reference 1', NOW()),
  ('Product 2', 14.99, 'This is product 2', ARRAY['Category B', 'Category C'], ARRAY['image3.jpg', 'image4.jpg'], 'Reference 2', NOW()),
  ('Product 3', 19.99, 'This is product 3', ARRAY['Category A', 'Category C'], ARRAY['image5.jpg', 'image6.jpg'], 'Reference 3', NOW());
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS products;
-- +goose StatementEnd
