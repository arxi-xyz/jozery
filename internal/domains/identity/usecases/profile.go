package usecases

import (
	"context"
	"errors"

	"jozery/internal/domains/identity/db"
	"jozery/internal/domains/identity/entity"
)

func (s *Service) Profile(ctx context.Context, userID int64) (*entity.User, error) {
	user, err := s.repo.FindUserByID(ctx, userID)
	if errors.Is(err, db.ErrNotFound) {
		return nil, err
	}
	return user, err
}
