package security

import (
	"testing"
)

func TestHashedAndCheckPassword(t *testing.T) {
	password := "1234565432test"
	hash, err := HashPassword(password)
	if err != nil {
		t.Errorf("Error hashing password: %v", err)
		return
	}
	if !CheckPasswordHash(password, hash) {
		t.Errorf("Password and hash do not match")
	}
	if CheckPasswordHash("wrongPassword!", hash) {
		t.Errorf("wrong password mathed the hash!!")
	}
}
