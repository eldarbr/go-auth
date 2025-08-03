package server

import (
	"net/http"

	"github.com/eldarbr/go-auth/internal/provider/storage"
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
	InitSession(w http.ResponseWriter, r *http.Request, _ httprouter.Params)
}

type ManageHandlingModule interface {
	CreateUser(w http.ResponseWriter, r *http.Request, _ httprouter.Params)
	GetUserInfo(w http.ResponseWriter, r *http.Request, _ httprouter.Params)
	MiddlewareAuthorizeAnyClaim(requestedClaims []encrypt.ClaimUserRole, next httprouter.Handle) httprouter.Handle
	MiddlewareRateLimit(next httprouter.Handle) httprouter.Handle
}

type RateLimitHandlingModule interface {
	MiddlewareIPRateLimit(next httprouter.Handle) httprouter.Handle
}

func NewRouter(common CommonHandlingModule, auth AuthHandlingModule,
	manage ManageHandlingModule, ratelimiter RateLimitHandlingModule) http.Handler {

	handler := httprouter.New()

	handler.HandleOPTIONS = false

	handler.MethodNotAllowed = http.HandlerFunc(common.MethodNotAllowed)
	handler.NotFound = http.HandlerFunc(common.NotFound)

	// authenticate
	handler.POST("/auth/authenticate", auth.Authenticate)
	handler.POST("/auth/initsession", auth.InitSession)

	// create a user.
	handler.POST("/manage/users", ratelimiter.MiddlewareIPRateLimit(manage.MiddlewareAuthorizeAnyClaim(
		[]encrypt.ClaimUserRole{{ServiceName: myOwnServiceName, UserRole: storage.UserRoleTypeRoot}},
		manage.MiddlewareRateLimit(manage.CreateUser),
	)))

	// get a user.
	handler.GET("/manage/users", ratelimiter.MiddlewareIPRateLimit(manage.MiddlewareAuthorizeAnyClaim(
		[]encrypt.ClaimUserRole{{ServiceName: myOwnServiceName, UserRole: storage.UserRoleTypeRoot}},
		manage.MiddlewareRateLimit(manage.GetUserInfo),
	)))

	return handler
}
