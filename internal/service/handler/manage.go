package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/eldarbr/go-auth/internal/model"
	"github.com/eldarbr/go-auth/internal/provider/database"
	"github.com/eldarbr/go-auth/internal/service/encrypt"
	"github.com/julienschmidt/httprouter"
)

type ManageHandl struct {
	dbInstance *database.Database
	jwtService *encrypt.JWTService
}

func NewManageHandl(dbInstance *database.Database, jwtService *encrypt.JWTService) ManageHandl {
	srv := ManageHandl{
		dbInstance: dbInstance,
		jwtService: jwtService,
	}

	return srv
}

func (manage ManageHandl) CreateUser(respWriter http.ResponseWriter, request *http.Request, _ httprouter.Params) {
	log.Printf("request CreateUser received")

	var parsedBody model.UserCreds

	err := json.NewDecoder(request.Body).Decode(&parsedBody)
	if err != nil || !parsedBody.ValidFormat() {
		writeJSONResponse(respWriter, model.ErrorResponse{Error: "bad request"}, http.StatusBadRequest)

		return
	}

	hashedPassword, err := encrypt.PasswordEncrypt(parsedBody.Password)
	if err != nil {
		log.Printf("CreateUser - hash password err: %s", err.Error())
		writeJSONResponse(respWriter, model.ErrorResponse{Error: "internal error"}, http.StatusInternalServerError)

		return
	}

	dbUser := database.User{
		Username: parsedBody.Username,
		Password: hashedPassword,
	}

	err = database.TableUsers.Add(request.Context(), manage.dbInstance.GetPool(), &dbUser)
	if err != nil {
		log.Printf("CreateUser - insert user err: %s", err.Error())
		writeJSONResponse(respWriter, model.ErrorResponse{Error: "internal error"}, http.StatusInternalServerError)

		return
	}

	writeJSONResponse(respWriter, model.UserUsernme{Username: dbUser.Username}, http.StatusOK)
}

func (manage ManageHandl) GetUserInfo(respWriter http.ResponseWriter, request *http.Request, _ httprouter.Params) {
	log.Printf("request GetUserInfo received")

	requestedUsername := request.URL.Query().Get("username")
	if requestedUsername == "" {
		writeJSONResponse(respWriter, model.ErrorResponse{Error: "bad request"}, http.StatusBadRequest)

		return
	}

	data, err := database.TableUsersRoles.GetByUsername(request.Context(), manage.dbInstance.GetPool(), requestedUsername)
	if err != nil {
		log.Printf("GetUserInfo - get user err: %s", err.Error())
		writeJSONResponse(respWriter, model.ErrorResponse{Error: "internal error"}, http.StatusInternalServerError)

		return
	}

	response := model.UserInfoResponse{
		Username: requestedUsername,
		Roles:    model.PrepareClaims(data),
	}

	writeJSONResponse(respWriter, response, http.StatusOK)
}

// Checks if the user has any of the claims.
func (manage ManageHandl) MiddlewareAuthorizeAnyClaim(requestedClaims []encrypt.ClaimUserRole,
	next httprouter.Handle) httprouter.Handle {
	return func(respWriter http.ResponseWriter, request *http.Request, routerParams httprouter.Params) {
		claims, err := manage.jwtService.ValidateToken(request.Header.Get("Authorization"))
		if err != nil {
			writeJSONResponse(respWriter, model.ErrorResponse{Error: "unauthorized"}, http.StatusUnauthorized)

			return
		}

		if !claims.ContainAny(requestedClaims) {
			writeJSONResponse(respWriter, model.ErrorResponse{Error: "forbidden"}, http.StatusForbidden)

			return
		}

		next(respWriter, request, routerParams)
	}
}
