package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// getProduct retrieves a single product from the database by ID and returns it as a JSON response.
//
// It expects the ID of the product to be provided as a URL parameter. If the ID is not a valid integer, it returns an HTTP 400 Bad Request error.
// If the product is not found in the database, it returns an HTTP 404 Not Found error.
// If there is an error while querying the database, it returns an HTTP 500 Internal Server Error.
//
// It uses the db variable, which should be a connection pool to a PostgreSQL database, to execute the query.
func (ph ProductsHandler) getProduct(w http.ResponseWriter, r *http.Request) {
	// Extract product ID from URL parameter
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	// Build SQL query
	sqlQuery := "SELECT id, name, price, description, categories, images, referenced_name, date_added FROM products WHERE id = $1"

	// Execute query
	row := ph.db.QueryRow(sqlQuery, id)

	// Scan product
	p := Product{}
	err = row.Scan(&p.ID, &p.Name, &p.Price, &p.Description, (*textArray)(&p.Categories), (*textArray)(&p.Images), &p.ReferencedName, &p.DateAdded)
	if err == sql.ErrNoRows {
		http.Error(w, "Product not found", http.StatusNotFound)
		return
	} else if err != nil {
		log.Println(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Encode and send response
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(p)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
