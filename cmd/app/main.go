package main

import (
	"api/internal/app/controller"
	"api/internal/app/middleware"
	"api/internal/app/routing"
)

func main() {
	router := routing.NewRouter()
	router.Use(middleware.Cors)
	router.Use(middleware.ContentTypeJSON)
	router.Handle("POST", "/api/users", controller.CreateUser)
	router.Handle("GET", "/api/users/:id", controller.GetUser)
	router.Handle("GET", "/api/users/:id/worlds", controller.GetUserWorlds)
	router.Handle("POST", "/api/users/:id/worlds", controller.CreateWorld)
	router.Handle("GET", "/api/worlds/:id", controller.GetWorld)
	router.Handle("GET", "/api/users/zones", controller.Test)
	router.Serve("8080", routing.ServeOptions{Message: "http://localhost:8080",})

}


