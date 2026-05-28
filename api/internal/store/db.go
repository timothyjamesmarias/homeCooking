package store

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jmoiron/sqlx"
	_ "github.com/jackc/pgx/v5/stdlib"
)

var ErrNotFound = errors.New("not found")

type DB struct {
	*sqlx.DB
}

func (d *DB) SqlDB() *sql.DB {
	return d.DB.DB
}

func Connect(databaseURL string) (*DB, error) {
	db, err := sqlx.Connect("pgx", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("connecting to database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	slog.Info("connected to database")
	return &DB{db}, nil
}
