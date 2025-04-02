package controller

import (
	"api/internal/app/models"
	neoModels "api/internal/app/models/neo"
	neo "api/internal/app/neo4j"
	"api/internal/app/postgres"
	"api/internal/app/rest"
	"api/internal/app/routing"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

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

func fetchUserWorlds(userID string) ([]*neoModels.World, error) {
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
			Return("w, c, o, z, l, ci").
			Build()

		records, err := tx.Run(ctx, query, params)
		if err != nil {
			return neoModels.World{}, err
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
	
	worlds, err := neo.BuildNodeTree[neoModels.World](records)
	if err != nil {
		return nil, err
	}

	if worlds == nil {
		return nil, fmt.Errorf("no worlds found for user")
	}
	return worlds, nil
}

func CreateUser(w http.ResponseWriter, r *http.Request, context routing.Context) {
    var user models.User
    db, err := postgres.Connect()
    if err != nil {
        rest.RespondWithError(w, http.StatusInternalServerError, "Failed to connect to database")
        return
    }

    err = json.NewDecoder(r.Body).Decode(&user)
    if err != nil {
        rest.RespondWithError(w, http.StatusBadRequest, "Failed to decode request body")
        return
    }

    res := db.Create(&user)
    if res.Error != nil {
        rest.RespondWithError(w, http.StatusInternalServerError, "Failed to create user")
        return
    }

    err = createNeoUser(user)
    if err != nil {
        rest.RespondWithError(w, http.StatusInternalServerError, "Failed to create user in Neo4j")
        return
    }

    data := map[string]interface{}{
        "username": user.Username,
        "id":       user.ID,
    }
    rest.RespondWithSuccess(w, http.StatusCreated, "User created successfully", data)
}

func GetUser(w http.ResponseWriter, r *http.Request, context routing.Context) {
    db, err := postgres.Connect()
    if err != nil {
        rest.RespondWithError(w, http.StatusInternalServerError, "Failed to connect to database")
        return
    }

    id := context.GetPathParam("id")
    if id == "" {
        rest.RespondWithError(w, http.StatusBadRequest, "Missing user ID")
        return
    }

    if _, err := strconv.ParseInt(id, 10, 64); err != nil {
        rest.RespondWithError(w, http.StatusBadRequest, "Invalid user ID")
        return
    }

    var user models.User
    res := db.First(&user, id)
    if res.Error != nil {
        rest.RespondWithError(w, http.StatusNotFound, "User not found")
        return
    }

    data := map[string]interface{}{
        "username": user.Username,
        "id":       user.ID,
    }
    rest.RespondWithSuccess(w, http.StatusOK, "User retrieved successfully", data)
}

func GetUserWorlds(w http.ResponseWriter, r *http.Request, context routing.Context) {
    id := context.GetPathParam("id")
    if id == "" {
        rest.RespondWithError(w, http.StatusBadRequest, "Missing user ID")
        return
    }

    worlds, err := fetchUserWorlds(id)
    if err != nil {
        if strings.Contains(err.Error(), "no worlds found for user") {
            rest.RespondWithError(w, http.StatusNotFound, "No worlds found for user")
            return
        }
        rest.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch user worlds")
        return
    }

    if len(worlds) == 1 {
        rest.RespondWithSuccess(w, http.StatusOK, "World retrieved successfully", worlds[0])
        return
    }

    if len(worlds) == 0 {
        rest.RespondWithError(w, http.StatusNotFound, "No worlds found")
        return
    }

    rest.RespondWithSuccess(w, http.StatusOK, "Worlds retrieved successfully", worlds)
}

func Login(w http.ResponseWriter, r *http.Request, context routing.Context) {
    var user models.User
    db, err := postgres.Connect()
    if err != nil {
        rest.RespondWithError(w, http.StatusInternalServerError, "Failed to connect to database")
        return
    }

    err = json.NewDecoder(r.Body).Decode(&user)
    if err != nil {
        rest.RespondWithError(w, http.StatusBadRequest, "Failed to decode request body")
        return
    }

    var dbUser models.User
		
    res := db.Where("username = ?", user.Username).First(&dbUser)
    if res.Error != nil {
        rest.RespondWithError(w, http.StatusNotFound, "User not found")
        return
    }

    if !dbUser.ComparePassword(user.Password) {
        rest.RespondWithError(w, http.StatusUnauthorized, "Invalid password")
        return
    }

    data := map[string]interface{}{
        "username": dbUser.Username,
        "id":       dbUser.ID,
    }
    rest.RespondWithSuccess(w, http.StatusOK, "Login successful", data)
}

func Test(w http.ResponseWriter, r *http.Request, context routing.Context) {
    rest.RespondWithSuccess(w, http.StatusOK, "Hello, World!", nil)
}

