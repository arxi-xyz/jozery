package usecases

import (
	"context"
	"errors"

	"jozery/internal/Domains/authentication/db"
)

func (s *Service) Logout(ctx context.Context, accessToken string) error {
	if _, err := s.UserID(ctx, accessToken); err != nil {
		return err
	}
	err := s.repo.RevokeAccess(ctx, accessToken)
	if errors.Is(err, db.ErrNotFound) {
		return ErrInvalidToken
	}
	return err
}

func (s *Service) LogoutAll(ctx context.Context, accessToken string) error {
	userID, err := s.UserID(ctx, accessToken)
	if err != nil {
		return err
	}
	return s.repo.RevokeUser(ctx, userID)
}
