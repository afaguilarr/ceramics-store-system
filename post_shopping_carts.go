package main

import (
	"encoding/json"
	"net/http"
	"time"
)

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
