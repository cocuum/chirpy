package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/golang-jwt/jwt/v5"
	"github.com/alexedwards/argon2id"
)

type TokenType string

const (
	// TokenTypeAccess -
	TokenTypeAccess TokenType = "chirpy-access"
)

func HashPassword(password string) (string, error) {
	hash, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return "", err
	}
	return hash, nil
}

func CheckPasswordHash(password, hash string) (bool, error) {
	match, err := argon2id.ComparePasswordAndHash(password, hash)
	if err != nil {
		return false, err
	}
	return match, nil
}

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	signInKey := []byte(tokenSecret)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:		string(TokenTypeAccess),
		IssuedAt:	jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt:	jwt.NewNumericDate(time.Now().Add(expiresIn)),
		Subject:	userID.String(),
	})
	return token.SignedString(signInKey)
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	claims := jwt.RegisteredClaims{}
	signInKey := []byte(tokenSecret)
	token, err := jwt.ParseWithClaims(tokenString, &claims, func(t *jwt.Token) (interface{}, error) {
		return signInKey,nil
	})

	if err != nil {
		return uuid.Nil,err
	}

	id, err := token.Claims.GetSubject()
	if err != nil {
		return uuid.Nil,err
	}

	issuer, err := token.Claims.GetIssuer()
	if err != nil {
		return uuid.Nil, err
	}
	if issuer != "chirpy-access" {
		return uuid.Nil, errors.New("invalid issuer")
	}

	userID, err := uuid.Parse(id)
	if err != nil {
		return uuid.Nil,fmt.Errorf("invalid user ID: %w", err)
	}
	return userID, nil
}