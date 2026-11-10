package db

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"jozery/internal/domains/authentication/entity"
)

type Postgres struct {
	pool *pgxpool.Pool
	q    querier
}

type querier interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
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
		INSERT INTO tokens (
			user_id, access_token, refresh_token, access_expires_at, refresh_expires_at,
			ip, user_agent, device, location, logged_in_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		token.UserID, token.AccessToken, token.RefreshToken, token.AccessExpiresAt, token.RefreshExpiresAt,
		token.IP, token.UserAgent, token.Device, token.Location, token.LoggedInAt)
	return err
}

func (p *Postgres) RotateToken(ctx context.Context, id int64, token entity.Token) error {
	tag, err := p.q.Exec(ctx, `
		UPDATE tokens
		SET access_token = $2, refresh_token = $3, access_expires_at = $4, refresh_expires_at = $5
		WHERE id = $1 AND revoked_at IS NULL`,
		id, token.AccessToken, token.RefreshToken, token.AccessExpiresAt, token.RefreshExpiresAt)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return ErrNotFound
	}
	return nil
}

func (p *Postgres) FindActiveRefresh(ctx context.Context, refreshToken string) (entity.Token, error) {
	var token entity.Token
	err := p.q.QueryRow(ctx, `
		SELECT id, user_id, ip, user_agent, device, location, logged_in_at
		FROM tokens
		WHERE refresh_token = $1 AND revoked_at IS NULL AND refresh_expires_at > now()
		FOR UPDATE`, refreshToken).Scan(
		&token.ID, &token.UserID, &token.IP, &token.UserAgent, &token.Device, &token.Location, &token.LoggedInAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return entity.Token{}, ErrNotFound
	}
	return token, err
}

func (p *Postgres) FindActiveAccess(ctx context.Context, accessToken string) (entity.Token, error) {
	var token entity.Token
	err := p.q.QueryRow(ctx, `
		SELECT id, user_id, ip, user_agent, device, location, logged_in_at
		FROM tokens
		WHERE access_token = $1 AND revoked_at IS NULL AND access_expires_at > now()`, accessToken).Scan(
		&token.ID, &token.UserID, &token.IP, &token.UserAgent, &token.Device, &token.Location, &token.LoggedInAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return entity.Token{}, ErrNotFound
	}
	return token, err
}

func (p *Postgres) ListSessions(ctx context.Context, userID int64) ([]entity.Token, error) {
	rows, err := p.q.Query(ctx, `
		SELECT id, user_id, ip, user_agent, device, location, logged_in_at
		FROM tokens
		WHERE user_id = $1 AND revoked_at IS NULL AND refresh_expires_at > now()
		ORDER BY logged_in_at DESC, id DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sessions := make([]entity.Token, 0)
	for rows.Next() {
		var token entity.Token
		if err := rows.Scan(&token.ID, &token.UserID, &token.IP, &token.UserAgent, &token.Device, &token.Location, &token.LoggedInAt); err != nil {
			return nil, err
		}
		sessions = append(sessions, token)
	}
	return sessions, rows.Err()
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

func (p *Postgres) RevokeSession(ctx context.Context, userID, sessionID int64) error {
	tag, err := p.q.Exec(ctx, `
		UPDATE tokens SET revoked_at = now()
		WHERE id = $1 AND user_id = $2 AND revoked_at IS NULL`, sessionID, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return ErrNotFound
	}
	return nil
}

func (p *Postgres) RevokeOtherSessions(ctx context.Context, userID, keepID int64) error {
	_, err := p.q.Exec(ctx, `
		UPDATE tokens SET revoked_at = now()
		WHERE user_id = $1 AND id <> $2 AND revoked_at IS NULL`, userID, keepID)
	return err
}

func (p *Postgres) RevokeUser(ctx context.Context, userID int64) error {
	_, err := p.q.Exec(ctx, `
		UPDATE tokens SET revoked_at = now()
		WHERE user_id = $1 AND revoked_at IS NULL`, userID)
	return err
}
