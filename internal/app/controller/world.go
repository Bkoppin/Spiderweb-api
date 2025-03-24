package controller

import (
	neo "api/internal/app/neo4j"
	"api/internal/app/routing"
	"context"
	"encoding/json"
	"net/http"
)

type World struct {
	Name string `json:"name"`
}

func CreateWorld(
	w http.ResponseWriter,
	r *http.Request,
	rctx routing.Context,
) {

	ctx := context.Background()

	var body struct {
		Name   string `json:"name"`
	}

	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	driver, err := neo.NewDriver()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer driver.Close(ctx)


	node, err := neo.CreateAndReturnNode(ctx, driver, "World", map[string]interface{}{
		"name": body.Name,
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(node)


		
	}