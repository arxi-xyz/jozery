package usecases

import (
	"context"
	"errors"

	"jozery/internal/authentication/db"
	"jozery/internal/authentication/entity"
)

var (
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrInvalidToken        = errors.New("invalid token")
	ErrCredentialsNotFound = errors.New("credentials not found")
)

type Credentials struct {
	UserID       int64
	PasswordHash string
}

type CredentialsFinder interface {
	FindCredentials(ctx context.Context, username string) (Credentials, error)
}

type Session struct {
	AccessToken  string
	RefreshToken string
	UserID       int64
}

type Service struct {
	repo  db.Repository
	users CredentialsFinder
}

func New(repo db.Repository, users CredentialsFinder) *Service {
	return &Service{repo: repo, users: users}
}

func (s *Service) issueSession(ctx context.Context, userID int64) (*Session, error) {
	access, refresh, accessExp, refreshExp, err := signTokenPair(userID)
	if err != nil {
		return nil, err
	}
	err = s.repo.SaveToken(ctx, entity.Token{
		UserID:           userID,
		AccessToken:      access,
		RefreshToken:     refresh,
		AccessExpiresAt:  accessExp,
		RefreshExpiresAt: refreshExp,
	})
	if err != nil {
		return nil, err
	}
	return &Session{AccessToken: access, RefreshToken: refresh, UserID: userID}, nil
}

func (s *Service) UserID(ctx context.Context, raw string) (int64, error) {
	claims, err := parseToken(raw, tokenTypeAccess)
	if err != nil {
		return 0, ErrInvalidToken
	}

	userID, err := s.repo.FindActiveAccessUserID(ctx, raw)
	if errors.Is(err, db.ErrNotFound) {
		return 0, ErrInvalidToken
	}
	if err != nil {
		return 0, err
	}
	if userID != claims.UserID {
		return 0, ErrInvalidToken
	}
	return userID, nil
}
