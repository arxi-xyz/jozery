package pgsql

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Open(ctx context.Context) (*pgxpool.Pool, error) {
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
	return pool, nil
}

func databaseURL() string {
	if url := os.Getenv("DATABASE_URL"); url != "" {
		return url
	}
	return "postgres://jozery:jozery@localhost:5432/jozery?sslmode=disable"
}
