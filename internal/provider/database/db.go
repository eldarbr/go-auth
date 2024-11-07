package database

import (
	"context"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file" // File driver import for .sql migrations.
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrAlreadyInitialized = errors.New("the database is initialized already")
	ErrDBNotInitilized    = errors.New("database was not initialized")
	ErrNilArgument        = errors.New("nil argument received")
	ErrNoRows             = errors.New("query has returned no rows")
)

type Database struct {
	dbPool *pgxpool.Pool
}

func Setup(ctx context.Context, dbURI, migrationsPath string) (Database, error) {
	var dbInstance Database
	return dbInstance, setup(ctx, dbURI, migrationsPath, &dbInstance)
}

// Setup prepares the connection pool and runs MigrateAuthUp.
func setup(ctx context.Context, dbURI, migrationsPath string, dbInstancePtr *Database) error {
	var err error

	if dbInstancePtr == nil {
		return ErrNilArgument
	}

	// Connect to the db.
	dbPool, err := pgxpool.New(ctx, dbURI)
	if err != nil {
		return fmt.Errorf("database.Setup Connection failed: %w", err)
	}

	if err = dbPool.Ping(ctx); err != nil {
		return fmt.Errorf("database.Setup Ping failed: %w", err)
	}

	dbInstancePtr.dbPool = dbPool

	// Migrate the db.
	err = MigrateAuthUp(dbURI, migrationsPath)
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		dbInstancePtr.ClosePool()

		return fmt.Errorf("database.Setup Migration failed: %w", err)
	}

	return nil
}

func MigrateAuthUp(dbURI, migrationsPath string) error {
	var err error

	postgres := &pgx.Postgres{}

	dbConn, err := postgres.Open(dbURI)
	if err != nil {
		return fmt.Errorf("database.migrateAuth Connection failed: %w", err)
	}
	defer dbConn.Close()

	migration, err := migrate.NewWithDatabaseInstance(migrationsPath, "pgx", dbConn)
	if err != nil {
		return fmt.Errorf("database.migrateAuth Migration failed: %w", err)
	}

	err = migration.Up()
	if err != nil {
		return fmt.Errorf("database.migrateAuth Migration.Up failed: %w", err)
	}

	return nil
}

func (db Database) GetPool() *pgxpool.Pool {
	return db.dbPool
}

func (db *Database) ClosePool() {
	if db.dbPool != nil {
		db.dbPool.Close()
		db.dbPool = nil
	}
}
