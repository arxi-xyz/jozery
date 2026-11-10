package entity

import "golang.org/x/crypto/bcrypt"

func PasswordMatches(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
