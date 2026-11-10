package app

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"jozery/internal/Domains/authentication"
	"jozery/internal/Domains/identity"
)

func Mount(e *echo.Echo, pool *pgxpool.Pool) {
	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())
	e.GET("/health", health(pool))

	sessions := &sessionAdapter{}
	users := identity.New(pool, sessions)
	authn := authentication.New(pool, credentialsAdapter{users: users})
	sessions.auth = authn

	protected := authenticate(authn)
	users.Routes(e, protected)
	authn.Routes(e, protected)
}
