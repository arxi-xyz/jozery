package main

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

//go:embed db/postgres/migrations/*.sql
var migrationFS embed.FS

func databaseURL() string {
	if url := os.Getenv("DATABASE_URL"); url != "" {
		return url
	}
	return "postgres://jozery:jozery@localhost:5432/jozery?sslmode=disable"
}

func openDB(ctx context.Context) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, databaseURL())
	if err != nil {
		return nil, fmt.Errorf("connect postgres: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	if err := migrate(ctx, pool); err != nil {
		pool.Close()
		return nil, err
	}

	return pool, nil
}

func migrate(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)`)
	if err != nil {
		return fmt.Errorf("migrate schema_migrations: %w", err)
	}

	entries, err := fs.ReadDir(migrationFS, "db/postgres/migrations")
	if err != nil {
		return fmt.Errorf("read migrations: %w", err)
	}

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".up.sql") {
			names = append(names, entry.Name())
		}
	}
	sort.Strings(names)

	for _, name := range names {
		if err := applyMigration(ctx, pool, name); err != nil {
			return err
		}
	}

	return seedAdmin(ctx, pool)
}

func applyMigration(ctx context.Context, pool *pgxpool.Pool, name string) error {
	var applied bool
	err := pool.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE version = $1)`, name).Scan(&applied)
	if err != nil {
		return fmt.Errorf("check migration %s: %w", name, err)
	}
	if applied {
		return nil
	}

	body, err := migrationFS.ReadFile("db/postgres/migrations/" + name)
	if err != nil {
		return fmt.Errorf("read migration %s: %w", name, err)
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin migration %s: %w", name, err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, string(body), pgx.QueryExecModeSimpleProtocol); err != nil {
		return fmt.Errorf("apply migration %s: %w", name, err)
	}
	if _, err := tx.Exec(ctx, `INSERT INTO schema_migrations (version) VALUES ($1)`, name); err != nil {
		return fmt.Errorf("record migration %s: %w", name, err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit migration %s: %w", name, err)
	}
	return nil
}

func seedAdmin(ctx context.Context, pool *pgxpool.Pool) error {
	hash, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash seed password: %w", err)
	}

	_, err = pool.Exec(ctx, `
		INSERT INTO users (username, email, password_hash)
		VALUES ($1, $2, $3)
		ON CONFLICT (username) DO NOTHING`, "admin", "admin@jozery.local", string(hash))
	if err != nil {
		return fmt.Errorf("seed admin: %w", err)
	}
	return nil
}

var (
	errInvalidCredentials = errors.New("invalid credentials")
	errUserExists         = errors.New("user already exists")
)

type rowQueryer interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

func authenticate(ctx context.Context, pool *pgxpool.Pool, username, password string) (*User, error) {
	var user User
	var hash string
	err := pool.QueryRow(ctx, `
		SELECT id, username, email, password_hash, created_at, updated_at
		FROM users WHERE username = $1`, username).Scan(&user.ID, &user.Username, &user.Email, &hash, &user.CreatedAt, &user.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errInvalidCredentials
	}
	if err != nil {
		return nil, err
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) != nil {
		return nil, errInvalidCredentials
	}
	return &user, nil
}

func register(ctx context.Context, pool *pgxpool.Pool, username, email, password string) (*User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}
	existing, err := getUser(ctx, pool, username, email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errUserExists
	}

	var user User
	err = pool.QueryRow(ctx, `
		INSERT INTO users (username, email, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id, username, email, created_at, updated_at`, username, email, string(hash)).
		Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("register user: %w", err)
	}
	return &user, nil
}

func getUser(ctx context.Context, pool *pgxpool.Pool, username string, email string) (*User, error) {
	var user User
	err := pool.QueryRow(ctx, `SELECT id, username, email, created_at, updated_at FROM users WHERE username = $1 OR email = $2`, username, email).Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt, &user.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	return &user, nil
}

func getUserByID(ctx context.Context, q rowQueryer, id int64) (*User, error) {
	var user User
	err := q.QueryRow(ctx, `SELECT id, username, email, created_at, updated_at FROM users WHERE id = $1`, id).Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

type User struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
