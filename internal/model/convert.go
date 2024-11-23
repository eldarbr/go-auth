package model

import (
	"github.com/eldarbr/go-auth/internal/provider/storage"
	"github.com/eldarbr/go-auth/internal/service/encrypt"
)

func PrepareClaims(dbRoles []storage.UserRole) []encrypt.ClaimUserRole {
	if len(dbRoles) == 0 {
		return nil
	}

	claimGroups := make([]encrypt.ClaimUserRole, 0, len(dbRoles))

	for _, dbEntry := range dbRoles {
		newClaim := encrypt.ClaimUserRole{
			ServiceName: dbEntry.ServiceName,
			UserRole:    dbEntry.UserRole,
		}
		claimGroups = append(claimGroups, newClaim)
	}

	return claimGroups
}
