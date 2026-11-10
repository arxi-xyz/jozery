package app

import (
	"context"
	"errors"

	"jozery/internal/domains/authentication"
	auth "jozery/internal/domains/authentication/usecases"
	"jozery/internal/domains/identity"
	"jozery/internal/domains/identity/usecases"
)

type credentialsAdapter struct {
	users *identity.Module
}

func (a credentialsAdapter) FindCredentials(ctx context.Context, username string) (auth.Credentials, error) {
	creds, err := a.users.FindCredentials(ctx, username)
	if errors.Is(err, identity.ErrNotFound) {
		return auth.Credentials{}, auth.ErrCredentialsNotFound
	}
	if err != nil {
		return auth.Credentials{}, err
	}
	return auth.Credentials{UserID: creds.UserID, PasswordHash: creds.PasswordHash}, nil
}

type sessionAdapter struct {
	auth *authentication.Module
}

func (a *sessionAdapter) Login(ctx context.Context, username, password string, client usecases.Client) (usecases.Session, error) {
	access, refresh, err := a.auth.Login(ctx, username, password, auth.Client{
		IP:        client.IP,
		UserAgent: client.UserAgent,
	})
	if err != nil {
		return usecases.Session{}, err
	}
	return usecases.Session{AccessToken: access, RefreshToken: refresh}, nil
}
