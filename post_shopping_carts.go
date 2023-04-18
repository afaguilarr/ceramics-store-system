package main

import (
	"encoding/json"
	"net/http"
	"time"
)

// upsertShoppingCartHandler handles the HTTP request for upserting a shopping cart into Redis.
//
// It decodes the request body into a `ShoppingCart` struct, and then upserts it into Redis using the
// `IPAddress` as the key. If the upsert succeeds, the function returns the saved shopping cart as a JSON
// response. If any errors occur during decoding, upserting, or encoding the response, the function returns
// an HTTP error with an appropriate status code and message.
func (sch ShoppingCartsHandler) upsertShoppingCartHandler(w http.ResponseWriter, r *http.Request) {
	var shoppingCart ShoppingCart
	err := json.NewDecoder(r.Body).Decode(&shoppingCart)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ipAddress := shoppingCart.IPAddress

	// Convert the shopping cart to a JSON string
	shoppingCartJSON, err := json.Marshal(shoppingCart)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Upsert the shopping cart record in Redis using the IP address as the key
	err = sch.redisClient.Set(r.Context(), ipAddress, shoppingCartJSON, 24*time.Hour).Err()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return the saved shopping cart
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(shoppingCart)
}
