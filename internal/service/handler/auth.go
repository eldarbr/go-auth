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
}

func NewAuthHandl(dbInstance *database.Database, jwtService *encrypt.JWTService) AuthHandl {
	srv := AuthHandl{
		dbInstance: dbInstance,
		jwtService: jwtService,
	}

	return srv
}

func (authHandl AuthHandl) Authenticate(respWriter http.ResponseWriter, request *http.Request, _ httprouter.Params) {
	log.Printf("request Authenticate received")

	var creds model.UserCreds

	err := json.NewDecoder(request.Body).Decode(&creds)
	if err != nil || creds.Password == "" || creds.Username == "" {
		writeJSONResponse(respWriter, model.ErrorResponse{Error: "bad request"}, http.StatusBadRequest)

		return
	}

	dbUser, err := database.TableUsers.GetByUsername(request.Context(), authHandl.dbInstance.GetPool(), creds.Username)
	if errors.Is(err, database.ErrNoRows) || !encrypt.PasswordCompare(creds.Password, dbUser.Password) {
		writeJSONResponse(respWriter, model.ErrorResponse{Error: "unauthorized"}, http.StatusUnauthorized)

		return
	}

	if err != nil {
		log.Printf("TableUsers.GetByUsername %s: %s", creds.Username, err.Error())
		writeJSONResponse(respWriter, model.ErrorResponse{Error: "internal error"}, http.StatusInternalServerError)

		return
	}

	dbUserGroups, err := database.TableUsersGroups.GetByUsername(request.Context(), authHandl.dbInstance.GetPool(),
		dbUser.Username)
	if err != nil && !errors.Is(err, database.ErrNoRows) {
		log.Printf("TableUsersGroups.GetByUsername %s: %s", creds.Username, err.Error())
		writeJSONResponse(respWriter, model.ErrorResponse{Error: "internal error"}, http.StatusInternalServerError)

		return
	}

	claimGroups := model.PrepareClaims(dbUserGroups)
	token, err := authHandl.jwtService.IssueToken(encrypt.AuthCustomClaims{
		Username: dbUser.Username,
		Groups:   claimGroups,
	})

	if err != nil {
		log.Printf("jwtService.IssueToken: %s", err.Error())
		writeJSONResponse(respWriter, model.ErrorResponse{Error: "internal error"}, http.StatusInternalServerError)

		return
	}

	resp := model.UserTokenResponse{
		AuthResponse: model.AuthResponse{Token: token},
	}

	writeJSONResponse(respWriter, resp, http.StatusOK)
}
