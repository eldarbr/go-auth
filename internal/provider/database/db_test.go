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

var testDB database.Database

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

	require.ErrorIs(t, database.TableUsers.Add(context.Background(), testDB.GetPool(), nil), database.ErrNilArgument)
	require.ErrorIs(t, database.TableServices.Add(context.Background(), testDB.GetPool(), nil),
		database.ErrNilArgument)

	_, err := database.TableUsersGroups.Add(context.Background(), testDB.GetPool(), nil)

	require.ErrorIs(t, err, database.ErrNilArgument)
	require.ErrorIs(t, database.TableUsersGroups.Insert(context.Background(), testDB.GetPool(), nil),
		database.ErrNilArgument)
}

func TestNilDB(t *testing.T) {
	t.Parallel() // Running all db tests in parallel.
	checkDB(t)

	// nil is an unitialized db.

	err := database.TableUsers.Add(context.Background(), nil, nil)
	require.ErrorIs(t, err, database.ErrDBNotInitilized)

	err = database.TableUsers.Update(context.Background(), nil, nil, "")
	require.ErrorIs(t, err, database.ErrDBNotInitilized)

	_, err = database.TableUsers.GetByUsername(context.Background(), nil, "")
	require.ErrorIs(t, err, database.ErrDBNotInitilized)

	err = database.TableUsers.Delete(context.Background(), nil, "")
	require.ErrorIs(t, err, database.ErrDBNotInitilized)

	err = database.TableServices.Add(context.Background(), nil, nil)
	require.ErrorIs(t, err, database.ErrDBNotInitilized)

	err = database.TableServices.Update(context.Background(), nil, nil, "")
	require.ErrorIs(t, err, database.ErrDBNotInitilized)

	_, err = database.TableServices.GetByServiceName(context.Background(), nil, "")
	require.ErrorIs(t, err, database.ErrDBNotInitilized)

	err = database.TableServices.Delete(context.Background(), nil, "")
	require.ErrorIs(t, err, database.ErrDBNotInitilized)

	err = database.TableUsersGroups.Insert(context.Background(), nil, nil)
	require.ErrorIs(t, err, database.ErrDBNotInitilized)

	_, err = database.TableUsersGroups.Add(context.Background(), nil, nil)
	require.ErrorIs(t, err, database.ErrDBNotInitilized)

	err = database.TableUsersGroups.UpdateByID(context.Background(), nil, nil, 0)
	require.ErrorIs(t, err, database.ErrDBNotInitilized)

	_, err = database.TableUsersGroups.GetByID(context.Background(), nil, 0)
	require.ErrorIs(t, err, database.ErrDBNotInitilized)

	_, err = database.TableUsersGroups.GetByUsername(context.Background(), nil, "")
	require.ErrorIs(t, err, database.ErrDBNotInitilized)

	err = database.TableUsersGroups.DeleteByID(context.Background(), nil, 0)
	require.ErrorIs(t, err, database.ErrDBNotInitilized)
}

func TestUsersValidAddAndGet(t *testing.T) {
	t.Parallel() // Running all db tests in parallel.
	checkDB(t)

	users := []database.User{
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
		err := database.TableUsers.Add(context.Background(), testDB.GetPool(), &user)
		require.NoError(t, err)
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

	users := []database.User{
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
		err := database.TableUsers.Add(context.Background(), testDB.GetPool(), &user)
		require.NoError(t, err)
	}

	newUsers := []database.User{
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
		err := database.TableUsers.Update(context.Background(), testDB.GetPool(), &newUsers[i], users[i].Username)
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
		err := database.TableUsers.Delete(context.Background(), testDB.GetPool(), newUsers[i].Username)
		require.NoError(t, err)
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

			var dummy database.User
			err = database.TableUsers.Update(context.Background(), testDB.GetPool(), &dummy, username)
			require.ErrorIs(t, database.ErrNoRows, err)

			err = database.TableUsers.Delete(context.Background(), testDB.GetPool(), username)
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

	setupForValidUserGroupsAddAndGet(t, "1")

	usersGroups := []database.AddUserGroup{
		{
			Username:    "username111",
			UserRole:    database.UserRoleTypeUser,
			ServiceName: "service111",
		},
		{
			Username:    "username111",
			UserRole:    database.UserRoleTypeAdmin,
			ServiceName: "service211",
		},
		{
			Username:    "username211",
			UserRole:    database.UserRoleTypeUser,
			ServiceName: "service211",
		},
		{
			Username:    "username311",
			UserRole:    database.UserRoleTypeUser,
			ServiceName: "service211",
		},
	}

	mapCntGroupsOfAUser := make(map[string]int)
	listInsertedEntries := make([]database.UserGroup, 0, len(usersGroups))

	for _, userGroup := range usersGroups {
		dbUserGroup, err := database.TableUsersGroups.Add(context.Background(), testDB.GetPool(), &userGroup)
		require.NoError(t, err)
		require.NotNil(t, dbUserGroup)
		assert.Equal(t, userGroup.Username, dbUserGroup.Username)
		assert.Equal(t, userGroup.UserRole, dbUserGroup.UserRole)
		assert.Equal(t, userGroup.ServiceName, dbUserGroup.ServiceName)
		assert.NotEqual(t, 0, dbUserGroup.ID)
		assert.NotEqualValues(t, 0, dbUserGroup.CreatedTS)

		mapCntGroupsOfAUser[userGroup.Username]++

		listInsertedEntries = append(listInsertedEntries, *dbUserGroup)
	}

	for _, userGroup := range usersGroups {
		dbUserGroups, err := database.TableUsersGroups.GetByUsername(context.Background(), testDB.GetPool(),
			userGroup.Username)
		require.NoError(t, err)
		require.NotNil(t, dbUserGroups)

		assert.Len(t, dbUserGroups, mapCntGroupsOfAUser[userGroup.Username])

		foundInDB := false

		for _, dbGroup := range dbUserGroups {
			if dbGroup.Username == userGroup.Username &&
				dbGroup.ServiceName == userGroup.ServiceName &&
				dbGroup.UserRole == userGroup.UserRole {
				foundInDB = true

				break
			}
		}

		assert.True(t, foundInDB)
	}

	for _, inserted := range listInsertedEntries {
		dbEntry, err := database.TableUsersGroups.GetByID(context.Background(), testDB.GetPool(), inserted.ID)
		require.NoError(t, err)
		require.NotNil(t, dbEntry)

		assert.ElementsMatch(t,
			[]any{inserted.ID, inserted.Username, inserted.ServiceName, inserted.CreatedTS.String(), inserted.UserRole},
			[]any{dbEntry.ID, dbEntry.Username, dbEntry.ServiceName, dbEntry.CreatedTS.String(), dbEntry.UserRole},
		)
	}
}

func setupForValidUserGroupsAddAndGet(t *testing.T, suffix string) {
	t.Helper()
	checkDB(t)

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

	users := []database.User{
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
		err := database.TableUsers.Add(context.Background(), testDB.GetPool(), &user)
		require.NoError(t, err)
	}
}

//nolint:funlen // Won't decompose.
func TestUsersGroupsValidUpdateAndDelete(t *testing.T) {
	t.Parallel() // Running all db tests in parallel.
	checkDB(t)

	setupForValidUserGroupsAddAndGet(t, "2")

	usersGroups := []database.AddUserGroup{
		{
			Username:    "username112",
			UserRole:    database.UserRoleTypeUser,
			ServiceName: "service312",
		},
		{
			Username:    "username112",
			UserRole:    database.UserRoleTypeAdmin,
			ServiceName: "service212",
		},
		{
			Username:    "username212",
			UserRole:    database.UserRoleTypeUser,
			ServiceName: "service212",
		},
		{
			Username:    "username312",
			UserRole:    database.UserRoleTypeUser,
			ServiceName: "service212",
		},
	}

	insertedEntries := make([]database.UserGroup, 0, len(usersGroups))
	insertedIDs := make([]uint, 0, len(usersGroups))

	// Add.
	for _, userGroup := range usersGroups {
		ins, err := database.TableUsersGroups.Add(context.Background(), testDB.GetPool(), &userGroup)
		require.NoError(t, err)
		require.NotNil(t, ins)

		insertedEntries = append(insertedEntries, *ins)
		insertedIDs = append(insertedIDs, ins.ID)
	}

	// Mutate.
	insertedEntries[0].UserRole = database.UserRoleTypeRoot
	insertedEntries[1].Username = "username112"
	insertedEntries[1].ID = 97654
	insertedEntries[2].ServiceName = "service112"
	insertedEntries[2].UserRole = database.UserRoleTypeAdmin
	insertedEntries[3].ServiceName = "service312"
	insertedEntries[3].CreatedTS = insertedEntries[3].CreatedTS.Add(time.Hour * 1000)

	// Update.
	for i := range insertedEntries {
		err := database.TableUsersGroups.UpdateByID(context.Background(), testDB.GetPool(),
			&insertedEntries[i], insertedIDs[i])
		require.NoError(t, err)
	}

	// Assert.
	for _, inserted := range insertedEntries {
		dbEntry, err := database.TableUsersGroups.GetByID(context.Background(), testDB.GetPool(), inserted.ID)
		require.NoError(t, err)
		require.NotNil(t, dbEntry)

		assert.ElementsMatch(t,
			[]any{inserted.ID, inserted.Username, inserted.ServiceName, inserted.CreatedTS.String(), inserted.UserRole},
			[]any{dbEntry.ID, dbEntry.Username, dbEntry.ServiceName, dbEntry.CreatedTS.String(), dbEntry.UserRole},
		)
	}

	// Delete.
	for i := range insertedEntries {
		err := database.TableUsersGroups.DeleteByID(context.Background(), testDB.GetPool(), insertedEntries[i].ID)
		require.NoError(t, err)
	}
}

func TestUsersGroupsInvalidGetUpdateDelete(t *testing.T) {
	t.Parallel() // Running all db tests in parallel.
	checkDB(t)

	usernames := []string{"123", "", "12345678901234567890", "true", "nameuser", "AAAAAAA", "bbbbbBBBBBB",
		"1234567890123456789012345678901234567890", "nothing to see here"}

	for _, username := range usernames {
		t.Run(username, func(t *testing.T) {
			t.Parallel()

			invalidID := uint(len(username) + 99999) //nolint:gosec // no overflow is possible.

			ret1, err := database.TableUsersGroups.GetByUsername(context.Background(), testDB.GetPool(), username)
			assert.Empty(t, ret1)
			require.NoError(t, err)

			ret2, err := database.TableUsersGroups.GetByID(context.Background(), testDB.GetPool(), invalidID)
			assert.Nil(t, ret2)
			require.ErrorIs(t, database.ErrNoRows, err)

			var dummy database.UserGroup
			dummy.UserRole = database.UserRoleTypeUser
			err = database.TableUsersGroups.UpdateByID(context.Background(), testDB.GetPool(), &dummy, invalidID)
			require.ErrorIs(t, database.ErrNoRows, err)

			err = database.TableUsersGroups.DeleteByID(context.Background(), testDB.GetPool(), invalidID)
			require.ErrorIs(t, database.ErrNoRows, err)
		})
	}
}
