package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)


var developmentSecret = "development"

/* CreateJWT is a function that creates a JWT token
 * It takes a username as a parameter and returns a string and an error
 * The string is the JWT token
 * The error is nil if the token is created successfully, otherwise it contains an error message
 */
func CreateJWT(username string) (string, error) {
	claims := jwt.MapClaims{}
	claims["username"] = username
	claims["exp"] = time.Now().Add(time.Hour * 24).Unix()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(developmentSecret))
	if err != nil {
		return "", fmt.Errorf("error creating JWT token: %w", err)
	}
	return tokenString, nil
}

/* VerifyJWT is a function that verifies a JWT token
 * It takes a tokenString as a parameter and returns a boolean and an error
 * The boolean is true if the token is valid, false otherwise
 * The error is nil if the token is valid, otherwise it contains an error message
 */
func VerifyJWT(tokenString string) (bool, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(developmentSecret), nil
	})
	if err != nil {
		return false, fmt.Errorf("error parsing JWT token: %w", err)
	}
	return token.Valid, nil
}

/* DecodeJWT is a function that decodes a JWT token and returns the claims
 * It takes a tokenString as a parameter and returns a map of claims and an error
 * The map of claims contains the information stored in the token
 * The error is nil if the token is decoded successfully, otherwise it contains an error message
 */
func DecodeJWT(tokenString string) (jwt.MapClaims, error) {
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(developmentSecret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("error parsing JWT token: %w", err)
	}
	if !token.Valid {
		return nil, fmt.Errorf("invalid JWT token")
	}
	return claims, nil
}
