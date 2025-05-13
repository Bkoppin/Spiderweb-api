package controller

import (
	"api/internal/app/models"
	neoModels "api/internal/app/models/neo"
	neo "api/internal/app/neo4j"
	"api/internal/app/postgres"
	"api/internal/app/routing"
	"encoding/json"
	"net/http"
	"strconv"
)

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

	res := db.Create(&user).Omit("password")
	if res.Error != nil {
		http.Error(w, res.Error.Error(), http.StatusInternalServerError)
		return
	}

	neoUser := neoModels.User{
		Username: user.Username,
		UserID:   int64(user.ID),
	}

	err = neoUser.Create(&neoUser, neo.CreateOptions{})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(neoUser)
	
	
}

func GetUser(w http.ResponseWriter, r *http.Request, context routing.Context) {
	db, err := postgres.Connect()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	id := context.GetPathParam("id")
	if id == "" {
		http.Error(w, "Missing user ID", http.StatusBadRequest)
		return
	}

	if _, err := strconv.ParseInt(id, 10, 64); err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var user models.User
	res := db.First(&user, id).Omit("password")

	if res.Error != nil {
		http.Error(w, res.Error.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)

}

func GetUserWorlds(w http.ResponseWriter, r *http.Request, context routing.Context) {
	id := context.GetPathParam("id")
	if id == "" {
		http.Error(w, "Missing user ID", http.StatusBadRequest)
		return
	}

	parsedID, err := strconv.ParseInt(id, 10, 64)

	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var user neoModels.User
	err = user.Find(&user, "userID", parsedID).Populate(neo.PopulateOptions{
		Depth: 1,
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(user.Worlds) == 0 {
		http.Error(w, "No worlds found for this user", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user.Worlds)
}

func Login(w http.ResponseWriter, r *http.Request, context routing.Context) {
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

	var dbUser models.User

	res := db.Where("username = ?", user.Username).First(&dbUser).Omit("password")
	if res.Error != nil {
		http.Error(w, "Invalid Credentials", http.StatusNotFound)
		return
	}

	if !dbUser.ComparePassword(user.Password) {
		http.Error(w, "Invalid Credentials", http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(dbUser)
}

func GetNeoUser(w http.ResponseWriter, r *http.Request, context routing.Context) {
	if context.GetPathParam("id") == "" {
		http.Error(w, "Missing user ID", http.StatusBadRequest)
		return
	}
	idParam := context.GetPathParam("id")

	id, err := strconv.ParseInt(idParam, 10, 64)

	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var user neoModels.User
	err = user.Find(&user, "userID", id).Populate(neo.PopulateOptions{
		Depth: 1,
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}
