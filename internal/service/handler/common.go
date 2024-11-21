package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/eldarbr/go-auth/internal/model"
	"github.com/julienschmidt/httprouter"
)

type CacheImpl interface {
	GetAndIncrease(key string) int
}

const (
	defaultRateLimiterIPSourceHeader = "X-Real-IP"
)

func writeJSONResponse(responseWriter http.ResponseWriter, response any, code int) {
	responseWriter.Header().Set("Content-Type", "application/json")

	resp, marshalErr := json.Marshal(response)
	if marshalErr != nil {
		responseWriter.WriteHeader(http.StatusInternalServerError)
		responseWriter.Write([]byte("{\"error\": \"response marshal error\"}")) //nolint:errcheck // won't check.

		return
	}

	responseWriter.WriteHeader(code)
	responseWriter.Write(resp) //nolint:errcheck // won't check.
}

type CommonHandl struct{}

func (CommonHandl) MethodNotAllowed(w http.ResponseWriter, _ *http.Request) {
	writeJSONResponse(w, model.ErrorResponse{Error: "method not allowed"}, http.StatusMethodNotAllowed)
}

func (CommonHandl) NotFound(w http.ResponseWriter, _ *http.Request) {
	writeJSONResponse(w, model.ErrorResponse{Error: "not found"}, http.StatusNotFound)
}

type RateLimitHandl struct {
	cache  CacheImpl
	Config RateLimitHandlConfig
}

type RateLimitHandlConfig struct {
	IPSourceHeader string
	Requests       int
}

func NewIPRateLimitHandl(requests int, ipCache CacheImpl) *RateLimitHandl {
	return &RateLimitHandl{
		cache: ipCache,
		Config: RateLimitHandlConfig{
			IPSourceHeader: defaultRateLimiterIPSourceHeader, // set defaults.
			Requests:       requests,
		},
	}
}

func (ratelimit RateLimitHandl) MiddlewareIPRateLimit(next httprouter.Handle) httprouter.Handle {
	return func(respWriter http.ResponseWriter, request *http.Request, routerParams httprouter.Params) {
		if ratelimit.cache == nil {
			log.Println("MiddlewareRateLimit uninitialized cache")
			writeJSONResponse(respWriter, model.ErrorResponse{Error: "internal error"}, http.StatusInternalServerError)

			return
		}

		ip := request.Header.Get(ratelimit.Config.IPSourceHeader)
		if ip != "" {
			ipRequests := ratelimit.cache.GetAndIncrease("ip:" + ip)
			if ipRequests > ratelimit.Config.Requests {
				writeJSONResponse(respWriter, model.ErrorResponse{Error: "rate limited"}, http.StatusTooManyRequests)

				return
			}
		} else {
			log.Println("ratelimit Handler MiddlewareIPRateLimit didn't make a cache lookup - empty ip")
		}

		next(respWriter, request, routerParams)
	}
}
