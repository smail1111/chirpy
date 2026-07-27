package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
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

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "chirpy-access",
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
		Subject:   userID.String(),
	})

	return token.SignedString([]byte(tokenSecret))
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	token, er := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(*jwt.Token) (any, error) {
		return []byte(tokenSecret), nil
	})

	if er != nil {
		return uuid.Nil, er
	}

	str, er := token.Claims.GetSubject()

	if er != nil {
		return uuid.Nil, er
	}

	return uuid.Parse(str)
}

func GetBearerToken(headers http.Header) (string, error) {
	auth := headers.Get("Authorization")

	if auth == "" {
		return auth, fmt.Errorf("No Authorization")
	}

	split := strings.Split(auth, " ")

	if len(split) < 2 || split[0] != "Bearer" {
		return "", fmt.Errorf("Invalid Header Format")
	}

	return split[1], nil
}

func MakeRefreshToken() string {
	key := make([]byte, 32)

	rand.Read(key)

	return hex.EncodeToString(key)
}

func GetAPIKey(headers http.Header) (string, error) {
	auth := headers.Get("Authorization")

	if auth == "" {
		return auth, fmt.Errorf("No Authorization")
	}

	split := strings.Split(auth, " ")

	if len(split) < 2 || split[0] != "ApiKey" {
		return "", fmt.Errorf("Invalid Header Format")
	}

	return split[1], nil
}
