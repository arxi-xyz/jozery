package usecases

import (
	"crypto/rand"
	"encoding/hex"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	accessTokenTTL   = 24 * time.Hour
	refreshTokenTTL  = 7 * 24 * time.Hour
	tokenTypeAccess  = "access"
	tokenTypeRefresh = "refresh"
)

var jwtSecret = []byte("secret")

type tokenClaims struct {
	UserID int64  `json:"user_id"`
	Type   string `json:"type"`
	jwt.RegisteredClaims
}

func signTokenPair(userID int64) (string, string, time.Time, time.Time, error) {
	access, accessExp, err := signToken(userID, tokenTypeAccess, accessTokenTTL)
	if err != nil {
		return "", "", time.Time{}, time.Time{}, err
	}
	refresh, refreshExp, err := signToken(userID, tokenTypeRefresh, refreshTokenTTL)
	if err != nil {
		return "", "", time.Time{}, time.Time{}, err
	}
	return access, refresh, accessExp, refreshExp, nil
}

func signToken(userID int64, tokenType string, ttl time.Duration) (string, time.Time, error) {
	now := time.Now()
	exp := now.Add(ttl)
	jti, err := newTokenID()
	if err != nil {
		return "", time.Time{}, err
	}

	claims := tokenClaims{
		UserID: userID,
		Type:   tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatInt(userID, 10),
			ID:        jti,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(exp),
		},
	}
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(jwtSecret)
	if err != nil {
		return "", time.Time{}, err
	}
	return signed, exp, nil
}

func parseToken(raw, tokenType string) (*tokenClaims, error) {
	claims := &tokenClaims{}
	token, err := jwt.ParseWithClaims(raw, claims, func(token *jwt.Token) (any, error) {
		if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, ErrInvalidToken
		}
		return jwtSecret, nil
	})
	if err != nil || !token.Valid || claims.Type != tokenType || claims.UserID == 0 {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

func newTokenID() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
