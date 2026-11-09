package db

import (
	"context"
	"errors"

	"jozery/internal/authentication/entity"
)

var ErrNotFound = errors.New("not found")

type Repository interface {
	WithinTx(ctx context.Context, fn func(Repository) error) error
	SaveToken(ctx context.Context, token entity.Token) error
	FindActiveRefreshUserID(ctx context.Context, refreshToken string) (int64, error)
	RevokeRefresh(ctx context.Context, refreshToken string) error
	FindActiveAccessUserID(ctx context.Context, accessToken string) (int64, error)
	RevokeAccess(ctx context.Context, accessToken string) error
}
