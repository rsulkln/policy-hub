package auth

import (
	"encoding/json"
	"net/http"
	"project/database/redis"
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

	if creds.UserName != "admin" || creds.Password != "test" {
		http.Error(w, "Unauthorized!", http.StatusUnauthorized)

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
