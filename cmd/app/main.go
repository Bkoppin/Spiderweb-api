package main

import (
	"api/internal/app/controller"
	"api/internal/app/middleware"
	"api/internal/app/routing"
)

func main() {
	router := routing.New()
	router.Use(middleware.Cors)
	router.Use(middleware.ContentTypeJSON)
	router.Handle("/api/users", controller.CreateUser, "POST")
	router.Handle("/api/users/{id}", controller.GetUser, "GET")
	router.Handle("/api/auth/login", controller.Login, "POST", middleware.WithAuth)
	router.Serve(8080, routing.ServeOptions{Message: "http://localhost:8080",})

}


