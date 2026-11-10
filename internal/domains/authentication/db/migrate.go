package db

import (
	"context"
	"embed"

	"github.com/jackc/pgx/v5/pgxpool"
	"jozery/lib/db/pgsql"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	return pgsql.Up(ctx, pool, migrationFiles, "migrations")
}
