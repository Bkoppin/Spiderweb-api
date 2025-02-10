package controller

import (
	"api/internal/app/models"
	"api/internal/app/postgres"
	"api/internal/app/routing"
	"encoding/json"
	"net/http"
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

	res := db.Create(&user)
	if res.Error != nil {
		http.Error(w, res.Error.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)

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
			"user": dbUser.Username,
			"id":   dbUser.ID,
		}
		json.NewEncoder(w).Encode(data)
}

