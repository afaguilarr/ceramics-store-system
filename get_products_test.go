package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

// getMockDB returns a new mock database connection and mock object for testing purposes.
// t is the testing.T object used for logging any errors that occur.
func getMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to set up mock database: %v", err)
	}
	return db, mock
}

// getExpectedProducts returns a slice of Product objects that can be used as expected values in tests.
func getExpectedProducts() []Product {
	return []Product{
		{ID: 1, Name: "Product A", Price: 10.0, Description: "Product A description", Categories: []string{"cat1", "cat2"}, Images: []string{"img1", "img2"}, ReferencedName: "Product B", DateAdded: time.Now().Add(time.Minute)},
		{ID: 2, Name: "Product B", Price: 20.0, Description: "Product B description", Categories: []string{"cat1", "cat3"}, Images: []string{"img3", "img4"}, ReferencedName: "Product C", DateAdded: time.Now()},
	}
}

// expectedQuery returns a SELECT query string for the 'products' table with the given order by clause.
func expectedQuery(orderBy string, nameFilter, refNameFilter bool, categoriesFiltered int) string {
	query := "SELECT id, name, price, description, categories, images, referenced_name, date_added FROM products"

	if nameFilter {
		query += " WHERE name ILIKE '%ARandomName%'"
	}

	if refNameFilter {
		if nameFilter {
			query += " AND "
		} else {
			query += " WHERE "
		}
		query += "referenced_name ILIKE '%ARandomReferencedName%'"
	}

	if categoriesFiltered > 0 {
		if nameFilter || refNameFilter {
			query += " AND "
		} else {
			query += " WHERE "
		}
		for i := 0; i < categoriesFiltered; i++ {
			if i > 0 {
				query += " OR "
			}
			// Note how we had to escape the parentheses here by actually escaping the backslash
			query = fmt.Sprintf("%s'Category%v' = ANY\\(categories\\)", query, i)
		}
	}

	query = fmt.Sprintf("%s ORDER BY %s", query, orderBy)
	return query
}

// getProductsURL returns a URL with the given parameters.
func getProductsURL(order string, nameFilter, refNameFilter bool, categoriesFiltered int) string {
	URL := "/products"
	if order == "" && !nameFilter && !refNameFilter && categoriesFiltered == 0 {
		return URL
	}

	URL = URL + "?"
	if order != "" {
		URL = fmt.Sprintf("%sorder=%s", URL, order)
	}

	if nameFilter {
		if order != "" {
			URL = fmt.Sprintf("%s&", URL)
		}
		URL = fmt.Sprintf("%sname=ARandomName", URL)
	}

	if refNameFilter {
		if order != "" || nameFilter {
			URL = fmt.Sprintf("%s&", URL)
		}
		URL = fmt.Sprintf("%sreferenced_name=ARandomReferencedName", URL)
	}

	if categoriesFiltered > 0 {
		if order != "" || nameFilter || refNameFilter {
			URL = fmt.Sprintf("%s&", URL)
		}
		for i := 0; i < categoriesFiltered; i++ {
			if i > 0 {
				URL = fmt.Sprintf("%s&", URL)
			}
			URL = fmt.Sprintf("%scategories=Category%v", URL, i)
		}
	}
	return URL
}

// getMockRows returns a mock sqlmock.Rows object populated with the given products slice.
func getMockRows(products []Product) *sqlmock.Rows {
	rows := sqlmock.NewRows([]string{"id", "name", "price", "description", "categories", "images", "referenced_name", "date_added"})
	for _, p := range products {
		rows.AddRow(p.ID, p.Name, p.Price, p.Description, sliceToPostgreSQLArray(p.Categories), sliceToPostgreSQLArray(p.Images), p.ReferencedName, p.DateAdded)
	}
	return rows
}

// makeRequest creates a new HTTP request with the GET method, target and nil body, serves it using the
// ProductsHandler with the provided mock database and records the response to a ResponseRecorder which is then returned.
// It is intended to be used in testing.
func makeRequest(t *testing.T, db *sql.DB, target string) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	ph := ProductsHandler{db: db}
	handler := http.HandlerFunc(ph.getProducts)
	req, err := http.NewRequest(http.MethodGet, target, nil)
	if err != nil {
		t.Fatalf("Failed to create HTTP request: %v", err)
	}
	handler.ServeHTTP(rr, req)
	return rr
}

// checkResponseCode compares the status code of the response with the expected status code and logs an error if they are not the same.
// t is the testing.T object used for logging any errors that occur.
func checkResponseCode(t *testing.T, status, expectedStatus int) {
	if status != expectedStatus {
		t.Errorf("Handler returned wrong status code: got %v, expected %v", status, expectedStatus)
	}
}

// checkResponseBody compares the response body string with the expected products slice and logs an error if they are not the same.
// t is the testing.T object used for logging any errors that occur.
func checkResponseBody(t *testing.T, body, expectedBody string, expectedProducts []Product) {
	if expectedProducts != nil && expectedBody == "" {
		expectedBodyBytes, err := json.Marshal(expectedProducts)
		if err != nil {
			t.Fatalf("failed to marshal expected body: %v", err)
		}
		// The rr.Body.String() expects a new line at the end for some reason
		expectedBodyBytes = append(expectedBodyBytes, []byte("\n")...)
		expectedBody = string(expectedBodyBytes)
	}

	if body != string(expectedBody) {
		t.Errorf("handler returned unexpected body: got %v want %v", body, expectedBody)
	}
}

// checkMockExpectations checks whether all expectations of the given mock object have been fulfilled and logs an error if not.
// t is the testing.T object used for logging any errors that occur.
func checkMockExpectations(t *testing.T, mock sqlmock.Sqlmock) {
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled mock expectations: %s", err)
	}
}

func TestGetProducts(t *testing.T) {
	tests := []struct {
		name               string
		order              string
		nameFilter         bool
		refNameFilter      bool
		categoriesFiltered int
		dbString           string
		expectedProducts   []Product
	}{
		{
			name:               "DEFAULT",
			order:              "",
			nameFilter:         false,
			refNameFilter:      false,
			categoriesFiltered: 0,
			dbString:           "date_added DESC",
			expectedProducts:   getExpectedProducts(),
		},
		{
			name:               "Filtering by name",
			order:              "",
			nameFilter:         true,
			refNameFilter:      false,
			categoriesFiltered: 0,
			dbString:           "date_added DESC",
			expectedProducts:   getExpectedProducts(),
		},
		{
			name:               "Filtering by referenced name",
			order:              "",
			nameFilter:         false,
			refNameFilter:      true,
			categoriesFiltered: 0,
			dbString:           "date_added DESC",
			expectedProducts:   getExpectedProducts(),
		},
		{
			name:               "Filtering by name and referenced name",
			order:              "",
			nameFilter:         true,
			refNameFilter:      true,
			categoriesFiltered: 0,
			dbString:           "date_added DESC",
			expectedProducts:   getExpectedProducts(),
		},
		{
			name:               "Filtering by a single category",
			order:              "",
			nameFilter:         false,
			refNameFilter:      false,
			categoriesFiltered: 1,
			dbString:           "date_added DESC",
			expectedProducts:   getExpectedProducts(),
		},
		{
			name:               "Filtering by multiple categories",
			order:              "",
			nameFilter:         false,
			refNameFilter:      false,
			categoriesFiltered: 2,
			dbString:           "date_added DESC",
			expectedProducts:   getExpectedProducts(),
		},
		{
			name:               "Filtering by multiple categories, name, and filter",
			order:              "",
			nameFilter:         true,
			refNameFilter:      true,
			categoriesFiltered: 2,
			dbString:           "date_added DESC",
			expectedProducts:   getExpectedProducts(),
		},
		{
			name:               "Price Descending",
			order:              "price_desc",
			nameFilter:         false,
			refNameFilter:      false,
			categoriesFiltered: 0,
			dbString:           "price DESC",
			expectedProducts:   getExpectedProducts(),
		},
		{
			name:               "Price Ascending",
			order:              "price_asc",
			nameFilter:         false,
			refNameFilter:      false,
			categoriesFiltered: 0,
			dbString:           "price ASC",
			expectedProducts:   getExpectedProducts(),
		},
		{
			name:               "Date Descending",
			order:              "date_desc",
			nameFilter:         false,
			refNameFilter:      false,
			categoriesFiltered: 0,
			dbString:           "date_added DESC",
			expectedProducts:   getExpectedProducts(),
		},
		{
			name:               "Date Ascending",
			order:              "date_asc",
			nameFilter:         false,
			refNameFilter:      false,
			categoriesFiltered: 0,
			dbString:           "date_added ASC",
			expectedProducts:   getExpectedProducts(),
		},
		{
			name:               "Filtering using all options",
			order:              "date_asc",
			nameFilter:         true,
			refNameFilter:      true,
			categoriesFiltered: 2,
			dbString:           "date_added ASC",
			expectedProducts:   getExpectedProducts(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := getMockDB(t)
			defer db.Close()

			mock.ExpectQuery(expectedQuery(tt.dbString, tt.nameFilter, tt.refNameFilter, tt.categoriesFiltered)).WillReturnRows(getMockRows(tt.expectedProducts))
			rr := makeRequest(t, db, getProductsURL(tt.order, tt.nameFilter, tt.refNameFilter, tt.categoriesFiltered))

			checkResponseCode(t, rr.Code, http.StatusOK)
			checkResponseBody(t, rr.Body.String(), "", tt.expectedProducts)
			checkMockExpectations(t, mock)
		})
	}
}

func TestGetProducts_DBError(t *testing.T) {
	db, mock := getMockDB(t)
	defer db.Close()

	expectedErr := errors.New("some error")
	mock.ExpectQuery("SELECT id, name, price, description, categories, images, referenced_name, date_added FROM products").
		WillReturnError(expectedErr)

	rr := makeRequest(t, db, getProductsURL("", false, false, 0))

	checkResponseCode(t, rr.Code, http.StatusInternalServerError)
	checkResponseBody(t, rr.Body.String(), "Internal server error\n", nil)
	checkMockExpectations(t, mock)
}

func TestGetProducts_ScanError(t *testing.T) {
	db, mock := getMockDB(t)
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "name", "price", "description", "categories", "images", "referenced_name", "date_added"}).
		AddRow(1, "Test Product", 9.99, "Test Description", nil, nil, nil, time.Now()).
		AddRow(2, "Invalid Product", "invalid price", "Invalid Description", nil, nil, nil, time.Now())
	mock.ExpectQuery("SELECT id, name, price, description, categories, images, referenced_name, date_added FROM products").WillReturnRows(rows)

	rr := makeRequest(t, db, getProductsURL("", false, false, 0))

	checkResponseCode(t, rr.Code, http.StatusInternalServerError)
	checkResponseBody(t, rr.Body.String(), "Internal server error\n", nil)
	checkMockExpectations(t, mock)
}
