package authentication

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"
	"jozery/internal/Domains/authentication/db"
	"jozery/internal/Domains/authentication/usecases"
)

type Module struct {
	service *usecases.Service
	handler *handler
}

func New(pool *pgxpool.Pool, users usecases.CredentialsFinder) *Module {
	repo := db.New(pool)
	service := usecases.New(repo, users)
	return &Module{service: service, handler: newHandler(service)}
}

func (m *Module) Routes(e *echo.Echo, protected echo.MiddlewareFunc) {
	m.handler.register(e, protected)
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
