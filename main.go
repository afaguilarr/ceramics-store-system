package main

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

type Product struct {
	ID             int       `json:"id"`
	Name           string    `json:"name"`
	Price          float64   `json:"price"`
	Description    string    `json:"description"`
	Categories     []string  `json:"categories"`
	Images         []string  `json:"images"`
	ReferencedName string    `json:"referenced_name"`
	DateAdded      time.Time `json:"date_added"`
}

type ShoppingCartItem struct {
	ID               int `json:"id,omitempty"`
	ShoppingCartID   int `json:"shopping_cart_id,omitempty"`
	ProductID        int `json:"product_id,omitempty"`
	NumberOfProducts int `json:"number_of_products,omitempty"`
}

type ShoppingCart struct {
	ID                int                `json:"id,omitempty"`
	UserID            int                `json:"user_id,omitempty"`
	IPAddress         string             `json:"ip_address,omitempty"`
	ShoppingCartItems []ShoppingCartItem `json:"shopping_cart_items,omitempty"`
}

type ProductsHandler struct {
	db *sql.DB
}

type ShoppingCartsHandler struct {
	db          *sql.DB
	redisClient *redis.Client
}

func main() {
	// Open database connection
	db, err := sql.Open("postgres", "postgres://user:password@postgres_db/products_db?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Set connection pool configuration
	db.SetMaxIdleConns(5)
	db.SetMaxOpenConns(10)
	db.SetConnMaxIdleTime(time.Minute)
	db.SetConnMaxLifetime(time.Minute * 3)

	redisClient := redis.NewClient(&redis.Options{
		Addr: "redis_db:6379",
	})

	// Initialize router
	r := mux.NewRouter()

	ph := ProductsHandler{db: db}
	sch := ShoppingCartsHandler{db: db, redisClient: redisClient}

	// Define endpoint for getting all products
	r.HandleFunc("/products", ph.getProducts).Methods(http.MethodGet)
	// Define endpoint for getting a single product by ID
	r.HandleFunc("/products/{id}", ph.getProduct).Methods(http.MethodGet)
	// Define endpoint for upserting a shopping cart in redis
	r.HandleFunc("/shopping_carts", sch.upsertShoppingCartHandler).Methods(http.MethodPost)
	// Define endpoint for getting a shopping cart from redis
	r.HandleFunc("/shopping_carts", sch.getShoppingCartHandler).Methods(http.MethodGet)

	// Start server
	log.Println("Server started on :8080")
	err = http.ListenAndServe(":8080", r)
	if err != nil {
		log.Fatal(err)
	}
}
