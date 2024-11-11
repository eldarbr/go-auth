package model

import "github.com/eldarbr/go-auth/internal/service/encrypt"

type UserUsernme struct {
	Username string `json:"username"`
}

type UserCreds struct {
	UserUsernme
	Password string `json:"password"`
}

type UserTokenResponse struct {
	AuthResponse
}

const (
	capUserCredsUsernameMin = 3
	capUserCredsUsernameMax = 20
	capUserCredsPasswordMin = 5
	capUserCredsPasswordMax = 70
)

func (creds UserCreds) ValidFormat() bool {
	valid := len(creds.Username) >= capUserCredsUsernameMin &&
		len(creds.Username) <= capUserCredsUsernameMax &&
		len(creds.Password) >= capUserCredsPasswordMin &&
		len(creds.Password) <= capUserCredsPasswordMax &&
		encrypt.IsPrintableASCII(creds.Username) &&
		encrypt.IsPrintableASCII(creds.Password)

	return valid
}
