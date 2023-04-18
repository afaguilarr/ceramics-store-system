package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gorilla/mux"
)

func sliceToPostgreSQLArray(elements []string) []uint8 {
	var quotedElements []string
	for _, element := range elements {
		quotedElements = append(quotedElements, fmt.Sprintf("'%s'", element))
	}
	return []uint8("{" + strings.Join(quotedElements, ",") + "}")
}

func TestGetProduct_Success(t *testing.T) {
	// Create mock DB
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}
	defer db.Close()

	// Expected product
	expectedProduct := Product{
		ID:             1,
		Name:           "Test Product",
		Price:          10.99,
		Description:    "This is a test product",
		Categories:     []string{"category1", "category2"},
		Images:         []string{"image1.jpg", "image2.jpg"},
		ReferencedName: "test-reference",
		DateAdded:      time.Now(),
	}

	// Set proper formats for SQL Arrays
	postgreSQLArrayCategories := sliceToPostgreSQLArray(expectedProduct.Categories)
	postgreSQLArrayImages := sliceToPostgreSQLArray(expectedProduct.Images)

	// Set expectations on mock
	rows := sqlmock.NewRows([]string{"id", "name", "price", "description", "categories", "images", "referenced_name", "date_added"}).
		AddRow(expectedProduct.ID, expectedProduct.Name, expectedProduct.Price, expectedProduct.Description, postgreSQLArrayCategories, postgreSQLArrayImages, expectedProduct.ReferencedName, expectedProduct.DateAdded)
	mock.ExpectQuery("SELECT id, name, price, description, categories, images, referenced_name, date_added FROM products WHERE id = ?").
		WithArgs(1).
		WillReturnRows(rows)

	// Make request to handler
	req, err := http.NewRequest(http.MethodGet, "/products/1", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	// Set the URL parameters for the request
	vars := map[string]string{"id": "1"}
	req = mux.SetURLVars(req, vars)

	// Set Products Handler
	ph := ProductsHandler{db: db}

	// Set up response recorder
	rr := httptest.NewRecorder()
	// Call function
	handler := http.HandlerFunc(ph.getProduct)
	handler.ServeHTTP(rr, req)

	// Check response status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check response body
	expectedBody, err := json.Marshal(expectedProduct)
	if err != nil {
		t.Fatalf("failed to marshal expected body: %v", err)
	}

	// The rr.Body.String() expects a new line at the end for some reason
	expectedBody = append(expectedBody, []byte("\n")...)
	if rr.Body.String() != string(expectedBody) {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), string(expectedBody))
	}

	// Check content-type header
	if contentType := rr.Header().Get("Content-Type"); contentType != "application/json" {
		t.Errorf("handler returned wrong content-type header: got %v want %v", contentType, "application/json")
	}

	// Check mock expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled mock expectations: %s", err)
	}
}

func TestGetProduct_ErrorWhileQuerying(t *testing.T) {
	// Create mock DB
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// Set up expected query and result
	expectedErr := errors.New("some error")
	mock.ExpectQuery("SELECT id, name, price, description, categories, images, referenced_name, date_added FROM products WHERE id = ?").
		WithArgs(1).
		WillReturnError(expectedErr)

	// Set up request
	req, err := http.NewRequest(http.MethodGet, "/products/1", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Set the URL parameters for the request
	vars := map[string]string{"id": "1"}
	req = mux.SetURLVars(req, vars)

	// Set Products Handler
	ph := ProductsHandler{db: db}

	// Set up response recorder
	rr := httptest.NewRecorder()
	// Call function
	handler := http.HandlerFunc(ph.getProduct)
	handler.ServeHTTP(rr, req)

	// Check response status code and body
	if rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected status code %v but got %v", http.StatusInternalServerError, rr.Code)
	}

	expectedBody := "Internal server error\n"
	if rr.Body.String() != expectedBody {
		t.Errorf("Expected body %q but got %q", expectedBody, rr.Body.String())
	}

	// Check mock expectations
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %s", err)
	}
}

func TestGetProduct_ProductNotFound(t *testing.T) {
	// Create mock DB
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// Set up expected query and result
	mock.ExpectQuery("SELECT id, name, price, description, categories, images, referenced_name, date_added FROM products WHERE id = ?").
		WithArgs(1).
		WillReturnError(sql.ErrNoRows)

	// Set up request
	req, err := http.NewRequest(http.MethodGet, "/products/1", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Set the URL parameters for the request
	vars := map[string]string{"id": "1"}
	req = mux.SetURLVars(req, vars)

	// Set Products Handler
	ph := ProductsHandler{db: db}

	// Set up response recorder
	rr := httptest.NewRecorder()
	// Call function
	handler := http.HandlerFunc(ph.getProduct)
	handler.ServeHTTP(rr, req)

	// Check response status code and body
	if rr.Code != http.StatusNotFound {
		t.Errorf("Expected status code %v but got %v", http.StatusNotFound, rr.Code)
	}

	expectedBody := "Product not found\n"
	if rr.Body.String() != expectedBody {
		t.Errorf("Expected body %q but got %q", expectedBody, rr.Body.String())
	}

	// Check mock expectations
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %s", err)
	}
}

func TestGetProduct_InvalidID(t *testing.T) {
	// Make request to handler with invalid ID
	req, err := http.NewRequest(http.MethodGet, "/products/invalid", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	// Set the URL parameters for the request
	vars := map[string]string{"id": "invalid"}
	req = mux.SetURLVars(req, vars)

	// Set Products Handler
	ph := ProductsHandler{db: nil}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(ph.getProduct)
	handler.ServeHTTP(rr, req)

	// Check response status code
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}

	// Check response body
	expectedBody := "Invalid product ID\n"
	if rr.Body.String() != expectedBody {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expectedBody)
	}

	// Check content-type header
	if contentType := rr.Header().Get("Content-Type"); contentType != "text/plain; charset=utf-8" {
		t.Errorf("handler returned wrong content-type header: got %v want %v", contentType, "text/plain; charset=utf-8")
	}
}
