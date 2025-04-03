package controller

import (
	"api/internal/app/models"
	neoModels "api/internal/app/models/neo"
	neo "api/internal/app/neo4j"
	"api/internal/app/postgres"
	"api/internal/app/rest"
	"api/internal/app/routing"
	"encoding/json"
	"net/http"
	"strconv"
)

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

	neoUser := neoModels.User{
		Username: user.Username,
		UserID:   int64(user.ID),
	}

	err = neoUser.Create(&neoUser, neo.CreateOptions{})

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
	res := db.First(&user, id).Omit("password")
	if res.Error != nil {
		rest.RespondWithError(w, http.StatusNotFound, "User not found")
		return
	}

	rest.RespondWithSuccess(w, http.StatusOK, "User retrieved successfully", user)
}

func GetUserWorlds(w http.ResponseWriter, r *http.Request, context routing.Context) {
	id := context.GetPathParam("id")
	if id == "" {
		rest.RespondWithError(w, http.StatusBadRequest, "Missing user ID")
		return
	}

	parsedID, err := strconv.ParseInt(id, 10, 64)

	if err != nil {
		rest.RespondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	var user neoModels.User
	err = user.Find(&user, "userID", parsedID).Populate(neo.PopulateOptions{
		Depth: 1,
	})

	if err != nil {
		rest.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if len(user.Worlds) == 0 {
		rest.RespondWithError(w, http.StatusNotFound, "No worlds found for user")
		return
	}
	rest.RespondWithSuccess(w, http.StatusOK, "Worlds retrieved successfully", user.Worlds)
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

func GetNeoUser(w http.ResponseWriter, r *http.Request, context routing.Context) {
	if context.GetPathParam("id") == "" {
		rest.RespondWithError(w, http.StatusBadRequest, "Missing user ID")
		return
	}
	idParam := context.GetPathParam("id")

	id, err := strconv.ParseInt(idParam, 10, 64)

	if err != nil {
		rest.RespondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	var user neoModels.User
	err = user.Find(&user, "userID", id).Populate(neo.PopulateOptions{
		Depth: 1,
	})

	if err != nil {
		rest.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	rest.RespondWithSuccess(w, http.StatusOK, "User retrieved successfully", user)
}
