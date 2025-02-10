package middleware

import (
	"net/http"
)

func Cors(w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	return nil
}

func ContentTypeJSON(w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Content-Type", "application/json")
	return nil
}
