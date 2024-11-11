package server

import (
	"net/http"

	"github.com/eldarbr/go-auth/internal/provider/database"
	"github.com/eldarbr/go-auth/internal/service/encrypt"
	"github.com/julienschmidt/httprouter"
)

const myOwnServiceName = "go-auth"

type CommonHandlingModule interface {
	MethodNotAllowed(w http.ResponseWriter, _ *http.Request)
	NotFound(w http.ResponseWriter, _ *http.Request)
}

type AuthHandlingModule interface {
	Authenticate(w http.ResponseWriter, r *http.Request, _ httprouter.Params)
}

type ManageHandlingModule interface {
	CreateUser(w http.ResponseWriter, r *http.Request, _ httprouter.Params)
	GetUserInfo(w http.ResponseWriter, r *http.Request, _ httprouter.Params)
	MiddlewareAuthorizeAnyClaim(requestedClaims []encrypt.ClaimUserRole, next httprouter.Handle) httprouter.Handle
}

func NewRouter(common CommonHandlingModule, auth AuthHandlingModule,
	manage ManageHandlingModule) http.Handler {

	handler := httprouter.New()

	handler.MethodNotAllowed = http.HandlerFunc(common.MethodNotAllowed)
	handler.NotFound = http.HandlerFunc(common.NotFound)

	handler.POST("/authapi/auth/authenticate", auth.Authenticate)

	handler.POST("/authapi/manage/users", manage.MiddlewareAuthorizeAnyClaim(
		[]encrypt.ClaimUserRole{{ServiceName: myOwnServiceName, UserRole: database.UserRoleTypeRoot}},
		manage.CreateUser,
	))
	handler.GET("/authapi/manage/users", manage.MiddlewareAuthorizeAnyClaim(
		[]encrypt.ClaimUserRole{{ServiceName: myOwnServiceName, UserRole: database.UserRoleTypeRoot}},
		manage.GetUserInfo,
	))

	return handler
}
