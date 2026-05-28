package main

import (
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"

	"home-cooking.timothymarias.com/internal/config"
	"home-cooking.timothymarias.com/internal/middleware"
	"home-cooking.timothymarias.com/internal/router"
	"home-cooking.timothymarias.com/internal/store"
)

func main() {
	migrateUp := flag.Bool("migrate-up", false, "Run all up migrations and exit")
	migrateDown := flag.Bool("migrate-down", false, "Roll back last migration and exit")
	flag.Parse()

	_ = godotenv.Load()

	cfg := config.Load()

	db, err := store.Connect(cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if *migrateUp || *migrateDown {
		if err := runMigration(db, *migrateUp); err != nil {
			slog.Error("migration failed", "error", err)
			os.Exit(1)
		}
		return
	}

	if err := runMigration(db, true); err != nil {
		slog.Error("auto-migration failed", "error", err)
		os.Exit(1)
	}

	mux := router.New(db)

	var h http.Handler = mux
	h = middleware.Logging(h)
	h = middleware.Recovery(h)

	addr := ":" + cfg.Port
	slog.Info("starting server", "addr", addr)
	if err := http.ListenAndServe(addr, h); err != nil {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}

func runMigration(db *store.DB, up bool) error {
	driver, err := postgres.WithInstance(db.SqlDB(), &postgres.Config{})
	if err != nil {
		return fmt.Errorf("creating migration driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance("file://migrations", "postgres", driver)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			slog.Info("no migrations found, skipping")
			return nil
		}
		return fmt.Errorf("creating migration instance: %w", err)
	}

	if up {
		err = m.Up()
	} else {
		err = m.Steps(-1)
	}

	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("running migration: %w", err)
	}

	v, dirty, _ := m.Version()
	slog.Info("migrations complete", "version", v, "dirty", dirty)
	return nil
}
