package auth

import (
	"github.com/alexedwards/argon2id"
)

func HashPassword(password string) (string, error) {
	hashed, er := argon2id.CreateHash(password, argon2id.DefaultParams)
	if er != nil {
		return "", er
	}
	return hashed, nil
}

func CheckPasswordHash(password, hash string) (bool, error) {
	match, er := argon2id.ComparePasswordAndHash(password, hash)
	if er != nil {
		return false, er
	}
	return match, nil
}
