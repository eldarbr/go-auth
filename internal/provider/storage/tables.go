package storage

import (
	"context"
	"time"

	"github.com/eldarbr/go-auth/pkg/database"
)

//nolint:gochecknoinits // Set default implementations.
func init() {
	TableUsers = implTableUsers{}
	TableServices = implTableServices{}
	TableUsersRoles = implTableUsersRoles{}
}

type UserRoleType = string

const (
	UserRoleTypeRoot  UserRoleType = "root"
	UserRoleTypeAdmin UserRoleType = "admin"
	UserRoleTypeUser  UserRoleType = "user"
)

type AddUser struct {
	Username string
	Password string
}

type User struct {
	AddUser
	ID string
}

type Service struct {
	Name string
}

type AddUserRole struct {
	UserID      string
	UserRole    UserRoleType
	ServiceName string
}

type UserRole struct {
	CreatedTS time.Time
	AddUserRole
	ID uint
}

type GroupUser struct {
	GroupName string
	Username  string
}

var TableUsers interface {
	Add(ctx context.Context, database database.Querier, user *AddUser) (*User, error)
	UpdateByUsername(ctx context.Context, database database.Querier, user *AddUser, username string) error
	GetByUsername(ctx context.Context, database database.Querier, username string) (*User, error)
	DeleteByUsername(ctx context.Context, database database.Querier, username string) error
}

var TableServices interface {
	Add(ctx context.Context, database database.Querier, service *Service) error
	Update(ctx context.Context, database database.Querier, service *Service, serviceName string) error
	GetByServiceName(ctx context.Context, database database.Querier, serviceName string) (*Service, error)
	Delete(ctx context.Context, database database.Querier, serviceName string) error
}

var TableUsersRoles interface {
	Add(ctx context.Context, database database.Querier, useRole *AddUserRole) (*UserRole, error)
	Insert(ctx context.Context, database database.Querier, useRole *UserRole) error
	UpdateByID(ctx context.Context, database database.Querier, useRole *UserRole, dbEntryID uint) error
	GetByUserID(ctx context.Context, database database.Querier, userID string) ([]UserRole, error)
	GetByID(ctx context.Context, database database.Querier, dbEntryID uint) (*UserRole, error)
	DeleteByID(ctx context.Context, database database.Querier, dbEntryID uint) error
}
