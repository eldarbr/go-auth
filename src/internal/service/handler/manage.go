package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/eldarbr/go-auth/internal/model"
	"github.com/eldarbr/go-auth/internal/provider/storage"
	"github.com/eldarbr/go-auth/internal/service/encrypt"
	"github.com/eldarbr/go-auth/pkg/database"
	"github.com/julienschmidt/httprouter"
)

type ManageHandl struct {
	dbInstance *database.Database
	jwtService *encrypt.JWTService
	cache      CacheImpl
	reqLimit   int
}

type ctxKey string

const (
	ctxKeyRequesterUsername ctxKey = "RequesterUsername"
)

func NewManageHandl(dbInstance *database.Database, jwtService *encrypt.JWTService,
	cache CacheImpl, limit int) ManageHandl {
	srv := ManageHandl{
		dbInstance: dbInstance,
		jwtService: jwtService,
		cache:      cache,
		reqLimit:   limit,
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

	dbUser := storage.AddUser{
		Username: parsedBody.Username,
		Password: hashedPassword,
	}

	dbCreatedUser, err := storage.TableUsers.Add(request.Context(), manage.dbInstance.GetPool(), &dbUser)
	if err != nil {
		log.Printf("CreateUser - insert user err: %s", err.Error())
		writeJSONResponse(respWriter, model.ErrorResponse{Error: "internal error"}, http.StatusInternalServerError)

		return
	}

	writeJSONResponse(respWriter, model.UserCreateResponse{
		UserID:      dbCreatedUser.ID,
		UserUsernme: model.UserUsernme{Username: dbCreatedUser.Username},
	}, http.StatusOK)
}

func (manage ManageHandl) GetUserInfo(respWriter http.ResponseWriter, request *http.Request, _ httprouter.Params) {
	log.Printf("request GetUserInfo received")

	requestedUsername := request.URL.Query().Get("username")
	if requestedUsername == "" {
		writeJSONResponse(respWriter, model.ErrorResponse{Error: "bad request"}, http.StatusBadRequest)

		return
	}

	userInfo, infoErr := storage.TableUsers.GetByUsername(request.Context(),
		manage.dbInstance.GetPool(), requestedUsername)
	if infoErr != nil {
		log.Printf("GetUserInfo - get user err: %s", infoErr.Error())
		writeJSONResponse(respWriter, model.ErrorResponse{Error: "internal error"}, http.StatusInternalServerError)

		return
	}

	roles, rolesErr := storage.TableUsersRoles.GetByUserID(request.Context(),
		manage.dbInstance.GetPool(), userInfo.ID)
	if rolesErr != nil && !errors.Is(rolesErr, database.ErrNoRows) {
		log.Printf("GetUserInfo - get user err: %s", rolesErr.Error())
		writeJSONResponse(respWriter, model.ErrorResponse{Error: "internal error"}, http.StatusInternalServerError)

		return
	}

	response := model.UserInfoResponse{
		Username: requestedUsername,
		UserID:   userInfo.ID,
		Roles:    model.PrepareClaims(roles),
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

		nextCtx := context.WithValue(request.Context(), ctxKeyRequesterUsername, claims.Username)

		next(respWriter, request.WithContext(nextCtx), routerParams)
	}
}

func (manage ManageHandl) MiddlewareRateLimit(next httprouter.Handle) httprouter.Handle {
	return func(respWriter http.ResponseWriter, request *http.Request, routerParams httprouter.Params) {
		if manage.cache == nil {
			log.Println("MiddlewareRateLimit uninitialized cache")
			writeJSONResponse(respWriter, model.ErrorResponse{Error: "internal error"}, http.StatusInternalServerError)

			return
		}

		username, usernameOk := request.Context().Value(ctxKeyRequesterUsername).(string)

		if usernameOk && username != "" {
			lookups := manage.cache.GetAndIncrease("usr:" + username)
			if lookups > manage.reqLimit {
				writeJSONResponse(respWriter, model.ErrorResponse{Error: "rate limited"}, http.StatusTooManyRequests)

				return
			}
		} else {
			log.Println("Manage Handler MiddlewareRateLimit didn't make a cache lookup - empty username")
		}

		next(respWriter, request, routerParams)
	}
}
