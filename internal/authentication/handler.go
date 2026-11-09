package authentication

import (
	"errors"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v5"
	"jozery/internal/authentication/usecases"
)

type handler struct {
	auth     *usecases.Service
	validate *validator.Validate
}

func newHandler(auth *usecases.Service) *handler {
	return &handler{auth: auth, validate: validator.New()}
}

func (h *handler) register(e *echo.Echo) {
	e.POST("/login", h.login)
	e.POST("/refresh-token", h.refreshToken)
	e.POST("/logout", h.logout)
}

func (h *handler) login(c *echo.Context) error {
	var request loginRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	if err := h.validate.Struct(request); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	session, err := h.auth.Login(c.Request().Context(), request.Username, request.Password)
	if errors.Is(err, usecases.ErrInvalidCredentials) {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid username or password"})
	}
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	request.Password = ""
	return c.JSON(http.StatusOK, tokenResponse{
		Code:         "0000",
		Message:      "Login successful",
		Data:         request,
		Token:        session.AccessToken,
		RefreshToken: session.RefreshToken,
	})
}

func (h *handler) refreshToken(c *echo.Context) error {
	var request refreshTokenRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	if err := h.validate.Struct(request); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	session, err := h.auth.Refresh(c.Request().Context(), request.RefreshToken)
	if errors.Is(err, usecases.ErrInvalidToken) {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid token"})
	}
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, tokenResponse{
		Code:         "0000",
		Message:      "Token refreshed successfully",
		Data:         session.UserID,
		Token:        session.AccessToken,
		RefreshToken: session.RefreshToken,
	})
}

func (h *handler) logout(c *echo.Context) error {
	raw, err := bearerToken(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid token"})
	}
	err = h.auth.Logout(c.Request().Context(), raw)
	if errors.Is(err, usecases.ErrInvalidToken) {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid token"})
	}
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{
		"code":    "0000",
		"message": "Logged out",
	})
}

func bearerToken(c *echo.Context) (string, error) {
	header := c.Request().Header.Get("Authorization")
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) || strings.TrimSpace(header[len(prefix):]) == "" {
		return "", usecases.ErrInvalidToken
	}
	return strings.TrimSpace(header[len(prefix):]), nil
}

type loginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password,omitempty" validate:"required"`
}

type refreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type tokenResponse struct {
	Code         string `json:"code"`
	Message      string `json:"message"`
	Data         any    `json:"data"`
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}
