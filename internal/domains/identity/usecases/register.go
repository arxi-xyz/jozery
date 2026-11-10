package usecases

import (
	"context"
	"errors"
	"fmt"

	"jozery/internal/domains/identity/db"
	"jozery/internal/domains/identity/entity"
)

func (s *Service) Register(ctx context.Context, username, email, password string, client Client) (*Session, error) {
	hash, err := entity.HashPassword(password)
	if err != nil {
		return nil, err
	}

	_, err = s.repo.FindUserByUsernameOrEmail(ctx, username, email)
	if err == nil {
		return nil, ErrUserExists
	}
	if !errors.Is(err, db.ErrNotFound) {
		return nil, err
	}

	if _, err := s.repo.CreateUser(ctx, username, email, hash); err != nil {
		return nil, fmt.Errorf("register user: %w", err)
	}
	session, err := s.tokens.Login(ctx, username, password, client)
	if err != nil {
		return nil, err
	}
	return &session, nil
}
