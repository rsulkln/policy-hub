package main

import (
	"fmt"
	jwt "github.com/golang-jwt/jwt/v5"
	"net/http"
	"project/auth"
)

func main() {
	http.HandleFunc("/login", auth.LoginHandler)

	token, jErr := auth.GenerateAccessToken("23", "admin")
	if jErr != nil {
		fmt.Println("you have an error to generate token !")

		return
	}
	fmt.Printf("token generated! %+v ", token)

	parsedToken, vErr := auth.ValidationToken(token)
	if vErr != nil {
		fmt.Printf("token is not valid :%+v", vErr)
	}
	if claims, ok := parsedToken.Claims.(jwt.MapClaims); ok && parsedToken.Valid {
		fmt.Println("Token is valid")
		fmt.Println("User ID:", claims["sub"])
		fmt.Println("Role:", claims["role"])
	} else {
		fmt.Println("Token is invalid")
	}
}
