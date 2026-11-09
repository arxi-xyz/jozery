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
	ErrCurrentSession      = errors.New("current session")
	ErrSessionNotFound     = errors.New("session not found")
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

func (s *Service) issueSession(ctx context.Context, userID int64, meta sessionMeta) (*Session, error) {
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
		IP:               meta.IP,
		UserAgent:        meta.UserAgent,
		Device:           meta.Device,
		Location:         meta.Location,
		LoggedInAt:       meta.LoggedInAt,
	})
	if err != nil {
		return nil, err
	}
	return &Session{AccessToken: access, RefreshToken: refresh, UserID: userID}, nil
}

func (s *Service) UserID(ctx context.Context, raw string) (int64, error) {
	token, err := s.currentSession(ctx, raw)
	if err != nil {
		return 0, err
	}
	return token.UserID, nil
}

func (s *Service) currentSession(ctx context.Context, raw string) (entity.Token, error) {
	claims, err := parseToken(raw, tokenTypeAccess)
	if err != nil {
		return entity.Token{}, ErrInvalidToken
	}

	token, err := s.repo.FindActiveAccess(ctx, raw)
	if errors.Is(err, db.ErrNotFound) {
		return entity.Token{}, ErrInvalidToken
	}
	if err != nil {
		return entity.Token{}, err
	}
	if token.UserID != claims.UserID {
		return entity.Token{}, ErrInvalidToken
	}
	return token, nil
}
