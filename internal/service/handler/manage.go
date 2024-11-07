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

const theServiceName = "go-auth"

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

	var parsedBody model.CraeteUserRequest

	err := json.NewDecoder(request.Body).Decode(&parsedBody)
	if err != nil || parsedBody.Auth.Token == "" || !parsedBody.NewUser.ValidFormat() {
		writeJSONResponse(respWriter, model.ErrorResponse{Error: "bad request"}, http.StatusBadRequest)

		return
	}

	claims, err := manage.jwtService.ValidateToken(parsedBody.Auth.Token)
	if err != nil {
		writeJSONResponse(respWriter, model.ErrorResponse{Error: "unauthorized"}, http.StatusUnauthorized)

		return
	}

	if !claims.Contain(encrypt.ClaimUserGroupRole{ServiceName: theServiceName, UserRole: database.UserRoleTypeRoot}) {
		writeJSONResponse(respWriter, model.ErrorResponse{Error: "forbidden"}, http.StatusForbidden)

		return
	}

	hashedPassword, err := encrypt.PasswordEncrypt(parsedBody.NewUser.Password)
	if err != nil {
		log.Printf("CreateUser - hash password err: %s", err.Error())
		writeJSONResponse(respWriter, model.ErrorResponse{Error: "internal error"}, http.StatusInternalServerError)

		return
	}

	dbUser := database.User{
		Username: parsedBody.NewUser.Username,
		Password: hashedPassword,
	}

	err = database.TableUsers.Add(request.Context(), manage.dbInstance.GetPool(), &dbUser)
	if err != nil {
		log.Printf("CreateUser - insert user err: %s", err.Error())
		writeJSONResponse(respWriter, model.ErrorResponse{Error: "internal error"}, http.StatusInternalServerError)

		return
	}

	writeJSONResponse(respWriter, parsedBody.NewUser, http.StatusOK)
}
