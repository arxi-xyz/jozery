package usecases

import (
	"context"
	"errors"
	"fmt"

	"jozery/internal/domains/identity/db"
	"jozery/internal/domains/identity/entity"
)

func SeedAdmin(ctx context.Context, repo db.Repository, username, email, password string) (bool, error) {
	_, err := repo.FindUserByUsernameOrEmail(ctx, username, email)
	if err == nil {
		return false, nil
	}
	if !errors.Is(err, db.ErrNotFound) {
		return false, err
	}

	hash, err := entity.HashPassword(password)
	if err != nil {
		return false, err
	}
	if _, err := repo.CreateUser(ctx, username, email, hash); err != nil {
		return false, fmt.Errorf("seed admin: %w", err)
	}
	return true, nil
}
