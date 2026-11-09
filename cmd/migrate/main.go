package main

import (
	"context"
	"log"

	"jozery/internal/authentication"
	"jozery/internal/identity"
	"jozery/lib/db/pgsql"
)

func main() {
	ctx := context.Background()
	pool, err := pgsql.Open(ctx)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer pool.Close()

	if err := identity.Migrate(ctx, pool); err != nil {
		log.Fatalf("identity migrations: %v", err)
	}
	if err := authentication.Migrate(ctx, pool); err != nil {
		log.Fatalf("authentication migrations: %v", err)
	}
	log.Print("migrations applied")
}
