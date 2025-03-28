package controller

import (
	"api/internal/app/models"
	neo "api/internal/app/neo4j"
	"api/internal/app/postgres"
	"api/internal/app/routing"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func createNeoUser(user models.User) (error) {
	ctx := context.Background()
	driver, err := neo.NewDriver()
	if err != nil {
		return err
	}
	defer driver.Close(ctx)

	session := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	tx, err := session.BeginTransaction(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	queryBuilder := neo.NewQueryBuilder()
	query, params := queryBuilder.Create("(u:User {username: $username, userID: $userID})").
		WithParam("username", user.Username).
		WithParam("userID", user.ID).
		Return("u").
		Build()

	_, err = tx.Run(ctx, query, params)
	if err != nil {
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}

func fetchUserWorlds(userID string) ([]models.World, error) {
	ctx := context.Background()
	driver, err := neo.NewDriver()
	if err != nil {
		return nil, err
	}
	defer driver.Close(ctx)

	session := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	tx, err := session.BeginTransaction(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	queryBuilder := neo.NewQueryBuilder()


	parsedUserID, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return nil, err
	}

	query, params := queryBuilder.Match("(u:User {userID: $userID})").
		WithParam("userID", parsedUserID).
		Match("(u)-[:OWNS]->(w:World)").
		Return("w").
		Build()

	res, err := tx.Run(ctx, query, params)
	if err != nil {
		return nil, err
	}

	var worlds []models.World
	for res.Next(ctx) {
		record := res.Record()
		worldValue, ok := record.Get("w")
		if !ok {
			return nil, fmt.Errorf("failed to retrieve 'w' from record")
		}
		worldNode := worldValue.(neo4j.Node)

		world := models.World{
			ID:          worldNode.ElementId,
			Name:        worldNode.Props["name"].(string),
			Type:        worldNode.Props["type"].(string),
			Description: worldNode.Props["description"].(string),
		}
		worlds = append(worlds, world)
	}

	return worlds, nil
}

func CreateUser(w http.ResponseWriter, r *http.Request, context routing.Context) {
	var user models.User
	db, err := postgres.Connect()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	res := db.Create(&user)
	if res.Error != nil {
		http.Error(w, res.Error.Error(), http.StatusInternalServerError)
		return
	}

	err = createNeoUser(user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"username": user.Username,
		"id":   user.ID,
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(data)

}

func GetUser(w http.ResponseWriter, r *http.Request, context routing.Context) {
		db, err := postgres.Connect()

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		id := context.GetPathParam("id")

		if id == "" {
			http.Error(w, "missing id", http.StatusBadRequest)
			return
		}

		if _, err := strconv.ParseInt(id, 10, 64); err != nil {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}

		var user models.User
		res := db.First(&user, id)
		if res.Error != nil {
			http.Error(w, res.Error.Error(), http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		data := map[string]interface{}{
			"username": user.Username,
			"id":   user.ID,
		}
		json.NewEncoder(w).Encode(data)
	
}

func GetUserWorlds(w http.ResponseWriter, r *http.Request, context routing.Context) {
		id := context.GetPathParam("id")
		if id == "" {
			http.Error(w, "missing id", http.StatusBadRequest)
			return
		}

		worlds, err := fetchUserWorlds(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(worlds)
}


func Login(w http.ResponseWriter, r *http.Request, context routing.Context) {
		var user models.User
		db , err := postgres.Connect()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		err = json.NewDecoder(r.Body).Decode(&user)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var dbUser models.User
		res := db.Where("username = ?", user.Username).First(&dbUser)
		if res.Error != nil {
			http.Error(w, res.Error.Error(), http.StatusNotFound)
			return
		}

		if !dbUser.ComparePassword(user.Password) {
			http.Error(w, "invalid password", http.StatusUnauthorized)
			return
		}
		

		w.WriteHeader(http.StatusOK)
		data := map[string]interface{}{
			"username": dbUser.Username,
			"id":   dbUser.ID,
		}
		json.NewEncoder(w).Encode(data)
}

func Test(w http.ResponseWriter, r *http.Request, context routing.Context) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Hello, World!")
}

