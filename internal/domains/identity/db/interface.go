package db

import (
	"context"
	"errors"

	"jozery/internal/Domains/identity/entity"
)

var ErrNotFound = errors.New("not found")

type Repository interface {
	FindUserByUsernameOrEmail(ctx context.Context, username, email string) (*entity.User, error)
	FindUserByID(ctx context.Context, id int64) (*entity.User, error)
	FindCredentialsByUsername(ctx context.Context, username string) (int64, string, error)
	CreateUser(ctx context.Context, username, email, passwordHash string) (*entity.User, error)
}
