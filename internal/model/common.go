package model

type ErrorResponse struct {
	Error string `json:"error"`
}

type AuthResponse struct {
	Token string `json:"token"`
}
