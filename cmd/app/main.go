package main

import (
	"api/internal/app/auth"
	"fmt"
)

func main() {
	token, err := auth.CreateJWT("admin")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(token)

	fmt.Println(auth.VerifyJWT(token))
	fmt.Println(auth.VerifyJWT("invalid token"))
	fmt.Println(auth.DecodeJWT(token))

}