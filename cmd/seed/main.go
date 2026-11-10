package main

import (
	"context"
	"log"
	"os"

	"jozery/internal/domains/identity/db"
	"jozery/internal/domains/identity/usecases"
	"jozery/lib/db/pgsql"
)

func main() {
	password := os.Getenv("ADMIN_PASSWORD")
	if password == "" {
		log.Print("ADMIN_PASSWORD is unset, seed skipped")
		return
	}

	ctx := context.Background()
	pool, err := pgsql.Open(ctx)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer pool.Close()

	created, err := usecases.SeedAdmin(ctx, db.New(pool), "admin", "admin@jozery.local", password)
	if err != nil {
		log.Fatalf("seed admin: %v", err)
	}
	if created {
		log.Print("admin seeded")
		return
	}
	log.Print("admin already exists")
}
