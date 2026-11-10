package app

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"
)

type healthResponse struct {
	Status   string `json:"status"`
	Ping     string `json:"ping"`
	Version  string `json:"version"`
	Database string `json:"database"`
}

func health(pool *pgxpool.Pool) echo.HandlerFunc {
	return func(c *echo.Context) error {
		database := "ok"
		status := "ok"
		code := http.StatusOK
		if err := pool.Ping(c.Request().Context()); err != nil {
			database = "down"
			status = "degraded"
			code = http.StatusServiceUnavailable
		}
		return c.JSON(code, healthResponse{Status: status, Ping: "pong", Version: "1.0.0", Database: database})
	}
}
