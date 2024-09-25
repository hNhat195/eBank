package token

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const minSecretKeySize = 32

type JWTMaker struct {
	secretKey string
}

type JWTPayloadClaims struct {
	Payload
	jwt.RegisteredClaims
}

func NewJWTPayloadClaims(payload *Payload) *JWTPayloadClaims {
	return &JWTPayloadClaims{
		Payload: *payload,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(payload.ExpiredAt),
			IssuedAt:  jwt.NewNumericDate(payload.IssuedAt),
			NotBefore: jwt.NewNumericDate(payload.IssuedAt),
			Subject:   payload.Username,
			ID:        payload.ID.String(),
			Audience:  []string{"clients"},
		},
	}
}
func NewJWTMaker(secretKey string) (Maker, error) {
	if len(secretKey) < minSecretKeySize {
		return nil, fmt.Errorf("invalid key size: must be at least %d characters", minSecretKeySize)
	}
	return &JWTMaker{secretKey}, nil
}

func (maker *JWTMaker) CreateToken(username string, duration time.Duration) (string, error) {
	payload, err := NewPayload(username, duration)
	if err != nil {
		return "", err
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, NewJWTPayloadClaims(payload))

	return jwtToken.SignedString([]byte(maker.secretKey))
}

func (maker *JWTMaker) VerifyToken(token string) (*Payload, error) {
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, ErrInvalidToken
		}
		return []byte(maker.secretKey), nil
	}

	jwtToken, err := jwt.ParseWithClaims(token, &JWTPayloadClaims{}, keyFunc)

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		} else if errors.Is(err, ErrInvalidToken) {
			return nil, ErrInvalidToken
		} else {
			return nil, err
		}
	}

	payloadClaims, ok := jwtToken.Claims.(*JWTPayloadClaims)
	if !ok {
		return nil, ErrInvalidToken
	}

	return &payloadClaims.Payload, nil

}
