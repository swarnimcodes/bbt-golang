package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"

	"github.com/swarnimcodes/bitespeed-backend-task/handlers"
	"github.com/swarnimcodes/bitespeed-backend-task/handlers/customer"
	"github.com/swarnimcodes/bitespeed-backend-task/middlewares"
)

// type alias for a func that takes in a handler and returns a handler
type Middleware func(http.Handler) http.Handler

// router struct to hold global and route specific middlewares
type Router struct {
	mux               *http.ServeMux
	globalMiddlewares []Middleware
	routes            map[string]http.Handler
}

// create a new Router instance
func NewRouter() *Router {
	return &Router{
		mux:    http.NewServeMux(),
		routes: make(map[string]http.Handler),
	}
}

// adds global middlewares
func (r *Router) AddGlobalMiddleware(middleware Middleware) {
	r.globalMiddlewares = append(r.globalMiddlewares, middleware)
}

// route specific middlewares
func (r *Router) Handle(pattern string, handler http.Handler, middlewares ...Middleware) {
	wrappedHandler := r.applyMiddlewares(handler, middlewares...)
	r.routes[pattern] = wrappedHandler
	r.mux.Handle(pattern, wrappedHandler)
}

func (r *Router) applyMiddlewares(handler http.Handler, middlewares ...Middleware) http.Handler {
	for _, middleware := range middlewares {
		handler = middleware(handler)
	}
	for _, middleware := range r.globalMiddlewares {
		handler = middleware(handler)
	}
	return handler
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}

func main() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Couldn't load `.env` file: %s\n", err)
	}
	router := NewRouter()

	// global middlewares

	router.AddGlobalMiddleware(middlewares.Auth)

	// handle request with route-specific middlewares
	router.Handle("GET /", http.HandlerFunc(handlers.Hello))

	// Bitespeed
	router.Handle("GET /customers", http.HandlerFunc(customer.GetAllCustomers))
	router.Handle("POST /identify", http.HandlerFunc(customer.IdentifyCustomer))

	port := os.Getenv("PORT")
	if port == "" {
		log.Println("No port specified in `.env`. Using default port `8080`.")
		port = "8080"
	}
	log.Printf("Server started at: %s\n", port)
	addr := fmt.Sprintf(":%s", port)
	log.Fatal(http.ListenAndServe(addr, router))
}
