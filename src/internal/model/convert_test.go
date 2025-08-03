package model_test

import (
	"flag"
	"testing"

	"github.com/eldarbr/go-auth/internal/model"
	"github.com/eldarbr/go-auth/internal/provider/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var _ = flag.String("t-db-uri", "", "perform sql tests on the `t-db-uri` database")

func TestPrepareClaimsEmpty(t *testing.T) {
	t.Parallel()

	dbRoles := []storage.UserRole{}

	converted := model.PrepareClaims(dbRoles)

	assert.Empty(t, converted)
}

func TestPrepareClaimsOneRole(t *testing.T) {
	t.Parallel()

	dbRoles := []storage.UserRole{
		{AddUserRole: storage.AddUserRole{UserID: "123", ServiceName: "service", UserRole: storage.UserRoleTypeUser}}, //nolint:exhaustruct,lll // other fields are not used.
	}

	converted := model.PrepareClaims(dbRoles)

	require.Len(t, converted, 1)
	assert.Equal(t, "service", converted[0].ServiceName)
	assert.Equal(t, storage.UserRoleTypeUser, converted[0].UserRole)
}

func TestPrepareClaimsManyRoles(t *testing.T) {
	t.Parallel()

	dbRoles := []storage.UserRole{
		{AddUserRole: storage.AddUserRole{UserID: "123", ServiceName: "service1", UserRole: storage.UserRoleTypeUser}},  //nolint:exhaustruct,lll // other fields are not used.
		{AddUserRole: storage.AddUserRole{UserID: "123", ServiceName: "service2", UserRole: storage.UserRoleTypeAdmin}}, //nolint:exhaustruct,lll // other fields are not used.
		{AddUserRole: storage.AddUserRole{UserID: "123", ServiceName: "service3", UserRole: storage.UserRoleTypeRoot}},  //nolint:exhaustruct,lll // other fields are not used.
		{AddUserRole: storage.AddUserRole{UserID: "123", ServiceName: "service4", UserRole: storage.UserRoleTypeUser}},  //nolint:exhaustruct,lll // other fields are not used.
	}

	converted := model.PrepareClaims(dbRoles)

	require.Len(t, converted, 4)
	assert.Equal(t, "service1", converted[0].ServiceName)
	assert.Equal(t, "service2", converted[1].ServiceName)
	assert.Equal(t, "service3", converted[2].ServiceName)
	assert.Equal(t, "service4", converted[3].ServiceName)
	assert.Equal(t, storage.UserRoleTypeUser, converted[0].UserRole)
	assert.Equal(t, storage.UserRoleTypeAdmin, converted[1].UserRole)
	assert.Equal(t, storage.UserRoleTypeRoot, converted[2].UserRole)
	assert.Equal(t, storage.UserRoleTypeUser, converted[3].UserRole)
}
