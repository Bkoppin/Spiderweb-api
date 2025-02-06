package controller

import (
	"api/internal/app/auth"
	"api/internal/app/models"
	"encoding/json"
	"net/http"
	"strconv"

	"gorm.io/gorm"
)

func Test() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		switch r.Method {
		case http.MethodGet:
			json.NewEncoder(w).Encode("Hello World")
		case http.MethodPost:
			json.NewEncoder(w).Encode("Hello World Created")
		case http.MethodPut:
			json.NewEncoder(w).Encode("Hello World Updated")
		case http.MethodDelete:
			json.NewEncoder(w).Encode("Hello World Deleted")
		}
	}
}

func CreateUser(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var user models.User
		err := json.NewDecoder(r.Body).Decode(&user)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		res := db.Create(&user)

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(res)
	}
}

func GetUser(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.URL.Query().Get("id"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var user models.User
		res := db.First(&user, id)
		if res.Error != nil {
			http.Error(w, res.Error.Error(), http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(user)
	}
}

func UpdateUser(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.URL.Query().Get("id"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var user models.User
		err = json.NewDecoder(r.Body).Decode(&user)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		user.ID = id
		res := db.Save(&user)
		if res.Error != nil {
			http.Error(w, res.Error.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(user)
	}
}

func DeleteUser(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.URL.Query().Get("id"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var user models.User
		res := db.Delete(&user, id)
		if res.Error != nil {
			http.Error(w, res.Error.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func LoginUser(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var user models.User
		err := json.NewDecoder(r.Body).Decode(&user)
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

		token, err := auth.CreateJWT(user.Username)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Authorization", token)
		json.NewEncoder(w).Encode(dbUser)
	}
}

func LogoutUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Del("Authorization")
		w.WriteHeader(http.StatusOK)
	}
}

func GetMe() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token == "" {
			http.Error(w, "missing token", http.StatusUnauthorized)
			return
		}

		valid, err := auth.VerifyJWT(token)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if !valid {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		claims, err := auth.DecodeJWT(token)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(claims)
	}
}

