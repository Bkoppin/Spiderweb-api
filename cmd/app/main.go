package main

import (
	controller "api/internal/app/http"
	"api/internal/app/router"
)

func main() {
	// Create a new router
	router := router.NewRouter()
	// Add routes to the router
	router.Handle("/api/v1/auth/login", controller.Test())

	// Start the server on port 8080
	router.Serve(8080)

}

