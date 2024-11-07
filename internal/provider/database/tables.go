package database

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

//nolint:gochecknoinits // Set default implementations.
func init() {
	TableUsers = implTableUsers{}
	TableServices = implTableServices{}
	TableUsersGroups = implTableUsersGroups{}
}

type UserRoleType = string

const (
	UserRoleTypeRoot  UserRoleType = "root"
	UserRoleTypeAdmin UserRoleType = "admin"
	UserRoleTypeUser  UserRoleType = "user"
)

type User struct {
	Username string
	Password string
}

type Service struct {
	Name string
}

type AddUserGroup struct {
	Username    string
	UserRole    UserRoleType
	ServiceName string
}

type UserGroup struct {
	ID        uint
	CreatedTS time.Time

	AddUserGroup
}

var TableUsers interface {
	Add(ctx context.Context, database *pgxpool.Pool, user *User) error
	Update(ctx context.Context, database *pgxpool.Pool, user *User, username string) error
	GetByUsername(ctx context.Context, database *pgxpool.Pool, username string) (*User, error)
	Delete(ctx context.Context, database *pgxpool.Pool, username string) error
}

var TableServices interface {
	Add(ctx context.Context, database *pgxpool.Pool, service *Service) error
	Update(ctx context.Context, database *pgxpool.Pool, service *Service, serviceName string) error
	GetByServiceName(ctx context.Context, database *pgxpool.Pool, serviceName string) (*Service, error)
	Delete(ctx context.Context, database *pgxpool.Pool, serviceName string) error
}

var TableUsersGroups interface {
	Add(ctx context.Context, database *pgxpool.Pool, userGroup *AddUserGroup) (*UserGroup, error)
	Insert(ctx context.Context, database *pgxpool.Pool, userGroup *UserGroup) error
	UpdateByID(ctx context.Context, database *pgxpool.Pool, userGroup *UserGroup, dbEntryID uint) error
	GetByUsername(ctx context.Context, database *pgxpool.Pool, username string) ([]UserGroup, error)
	GetByID(ctx context.Context, database *pgxpool.Pool, dbEntryID uint) (*UserGroup, error)
	DeleteByID(ctx context.Context, database *pgxpool.Pool, dbEntryID uint) error
}
