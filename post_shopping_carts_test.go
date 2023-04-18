package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-redis/redismock/v8"
	"github.com/gorilla/mux"
)

var shoppingCart = ShoppingCart{
	ID:        1,
	UserID:    1,
	IPAddress: "127.0.0.1",
	ShoppingCartItems: []ShoppingCartItem{
		{
			ID:               1,
			ProductID:        1,
			ShoppingCartID:   1,
			NumberOfProducts: 2,
		},
		{
			ID:               2,
			ProductID:        2,
			ShoppingCartID:   1,
			NumberOfProducts: 1,
		},
	},
}

func TestUpsertShoppingCartSuccess(t *testing.T) {
	// create a mock Redis DB
	redisDB, mock := redismock.NewClientMock()

	// marshal shopping cart to JSON
	shoppingCartJSON, err := json.Marshal(shoppingCart)
	if err != nil {
		t.Fatal(err)
	}

	// set expectations for the shopping cart being stored in the Redis DB with a TTL of 24 hours
	key := shoppingCart.IPAddress
	expectedTTL := 24 * time.Hour
	mockExpect := mock.ExpectSet(key, shoppingCartJSON, expectedTTL)
	mockExpect.SetVal("OK")

	// create a new request with the shopping cart JSON as the body
	req := httptest.NewRequest(http.MethodPost, "/shopping_carts", strings.NewReader(string(shoppingCartJSON)))

	// create a new response recorder
	rr := httptest.NewRecorder()

	// create a new router and add the upsertShoppingCartHandler handler function
	r := mux.NewRouter()
	sch := ShoppingCartsHandler{db: nil, redisClient: redisDB}
	r.HandleFunc("/shopping_carts", sch.upsertShoppingCartHandler).Methods(http.MethodPost)

	// serve the request
	r.ServeHTTP(rr, req)

	// check the response status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// check the response content type
	if contentType := rr.Header().Get("Content-Type"); contentType != "application/json" {
		t.Errorf("handler returned wrong content type: got %v want application/json", contentType)
	}

	// unmarshal the response body into a shopping cart struct
	var savedShoppingCart ShoppingCart
	if err := json.NewDecoder(rr.Body).Decode(&savedShoppingCart); err != nil {
		t.Fatal(err)
	}

	// check the saved shopping cart's IP address matches the original shopping cart's IP address
	if savedShoppingCart.IPAddress != shoppingCart.IPAddress {
		t.Errorf("saved shopping cart IP address = %v, want %v", savedShoppingCart.IPAddress, shoppingCart.IPAddress)
	}

	// check the saved shopping cart has the correct number of shopping cart items
	if len(savedShoppingCart.ShoppingCartItems) != len(shoppingCart.ShoppingCartItems) {
		t.Errorf("saved shopping cart has %v shopping cart items, want %v", len(savedShoppingCart.ShoppingCartItems), len(shoppingCart.ShoppingCartItems))
	}

	// wait for the expectations to be met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("redis expectations were not met: %s", err.Error())
	}
}

func TestUpsertShoppingCartInvalidJSON(t *testing.T) {
	// create a new request with invalid shopping cart JSON as the body
	req := httptest.NewRequest(http.MethodPost, "/shopping_carts", strings.NewReader("invalid JSON"))

	// create a new response recorder
	rr := httptest.NewRecorder()

	// create a new router and add the upsertShoppingCartHandler handler function
	r := mux.NewRouter()
	sch := ShoppingCartsHandler{db: nil, redisClient: nil}
	r.HandleFunc("/shopping_carts", sch.upsertShoppingCartHandler).Methods(http.MethodPost)

	// serve the request
	r.ServeHTTP(rr, req)

	// check the response status code
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}
}

func TestUpsertShoppingCartRedisError(t *testing.T) {
	// create a mock Redis DB
	redisDB, mock := redismock.NewClientMock()

	// marshal shopping cart to JSON
	shoppingCartJSON, err := json.Marshal(shoppingCart)
	if err != nil {
		t.Fatal(err)
	}

	// create a new request with a valid shopping cart JSON as the body
	req := httptest.NewRequest(http.MethodPost, "/shopping_carts", strings.NewReader(string(shoppingCartJSON)))

	// create a new response recorder
	rr := httptest.NewRecorder()

	// set expectations for the shopping cart being stored in the Redis DB with a TTL of 24 hours
	key := shoppingCart.IPAddress
	expectedTTL := 24 * time.Hour
	mockExpect := mock.ExpectSet(key, shoppingCartJSON, expectedTTL)
	mockExpect.SetErr(errors.New("Redis command error"))

	// create a new router and add the upsertShoppingCartHandler handler function
	r := mux.NewRouter()
	sch := ShoppingCartsHandler{db: nil, redisClient: redisDB}
	r.HandleFunc("/shopping_carts", sch.upsertShoppingCartHandler).Methods(http.MethodPost)

	// serve the request
	r.ServeHTTP(rr, req)

	// check the response status code
	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusInternalServerError)
	}
}
