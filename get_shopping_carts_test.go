package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-redis/redismock/v8"
)

func TestGetShoppingCartHandler_Success(t *testing.T) {
	// Create a new Redis mock
	redisDB, mock := redismock.NewClientMock()

	// Set up the expected Redis GET response
	shoppingCartJSON, _ := json.Marshal(shoppingCart)
	mock.ExpectGet(shoppingCart.IPAddress).SetVal(string(shoppingCartJSON))

	// Create a new request with the IP address query parameter
	req, err := http.NewRequest("GET", "/shopping_cart?ip_address="+shoppingCart.IPAddress, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a new response recorder
	rr := httptest.NewRecorder()

	// Call the handler function
	sch := ShoppingCartsHandler{db: nil, redisClient: redisDB}
	sch.getShoppingCartHandler(rr, req)

	// Check the response status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check the response body
	expectedResponseBody, _ := json.Marshal(shoppingCart)
	// Adding line jump to match the expected body string
	expectedResponseBodyString := string(expectedResponseBody) + "\n"
	if rr.Body.String() != expectedResponseBodyString {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expectedResponseBodyString)
	}

	// Verify that the Redis GET command was called with the correct IP address
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestGetShoppingCartHandler_BadRequest(t *testing.T) {
	// Create a new Redis mock
	redisDB, _ := redismock.NewClientMock()

	// Create a new request without the IP address query parameter
	req, err := http.NewRequest("GET", "/shopping_cart", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a new response recorder
	rr := httptest.NewRecorder()

	// Call the handler function
	sch := ShoppingCartsHandler{db: nil, redisClient: redisDB}
	sch.getShoppingCartHandler(rr, req)

	// Check the response status code
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}
}

func TestGetShoppingCartHandler_NotFound(t *testing.T) {
	// Create a new Redis mock
	redisDB, mock := redismock.NewClientMock()

	// Set up the expected Redis GET response
	mock.ExpectGet(shoppingCart.IPAddress).RedisNil()

	// Create a new request with the IP address query parameter
	req, err := http.NewRequest("GET", "/shopping_cart?ip_address="+shoppingCart.IPAddress, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a new response recorder
	rr := httptest.NewRecorder()

	// Call the handler function
	sch := ShoppingCartsHandler{db: nil, redisClient: redisDB}
	sch.getShoppingCartHandler(rr, req)

	// Check the response status code
	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusNotFound)
	}

	// Check the response body
	expectedBody := "Shopping cart not found\n"
	if body := rr.Body.String(); body != expectedBody {
		t.Errorf("handler returned unexpected body: got %v want %v", body, expectedBody)
	}

	// Verify that all Redis expectations were met
	err = mock.ExpectationsWereMet()
	if err != nil {
		t.Errorf("there were unfulfilled Redis expectations: %s", err)
	}
}

func TestGetShoppingCartHandler_InternalError(t *testing.T) {
	// Create a new Redis mock
	redisDB, mock := redismock.NewClientMock()

	// Set up the expected Redis GET response
	mock.ExpectGet(shoppingCart.IPAddress).SetErr(errors.New("error"))

	// Create a new request with the IP address query parameter
	req, err := http.NewRequest("GET", "/shopping_cart?ip_address="+shoppingCart.IPAddress, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a new response recorder
	rr := httptest.NewRecorder()

	// Call the handler function
	sch := ShoppingCartsHandler{db: nil, redisClient: redisDB}
	sch.getShoppingCartHandler(rr, req)

	// Check the response status code
	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusInternalServerError)
	}

	// Check the response body
	expectedBody := "error\n"
	if body := rr.Body.String(); body != expectedBody {
		t.Errorf("handler returned unexpected body: got %v want %v", body, expectedBody)
	}

	// Verify that all Redis expectations were met
	err = mock.ExpectationsWereMet()
	if err != nil {
		t.Errorf("there were unfulfilled Redis expectations: %s", err)
	}
}
