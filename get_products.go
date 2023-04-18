package main

import (
	"encoding/json"
	"log"
	"net/http"
)

// getProducts retrieves a list of products from the database and sends a JSON response.
//
// Query parameters can be used to filter the results by name, referenced name, category,
// or a list of categories. The results can also be ordered by price or date added.
//
// If there is an internal server error, it returns a 500 Internal Server Error.
func (ph ProductsHandler) getProducts(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	name := r.URL.Query().Get("name")
	refName := r.URL.Query().Get("referenced_name")
	categories := r.URL.Query()["categories"]
	order := r.URL.Query().Get("order")

	// Build SQL query
	sqlQuery := "SELECT id, name, price, description, categories, images, referenced_name, date_added FROM products"

	// Add filters
	//
	// Add name filter
	if name != "" {
		sqlQuery += " WHERE name ILIKE '%" + name + "%'"
	}

	// Add refname filter
	if refName != "" {
		if name == "" {
			sqlQuery += " WHERE referenced_name ILIKE '%" + refName + "%'"
		} else {
			sqlQuery += " AND referenced_name ILIKE '%" + refName + "%'"
		}
	}

	// Add categories filter
	if len(categories) > 0 {
		if name == "" && refName == "" {
			sqlQuery += " WHERE "
		} else {
			sqlQuery += " AND "
		}
		for i, category := range categories {
			sqlQuery += "'" + category + "' = ANY(categories)"
			if i < len(categories)-1 {
				sqlQuery += " OR "
			}
		}
	}

	// Add order by
	if order != "" {
		switch order {
		case "price_asc":
			sqlQuery += " ORDER BY price ASC"
		case "price_desc":
			sqlQuery += " ORDER BY price DESC"
		case "date_asc":
			sqlQuery += " ORDER BY date_added ASC"
		default:
			sqlQuery += " ORDER BY date_added DESC"
		}
	} else {
		sqlQuery += " ORDER BY date_added DESC"
	}

	// Execute query
	rows, err := ph.db.Query(sqlQuery)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Collect products
	products := []Product{}
	for rows.Next() {
		p := Product{}
		err := rows.Scan(&p.ID, &p.Name, &p.Price, &p.Description, (*textArray)(&p.Categories), (*textArray)(&p.Images), &p.ReferencedName, &p.DateAdded)
		if err != nil {
			log.Println(err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		products = append(products, p)
	}

	// Encode and send response
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(products)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
