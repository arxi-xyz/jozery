package usecases

import (
	"context"
	"errors"

	"jozery/internal/identity/db"
)

var (
	ErrUserExists   = errors.New("user already exists")
	ErrInvalidToken = errors.New("invalid token")
)

type Session struct {
	AccessToken  string
	RefreshToken string
}

type Client struct {
	IP        string
	UserAgent string
}

type Tokens interface {
	Login(ctx context.Context, username, password string, client Client) (Session, error)
	UserID(ctx context.Context, accessToken string) (int64, error)
}

type Service struct {
	repo   db.Repository
	tokens Tokens
}

func New(repo db.Repository, tokens Tokens) *Service {
	return &Service{repo: repo, tokens: tokens}
}
