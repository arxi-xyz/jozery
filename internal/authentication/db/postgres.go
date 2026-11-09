package db

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"jozery/internal/authentication/entity"
)

type Postgres struct {
	pool *pgxpool.Pool
	q    querier
}

type querier interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

func New(pool *pgxpool.Pool) Repository {
	return &Postgres{pool: pool, q: pool}
}

func (p *Postgres) WithinTx(ctx context.Context, fn func(Repository) error) error {
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := fn(&Postgres{pool: p.pool, q: tx}); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (p *Postgres) SaveToken(ctx context.Context, token entity.Token) error {
	_, err := p.q.Exec(ctx, `
		INSERT INTO tokens (user_id, access_token, refresh_token, access_expires_at, refresh_expires_at)
		VALUES ($1, $2, $3, $4, $5)`,
		token.UserID, token.AccessToken, token.RefreshToken, token.AccessExpiresAt, token.RefreshExpiresAt)
	return err
}

func (p *Postgres) FindActiveRefreshUserID(ctx context.Context, refreshToken string) (int64, error) {
	var userID int64
	err := p.q.QueryRow(ctx, `
		SELECT user_id FROM tokens
		WHERE refresh_token = $1 AND revoked_at IS NULL AND refresh_expires_at > now()
		FOR UPDATE`, refreshToken).Scan(&userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, ErrNotFound
	}
	return userID, err
}

func (p *Postgres) RevokeRefresh(ctx context.Context, refreshToken string) error {
	tag, err := p.q.Exec(ctx, `
		UPDATE tokens SET revoked_at = now()
		WHERE refresh_token = $1 AND revoked_at IS NULL`, refreshToken)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return ErrNotFound
	}
	return nil
}

func (p *Postgres) FindActiveAccessUserID(ctx context.Context, accessToken string) (int64, error) {
	var userID int64
	err := p.q.QueryRow(ctx, `
		SELECT user_id FROM tokens
		WHERE access_token = $1 AND revoked_at IS NULL AND access_expires_at > now()`, accessToken).Scan(&userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, ErrNotFound
	}
	return userID, err
}

func (p *Postgres) RevokeAccess(ctx context.Context, accessToken string) error {
	tag, err := p.q.Exec(ctx, `
		UPDATE tokens SET revoked_at = now()
		WHERE access_token = $1 AND revoked_at IS NULL`, accessToken)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return ErrNotFound
	}
	return nil
}
