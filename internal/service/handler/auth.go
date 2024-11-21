package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/eldarbr/go-auth/internal/model"
	"github.com/eldarbr/go-auth/internal/provider/database"
	"github.com/eldarbr/go-auth/internal/service/encrypt"
	"github.com/julienschmidt/httprouter"
)

type AuthHandl struct {
	dbInstance *database.Database
	jwtService *encrypt.JWTService
	cache      CacheImpl
	reqLimit   int
}

func NewAuthHandl(dbInstance *database.Database, jwtService *encrypt.JWTService,
	cache CacheImpl, limit int) AuthHandl {
	srv := AuthHandl{
		dbInstance: dbInstance,
		jwtService: jwtService,
		cache:      cache,
		reqLimit:   limit,
	}

	return srv
}

func (authHandl AuthHandl) Authenticate(respWriter http.ResponseWriter, request *http.Request, _ httprouter.Params) {
	log.Printf("request Authenticate received")

	var creds model.UserCreds

	// Decode the request body.
	err := json.NewDecoder(request.Body).Decode(&creds)
	if err != nil || creds.Password == "" || creds.Username == "" {
		writeJSONResponse(respWriter, model.ErrorResponse{Error: "bad request"}, http.StatusBadRequest)

		return
	}

	lookups := authHandl.cache.GetAndIncrease("usr:" + creds.Username)
	if lookups > authHandl.reqLimit {
		writeJSONResponse(respWriter, model.ErrorResponse{Error: "rate limited"}, http.StatusTooManyRequests)

		return
	}

	// Get username entry from the db.
	dbUser, err := database.TableUsers.GetByUsername(request.Context(), authHandl.dbInstance.GetPool(), creds.Username)
	if errors.Is(err, database.ErrNoRows) || !encrypt.PasswordCompare(creds.Password, dbUser.Password) {
		// ErrNoRows or wrong hash -> unauthorized.
		writeJSONResponse(respWriter, model.ErrorResponse{Error: "unauthorized"}, http.StatusUnauthorized)

		return
	}

	// Handle db query error.
	if err != nil {
		log.Printf("TableUsers.GetByUsername %s: %s", creds.Username, err.Error())
		writeJSONResponse(respWriter, model.ErrorResponse{Error: "internal error"}, http.StatusInternalServerError)

		return
	}

	// Get user's roles.
	dbUserRoles, err := database.TableUsersRoles.GetByUserID(request.Context(), authHandl.dbInstance.GetPool(),
		dbUser.ID)
	if err != nil && !errors.Is(err, database.ErrNoRows) {
		log.Printf("TableUsersGroups.GetByUsername %s: %s", creds.Username, err.Error())
		writeJSONResponse(respWriter, model.ErrorResponse{Error: "internal error"}, http.StatusInternalServerError)

		return
	}

	// Convert the roles to custom jwt claims.
	claims := model.PrepareClaims(dbUserRoles)

	// Issue a token.
	token, err := authHandl.jwtService.IssueToken(encrypt.AuthCustomClaims{
		Username: dbUser.Username,
		Roles:    claims,
		UserID:   dbUser.ID,
	})
	if err != nil {
		log.Printf("jwtService.IssueToken: %s", err.Error())
		writeJSONResponse(respWriter, model.ErrorResponse{Error: "internal error"}, http.StatusInternalServerError)

		return
	}

	// Prepare the response body.
	resp := model.UserTokenResponse{
		AuthResponse: model.AuthResponse{Token: token},
	}

	writeJSONResponse(respWriter, resp, http.StatusOK)
}
