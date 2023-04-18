package main

import (
	"encoding/json"
	"net/http"

	"github.com/go-redis/redis/v8"
)

// getShoppingCartHandler is an HTTP handler function that retrieves a shopping cart record
// from Redis based on the IP address query parameter and returns it as a JSON response.
//
// If the IP address query parameter is missing or empty, it will return an HTTP bad request error (400).
// If the shopping cart record is not found in Redis, it will return an HTTP not found error (404).
// If there is an error while retrieving or unmarshaling the shopping cart record, it will return an HTTP internal server error (500).
func (sch ShoppingCartsHandler) getShoppingCartHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the IP address from the query parameters
	ipAddress := r.URL.Query().Get("ip_address")
	if ipAddress == "" {
		http.Error(w, "ip_address query parameter is required", http.StatusBadRequest)
		return
	}

	// Get the shopping cart record from Redis
	shoppingCartJSON, err := sch.redisClient.Get(r.Context(), ipAddress).Bytes()
	if err != nil {
		if err == redis.Nil {
			http.Error(w, "Shopping cart not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Unmarshal the shopping cart record from JSON
	var shoppingCart ShoppingCart
	err = json.Unmarshal(shoppingCartJSON, &shoppingCart)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return the retrieved shopping cart
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(shoppingCart)
}
