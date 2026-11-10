package app

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/labstack/echo/v5"
	"jozery/internal/domains/authentication/usecases"
	"jozery/internal/requestctx"
)

type sessionReader interface {
	UserID(ctx context.Context, accessToken string) (int64, error)
}

func authenticate(sessions sessionReader) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			raw, ok := bearerToken(c)
			if !ok {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid token"})
			}
			userID, err := sessions.UserID(c.Request().Context(), raw)
			if errors.Is(err, usecases.ErrInvalidToken) {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid token"})
			}
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
			}
			c.SetRequest(c.Request().WithContext(requestctx.With(c.Request().Context(), userID, raw)))
			return next(c)
		}
	}
}

func bearerToken(c *echo.Context) (string, bool) {
	header := c.Request().Header.Get("Authorization")
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return "", false
	}
	raw := strings.TrimSpace(header[len(prefix):])
	if raw == "" {
		return "", false
	}
	return raw, true
}
