package auth

type RegisterDTO struct {
	Email    string
	Password string
	Location string
	Language string
}

type AuthDTO struct {
	Email    string
	Password string
}
