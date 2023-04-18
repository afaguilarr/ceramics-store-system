# ceramics-store-system
This repository contains all the necessary code to build and run the ceramics store.

# Docs
All documentation is stored here: https://drive.google.com/drive/folders/1NPq6T9mQcRmBjX2bb0PuO5_lxnBxXCht
If you need access, please send an email to afaguilarr@unal.edu.co

# Apply migrations

Outside the container:
```
go run github.com/pressly/goose/v3/cmd/goose postgres "postgres://user:password@localhost/products_db?sslmode=disable" up
```

Inside the container:
```
go run github.com/pressly/goose/v3/cmd/goose postgres "postgres://user:password@postgres_db/products_db?sslmode=disable" up
```

# Add new migrations
```
cd db_migrations
go run github.com/pressly/goose/v3/cmd/goose create shopping_cart_tables sql
```

# Run unit tests with coverage

```
go test -coverprofile=coverage.out ./...
```
```
go tool cover -html=coverage.out -o coverage.html
```

# Test endpoints

name: Return all products that have a name containing the specified string. For example, /products?name=apple would return all products that have "apple" in their name.

referenced_name: Return all products that have a referenced_name containing the specified string. For example, /products?referenced_name=John would return all products that have "John" in their referenced_name.

category: Return all products that belong to the specified category. For example, /products?category=Electronics would return all products that belong to the "Electronics" category.

categories: Return all products that belong to any of the specified categories. This parameter can be repeated to search for multiple categories. For example, /products?categories=Electronics&categories=Computers would return all products that belong to either the "Electronics" or "Computers" category.

order: Return all products sorted by price in either ascending or descending order. For example, /products?order=asc would return all products sorted by price in ascending order.
