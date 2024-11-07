package model

type CraeteUserRequest struct {
	Auth    AuthResponse `json:"authorization"`
	NewUser UserCreds    `json:"user"`
}

type UserGroup struct {
	Username    string `json:"username"`
	UserRole    string `json:"userRole"`
	ServiceName string `json:"serviceName"`
}
