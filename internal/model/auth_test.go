package model_test

import (
	"testing"

	"github.com/eldarbr/go-auth/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestCredsValidationEmpty(t *testing.T) {
	t.Parallel()

	creds := model.UserCreds{
		UserUsernme: model.UserUsernme{Username: ""},
		Password:    "",
	}

	assert.False(t, creds.ValidFormat())
}

func TestCredsValidationEmptyUsername(t *testing.T) {
	t.Parallel()

	creds := model.UserCreds{
		UserUsernme: model.UserUsernme{Username: ""},
		Password:    "superpassword",
	}

	assert.False(t, creds.ValidFormat())
}

func TestCredsValidationEmptyPassword(t *testing.T) {
	t.Parallel()

	creds := model.UserCreds{
		UserUsernme: model.UserUsernme{Username: "superusername"},
		Password:    "",
	}

	assert.False(t, creds.ValidFormat())
}

func TestCredsValidationShortPassword(t *testing.T) {
	t.Parallel()

	creds := model.UserCreds{
		UserUsernme: model.UserUsernme{Username: "superusername"},
		Password:    "shrt",
	}

	assert.False(t, creds.ValidFormat())
}

func TestCredsValidationUsernameWrongChars(t *testing.T) {
	t.Parallel()

	creds := model.UserCreds{
		UserUsernme: model.UserUsernme{Username: "!WHAT?~"},
		Password:    "superpassword!!!./zxc$",
	}

	assert.False(t, creds.ValidFormat())
}

func TestCredsValidationPasswordWrongChars(t *testing.T) {
	t.Parallel()

	creds := model.UserCreds{
		UserUsernme: model.UserUsernme{Username: "WHAT"},
		Password:    "\000superpassword!!!./zxc$\x7f",
	}

	assert.False(t, creds.ValidFormat())
}
func TestCredsValidationValid(t *testing.T) {
	t.Parallel()

	creds := model.UserCreds{
		UserUsernme: model.UserUsernme{Username: "dougiela"},
		Password:    "superpassword!!!./zxc$",
	}

	assert.True(t, creds.ValidFormat())
}
