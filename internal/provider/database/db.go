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

// Database object should be passed from the owner to users via reference, so after the owner closes
// the pool, the users might see the change.
type Database struct {
	dbPool *pgxpool.Pool
}

// Setup prepares the connection pool and runs MigrateAuthUp.
func Setup(ctx context.Context, dbURI, migrationsPath string) (*Database, error) {
	var err error

	// Connect to the db.
	dbPool, err := pgxpool.New(ctx, dbURI)
	if err != nil {
		return nil, fmt.Errorf("database.Setup Connection failed: %w", err)
	}

	if err = dbPool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("database.Setup Ping failed: %w", err)
	}

	dbInstancePtr := &Database{
		dbPool: dbPool,
	}

	// Migrate the db.
	err = MigrateAuthUp(dbURI, migrationsPath)
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		dbInstancePtr.ClosePool()

		return nil, fmt.Errorf("database.Setup Migration failed: %w", err)
	}

	return dbInstancePtr, nil
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

func (db *Database) GetPool() *pgxpool.Pool {
	if db == nil {
		return nil
	}

	return db.dbPool
}

func (db *Database) ClosePool() {
	if db == nil {
		return
	}

	if db.dbPool != nil {
		db.dbPool.Close()
		db.dbPool = nil
	}
}
