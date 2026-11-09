package main

import (
	"context"
	"errors"
	"log"
	"net/http"

	"jozery/internal/authentication"
	auth "jozery/internal/authentication/usecases"
	"jozery/internal/identity"
	"jozery/internal/identity/usecases"
	"jozery/lib/db/pgsql"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

func main() {
	db, err := pgsql.Open(context.Background())
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer db.Close()

	e := echo.New()

	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())

	e.GET("/health", func(c *echo.Context) error {
		return health(c, db)
	})

	sessions := &sessionAdapter{}
	users := identity.New(e, db, sessions)
	sessions.auth = authentication.New(e, db, credentialsAdapter{users: users})

	if err := e.Start(":8080"); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}

type HealthResponse struct {
	Status   string `json:"status"`
	Ping     string `json:"ping"`
	Version  string `json:"version"`
	Database string `json:"database"`
}

type credentialsAdapter struct {
	users *identity.Module
}

func (a credentialsAdapter) FindCredentials(ctx context.Context, username string) (auth.Credentials, error) {
	creds, err := a.users.FindCredentials(ctx, username)
	if errors.Is(err, identity.ErrNotFound) {
		return auth.Credentials{}, auth.ErrCredentialsNotFound
	}
	if err != nil {
		return auth.Credentials{}, err
	}
	return auth.Credentials{UserID: creds.UserID, PasswordHash: creds.PasswordHash}, nil
}

type sessionAdapter struct {
	auth *authentication.Module
}

func (a *sessionAdapter) Login(ctx context.Context, username, password string, client usecases.Client) (usecases.Session, error) {
	access, refresh, err := a.auth.Login(ctx, username, password, auth.Client{
		IP:        client.IP,
		UserAgent: client.UserAgent,
	})
	if err != nil {
		return usecases.Session{}, err
	}
	return usecases.Session{AccessToken: access, RefreshToken: refresh}, nil
}

func (a *sessionAdapter) UserID(ctx context.Context, accessToken string) (int64, error) {
	userID, err := a.auth.UserID(ctx, accessToken)
	if errors.Is(err, auth.ErrInvalidToken) {
		return 0, usecases.ErrInvalidToken
	}
	return userID, err
}

func health(c *echo.Context, db *pgxpool.Pool) error {
	database := "ok"
	status := "ok"
	code := http.StatusOK
	if err := db.Ping(c.Request().Context()); err != nil {
		database = "down"
		status = "degraded"
		code = http.StatusServiceUnavailable
	}
	return c.JSON(code, HealthResponse{Status: status, Ping: "pong", Version: "1.0.0", Database: database})
}
