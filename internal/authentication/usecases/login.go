package usecases

import (
	"context"
	"errors"

	"jozery/internal/authentication/entity"
)

func (s *Service) Login(ctx context.Context, username, password string, client Client) (*Session, error) {
	creds, err := s.users.FindCredentials(ctx, username)
	if errors.Is(err, ErrCredentialsNotFound) {
		return nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, err
	}
	if !entity.PasswordMatches(creds.PasswordHash, password) {
		return nil, ErrInvalidCredentials
	}
	return s.issueSession(ctx, creds.UserID, describeClient(ctx, client))
}
