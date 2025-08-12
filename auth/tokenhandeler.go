package auth

import (
	"encoding/json"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"project/database/mongo"
	redisd "project/database/redis"
	"project/repository"
	"strings"
	"time"
)

type Credentials struct {
	UserName string `json:"username"`
	Password string `json:"password"`
}

// LoginHandler godoc
// @Summary Login and get JWT tokens
// @Description Authenticate user and return access & refresh tokens
// @Tags Auth
// @Accept json
// @Produce json
// @Param credentials body Credentials true "User credentials"
// @Success 200 {object} map[string]string
// @Failure 400 {string} string "Invalid request body"
// @Failure 401 {string} string "Unauthorized"

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var creds Credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Invalid request body!", http.StatusBadRequest)

		return
	}

	repo := repository.NewMongoUserRepository(mongo.Client, "policy-hub", "users")

	mUser, gErr := repo.GetUserByUsername(r.Context(), creds.UserName)
	if gErr != nil {
		http.Error(w, "unauthorized!", http.StatusUnauthorized)

		return
	}

	if bcrypt.CompareHashAndPassword([]byte(mUser.Password), []byte(creds.Password)) != nil {
		http.Error(w, "unauthorized!", http.StatusBadRequest)

		return
	}

	token, err := GenerateAccessToken(mUser.ID, mUser.Role)
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)

		return
	}
	refreshToken, rErr := GenerateRefreshToken(mUser.ID, mUser.Role)
	if rErr != nil {
		http.Error(w, "Error refreshing token", http.StatusInternalServerError)

		return
	}

	err = redisd.RDB.Set(
		redisd.Ctx,
		fmt.Sprintf("refresh_%s", mUser.ID),
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

// RefreshTokenHandler godoc
// @Summary      Refresh access token
// @Description  Create a new access token using a valid refresh token.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        refresh_token  body  map[string]string  true  "Refresh token"
// @Success      200  {object}  map[string]string "new access_token"
// @Failure      400  {string}  string  "Invalid request body"
// @Failure      401  {string}  string  "Invalid or expired refresh token"
// @Failure      500  {string}  string  "Internal server error"
// @Router       /refresh-token [post]
func RefreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// 1) Validate JWT
	token, err := ValidationToken(req.RefreshToken)
	if err != nil || !token.Valid {
		http.Error(w, "invalid or expired refresh token", http.StatusUnauthorized)
		return
	}

	claims, ok := token.Claims.(*MyCustomClaims)
	if !ok {
		http.Error(w, "invalid token claims", http.StatusUnauthorized)
		return
	}

	redisKey := fmt.Sprintf("refresh_%s", claims.UserID)
	storedToken, rErr := redisd.RDB.Get(redisd.Ctx, redisKey).Result()
	if rErr != nil || storedToken != req.RefreshToken {
		http.Error(w, "refresh token not found or revoked", http.StatusUnauthorized)
		return
	}

	newAccessToken, err := GenerateAccessToken(claims.UserID, claims.Role)
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"access_token": newAccessToken,
	})
}

// LogoutTokenHandler godoc
// @Summary      Logout and invalidate tokens
// @Description  Deletes the refresh token from Redis and blacklists the current access token until it expires.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        refresh_token  body  map[string]string  true  "Refresh token"
// @Success      200  {string}  string  "Successfully logged out"
// @Failure      400  {string}  string  "Invalid request body"
// @Failure      401  {string}  string  "Invalid refresh token"
// @Failure      500  {string}  string  "Internal server error"
// @Router       /logout [post]
func LogoutTokenHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	token, err := ValidationToken(req.RefreshToken)
	if err != nil || !token.Valid {
		http.Error(w, "invalid refresh token", http.StatusUnauthorized)
		return
	}

	claims, ok := token.Claims.(*MyCustomClaims)
	if !ok {
		http.Error(w, "invalid token claims", http.StatusUnauthorized)
		return
	}

	redisKey := fmt.Sprintf("refresh_%s", claims.UserID)
	storedToken, rErr := redisd.RDB.Get(redisd.Ctx, redisKey).Result()
	if rErr != nil || storedToken != req.RefreshToken {
		http.Error(w, "invalid refresh token", http.StatusUnauthorized)
		return
	}

	if dErr := redisd.RDB.Del(redisd.Ctx, redisKey).Err(); dErr != nil {
		http.Error(w, "Error deleting refresh token", http.StatusInternalServerError)
		return
	}

	accessToken := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	if accessToken != "" {
		token, err := ValidationToken(accessToken)
		if err == nil && token.Valid {
			if claims, ok := token.Claims.(*MyCustomClaims); ok {
				if claims.ExpiresAt != nil {
					ttl := time.Until(claims.ExpiresAt.Time)
					if ttl > 0 {
						redisd.RDB.Set(redisd.Ctx, "blacklist_"+accessToken, "true", ttl)
					}
				}
			}
		}
	}
}
