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

func fetchUserWorlds(userID string) (neo.Node, error) {
	ctx := context.Background()
	driver, err := neo.NewDriver()
	if err != nil {
		return nil, err
	}
	defer driver.Close(ctx)

	session := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})

	defer session.Close(ctx)

	parsedUserID, err := strconv.Atoi(userID)
	if err != nil {
		return nil, err
	}

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		
		queryBuilder := neo.NewQueryBuilder()
		query, params := queryBuilder.Match("(u:User {userID: $userID})").
			WithParam("userID", parsedUserID).
			OptionalMatch("(u)-[:OWNS]->(w:World)").
			OptionalMatch("(w)-[:HAS]->(c:Continent)").
			OptionalMatch("(w)-[:HAS]->(o:Ocean)").
			OptionalMatch("(w)-[:HAS]->(z:Zone)").
			OptionalMatch("(z)-[:HAS]->(l:Location)").
			OptionalMatch("(z)-[:HAS]->(ci:City)").
			Return("u, w, c, o, z, l, ci").
			Build()

		records, err := tx.Run(ctx, query, params)
		if err != nil {
						return nil, err
		}
		recordList := []neo4j.Record{}
		for records.Next(ctx) {
						recordList = append(recordList, *records.Record())
		}
		return recordList, nil
	})

	if err != nil {
		return nil, err
	}
	records, ok := result.([]neo4j.Record)
	if !ok {
		return nil, fmt.Errorf("failed to convert result to []neo4j.Record")
	}
	return neo.BuildNodeTree(records), nil


	
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
		
		data := worlds.BuildTree()

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(data)
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

