package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

func main() {
	db, err := openDB(context.Background())
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer db.Close()

	e := echo.New()

	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())

	e.GET("/", func(c *echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!\n")
	})

	e.GET("/health", func(c *echo.Context) error {
		return health(c, db)
	})

	e.POST("/login", func(c *echo.Context) error {
		var loginRequest LoginRequest

		if err := c.Bind(&loginRequest); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}

		var validator = validator.New()
		if err := validator.Struct(loginRequest); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}

		user, err := authenticate(c.Request().Context(), db, loginRequest.Username, loginRequest.Password)
		if errors.Is(err, errInvalidCredentials) {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid username or password"})
		}
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		loginRequest.Password = ""

		token, refreshToken, err := issueTokens(c.Request().Context(), db, user)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		return c.JSON(http.StatusOK, LoginResponse{
			Code:         "0000",
			Message:      "Login successful",
			Data:         loginRequest,
			Token:        token,
			RefreshToken: refreshToken,
		})
	})

	e.POST("/register", func(c *echo.Context) error {
		var registerRequest RegisterRequest

		if err := c.Bind(&registerRequest); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}

		var validator = validator.New()
		if err := validator.Struct(registerRequest); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}

		user, err := register(c.Request().Context(), db, registerRequest.Username, registerRequest.Email, registerRequest.Password)
		if errors.Is(err, errUserExists) {
			return c.JSON(http.StatusConflict, map[string]string{"error": "User already exists"})
		}
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		token, refreshToken, err := issueTokens(c.Request().Context(), db, user)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		registerRequest.Password = ""

		return c.JSON(http.StatusOK, RegisterResponse{
			Code:         "0000",
			Message:      "User registered successfully",
			Data:         registerRequest,
			Token:        token,
			RefreshToken: refreshToken,
		})
	})

	e.POST("/refresh-token", func(c *echo.Context) error {
		var refreshTokenRequest RefreshTokenRequest
		if err := c.Bind(&refreshTokenRequest); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		if err := validator.New().Struct(refreshTokenRequest); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}

		token, refreshToken, user, err := refreshTokens(c.Request().Context(), db, refreshTokenRequest.RefreshToken)
		if errors.Is(err, errInvalidToken) {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid token"})
		}
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusOK, RefreshTokenResponse{
			Code:         "0000",
			Message:      "Token refreshed successfully",
			Data:         user,
			Token:        token,
			RefreshToken: refreshToken,
		})
	})

	e.GET("/profile", func(c *echo.Context) error {
		raw, err := bearerToken(c)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid token"})
		}
		user, err := userFromAccessToken(c.Request().Context(), db, raw)
		if errors.Is(err, errInvalidToken) {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid token"})
		}
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusOK, ProfileResponse{
			Code:    "0000",
			Message: "Profile",
			Data:    user,
		})
	})

	if err := e.Start(":8080"); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}

func bearerToken(c *echo.Context) (string, error) {
	header := c.Request().Header.Get("Authorization")
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) || strings.TrimSpace(header[len(prefix):]) == "" {
		return "", errInvalidToken
	}
	return strings.TrimSpace(header[len(prefix):]), nil
}

type ProfileResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    *User  `json:"data"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type RefreshTokenResponse struct {
	Code         string      `json:"code"`
	Message      string      `json:"message"`
	Data         interface{} `json:"data"`
	Token        string      `json:"token"`
	RefreshToken string      `json:"refresh_token"`
}

type RegisterResponse struct {
	Code         string      `json:"code"`
	Message      string      `json:"message"`
	Data         interface{} `json:"data"`
	Token        string      `json:"token"`
	RefreshToken string      `json:"refresh_token"`
}

type RegisterRequest struct {
	Username string `json:"username" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password,omitempty" validate:"required"`
}

type HealthResponse struct {
	Status   string `json:"status"`
	Ping     string `json:"ping"`
	Version  string `json:"version"`
	Database string `json:"database"`
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

type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password,omitempty" validate:"required"`
}

type LoginResponse struct {
	Code         string      `json:"code"`
	Message      string      `json:"message"`
	Data         interface{} `json:"data"`
	Token        string      `json:"token"`
	RefreshToken string      `json:"refresh_token"`
}
