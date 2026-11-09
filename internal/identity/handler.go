package identity

import (
	"errors"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v5"
	"jozery/internal/identity/entity"
	"jozery/internal/identity/usecases"
)

type handler struct {
	identity *usecases.Service
	validate *validator.Validate
}

func newHandler(identity *usecases.Service) *handler {
	return &handler{identity: identity, validate: validator.New()}
}

func (h *handler) register(e *echo.Echo) {
	e.POST("/register", h.registerUser)
	e.GET("/profile", h.profile)
}

func (h *handler) registerUser(c *echo.Context) error {
	var request registerRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	if err := h.validate.Struct(request); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	session, err := h.identity.Register(c.Request().Context(), request.Username, request.Email, request.Password)
	if errors.Is(err, usecases.ErrUserExists) {
		return c.JSON(http.StatusConflict, map[string]string{"error": "User already exists"})
	}
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	request.Password = ""
	return c.JSON(http.StatusOK, registerResponse{
		Code:         "0000",
		Message:      "User registered successfully",
		Data:         request,
		Token:        session.AccessToken,
		RefreshToken: session.RefreshToken,
	})
}

func (h *handler) profile(c *echo.Context) error {
	raw, err := bearerToken(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid token"})
	}
	profile, err := h.identity.Profile(c.Request().Context(), raw)
	if errors.Is(err, usecases.ErrInvalidToken) {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid token"})
	}
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, profileResponse{
		Code:    "0000",
		Message: "Profile",
		Data:    profile,
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

type registerRequest struct {
	Username string `json:"username" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password,omitempty" validate:"required"`
}

type registerResponse struct {
	Code         string `json:"code"`
	Message      string `json:"message"`
	Data         any    `json:"data"`
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

type profileResponse struct {
	Code    string       `json:"code"`
	Message string       `json:"message"`
	Data    *entity.User `json:"data"`
}
