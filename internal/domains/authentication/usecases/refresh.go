package usecases

import (
	"context"
	"errors"

	"jozery/internal/Domains/authentication/db"
	"jozery/internal/Domains/authentication/entity"
)

func (s *Service) Refresh(ctx context.Context, raw string) (*Session, error) {
	claims, err := parseToken(raw, tokenTypeRefresh)
	if err != nil {
		return nil, ErrInvalidToken
	}

	var session *Session
	err = s.repo.WithinTx(ctx, func(tx db.Repository) error {
		current, err := tx.FindActiveRefresh(ctx, raw)
		if errors.Is(err, db.ErrNotFound) {
			return ErrInvalidToken
		}
		if err != nil {
			return err
		}
		if current.UserID != claims.UserID {
			return ErrInvalidToken
		}

		access, refresh, accessExp, refreshExp, err := signTokenPair(current.UserID)
		if err != nil {
			return err
		}
		if err := tx.RotateToken(ctx, current.ID, entity.Token{
			AccessToken:      access,
			RefreshToken:     refresh,
			AccessExpiresAt:  accessExp,
			RefreshExpiresAt: refreshExp,
		}); err != nil {
			if errors.Is(err, db.ErrNotFound) {
				return ErrInvalidToken
			}
			return err
		}

		session = &Session{AccessToken: access, RefreshToken: refresh, UserID: current.UserID}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return session, nil
}
