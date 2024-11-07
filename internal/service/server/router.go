package server

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type CommonHandlingModule interface {
	MethodNotAllowed(w http.ResponseWriter, _ *http.Request)
	NotFound(w http.ResponseWriter, _ *http.Request)
}

type AuthHandlingModule interface {
	Authenticate(w http.ResponseWriter, r *http.Request, _ httprouter.Params)
}

type ManageHandlingModule interface {
	CreateUser(w http.ResponseWriter, r *http.Request, _ httprouter.Params)
}

func NewRouter(common CommonHandlingModule, auth AuthHandlingModule,
	manage ManageHandlingModule) http.Handler {

	handler := httprouter.New()

	handler.MethodNotAllowed = http.HandlerFunc(common.MethodNotAllowed)
	handler.NotFound = http.HandlerFunc(common.NotFound)

	handler.POST("/auth/authenticate", auth.Authenticate)

	handler.POST("/manage/users", manage.CreateUser)

	return handler
}
