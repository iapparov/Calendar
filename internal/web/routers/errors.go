package routers

type authError struct {
	Error string `json:"error"`
}

var (
	errAuthMissing  = &authError{Error: "authorization header missing"}
	errAuthScheme   = &authError{Error: "invalid authorization scheme"}
	errTokenEmpty   = &authError{Error: "token is empty"}
	errTokenInvalid = &authError{Error: "invalid or expired token"}
)
