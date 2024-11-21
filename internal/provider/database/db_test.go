package database_test

import (
	"context"
	"flag"
	"testing"
	"time"

	"github.com/eldarbr/go-auth/internal/provider/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testDBUri = flag.String("t-db-uri", "", "perform sql tests on the `t-db-uri` database")

var testDB *database.Database

func TestMain(m *testing.M) {
	flag.Parse()

	if testDBUri != nil && *testDBUri != "" {
		// Not checking the error as if there is an error, the tests won't run.
		testDB, _ = database.Setup(context.Background(), *testDBUri, "file://./sql")

		defer testDB.ClosePool()
	}

	m.Run()
}

func checkDB(t *testing.T) {
	t.Helper()

	pool := testDB.GetPool()
	if pool == nil {
		t.Skip("database was not initialized")
	}
}

func TestNilArguments(t *testing.T) {
	t.Parallel() // Running all db tests in parallel.
	checkDB(t)

	_, err := database.TableUsers.Add(context.Background(), testDB.GetPool(), nil)

	require.ErrorIs(t, err, database.ErrNilArgument)

	require.ErrorIs(t, database.TableServices.Add(context.Background(), testDB.GetPool(), nil),
		database.ErrNilArgument)

	_, err = database.TableUsersRoles.Add(context.Background(), testDB.GetPool(), nil)

	require.ErrorIs(t, err, database.ErrNilArgument)
	require.ErrorIs(t, database.TableUsersRoles.Insert(context.Background(), testDB.GetPool(), nil),
		database.ErrNilArgument)
}

func TestNilDB(t *testing.T) {
	t.Parallel() // Running all db tests in parallel.
	checkDB(t)

	// nil is an unitialized db.

	_, err := database.TableUsers.Add(context.Background(), nil, nil)
	require.ErrorIs(t, err, database.ErrDBNotInitilized)

	err = database.TableUsers.UpdateByUsername(context.Background(), nil, nil, "")
	require.ErrorIs(t, err, database.ErrDBNotInitilized)

	_, err = database.TableUsers.GetByUsername(context.Background(), nil, "")
	require.ErrorIs(t, err, database.ErrDBNotInitilized)

	err = database.TableUsers.DeleteByUsername(context.Background(), nil, "")
	require.ErrorIs(t, err, database.ErrDBNotInitilized)

	err = database.TableServices.Add(context.Background(), nil, nil)
	require.ErrorIs(t, err, database.ErrDBNotInitilized)

	err = database.TableServices.Update(context.Background(), nil, nil, "")
	require.ErrorIs(t, err, database.ErrDBNotInitilized)

	_, err = database.TableServices.GetByServiceName(context.Background(), nil, "")
	require.ErrorIs(t, err, database.ErrDBNotInitilized)

	err = database.TableServices.Delete(context.Background(), nil, "")
	require.ErrorIs(t, err, database.ErrDBNotInitilized)

	err = database.TableUsersRoles.Insert(context.Background(), nil, nil)
	require.ErrorIs(t, err, database.ErrDBNotInitilized)

	_, err = database.TableUsersRoles.Add(context.Background(), nil, nil)
	require.ErrorIs(t, err, database.ErrDBNotInitilized)

	err = database.TableUsersRoles.UpdateByID(context.Background(), nil, nil, 0)
	require.ErrorIs(t, err, database.ErrDBNotInitilized)

	_, err = database.TableUsersRoles.GetByID(context.Background(), nil, 0)
	require.ErrorIs(t, err, database.ErrDBNotInitilized)

	_, err = database.TableUsersRoles.GetByUserID(context.Background(), nil, "")
	require.ErrorIs(t, err, database.ErrDBNotInitilized)

	err = database.TableUsersRoles.DeleteByID(context.Background(), nil, 0)
	require.ErrorIs(t, err, database.ErrDBNotInitilized)
}

func TestUsersValidAddAndGet(t *testing.T) {
	t.Parallel() // Running all db tests in parallel.
	checkDB(t)

	users := []database.AddUser{
		{
			Username: "username1",
			Password: "password1",
		},
		{
			Username: "username2",
			Password: "password2",
		},
		{
			Username: "username3",
			Password: "password3",
		},
	}

	for _, user := range users {
		added, err := database.TableUsers.Add(context.Background(), testDB.GetPool(), &user)
		require.NoError(t, err)
		assert.Equal(t, user.Username, added.Username)
		assert.Equal(t, user.Password, added.Password)
	}

	for _, user := range users {
		dbUser, err := database.TableUsers.GetByUsername(context.Background(), testDB.GetPool(), user.Username)
		require.NoError(t, err)
		require.NotNil(t, dbUser)
		assert.Equal(t, user.Username, dbUser.Username)
		assert.Equal(t, user.Password, dbUser.Password)
	}
}

func TestUsersValidUpdateAndDelete(t *testing.T) {
	t.Parallel() // Running all db tests in parallel.
	checkDB(t)

	users := []database.AddUser{
		{
			Username: "1username1",
			Password: "1password1",
		},
		{
			Username: "1username2",
			Password: "1password2",
		},
		{
			Username: "1username3",
			Password: "1password3",
		},
	}

	// Add.
	for _, user := range users {
		_, err := database.TableUsers.Add(context.Background(), testDB.GetPool(), &user)
		require.NoError(t, err)
	}

	newUsers := []database.AddUser{
		{
			Username: "1username1",        // Same as before.
			Password: "passsssssssssssss", // New password.
		},
		{
			Username: "uuuuu",      // New username.
			Password: "1password2", // Same as before.
		},
		{
			Username: "uuu",     // New username.
			Password: "ppppppp", // New password.
		},
	}

	// Update.
	for i := range users {
		err := database.TableUsers.UpdateByUsername(context.Background(), testDB.GetPool(), &newUsers[i], users[i].Username)
		require.NoError(t, err)
	}

	// Assert.
	for i := range users {
		dbUser, err := database.TableUsers.GetByUsername(context.Background(), testDB.GetPool(), newUsers[i].Username)
		require.NoError(t, err)
		require.NotNil(t, dbUser)
		assert.Equal(t, newUsers[i].Username, dbUser.Username)
		assert.Equal(t, newUsers[i].Password, dbUser.Password)
	}

	// Delete.
	for i := range newUsers {
		err := database.TableUsers.DeleteByUsername(context.Background(), testDB.GetPool(), newUsers[i].Username)
		require.NoError(t, err)
	}

	// Assert deletion.
	for i := range users {
		_, err := database.TableUsers.GetByUsername(context.Background(), testDB.GetPool(), newUsers[i].Username)
		require.ErrorIs(t, err, database.ErrNoRows)
	}
}

//nolint:dupl // It is indeed almost a duplicate.
func TestUsersInvalidGetUpdateDelete(t *testing.T) {
	t.Parallel() // Running all db tests in parallel.
	checkDB(t)

	usernames := []string{"123", "", "12345678901234567890", "true", "nameuser", "AAAAAAA", "bbbbbBBBBBB",
		"1234567890123456789012345678901234567890", "nothing to see here"}

	for _, username := range usernames {
		t.Run(username, func(t *testing.T) {
			t.Parallel()

			ret, err := database.TableUsers.GetByUsername(context.Background(), testDB.GetPool(), username)
			assert.Nil(t, ret)
			require.ErrorIs(t, database.ErrNoRows, err)

			var dummy database.AddUser
			err = database.TableUsers.UpdateByUsername(context.Background(), testDB.GetPool(), &dummy, username)
			require.ErrorIs(t, database.ErrNoRows, err)

			err = database.TableUsers.DeleteByUsername(context.Background(), testDB.GetPool(), username)
			require.ErrorIs(t, database.ErrNoRows, err)
		})
	}
}

func TestServicesValidAddAndGet(t *testing.T) {
	t.Parallel() // Running all db tests in parallel.
	checkDB(t)

	services := []database.Service{
		{
			Name: "service1",
		},
		{
			Name: "service2",
		},
		{
			Name: "service3",
		},
	}

	for _, service := range services {
		err := database.TableServices.Add(context.Background(), testDB.GetPool(), &service)
		require.NoError(t, err)
	}

	for _, service := range services {
		dbService, err := database.TableServices.GetByServiceName(context.Background(), testDB.GetPool(), service.Name)
		require.NoError(t, err)
		require.NotNil(t, dbService)
		assert.Equal(t, service.Name, dbService.Name)
	}
}

func TestServicesValidUpdateAndDelete(t *testing.T) {
	t.Parallel() // Running all db tests in parallel.
	checkDB(t)

	services := []database.Service{
		{
			Name: "1service1",
		},
		{
			Name: "1service2",
		},
		{
			Name: "1service3",
		},
	}

	// Add.
	for _, service := range services {
		err := database.TableServices.Add(context.Background(), testDB.GetPool(), &service)
		require.NoError(t, err)
	}

	newServices := []database.Service{
		{
			Name: "1service1", // Same as before.
		},
		{
			Name: "1  service   2", // New service name.
		},
		{
			Name: "13", // New service name.
		},
	}

	// Update.
	for i := range services {
		err := database.TableServices.Update(context.Background(), testDB.GetPool(), &newServices[i], services[i].Name)
		require.NoError(t, err)
	}

	// Assert.
	for i := range services {
		dbSerbice, err := database.TableServices.GetByServiceName(context.Background(), testDB.GetPool(),
			newServices[i].Name)
		require.NoError(t, err)
		require.NotNil(t, dbSerbice)
		assert.Equal(t, newServices[i].Name, dbSerbice.Name)
	}

	// Delete.
	for i := range newServices {
		err := database.TableServices.Delete(context.Background(), testDB.GetPool(), newServices[i].Name)
		require.NoError(t, err)
	}
}

//nolint:dupl // It is indeed almost a duplicate. Almost.
func TestServicesInvalidGetUpdateDelete(t *testing.T) {
	t.Parallel() // Running all db tests in parallel.
	checkDB(t)

	serviceNames := []string{"123", "", "12345678901234567890", "true", "nameuser", "AAAAAAA", "bbbbbBBBBBB",
		"1234567890123456789012345678901234567890", "nothing to see here"}

	for _, serviceName := range serviceNames {
		t.Run(serviceName, func(t *testing.T) {
			t.Parallel()

			ret, err := database.TableServices.GetByServiceName(context.Background(), testDB.GetPool(), serviceName)
			assert.Nil(t, ret)
			require.ErrorIs(t, database.ErrNoRows, err)

			var dummy database.Service
			err = database.TableServices.Update(context.Background(), testDB.GetPool(), &dummy, serviceName)
			require.ErrorIs(t, database.ErrNoRows, err)

			err = database.TableServices.Delete(context.Background(), testDB.GetPool(), serviceName)
			require.ErrorIs(t, database.ErrNoRows, err)
		})
	}
}

//nolint:funlen // Won't decompose.
func TestUsersGroupsValidAddAndGet(t *testing.T) {
	t.Parallel() // Running all db tests in parallel.
	checkDB(t)

	idsMap := setupForValidUserGroupsAddAndGet(t, "1")

	usersGroups := []database.AddUserRole{
		{
			UserID:      idsMap["username111"],
			UserRole:    database.UserRoleTypeUser,
			ServiceName: "service111",
		},
		{
			UserID:      idsMap["username111"],
			UserRole:    database.UserRoleTypeAdmin,
			ServiceName: "service211",
		},
		{
			UserID:      idsMap["username211"],
			UserRole:    database.UserRoleTypeUser,
			ServiceName: "service211",
		},
		{
			UserID:      idsMap["username311"],
			UserRole:    database.UserRoleTypeUser,
			ServiceName: "service211",
		},
	}

	mapCntGroupsOfAUser := make(map[string]int)
	listInsertedEntries := make([]database.UserRole, 0, len(usersGroups))

	for _, userGroup := range usersGroups {
		dbUserGroup, err := database.TableUsersRoles.Add(context.Background(), testDB.GetPool(), &userGroup)
		require.NoError(t, err)
		require.NotNil(t, dbUserGroup)
		assert.Equal(t, userGroup.UserID, dbUserGroup.UserID)
		assert.Equal(t, userGroup.UserRole, dbUserGroup.UserRole)
		assert.Equal(t, userGroup.ServiceName, dbUserGroup.ServiceName)
		assert.NotEqual(t, 0, dbUserGroup.ID)
		assert.NotEqualValues(t, 0, dbUserGroup.CreatedTS)

		mapCntGroupsOfAUser[userGroup.UserID]++

		listInsertedEntries = append(listInsertedEntries, *dbUserGroup)
	}

	for _, userGroup := range usersGroups {
		dbUserGroups, err := database.TableUsersRoles.GetByUserID(context.Background(), testDB.GetPool(),
			userGroup.UserID)
		require.NoError(t, err)
		require.NotNil(t, dbUserGroups)

		assert.Len(t, dbUserGroups, mapCntGroupsOfAUser[userGroup.UserID])

		foundInDB := false

		for _, dbGroup := range dbUserGroups {
			if dbGroup.UserID == userGroup.UserID &&
				dbGroup.ServiceName == userGroup.ServiceName &&
				dbGroup.UserRole == userGroup.UserRole {
				foundInDB = true

				break
			}
		}

		assert.True(t, foundInDB)
	}

	for _, inserted := range listInsertedEntries {
		dbEntry, err := database.TableUsersRoles.GetByID(context.Background(), testDB.GetPool(), inserted.ID)
		require.NoError(t, err)
		require.NotNil(t, dbEntry)

		assert.ElementsMatch(t,
			[]any{inserted.ID, inserted.UserID, inserted.ServiceName, inserted.CreatedTS.String(), inserted.UserRole},
			[]any{dbEntry.ID, dbEntry.UserID, dbEntry.ServiceName, dbEntry.CreatedTS.String(), dbEntry.UserRole},
		)
	}
}

func setupForValidUserGroupsAddAndGet(t *testing.T, suffix string) map[string]string {
	t.Helper()
	checkDB(t)

	resultingMap := make(map[string]string)

	// START OF SETUP.
	services := []database.Service{
		{
			Name: "service11" + suffix,
		},
		{
			Name: "service21" + suffix,
		},
		{
			Name: "service31" + suffix,
		},
	}

	for _, service := range services {
		err := database.TableServices.Add(context.Background(), testDB.GetPool(), &service)
		require.NoError(t, err)
	}

	users := []database.AddUser{
		{
			Username: "username11" + suffix,
			Password: "password1" + suffix,
		},
		{
			Username: "username21" + suffix,
			Password: "password2" + suffix,
		},
		{
			Username: "username31" + suffix,
			Password: "password3" + suffix,
		},
	}

	for _, user := range users {
		added, err := database.TableUsers.Add(context.Background(), testDB.GetPool(), &user)
		require.NoError(t, err)

		resultingMap[added.Username] = added.ID
	}

	return resultingMap
}

//nolint:funlen // Won't decompose.
func TestUsersGroupsValidUpdateAndDelete(t *testing.T) {
	t.Parallel() // Running all db tests in parallel.
	checkDB(t)

	idsMap := setupForValidUserGroupsAddAndGet(t, "2")

	usersGroups := []database.AddUserRole{
		{
			UserID:      idsMap["username112"],
			UserRole:    database.UserRoleTypeUser,
			ServiceName: "service312",
		},
		{
			UserID:      idsMap["username112"],
			UserRole:    database.UserRoleTypeAdmin,
			ServiceName: "service212",
		},
		{
			UserID:      idsMap["username212"],
			UserRole:    database.UserRoleTypeUser,
			ServiceName: "service212",
		},
		{
			UserID:      idsMap["username312"],
			UserRole:    database.UserRoleTypeUser,
			ServiceName: "service212",
		},
	}

	insertedEntries := make([]database.UserRole, 0, len(usersGroups))
	insertedIDs := make([]uint, 0, len(usersGroups))

	// Add.
	for _, userGroup := range usersGroups {
		ins, err := database.TableUsersRoles.Add(context.Background(), testDB.GetPool(), &userGroup)
		require.NoError(t, err)
		require.NotNil(t, ins)

		insertedEntries = append(insertedEntries, *ins)
		insertedIDs = append(insertedIDs, ins.ID)
	}

	// Mutate.
	insertedEntries[0].UserRole = database.UserRoleTypeRoot
	insertedEntries[1].ID = 97654
	insertedEntries[2].ServiceName = "service112"
	insertedEntries[2].UserRole = database.UserRoleTypeAdmin
	insertedEntries[3].ServiceName = "service312"
	insertedEntries[3].CreatedTS = insertedEntries[3].CreatedTS.Add(time.Hour * 1000)

	// Update.
	for i := range insertedEntries {
		err := database.TableUsersRoles.UpdateByID(context.Background(), testDB.GetPool(),
			&insertedEntries[i], insertedIDs[i])
		require.NoError(t, err)
	}

	// Assert.
	for _, inserted := range insertedEntries {
		dbEntry, err := database.TableUsersRoles.GetByID(context.Background(), testDB.GetPool(), inserted.ID)
		require.NoError(t, err)
		require.NotNil(t, dbEntry)

		assert.ElementsMatch(t,
			[]any{inserted.ID, inserted.UserID, inserted.ServiceName, inserted.CreatedTS.String(), inserted.UserRole},
			[]any{dbEntry.ID, dbEntry.UserID, dbEntry.ServiceName, dbEntry.CreatedTS.String(), dbEntry.UserRole},
		)
	}

	// Delete.
	for i := range insertedEntries {
		err := database.TableUsersRoles.DeleteByID(context.Background(), testDB.GetPool(), insertedEntries[i].ID)
		require.NoError(t, err)
	}
}
