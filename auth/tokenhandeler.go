package auth

import (
	"encoding/json"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"project/database/mongo"
	"project/database/redis"
	"project/repository"
	"strings"
	"time"
)

type Credentials struct {
	UserName string `json:"username"`
	Password string `json:"password"`
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var creds Credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Invalid request body!", http.StatusBadRequest)

		return
	}

	repo := repository.NewMongoUserRepository(mongo.Client, "policy-hub", "users")

	mUser, gErr := repo.GetUserByID(r.Context(), creds.UserName)
	if gErr != nil {
		http.Error(w, "unauthorized!", http.StatusBadRequest)

		return
	}

	if bcrypt.CompareHashAndPassword([]byte(mUser.Password), []byte(creds.Password)) != nil {
		http.Error(w, "unauthorized!", http.StatusBadRequest)

		return
	}

	token, err := GenerateAccessToken("userID-1", "admin")
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)

		return
	}
	refreshToken, rErr := GenerateRefreshToken()
	if rErr != nil {
		http.Error(w, "Error refreshing token", http.StatusInternalServerError)

		return
	}

	err = redisd.RDB.Set(
		redisd.Ctx,
		"refresh_userID-1",
		refreshToken,
		7*24*time.Hour,
	).Err()
	if err != nil {
		http.Error(w, "Error setting refresh token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"access_token":  token,
		"refresh_token": refreshToken,
	})
}

func RefreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)

		return
	}
	storedToken, rErr := redisd.RDB.Get(redisd.Ctx, "refresh_userID-1").Result()
	if rErr != nil || storedToken != req.RefreshToken {
		http.Error(w, "invalid refresh token", http.StatusUnauthorized)

		return
	}
	newAccessToken, err := GenerateAccessToken("userID-1", "admin")
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)

		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"access_token": newAccessToken,
	})
}

func LogoutTokenHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)

		return
	}
	storedToken, rErr := redisd.RDB.Get(redisd.Ctx, "refresh_userID-1").Result()

	if rErr != nil || storedToken != req.RefreshToken {
		http.Error(w, "invalid refresh token", http.StatusUnauthorized)

		return
	}
	dErr := redisd.RDB.Del(redisd.Ctx, "refresh_userID-1")
	if dErr != nil {
		http.Error(w, "Error deleting refresh token", http.StatusInternalServerError)

		return
	}

	accsecToken := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	if accsecToken != "" {
		ttl := time.Hour
		redisd.RDB.Set(redisd.Ctx, "blacklist_"+accsecToken, "true", ttl)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"massage": "logout successfully",
	})
}
