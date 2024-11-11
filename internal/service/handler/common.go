package handler

import (
	"encoding/json"
	"net/http"

	"github.com/eldarbr/go-auth/internal/model"
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
