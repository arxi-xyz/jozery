package identity

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"
	"jozery/internal/identity/db"
	"jozery/internal/identity/usecases"
)

var ErrNotFound = db.ErrNotFound

type Credentials struct {
	UserID       int64
	PasswordHash string
}

type Module struct {
	repo db.Repository
}

func New(e *echo.Echo, pool *pgxpool.Pool, tokens usecases.Tokens) *Module {
	repo := db.New(pool)
	module := &Module{repo: repo}
	newHandler(usecases.New(repo, tokens)).register(e)
	return module
}

func (m *Module) FindCredentials(ctx context.Context, username string) (Credentials, error) {
	userID, passwordHash, err := m.repo.FindCredentialsByUsername(ctx, username)
	if errors.Is(err, db.ErrNotFound) {
		return Credentials{}, ErrNotFound
	}
	if err != nil {
		return Credentials{}, err
	}
	return Credentials{UserID: userID, PasswordHash: passwordHash}, nil
}
