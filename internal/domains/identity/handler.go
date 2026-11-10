package identity

import (
	"errors"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v5"
	"jozery/internal/domains/identity/entity"
	"jozery/internal/domains/identity/usecases"
	"jozery/internal/requestctx"
)

type handler struct {
	identity *usecases.Service
	validate *validator.Validate
}

func newHandler(identity *usecases.Service) *handler {
	return &handler{identity: identity, validate: validator.New()}
}

func (h *handler) register(e *echo.Echo, protected echo.MiddlewareFunc) {
	e.POST("/register", h.registerUser)
	e.GET("/profile", h.profile, protected)
}

func (h *handler) registerUser(c *echo.Context) error {
	var request registerRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	if err := h.validate.Struct(request); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	session, err := h.identity.Register(c.Request().Context(), request.Username, request.Email, request.Password, usecases.Client{
		IP:        c.RealIP(),
		UserAgent: c.Request().UserAgent(),
	})
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
	userID, ok := requestctx.UserID(c.Request().Context())
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid token"})
	}
	profile, err := h.identity.Profile(c.Request().Context(), userID)
	if errors.Is(err, ErrNotFound) {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "User not found"})
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
