package auth_request

// SendCodeRequest starts passwordless auth: a one-time code is emailed to Email.
type SendCodeRequest struct {
	Email string `json:"email" validate:"required,email"`
} //	@name	SendCodeRequest

// VerifyCodeRequest completes passwordless auth with the code from the email.
type VerifyCodeRequest struct {
	Email string `json:"email" validate:"required,email"`
	Code  string `json:"code" validate:"required,len=6,number"`
} //	@name	VerifyCodeRequest
