package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/eldarbr/go-auth/internal/model"
	"github.com/eldarbr/go-auth/internal/provider/storage"
	"github.com/eldarbr/go-auth/internal/service/encrypt"
	"github.com/eldarbr/go-auth/pkg/database"
	"github.com/julienschmidt/httprouter"
)

type AuthHandl struct {
	cache         CacheImpl
	dbInstance    *database.Database
	jwtService    *encrypt.JWTService
	sessionDomain string
	reqLimit      int
}

func NewAuthHandl(dbInstance *database.Database, jwtService *encrypt.JWTService,
	cache CacheImpl, limit int, sessionDomain string) AuthHandl {
	srv := AuthHandl{
		dbInstance:    dbInstance,
		jwtService:    jwtService,
		cache:         cache,
		reqLimit:      limit,
		sessionDomain: sessionDomain,
	}

	return srv
}

func (authHandl AuthHandl) Authenticate(respWriter http.ResponseWriter,
	request *http.Request, params httprouter.Params) {
	log.Printf("request Authenticate received")

	token, _ := authHandl.getToken(respWriter, request, params)
	if token == "" {
		return
	}

	// Prepare the response body.
	resp := model.UserTokenResponse{
		AuthResponse: model.AuthResponse{Token: token},
	}

	writeJSONResponse(respWriter, resp, http.StatusOK)
}

func (authHandl AuthHandl) InitSession(respWriter http.ResponseWriter,
	request *http.Request, params httprouter.Params) {
	log.Printf("request InitSession received")

	token, expires := authHandl.getToken(respWriter, request, params)
	if token == "" {
		return
	}

	responseCookie := http.Cookie{
		Name:     "tokenid",
		Value:    token,
		Domain:   authHandl.sessionDomain,
		Secure:   true,
		HttpOnly: true,
		Expires:  *expires,
		Path:     "/",
	}

	http.SetCookie(respWriter, &responseCookie)

	writeJSONResponse(respWriter, model.ErrorResponse{Error: ""}, http.StatusOK)
}

func (authHandl AuthHandl) getToken(respWriter http.ResponseWriter, request *http.Request, _ httprouter.Params) (string, *time.Time) {
	var creds model.UserCreds

	// Decode the request body.
	err := json.NewDecoder(request.Body).Decode(&creds)
	if err != nil || creds.Password == "" || creds.Username == "" {
		writeJSONResponse(respWriter, model.ErrorResponse{Error: "bad request"}, http.StatusBadRequest)

		return "", nil
	}

	lookups := authHandl.cache.GetAndIncrease("usr:" + creds.Username)
	if lookups > authHandl.reqLimit {
		writeJSONResponse(respWriter, model.ErrorResponse{Error: "rate limited"}, http.StatusTooManyRequests)

		return "", nil
	}

	// Get username entry from the db.
	dbUser, err := storage.TableUsers.GetByUsername(request.Context(), authHandl.dbInstance.GetPool(), creds.Username)
	if errors.Is(err, database.ErrNoRows) || !encrypt.PasswordCompare(creds.Password, dbUser.Password) {
		// ErrNoRows or wrong hash -> unauthorized.
		writeJSONResponse(respWriter, model.ErrorResponse{Error: "unauthorized"}, http.StatusUnauthorized)

		return "", nil
	}

	// Handle db query error.
	if err != nil {
		log.Printf("TableUsers.GetByUsername %s: %s", creds.Username, err.Error())
		writeJSONResponse(respWriter, model.ErrorResponse{Error: "internal error"}, http.StatusInternalServerError)

		return "", nil
	}

	// Get user's roles.
	dbUserRoles, err := storage.TableUsersRoles.GetByUserID(request.Context(), authHandl.dbInstance.GetPool(),
		dbUser.ID)
	if err != nil && !errors.Is(err, database.ErrNoRows) {
		log.Printf("TableUsersGroups.GetByUsername %s: %s", creds.Username, err.Error())
		writeJSONResponse(respWriter, model.ErrorResponse{Error: "internal error"}, http.StatusInternalServerError)

		return "", nil
	}

	// Convert the roles to custom jwt claims.
	claims := model.PrepareClaims(dbUserRoles)

	// Issue a token.
	token, expires, err := authHandl.jwtService.IssueToken(encrypt.AuthCustomClaims{
		Username: dbUser.Username,
		Roles:    claims,
		UserID:   dbUser.ID,
	})
	if err != nil {
		log.Printf("jwtService.IssueToken: %s", err.Error())
		writeJSONResponse(respWriter, model.ErrorResponse{Error: "internal error"}, http.StatusInternalServerError)

		return "", nil
	}

	return token, expires
}
