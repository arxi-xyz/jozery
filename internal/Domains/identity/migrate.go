package identity

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"jozery/internal/Domains/identity/db"
)

func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	return db.Migrate(ctx, pool)
}
