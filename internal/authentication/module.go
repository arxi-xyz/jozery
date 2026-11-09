package authentication

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"
	"jozery/internal/authentication/db"
	"jozery/internal/authentication/usecases"
)

type Module struct {
	service *usecases.Service
}

func New(e *echo.Echo, pool *pgxpool.Pool, users usecases.CredentialsFinder) *Module {
	repo := db.New(pool)
	service := usecases.New(repo, users)
	newHandler(service).register(e)
	return &Module{service: service}
}

func (m *Module) Login(ctx context.Context, username, password string, client usecases.Client) (string, string, error) {
	session, err := m.service.Login(ctx, username, password, client)
	if err != nil {
		return "", "", err
	}
	return session.AccessToken, session.RefreshToken, nil
}

func (m *Module) UserID(ctx context.Context, accessToken string) (int64, error) {
	return m.service.UserID(ctx, accessToken)
}
