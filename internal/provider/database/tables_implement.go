package database

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
)

type implTableUsers struct{}

type implTableServices struct{}

type implTableUsersRoles struct{}

func (s implTableUsers) Add(ctx context.Context, database Querier, user *AddUser) (*User, error) {
	if database == nil {
		return nil, ErrDBNotInitilized
	}

	if user == nil {
		return nil, ErrNilArgument
	}

	query := `
INSERT INTO "users"
  ("username",
  "password")
VALUES
  ($1, $2)
RETURNING
  "username",
  "password",
  "id"
	`

	var dst User

	queryResult := database.QueryRow(ctx, query, user.Username, user.Password)
	err := queryResult.Scan(&dst.Username, &dst.Password, &dst.ID)

	if err != nil && strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
		return nil, ErrUniqueKeyViolation
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNoRows
	}

	if err != nil {
		return nil, fmt.Errorf("TableUsers.Add failed on INSERT: %w", err)
	}

	return &dst, nil
}

func (s implTableUsers) UpdateByUsername(ctx context.Context, database Querier, user *AddUser, username string) error {
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
	if err != nil && strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
		return ErrUniqueKeyViolation
	}

	if err != nil {
		return fmt.Errorf("TableUsers.Update failed on UPDATE: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrNoRows
	}

	return nil
}

func (s implTableUsers) GetByUsername(ctx context.Context, database Querier, username string) (*User, error) {
	if database == nil {
		return nil, ErrDBNotInitilized
	}

	query := `
SELECT
  "username",
  "password",
  "id"
FROM "users"
WHERE "username" = $1
	`

	var dst User

	queryResult := database.QueryRow(ctx, query, username)
	err := queryResult.Scan(&dst.Username, &dst.Password, &dst.ID)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNoRows
	}

	if err != nil {
		return nil, fmt.Errorf("TableUsers.GetByUsername failed on SELECT: %w", err)
	}

	return &dst, nil
}

func (s implTableUsers) DeleteByUsername(ctx context.Context, database Querier, username string) error {
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

func (s implTableServices) Add(ctx context.Context, database Querier, service *Service) error {
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
	if err != nil && strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
		return ErrUniqueKeyViolation
	}

	if err != nil {
		return fmt.Errorf("TableServices.Add failed on INSERT: %w", err)
	}

	return nil
}

func (s implTableServices) Update(ctx context.Context, database Querier, service *Service,
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
	if err != nil && strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
		return ErrUniqueKeyViolation
	}

	if err != nil {
		return fmt.Errorf("TableServices.Update failed on UPDATE: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrNoRows
	}

	return nil
}

func (s implTableServices) GetByServiceName(ctx context.Context, database Querier,
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

func (s implTableServices) Delete(ctx context.Context, database Querier,
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
func (s implTableUsersRoles) Add(ctx context.Context, database Querier,
	userRole *AddUserRole) (*UserRole, error,
) {
	if database == nil {
		return nil, ErrDBNotInitilized
	}

	if userRole == nil {
		return nil, ErrNilArgument
	}

	query := `
INSERT INTO "users_roles"
  ("user_id",
  "user_role",
  "service_name")
VALUES
  ($1, $2, $3)
RETURNING
  "id",
  "user_id",
  "user_role",
  "service_name",
  "created_ts"
	`

	var dst UserRole

	queryResult := database.QueryRow(ctx, query, userRole.UserID, userRole.UserRole, userRole.ServiceName)
	err := queryResult.Scan(&dst.ID, &dst.UserID, &dst.UserRole, &dst.ServiceName, &dst.CreatedTS)

	if err != nil && strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
		return nil, ErrUniqueKeyViolation
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNoRows
	}

	if err != nil {
		return nil, fmt.Errorf("TableUsersGroups.Add failed on INSERT: %w", err)
	}

	return &dst, nil
}

// TODO: check if the user_role is adequate.
func (s implTableUsersRoles) Insert(ctx context.Context, database Querier, userRole *UserRole) error {
	if database == nil {
		return ErrDBNotInitilized
	}

	if userRole == nil {
		return ErrNilArgument
	}

	query := `
INSERT INTO "users_roles"
  ("id",
  "user_id",
  "user_role",
  "service_name",
  "created_ts")
VALUES
  ($1, $2, $3, $4, $5)
	`

	_, err := database.Exec(ctx, query, userRole.ID, userRole.UserID, userRole.UserRole,
		userRole.ServiceName, userRole.CreatedTS)
	if err != nil && strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
		return ErrUniqueKeyViolation
	}

	if err != nil {
		return fmt.Errorf("TableUsersGroups.Insert failed on INSERT: %w", err)
	}

	return nil
}

// TODO: check if the user_role is adequate.
func (s implTableUsersRoles) UpdateByID(ctx context.Context, database Querier, userRole *UserRole,
	dbEntryID uint) error {

	if database == nil {
		return ErrDBNotInitilized
	}

	if userRole == nil {
		return ErrNilArgument
	}

	query := `
UPDATE "users_roles"
SET
  "id" = $1,
  "user_id" = $2,
  "user_role" = $3,
  "service_name" = $4,
  "created_ts" = $5
WHERE "id" = $6
	`

	result, err := database.Exec(ctx, query, userRole.ID, userRole.UserID, userRole.UserRole,
		userRole.ServiceName, userRole.CreatedTS, dbEntryID)
	if err != nil && strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
		return ErrUniqueKeyViolation
	}

	if err != nil {
		return fmt.Errorf("TableUsersGroups.UpdateByID failed on UPDATE: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrNoRows
	}

	return nil
}

func (s implTableUsersRoles) GetByUserID(ctx context.Context, database Querier,
	userID string) ([]UserRole, error,
) {
	if database == nil {
		return nil, ErrDBNotInitilized
	}

	query := `
SELECT
  "id",
  "user_id",
  "user_role",
  "service_name",
  "created_ts"
FROM "users_roles"
WHERE "user_id" = $1
	`

	var (
		dst []UserRole
		err error
	)

	queryResult, err := database.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("TableUsersGroups.GetByUserID failed on SELECT: %w", err)
	}

	dst, err = pgx.CollectRows(queryResult, func(row pgx.CollectableRow) (UserRole, error) {
		var nextDst UserRole
		err = row.Scan(&nextDst.ID, &nextDst.UserID, &nextDst.UserRole, &nextDst.ServiceName, &nextDst.CreatedTS)

		return nextDst, err //nolint:wrapcheck // not an actual return
	})
	if err != nil {
		return nil, fmt.Errorf("TableUsersGroups.GetByUserID failed on Scan: %w", err)
	}

	return dst, nil
}

func (s implTableUsersRoles) GetByID(ctx context.Context, database Querier, dbEntryID uint) (*UserRole, error) {
	if database == nil {
		return nil, ErrDBNotInitilized
	}

	query := `
SELECT
  "id",
  "user_id",
  "user_role",
  "service_name",
  "created_ts"
FROM "users_roles"
WHERE "id" = $1
	`

	var dst UserRole

	queryResult := database.QueryRow(ctx, query, dbEntryID)
	err := queryResult.Scan(&dst.ID, &dst.UserID, &dst.UserRole, &dst.ServiceName, &dst.CreatedTS)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNoRows
	}

	if err != nil {
		return nil, fmt.Errorf("TableServices.GetByID failed on SELECT: %w", err)
	}

	return &dst, nil
}

func (s implTableUsersRoles) DeleteByID(ctx context.Context, database Querier, dbEntryID uint) error {
	if database == nil {
		return ErrDBNotInitilized
	}

	query := `
DELETE FROM "users_roles"
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
