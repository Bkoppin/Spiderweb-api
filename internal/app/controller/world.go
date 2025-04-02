package controller

import (
	neoModels "api/internal/app/models/neo"
	neo "api/internal/app/neo4j"
	"api/internal/app/routing"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func createWorld(query string, params map[string]interface{}) (*neoModels.World, error) {
	ctx := context.Background()
	driver, err := neo.NewDriver()
	if err != nil {
		return nil, err
	}
	defer driver.Close(ctx)

	session := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	tx, err := session.BeginTransaction(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	res, err := tx.Run(ctx, query, params)
	if err != nil {
		return nil, err
	}

	record, err := res.Single(ctx)
	if err != nil {
		return nil, err
	}

	worldValue, ok := record.Get("w")
	if !ok {
		return nil, fmt.Errorf("failed to retrieve 'w' from record")
	}
	worldNode := worldValue.(neo4j.Node)
	
	world := neoModels.World{
		ID:					worldNode.ElementId,
		Name:        worldNode.Props["name"].(string),
		Type:        worldNode.Props["type"].(string),
		Description: worldNode.Props["description"].(string),
	}

	err = tx.Commit(ctx)

	if err != nil {
		return nil, err
	}
	
	return &world, nil

}

func fetchWorld(query string, params map[string]interface{}) (*neoModels.World, error) {
	ctx := context.Background()
	driver, err := neo.NewDriver()
	if err != nil {
		return nil, err
	}
	defer driver.Close(ctx)

	session := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	res, err := session.Run(ctx, query, params)
	if err != nil {
		return nil, err
	}

	record, err := res.Single(ctx)
	if err != nil {
		return nil, err
	}

	worldValue, ok := record.Get("w")
	if !ok {
		return nil, fmt.Errorf("failed to retrieve 'w' from record")
	}
	worldNode := worldValue.(neo4j.Node)
	
	world := neoModels.World{
		ID:					worldNode.ElementId,
		Name:        worldNode.Props["name"].(string),
		Type:        worldNode.Props["type"].(string),
		Description: worldNode.Props["description"].(string),
	}

	return &world, nil
}


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

	queryBuilder := neo.NewQueryBuilder()
	
	userIDInt, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		http.Error(w, "invalid userID", http.StatusBadRequest)
		return
	}

	query, params := queryBuilder.
		Match("(u:User) WHERE u.userID = $userID").
		With("u").
		Create("(w:World {name: $name, type: $type, description: $description, userID: $userID})").
		Create("(u)-[:OWNS]->(w)").
		Return("w").
		WithParam("name", world.Name).
		WithParam("type", world.Type).
		WithParam("description", world.Description).
		WithParam("userID", userIDInt).
		Build()

	node, err := createWorld(query, params)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(node)
}

func GetWorld(w http.ResponseWriter, r *http.Request, rctx routing.Context) {
	id := rctx.GetPathParam("id")
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}

	queryBuilder := neo.NewQueryBuilder()

	query, params := queryBuilder.
		Match("(w) WHERE elementId(w) = $elementId").
		Return("w").
		WithParam("elementId", id).
		Build()

	node, err := fetchWorld(query, params)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(node)
}





