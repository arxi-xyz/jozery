package db

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"jozery/internal/identity/entity"
)

type Postgres struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) Repository {
	return &Postgres{pool: pool}
}

func (p *Postgres) FindUserByUsernameOrEmail(ctx context.Context, username, email string) (*entity.User, error) {
	var user entity.User
	err := p.pool.QueryRow(ctx, `
		SELECT id, username, email, created_at, updated_at
		FROM users WHERE username = $1 OR email = $2`, username, email).
		Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt, &user.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (p *Postgres) FindUserByID(ctx context.Context, id int64) (*entity.User, error) {
	var user entity.User
	err := p.pool.QueryRow(ctx, `
		SELECT id, username, email, created_at, updated_at
		FROM users WHERE id = $1`, id).
		Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt, &user.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (p *Postgres) FindCredentialsByUsername(ctx context.Context, username string) (int64, string, error) {
	var userID int64
	var passwordHash string
	err := p.pool.QueryRow(ctx, `
		SELECT id, password_hash FROM users WHERE username = $1`, username).
		Scan(&userID, &passwordHash)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, "", ErrNotFound
	}
	if err != nil {
		return 0, "", err
	}
	return userID, passwordHash, nil
}

func (p *Postgres) CreateUser(ctx context.Context, username, email, passwordHash string) (*entity.User, error) {
	var user entity.User
	err := p.pool.QueryRow(ctx, `
		INSERT INTO users (username, email, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id, username, email, created_at, updated_at`, username, email, passwordHash).
		Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
