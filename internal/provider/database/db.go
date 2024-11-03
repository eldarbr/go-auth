package database

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

var dbPool *pgxpool.Pool

func Setup(ctx context.Context, uri string) (err error) {
	dbPool, err = pgxpool.New(ctx, uri)
	if err != nil {
		return err
	}
	if err = dbPool.Ping(ctx); err != nil {
		return err
	}
	return nil
}

func Pool() *pgxpool.Pool {
	return dbPool
}
