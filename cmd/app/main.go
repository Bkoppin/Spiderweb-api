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
	router.Handle("POST", "/api/users", controller.CreateUser)
	router.Handle("GET", "/api/users/:id", controller.GetUser)
	router.Handle("GET", "/api/users/:id/worlds", controller.GetUserWorlds)
	router.Handle("GET", "/api/users/:id/neo", controller.GetNeoUser)
	router.Handle("POST", "/api/users/:id/worlds", controller.CreateWorld)
	router.Handle("GET", "/api/worlds/:id", controller.GetWorld)
	router.Handle("POST", "/api/worlds/:id/continents", controller.CreateContinent)
	router.Serve("8080", routing.ServeOptions{Message: "http://localhost:8080"})

}


