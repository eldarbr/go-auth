package model

import (
	"regexp"

	"github.com/eldarbr/go-auth/internal/service/encrypt"
)

type UserUsernme struct {
	Username string `json:"username"`
}

type UserCreateResponse struct {
	UserUsernme
	UserID string `json:"userId"`
}

type UserCreds struct {
	UserUsernme
	Password string `json:"password"`
}

type UserTokenResponse struct {
	AuthResponse
}

const (
	CapUserCredsUsernameMinlen = 4
	CapUserCredsUsernameMaxlen = 20
	CapUserCredsPasswordMinlen = 6
	CapUserCredsPasswordMaxlen = 70
)

// ValidFormat tests if the userCreds are valid.
// A valid userCreds is a cred with a username and password,
// 1) which lengths are not less than MINlen and are not greater than MAXlen;
// 2) which consist of printable ascii character.
// 3) username only consists of digits or letters.
func (creds UserCreds) ValidFormat() bool {
	valid := len(creds.Username) >= CapUserCredsUsernameMinlen &&
		len(creds.Username) <= CapUserCredsUsernameMaxlen &&
		len(creds.Password) >= CapUserCredsPasswordMinlen &&
		len(creds.Password) <= CapUserCredsPasswordMaxlen &&
		validateUsername(creds.Username) &&
		encrypt.IsPrintableASCII(creds.Password)

	return valid
}

var regexpValidUsername = regexp.MustCompile("^[0-9A-z]+$")

func validateUsername(username string) bool {
	return regexpValidUsername.MatchString(username)
}
