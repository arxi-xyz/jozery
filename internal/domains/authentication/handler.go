package authentication

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v5"
	"jozery/internal/domains/authentication/usecases"
	"jozery/internal/requestctx"
)

type handler struct {
	auth     *usecases.Service
	validate *validator.Validate
}

func newHandler(auth *usecases.Service) *handler {
	return &handler{auth: auth, validate: validator.New()}
}

func (h *handler) register(e *echo.Echo, protected echo.MiddlewareFunc) {
	e.POST("/login", h.login)
	e.POST("/refresh-token", h.refreshToken)
	e.POST("/logout", h.logout, protected)
	e.POST("/logout-all", h.logoutAll, protected)
	e.GET("/sessions", h.sessions, protected)
	e.POST("/sessions/revoke-others", h.revokeOtherSessions, protected)
	e.POST("/sessions/:id/revoke", h.revokeSession, protected)
}

func (h *handler) login(c *echo.Context) error {
	var request loginRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	if err := h.validate.Struct(request); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	session, err := h.auth.Login(c.Request().Context(), request.Username, request.Password, clientFrom(c))
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
	raw, err := accessToken(c)
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

func (h *handler) logoutAll(c *echo.Context) error {
	raw, err := accessToken(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid token"})
	}
	err = h.auth.LogoutAll(c.Request().Context(), raw)
	if errors.Is(err, usecases.ErrInvalidToken) {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid token"})
	}
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{
		"code":    "0000",
		"message": "Logged out of all sessions",
	})
}

func (h *handler) sessions(c *echo.Context) error {
	raw, err := accessToken(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid token"})
	}
	sessions, err := h.auth.Sessions(c.Request().Context(), raw)
	if errors.Is(err, usecases.ErrInvalidToken) {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid token"})
	}
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, sessionsResponse{
		Code:    "0000",
		Message: "Sessions",
		Data:    sessions,
	})
}

func (h *handler) revokeSession(c *echo.Context) error {
	raw, err := accessToken(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid token"})
	}
	sessionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || sessionID <= 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid session"})
	}
	err = h.auth.RevokeSession(c.Request().Context(), raw, sessionID)
	if errors.Is(err, usecases.ErrInvalidToken) {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid token"})
	}
	if errors.Is(err, usecases.ErrCurrentSession) {
		return c.JSON(http.StatusConflict, map[string]string{"error": "Cannot revoke the current session"})
	}
	if errors.Is(err, usecases.ErrSessionNotFound) {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Session not found"})
	}
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{
		"code":    "0000",
		"message": "Session revoked",
	})
}

func (h *handler) revokeOtherSessions(c *echo.Context) error {
	raw, err := accessToken(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid token"})
	}
	err = h.auth.RevokeOtherSessions(c.Request().Context(), raw)
	if errors.Is(err, usecases.ErrInvalidToken) {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid token"})
	}
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{
		"code":    "0000",
		"message": "Other sessions revoked",
	})
}

func clientFrom(c *echo.Context) usecases.Client {
	return usecases.Client{IP: c.RealIP(), UserAgent: c.Request().UserAgent()}
}

func accessToken(c *echo.Context) (string, error) {
	raw, ok := requestctx.AccessToken(c.Request().Context())
	if !ok {
		return "", usecases.ErrInvalidToken
	}
	return raw, nil
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

type sessionsResponse struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Data    []usecases.SessionInfo `json:"data"`
}
