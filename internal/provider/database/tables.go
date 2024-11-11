package database

import (
	"context"
	"time"
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

type User struct {
	Username string
	Password string
}

type Service struct {
	Name string
}

type AddUserRole struct {
	Username    string
	UserRole    UserRoleType
	ServiceName string
}

type UserRole struct {
	ID        uint
	CreatedTS time.Time

	AddUserRole
}

type GroupUser struct {
	GroupName string
	Username  string
}

var TableUsers interface {
	Add(ctx context.Context, database Querier, user *User) error
	Update(ctx context.Context, database Querier, user *User, username string) error
	GetByUsername(ctx context.Context, database Querier, username string) (*User, error)
	Delete(ctx context.Context, database Querier, username string) error
}

var TableServices interface {
	Add(ctx context.Context, database Querier, service *Service) error
	Update(ctx context.Context, database Querier, service *Service, serviceName string) error
	GetByServiceName(ctx context.Context, database Querier, serviceName string) (*Service, error)
	Delete(ctx context.Context, database Querier, serviceName string) error
}

var TableUsersRoles interface {
	Add(ctx context.Context, database Querier, useRole *AddUserRole) (*UserRole, error)
	Insert(ctx context.Context, database Querier, useRole *UserRole) error
	UpdateByID(ctx context.Context, database Querier, useRole *UserRole, dbEntryID uint) error
	GetByUsername(ctx context.Context, database Querier, username string) ([]UserRole, error)
	GetByID(ctx context.Context, database Querier, dbEntryID uint) (*UserRole, error)
	DeleteByID(ctx context.Context, database Querier, dbEntryID uint) error
}
