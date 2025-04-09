package controller

import (
	neoModels "api/internal/app/models/neo"
	neo "api/internal/app/neo4j"
	"api/internal/app/rest"
	"api/internal/app/routing"
	"encoding/json"
	"net/http"
	"strconv"
)

func CreateWorld(w http.ResponseWriter, r *http.Request, rctx routing.Context) {
	var world neoModels.World

	userID := rctx.GetPathParam("id")
	if userID == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}

	err := json.NewDecoder(r.Body).Decode(&world)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userIDInt, err := strconv.ParseInt(userID, 10, 64)

	if err != nil {
		http.Error(w, "invalid userID", http.StatusBadRequest)
		return
	}

	err = world.Create(&world, neo.CreateOptions{
		Rel:          "OWNS",
		RelDirection: "<-",
		Label:        "User",
		Field:        "userID",
		Value:        userIDInt,
	})

	if err != nil {
		rest.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	rest.RespondWithSuccess(w, http.StatusCreated, "World created successfully", world)

}

func GetWorld(w http.ResponseWriter, r *http.Request, rctx routing.Context) {
	id := rctx.GetPathParam("id")
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}

	var world neoModels.World
	err := world.Find(&world, "elementID", id).Populate(neo.PopulateOptions{
		Depth: 0,
	})

	if err != nil {
		rest.RespondWithError(w, http.StatusInternalServerError, err.Error())
	}

	rest.RespondWithSuccess(w, http.StatusOK, "World retrieved successfully", world)
}

func PutWorld(w http.ResponseWriter, r *http.Request, rctx routing.Context) {
	var world neoModels.World
	worldID := rctx.GetPathParam("id")

	if worldID == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}
	err := json.NewDecoder(r.Body).Decode(&world)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	world.ID = worldID

	err = world.Update(&world, neo.CreateOptions{})

	if err != nil {
		rest.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	rest.RespondWithSuccess(w, http.StatusOK, "World updated successfully", world)
}

func DeleteWorld(w http.ResponseWriter, r *http.Request, rctx routing.Context) {
	id := rctx.GetPathParam("id")
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}

	var world neoModels.World
	err := world.Delete(&world, "elementID", id, neo.DeleteOptions{
		Detach: true,
	})

	if err != nil {
		rest.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	rest.RespondWithSuccess(w, http.StatusOK, "World deleted successfully", world)
}
