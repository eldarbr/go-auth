package model

import (
	"github.com/eldarbr/go-auth/internal/provider/database"
	"github.com/eldarbr/go-auth/internal/service/encrypt"
)

func PrepareClaims(dbGroups []database.UserGroup) []encrypt.ClaimUserGroupRole {
	if len(dbGroups) == 0 {
		return nil
	}

	claimGroups := make([]encrypt.ClaimUserGroupRole, 0, len(dbGroups))

	for _, dbEntry := range dbGroups {
		newClaim := encrypt.ClaimUserGroupRole{
			ServiceName: dbEntry.ServiceName,
			UserRole:    dbEntry.UserRole,
		}
		claimGroups = append(claimGroups, newClaim)
	}

	return claimGroups
}
