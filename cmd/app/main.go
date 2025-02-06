package main

import (
	"api/internal/app/controller"
	"api/internal/app/routing"
)

func main() {
	// Create a new router
	router := routing.New()
	// Add routes to the router

	router.Handle("/api/v1/auth/login", controller.Test(), "GET")

	// Start the server on port 8080

	router.Serve(8080, routing.ServeOptions{Message: "http://localhost:8080",})

}


