package auth

import (
	"encoding/json"
	"net/http"
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"access_token": token})
}
