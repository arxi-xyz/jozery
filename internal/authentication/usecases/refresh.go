package usecases

import (
	"context"
	"errors"

	"jozery/internal/authentication/db"
	"jozery/internal/authentication/entity"
)

func (s *Service) Refresh(ctx context.Context, raw string) (*Session, error) {
	claims, err := parseToken(raw, tokenTypeRefresh)
	if err != nil {
		return nil, ErrInvalidToken
	}

	var session *Session
	err = s.repo.WithinTx(ctx, func(tx db.Repository) error {
		userID, err := tx.FindActiveRefreshUserID(ctx, raw)
		if errors.Is(err, db.ErrNotFound) {
			return ErrInvalidToken
		}
		if err != nil {
			return err
		}
		if userID != claims.UserID {
			return ErrInvalidToken
		}

		access, refresh, accessExp, refreshExp, err := signTokenPair(userID)
		if err != nil {
			return err
		}
		if err := tx.RevokeRefresh(ctx, raw); err != nil {
			if errors.Is(err, db.ErrNotFound) {
				return ErrInvalidToken
			}
			return err
		}
		if err := tx.SaveToken(ctx, entity.Token{
			UserID:           userID,
			AccessToken:      access,
			RefreshToken:     refresh,
			AccessExpiresAt:  accessExp,
			RefreshExpiresAt: refreshExp,
		}); err != nil {
			return err
		}

		session = &Session{AccessToken: access, RefreshToken: refresh, UserID: userID}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return session, nil
}
