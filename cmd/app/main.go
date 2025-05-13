package main

import (
	"api/internal/app/controller"
	"api/internal/app/middleware"
	neoModels "api/internal/app/models/neo"
	neo "api/internal/app/neo4j"
	"api/internal/app/routing"
)

func main() {

	neo.RegisterModel("User", &neoModels.User{})
	neo.RegisterModel("World", &neoModels.World{})
	neo.RegisterModel("Ocean", &neoModels.Ocean{})
	neo.RegisterModel("Continent", &neoModels.Continent{})
	neo.RegisterModel("Zone", &neoModels.Zone{})
	neo.RegisterModel("Location", &neoModels.Location{})
	neo.RegisterModel("City", &neoModels.City{})

	router := routing.NewRouter()
	router.Use(middleware.Cors)
	router.Use(middleware.ContentTypeJSON)
	router.Handle("POST", "/api/auth/login", controller.Login)
	router.Handle("POST", "/api/user", controller.CreateUser)
	router.Handle("GET", "/api/user/:id", controller.GetUser)
	router.Handle("GET", "/api/user/:id/worlds", controller.GetUserWorlds)
	router.Handle("GET", "/api/user/:id/neo", controller.GetNeoUser)
	router.Handle("POST", "/api/user/:id/world", controller.CreateWorld)
	router.Handle("GET", "/api/world/:id", controller.GetWorld)
	router.Handle("PUT", "/api/world/:id", controller.PutWorld)
	router.Handle("DELETE", "/api/world/:id", controller.DeleteWorld)
	router.Serve("8080", routing.ServeOptions{Message: "http://localhost:8080", Logging: true})

}
