package auth_response

type RegisterUserResponse struct {
	Token string `json:"token"`
} // @name RegisterUserResponse

type AuthResponse struct {
	Token string `json:"token"`
} // @name AuthResponse
