package jwt

import (
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/icoder-new/avito-shop/internal/config"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("token is invalid")
	ErrExpiredToken = errors.New("token has expired")
)

type TokenManager struct {
	signingKey []byte
	ttl        time.Duration
}

type Claims struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func NewTokenManager(jwt config.JWTCredentials) (*TokenManager, error) {
	return &TokenManager{
		signingKey: []byte(jwt.SecretKey),
		ttl:        jwt.ExpiresIn,
	}, nil
}

func (m *TokenManager) NewJWT(userID int64, username string) (string, error) {
	claims := Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(m.signingKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, nil
}

func (m *TokenManager) Parse(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&Claims{},
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return m.signingKey, nil
		},

		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}),
		jwt.WithLeeway(5*time.Second),
	)

	if err != nil {
		switch {
		case errors.Is(err, jwt.ErrTokenExpired):
			return nil, ErrExpiredToken
		case errors.Is(err, jwt.ErrTokenNotValidYet):
			return nil, ErrInvalidToken
		case errors.Is(err, jwt.ErrTokenInvalidClaims):
			return nil, ErrInvalidToken
		case errors.Is(err, jwt.ErrSignatureInvalid):
			return nil, ErrInvalidToken
		default:
			return nil, fmt.Errorf("failed to parse token: %w", err)
		}
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, ErrInvalidToken
	}

	if claims.UserID == 0 || claims.Username == "" {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

func (m *TokenManager) GetSigningKey() string {
	return base64.StdEncoding.EncodeToString(m.signingKey)
}
