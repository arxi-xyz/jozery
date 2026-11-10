package app

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"jozery/internal/domains/authentication"
	"jozery/internal/domains/identity"
)

type migration struct {
	name string
	run  func(context.Context, *pgxpool.Pool) error
}

func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	migrations := []migration{
		{name: "identity", run: identity.Migrate},
		{name: "authentication", run: authentication.Migrate},
	}
	for _, step := range migrations {
		if err := step.run(ctx, pool); err != nil {
			return fmt.Errorf("%s migrations: %w", step.name, err)
		}
	}
	return nil
}
