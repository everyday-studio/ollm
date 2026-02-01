package security

import (
	"crypto/rsa"
	"encoding/base64"
	"errors"
	"time"

	"github.com/everyday-studio/ollm/internal/domain"
	"github.com/golang-jwt/jwt/v4"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
	ErrParsingKey   = errors.New("error parsing RSA key")
)

type TokenType string

const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)

type JWTClaims struct {
	UserID int64     `json:"user_id"`
	Email  string    `json:"email"`
	Role   string    `json:"role"`
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

func GenerateToken(userID int64, email string, role domain.Role, privateKey *rsa.PrivateKey, expiratinTime time.Duration, tokenType TokenType) (string, error) {
	claims := &JWTClaims{
		UserID: userID,
		Email:  email,
		Role:   string(role),
		Type:   tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiratinTime)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func GenerateAccessToken(userID int64, email string, role domain.Role, privateKey *rsa.PrivateKey, expirationTime time.Duration) (string, error) {
	return GenerateToken(userID, email, role, privateKey, expirationTime, AccessToken)
}

func GenerateRefreshToken(userID int64, email string, role domain.Role, privateKey *rsa.PrivateKey, expirationTime time.Duration) (string, error) {
	return GenerateToken(userID, email, role, privateKey, expirationTime, RefreshToken)
}

// ValidateToken validates a JWT token and returns the claims
func ValidateToken(tokenString string, publicKey *rsa.PublicKey) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, ErrInvalidToken
		}
		return publicKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// ValidateAccessToken validates an access token
func ValidateAccessToken(tokenString string, publicKey *rsa.PublicKey) (*JWTClaims, error) {
	claims, err := ValidateToken(tokenString, publicKey)
	if err != nil {
		return nil, err
	}

	if claims.Type != AccessToken {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// ValidateRefreshToken validates a refresh token
func ValidateRefreshToken(tokenString string, publicKey *rsa.PublicKey) (*JWTClaims, error) {
	claims, err := ValidateToken(tokenString, publicKey)
	if err != nil {
		return nil, err
	}

	if claims.Type != RefreshToken {
		return nil, ErrInvalidToken
	}

	return claims, nil
}
