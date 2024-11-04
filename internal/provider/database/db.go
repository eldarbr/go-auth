package database

import (
	"context"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/pgx/v5"

	// File driver import for .sql migrations.
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
)

var dbPool *pgxpool.Pool

func Setup(ctx context.Context, uri string) error {
	var err error

	// Connect to the db.
	dbPool, err = pgxpool.New(ctx, uri)
	if err != nil {
		return fmt.Errorf("database.Setup Connection failed: %w", err)
	}

	if err = dbPool.Ping(ctx); err != nil {
		return fmt.Errorf("database.Setup Ping failed: %w", err)
	}

	// Migrate the db.
	err = migrateAuth(uri)
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("database.Setup Migration failed: %w", err)
	}

	return nil
}

func migrateAuth(uri string) error {
	var err error

	postgres := &pgx.Postgres{}

	dbConn, err := postgres.Open(uri)
	if err != nil {
		return fmt.Errorf("database.migrateAuth Connection failed: %w", err)
	}
	defer dbConn.Close()

	migration, err := migrate.NewWithDatabaseInstance("file://./internal/provider/database/sql", "pgx", dbConn)
	if err != nil {
		return fmt.Errorf("database.migrateAuth Migration failed: %w", err)
	}

	err = migration.Up()
	if err != nil {
		return fmt.Errorf("database.migrateAuth Migration.Up failed: %w", err)
	}

	return nil
}

func GetPool() *pgxpool.Pool {
	return dbPool
}
