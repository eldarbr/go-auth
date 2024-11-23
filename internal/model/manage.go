package model

import "github.com/eldarbr/go-auth/internal/service/encrypt"

type UserInfoResponse struct {
	UserID   string                  `json:"userId"`
	Username string                  `json:"username"`
	Roles    []encrypt.ClaimUserRole `json:"roles"`
}
