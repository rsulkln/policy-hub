package main

import (
	"context"
	"fmt"
	jwt "github.com/golang-jwt/jwt/v5"
	"log"
	"net/http"
	"project/auth"
	"project/database"
	"project/model"
	"time"
)

func main() {
	err := database.InitMongo()
	if err != nil {
		fmt.Println("i can't connection to mongoDB!", err)
	}
	database.NewMongoUserRepository(database.Client, "CRUDapplication", "users")

	newUser := &model.User{
		ID:       "user1",
		Name:     "test",
		Password: "test",
		Role:     "user",
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	repo := database.NewMongoUserRepository(database.Client, "CRUDapplication", "users")
	nErr := repo.CreateUser(ctx, newUser)
	if nErr != nil {
		log.Fatal("you have an error to create a new user!", nErr)
	}
	
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
