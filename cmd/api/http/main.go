package main

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/labstack/echo/v5"
	"jozery/internal/app"
	"jozery/lib/db/pgsql"
)

func main() {
	db, err := pgsql.Open(context.Background())
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer db.Close()

	e := echo.New()
	app.Mount(e, db)
	if err := e.Start(":8080"); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}
