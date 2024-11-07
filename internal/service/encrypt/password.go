package encrypt

import (
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const hashCost = 11

var (
	ErrCharNotAllowed = errors.New("character not allowed")
)

func PasswordEncrypt(password string) (string, error) {
	if !IsPrintableASCII(password) {
		return "", ErrCharNotAllowed
	}

	byteStr := []byte(password)

	hash, err := bcrypt.GenerateFromPassword(byteStr, hashCost)
	if err != nil {
		return "", fmt.Errorf("hash operation failed: %w", err)
	}

	return string(hash), nil
}

func PasswordCompare(password, hash string) bool {
	bytePassword := []byte(password)
	byteHash := []byte(hash)

	err := bcrypt.CompareHashAndPassword(byteHash, bytePassword)

	return err == nil
}

func IsPrintableASCII(s string) bool {
	for _, r := range s {
		if r < 32 || r > 126 {
			return false
		}
	}

	return true
}
