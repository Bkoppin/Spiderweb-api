package controller

import (
	neoModels "api/internal/app/models/neo"
	neo "api/internal/app/neo4j"
	"api/internal/app/routing"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func createContinent(query string, params map[string]interface{}) (*neoModels.Continent, error) {
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

	continentValue, ok := record.Get("c")
	if !ok {
		return nil, fmt.Errorf("failed to retrieve 'c' from record")
	}
	continentNode := continentValue.(neo4j.Node)

	continent := neoModels.Continent{
		ID:          continentNode.ElementId,
		Name:        continentNode.Props["name"].(string),
		Description: continentNode.Props["description"].(string),
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
	}

	return &continent, nil
}

func CreateContinent(w http.ResponseWriter, r *http.Request, rctx routing.Context) {
	var continent *neoModels.Continent
	if err := json.NewDecoder(r.Body).Decode(&continent); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	worldID := rctx.GetPathParam("id")
	if worldID == "" {
		http.Error(w, "missing worldID", http.StatusBadRequest)
		return
	}

	query, params := neo.NewQueryBuilder().
		Match("(w:World) WHERE elementId(w) = $id").
		With("w").
		Create("(c:Continent {name: $name, description: $description, type: $type})").
		Create("(w)-[:HAS]->(c)").
		WithParam("name", continent.Name).
		WithParam("description", continent.Description).
		WithParam("type", continent.Type).
		WithParam("id", worldID).
		Return("c").
		Build()

	continent, err := createContinent(query, params)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(continent)

}
