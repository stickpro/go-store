package auth_response

// AuthResponse carries the bearer token issued after a successful login or code
// verification.
type AuthResponse struct {
	Token string `json:"token"`
} //	@name	AuthResponse

// SendCodeResponse is returned from the code request step. It is intentionally
// contentless (always {"sent": true}) so callers can't distinguish existing from
// new or admin accounts.
type SendCodeResponse struct {
	Sent bool `json:"sent"`
} //	@name	SendCodeResponse
