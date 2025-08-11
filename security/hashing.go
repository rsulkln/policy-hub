package security

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	if bytes, hErr := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost); hErr != nil {
		fmt.Println("Error to hashing password:")
		return "", hErr
	} else {
		return string(bytes), nil
	}
}
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
