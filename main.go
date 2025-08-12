package main

import (
	"context"
	"fmt"
	jwt "github.com/golang-jwt/jwt/v5"
	"log"
	"net/http"
	"project/auth"
	"project/database/mongo"
	"project/middleware"
	"project/model"
	"project/repository"
	"project/security"
	"time"
)

func main() {
	err := mongo.InitMongo()
	if err != nil {
		panic(err)
	}
	repository.NewMongoUserRepository(mongo.Client, "CRUDapplication", "users")

	newUser := &model.User{
		ID:       "rasol",
		Name:     "kln",
		Password: "123456",
		Role:     "admin",
	}

	if hashPAss, hashErr := security.HashPassword(newUser.Password); hashErr != nil {
		panic(hashErr)
	} else {
		newUser.Password = hashPAss
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	repo := repository.NewMongoUserRepository(mongo.Client, "CRUDapplication", "users")
	nErr := repo.CreateUser(ctx, newUser)
	if nErr != nil {
		log.Fatal("you have an error to create a new user!", nErr)
	}

	http.HandleFunc("/login", auth.LoginHandler)

	token, jErr := auth.GenerateAccessToken(newUser.ID, newUser.Role)
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

	// مسیر تولید توکن (login)
	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		// اینجا میشه واقعی به دیتابیس وصل شد ولی فعلاً هاردکد می‌کنیم
		userID := "12345"
		role := "admin"

		token, err := auth.GenerateAccessToken(userID, role)
		if err != nil {
			http.Error(w, "Error generating token", http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "Your token: %s", token)
	})

	// مسیر محافظت‌شده
	http.Handle("/protected", middleware.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("✅ You are authorized!"))
	})))

	fmt.Println("Server is running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// test baraye kar kardan sahih JWT :

//http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {

//	userID := "12345"
//	role := "admin"
//
//	token, err := auth.GenerateAccessToken(userID, role)
//	if err != nil {
//		http.Error(w, "Error generating token", http.StatusInternalServerError)
//		return
//	}
//
//	fmt.Fprintf(w, "Your token: %s", token)
//})
//

//http.Handle("/protected", middleware.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//	w.Write([]byte("✅ You are authorized!"))
//})))
//
//fmt.Println("Server is running on :8080")
//log.Fatal(http.ListenAndServe(":8080", nil))
