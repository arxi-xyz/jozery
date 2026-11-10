package app

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	"jozery/internal/domains/authentication/usecases"
	"jozery/internal/requestctx"
)

func TestAuthenticateStoresUser(t *testing.T) {
	e := echo.New()
	e.GET("/profile", func(c *echo.Context) error {
		userID, ok := requestctx.UserID(c.Request().Context())
		token, tokenOK := requestctx.AccessToken(c.Request().Context())
		if !ok || !tokenOK || userID != 7 || token != "access" {
			return c.NoContent(http.StatusForbidden)
		}
		return c.NoContent(http.StatusNoContent)
	}, authenticate(lookup{userID: 7}))

	req := httptest.NewRequest(http.MethodGet, "/profile", nil)
	req.Header.Set("Authorization", "Bearer access")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
}

func TestAuthenticateRejectsInvalidToken(t *testing.T) {
	e := echo.New()
	e.GET("/profile", func(c *echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	}, authenticate(lookup{err: usecases.ErrInvalidToken}))

	req := httptest.NewRequest(http.MethodGet, "/profile", nil)
	req.Header.Set("Authorization", "Bearer nope")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status %d", rec.Code)
	}
}

type lookup struct {
	userID int64
	err    error
}

func (l lookup) UserID(context.Context, string) (int64, error) {
	if l.err != nil {
		return 0, l.err
	}
	return l.userID, nil
}
