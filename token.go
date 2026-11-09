package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	accessTokenTTL   = 24 * time.Hour
	refreshTokenTTL  = 7 * 24 * time.Hour
	tokenTypeAccess  = "access"
	tokenTypeRefresh = "refresh"
)

var (
	jwtSecret       = []byte("secret")
	errInvalidToken = errors.New("invalid token")
)

type tokenClaims struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Type     string `json:"type"`
	jwt.RegisteredClaims
}

func issueTokens(ctx context.Context, pool *pgxpool.Pool, user *User) (string, string, error) {
	access, refresh, accessExp, refreshExp, err := signTokenPair(user)
	if err != nil {
		return "", "", err
	}

	_, err = pool.Exec(ctx, `
		INSERT INTO tokens (user_id, access_token, refresh_token, access_expires_at, refresh_expires_at)
		VALUES ($1, $2, $3, $4, $5)`,
		user.ID, access, refresh, accessExp, refreshExp)
	if err != nil {
		return "", "", fmt.Errorf("store tokens: %w", err)
	}
	return access, refresh, nil
}

func refreshTokens(ctx context.Context, pool *pgxpool.Pool, raw string) (string, string, *User, error) {
	claims, err := parseToken(raw, tokenTypeRefresh)
	if err != nil {
		return "", "", nil, errInvalidToken
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return "", "", nil, err
	}
	defer tx.Rollback(ctx)

	var userID int64
	err = tx.QueryRow(ctx, `
		SELECT user_id FROM tokens
		WHERE refresh_token = $1 AND revoked_at IS NULL AND refresh_expires_at > now()
		FOR UPDATE`, raw).Scan(&userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", "", nil, errInvalidToken
	}
	if err != nil {
		return "", "", nil, err
	}
	if userID != claims.UserID {
		return "", "", nil, errInvalidToken
	}

	user, err := getUserByID(ctx, tx, userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", "", nil, errInvalidToken
	}
	if err != nil {
		return "", "", nil, err
	}

	access, refresh, accessExp, refreshExp, err := signTokenPair(user)
	if err != nil {
		return "", "", nil, err
	}

	tag, err := tx.Exec(ctx, `
		UPDATE tokens SET revoked_at = now()
		WHERE refresh_token = $1 AND revoked_at IS NULL`, raw)
	if err != nil {
		return "", "", nil, err
	}
	if tag.RowsAffected() != 1 {
		return "", "", nil, errInvalidToken
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO tokens (user_id, access_token, refresh_token, access_expires_at, refresh_expires_at)
		VALUES ($1, $2, $3, $4, $5)`,
		user.ID, access, refresh, accessExp, refreshExp)
	if err != nil {
		return "", "", nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return "", "", nil, err
	}
	return access, refresh, user, nil
}

func userFromAccessToken(ctx context.Context, pool *pgxpool.Pool, raw string) (*User, error) {
	claims, err := parseToken(raw, tokenTypeAccess)
	if err != nil {
		return nil, errInvalidToken
	}

	var userID int64
	err = pool.QueryRow(ctx, `
		SELECT user_id FROM tokens
		WHERE access_token = $1 AND revoked_at IS NULL AND access_expires_at > now()`, raw).Scan(&userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errInvalidToken
	}
	if err != nil {
		return nil, err
	}
	if userID != claims.UserID {
		return nil, errInvalidToken
	}

	user, err := getUserByID(ctx, pool, userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errInvalidToken
	}
	return user, err
}

func signTokenPair(user *User) (string, string, time.Time, time.Time, error) {
	access, accessExp, err := signToken(user, tokenTypeAccess, accessTokenTTL)
	if err != nil {
		return "", "", time.Time{}, time.Time{}, err
	}
	refresh, refreshExp, err := signToken(user, tokenTypeRefresh, refreshTokenTTL)
	if err != nil {
		return "", "", time.Time{}, time.Time{}, err
	}
	return access, refresh, accessExp, refreshExp, nil
}

func signToken(user *User, tokenType string, ttl time.Duration) (string, time.Time, error) {
	now := time.Now()
	exp := now.Add(ttl)
	jti, err := newTokenID()
	if err != nil {
		return "", time.Time{}, err
	}

	claims := tokenClaims{
		UserID:   user.ID,
		Username: user.Username,
		Email:    user.Email,
		Type:     tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatInt(user.ID, 10),
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
			return nil, errInvalidToken
		}
		return jwtSecret, nil
	})
	if err != nil || !token.Valid || claims.Type != tokenType || claims.UserID == 0 {
		return nil, errInvalidToken
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
