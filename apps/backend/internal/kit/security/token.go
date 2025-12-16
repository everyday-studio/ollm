package security

import (
	"crypto/rsa"
	"encoding/base64"
	"errors"

	"github.com/golang-jwt/jwt/v4"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
	ErrParsingKey   = errors.New("error parsing RSA key")
)

type TokenType string

const (
	AccessToken TokenType = "access"
	RefresToken TokenType = "refresh"
)

type JWTClaims struct {
	UserID int64     `json:"user_id"`
	Email  string    `json:"email"`
	Roles  []string  `json:"roles"`
	Type   TokenType `json:"type"`
	jwt.RegisteredClaims
}

func ParseRSAPrivateKeyFromBase64(base64Key string) (*rsa.PrivateKey, error) {
	pemBytes, err := base64.StdEncoding.DecodeString(base64Key)
	if err != nil {
		return nil, ErrParsingKey
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(pemBytes)
	if err != nil {
		return nil, ErrParsingKey
	}
	return privateKey, nil
}

func ParseRSAPublicKeyFromBase64(base64Key string) (*rsa.PublicKey, error) {
	pemBytes, err := base64.StdEncoding.DecodeString(base64Key)
	if err != nil {
		return nil, ErrParsingKey
	}

	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(pemBytes)
	if err != nil {
		return nil, ErrParsingKey
	}
	return publicKey, nil
}
