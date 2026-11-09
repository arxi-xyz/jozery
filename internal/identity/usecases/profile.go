package usecases

import (
	"context"
	"errors"

	"jozery/internal/identity/db"
	"jozery/internal/identity/entity"
)

func (s *Service) Profile(ctx context.Context, accessToken string) (*entity.User, error) {
	userID, err := s.tokens.UserID(ctx, accessToken)
	if err != nil {
		return nil, err
	}
	user, err := s.repo.FindUserByID(ctx, userID)
	if errors.Is(err, db.ErrNotFound) {
		return nil, ErrInvalidToken
	}
	return user, err
}
