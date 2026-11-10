package db

import (
	"context"
	"errors"

	"jozery/internal/domains/authentication/entity"
)

var ErrNotFound = errors.New("not found")

type Repository interface {
	WithinTx(ctx context.Context, fn func(Repository) error) error
	SaveToken(ctx context.Context, token entity.Token) error
	RotateToken(ctx context.Context, id int64, token entity.Token) error
	FindActiveRefresh(ctx context.Context, refreshToken string) (entity.Token, error)
	FindActiveAccess(ctx context.Context, accessToken string) (entity.Token, error)
	ListSessions(ctx context.Context, userID int64) ([]entity.Token, error)
	RevokeAccess(ctx context.Context, accessToken string) error
	RevokeSession(ctx context.Context, userID, sessionID int64) error
	RevokeOtherSessions(ctx context.Context, userID, keepID int64) error
	RevokeUser(ctx context.Context, userID int64) error
}
