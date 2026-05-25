package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Claims struct {
	UserID string `json:"user_id"`

	jwt.RegisteredClaims
}

func NewJWTService(secret string, ttlMinutes int) *JWTService {
	if ttlMinutes <= 0 {
		ttlMinutes = 60
	}

	return &JWTService{
		secret: []byte(secret),
		ttl:    time.Duration(ttlMinutes) * time.Minute,
	}
}

type JWTService struct {
	secret []byte
	ttl    time.Duration
}

func (s *JWTService) GenerateAccessToken(userID uuid.UUID) (string, error) {
	now := time.Now().UTC()

	claims := Claims{
		UserID: userID.String(),

		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.ttl)),
		},
	}

	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.secret)
}

func (s *JWTService) ParseAccessToken(tokenString string) (uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&Claims{},
		func(token *jwt.Token) (any, error) {
			if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
				return nil, errors.New("invalid token signing method")
			}

			return s.secret, nil
		},
	)

	if err != nil {
		return uuid.Nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return uuid.Nil, errors.New("invalid token")
	}

	return uuid.Parse(claims.UserID)
}
