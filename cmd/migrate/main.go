package main

import (
	"context"
	"log"

	"jozery/internal/app"
	"jozery/lib/db/pgsql"
)

func main() {
	ctx := context.Background()
	pool, err := pgsql.Open(ctx)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer pool.Close()

	if err := app.Migrate(ctx, pool); err != nil {
		log.Fatalf("migrate: %v", err)
	}
	log.Print("migrations applied")
}
