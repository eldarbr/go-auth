package database

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type implTableUsers struct{}

type implTableServices struct{}

type implTableUsersGroups struct{}

func (s implTableUsers) Add(ctx context.Context, database *pgxpool.Pool, user *User) error {
	if database == nil {
		return ErrDBNotInitilized
	}

	if user == nil {
		return ErrNilArgument
	}

	query := `
INSERT INTO "users"
  ("username",
  "password")
VALUES
  ($1, $2)
	`

	_, err := database.Exec(ctx, query, user.Username, user.Password)
	if err != nil {
		return fmt.Errorf("TableUsers.Add failed on INSERT INTO: %w", err)
	}

	return nil
}

func (s implTableUsers) Update(ctx context.Context, database *pgxpool.Pool, user *User, username string) error {
	if database == nil {
		return ErrDBNotInitilized
	}

	if user == nil {
		return ErrNilArgument
	}

	query := `
UPDATE "users"
SET
  "username" = $1,
  "password" = $2
WHERE "username" = $3
	`

	result, err := database.Exec(ctx, query, user.Username, user.Password, username)
	if err != nil {
		return fmt.Errorf("TableUsers.Update failed on UPDATE: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrNoRows
	}

	return nil
}

func (s implTableUsers) GetByUsername(ctx context.Context, database *pgxpool.Pool, username string) (*User, error) {
	if database == nil {
		return nil, ErrDBNotInitilized
	}

	query := `
SELECT
  "username",
  "password"
FROM "users"
WHERE "username" = $1
	`

	var dst User

	queryResult := database.QueryRow(ctx, query, username)
	err := queryResult.Scan(&dst.Username, &dst.Password)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNoRows
	}

	if err != nil {
		return nil, fmt.Errorf("TableUsers.GetByUsername failed on SELECT: %w", err)
	}

	return &dst, nil
}

func (s implTableUsers) Delete(ctx context.Context, database *pgxpool.Pool, username string) error {
	if database == nil {
		return ErrDBNotInitilized
	}

	query := `
DELETE FROM "users"
WHERE "username" = $1
	`

	result, err := database.Exec(ctx, query, username)
	if err != nil {
		return fmt.Errorf("TableUsers.Delete failed on DELETE: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrNoRows
	}

	return nil
}

func (s implTableServices) Add(ctx context.Context, database *pgxpool.Pool, service *Service) error {
	if database == nil {
		return ErrDBNotInitilized
	}

	if service == nil {
		return ErrNilArgument
	}

	query := `
INSERT INTO "services"
  ("name")
VALUES
	($1)
	`

	_, err := database.Exec(ctx, query, service.Name)
	if err != nil {
		return fmt.Errorf("TableServices.Add failed on INSERT: %w", err)
	}

	return nil
}

func (s implTableServices) Update(ctx context.Context, database *pgxpool.Pool, service *Service,
	serviceName string) error {

	if database == nil {
		return ErrDBNotInitilized
	}

	if service == nil {
		return ErrNilArgument
	}

	query := `
UPDATE "services"
SET
  "name" = $1
WHERE "name" = $2
	`

	result, err := database.Exec(ctx, query, service.Name, serviceName)
	if err != nil {
		return fmt.Errorf("TableServices.Update failed on UPDATE: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrNoRows
	}

	return nil
}

func (s implTableServices) GetByServiceName(ctx context.Context, database *pgxpool.Pool,
	serviceName string) (*Service, error,
) {
	if database == nil {
		return nil, ErrDBNotInitilized
	}

	query := `
SELECT
  "name"
FROM "services"
WHERE "name" = $1
	`

	var dst Service

	queryResult := database.QueryRow(ctx, query, serviceName)
	err := queryResult.Scan(&dst.Name)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNoRows
	}

	if err != nil {
		return nil, fmt.Errorf("TableServices.GetByServiceName failed on SELECT: %w", err)
	}

	return &dst, nil
}

func (s implTableServices) Delete(ctx context.Context, database *pgxpool.Pool,
	serviceName string) error {

	if database == nil {
		return ErrDBNotInitilized
	}

	query := `
DELETE FROM "services"
WHERE "name" = $1
	`

	result, err := database.Exec(ctx, query, serviceName)
	if err != nil {
		return fmt.Errorf("TableServices.Delete failed on DELETE: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrNoRows
	}

	return nil
}

// TODO: check if the user_role is adequate.
func (s implTableUsersGroups) Add(ctx context.Context, database *pgxpool.Pool,
	userGroup *AddUserGroup) (*UserGroup, error,
) {
	if database == nil {
		return nil, ErrDBNotInitilized
	}

	if userGroup == nil {
		return nil, ErrNilArgument
	}

	query := `
INSERT INTO "users_groups"
  ("username",
  "user_role",
  "service_name")
VALUES
  ($1, $2, $3)
RETURNING
  "id",
  "username",
  "user_role",
  "service_name",
  "created_ts"
	`

	var dst UserGroup

	queryResult := database.QueryRow(ctx, query, userGroup.Username, userGroup.UserRole, userGroup.ServiceName)
	err := queryResult.Scan(&dst.ID, &dst.Username, &dst.UserRole, &dst.ServiceName, &dst.CreatedTS)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNoRows
	}

	if err != nil {
		return nil, fmt.Errorf("TableUsersGroups.Add failed on INSERT: %w", err)
	}

	return &dst, nil
}

// TODO: check if the user_role is adequate.
func (s implTableUsersGroups) Insert(ctx context.Context, database *pgxpool.Pool, userGroup *UserGroup) error {
	if database == nil {
		return ErrDBNotInitilized
	}

	if userGroup == nil {
		return ErrNilArgument
	}

	query := `
INSERT INTO "users_groups"
  ("id",
  "username",
  "user_role",
  "service_name",
  "created_ts")
VALUES
	($1, $2, $3, $4, $5)
	`

	_, err := database.Exec(ctx, query, userGroup.ID, userGroup.Username, userGroup.UserRole,
		userGroup.ServiceName, userGroup.CreatedTS)
	if err != nil {
		return fmt.Errorf("TableUsersGroups.Insert failed on INSERT: %w", err)
	}

	return nil
}

// TODO: check if the user_role is adequate.
func (s implTableUsersGroups) UpdateByID(ctx context.Context, database *pgxpool.Pool, userGroup *UserGroup,
	dbEntryID uint) error {

	if database == nil {
		return ErrDBNotInitilized
	}

	if userGroup == nil {
		return ErrNilArgument
	}

	query := `
UPDATE "users_groups"
SET
  "id" = $1,
  "username" = $2,
  "user_role" = $3,
  "service_name" = $4,
  "created_ts" = $5
WHERE "id" = $6
	`

	result, err := database.Exec(ctx, query, userGroup.ID, userGroup.Username, userGroup.UserRole,
		userGroup.ServiceName, userGroup.CreatedTS, dbEntryID)
	if err != nil {
		return fmt.Errorf("TableUsersGroups.UpdateByID failed on UPDATE: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrNoRows
	}

	return nil
}

func (s implTableUsersGroups) GetByUsername(ctx context.Context, database *pgxpool.Pool,
	username string) ([]UserGroup, error,
) {
	if database == nil {
		return nil, ErrDBNotInitilized
	}

	query := `
SELECT
  "id",
  "username",
  "user_role",
  "service_name",
  "created_ts"
FROM "users_groups"
WHERE "username" = $1
	`

	var (
		dst []UserGroup
		err error
	)

	queryResult, err := database.Query(ctx, query, username)
	if err != nil {
		return nil, fmt.Errorf("TableUsersGroups.GetByUsername failed on SELECT: %w", err)
	}

	dst, err = pgx.CollectRows(queryResult, func(row pgx.CollectableRow) (UserGroup, error) {
		var nextDst UserGroup
		err = row.Scan(&nextDst.ID, &nextDst.Username, &nextDst.UserRole, &nextDst.ServiceName, &nextDst.CreatedTS)

		return nextDst, err //nolint:wrapcheck // not an actual return
	})
	if err != nil {
		return nil, fmt.Errorf("TableUsersGroups.GetByUsername failed on Scan: %w", err)
	}

	return dst, nil
}

func (s implTableUsersGroups) GetByID(ctx context.Context, database *pgxpool.Pool, dbEntryID uint) (*UserGroup, error) {
	if database == nil {
		return nil, ErrDBNotInitilized
	}

	query := `
SELECT
  "id",
  "username",
  "user_role",
  "service_name",
  "created_ts"
FROM "users_groups"
WHERE "id" = $1
		`

	var dst UserGroup

	queryResult := database.QueryRow(ctx, query, dbEntryID)
	err := queryResult.Scan(&dst.ID, &dst.Username, &dst.UserRole, &dst.ServiceName, &dst.CreatedTS)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNoRows
	}

	if err != nil {
		return nil, fmt.Errorf("TableServices.GetByID failed on SELECT: %w", err)
	}

	return &dst, nil
}

func (s implTableUsersGroups) DeleteByID(ctx context.Context, database *pgxpool.Pool, dbEntryID uint) error {
	if database == nil {
		return ErrDBNotInitilized
	}

	query := `
DELETE FROM "users_groups"
WHERE "id" = $1
		`

	result, err := database.Exec(ctx, query, dbEntryID)
	if err != nil {
		return fmt.Errorf("TableUsersGroups.DeleteByID failed on DELETE: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrNoRows
	}

	return nil
}
