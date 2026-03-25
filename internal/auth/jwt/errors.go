package jwt

import "errors"

var (
	errUnexpectedSigningMethod = errors.New("unexpected signing method")
	errInvalidAccessToken      = errors.New("invalid access token")
	errInvalidRefreshToken     = errors.New("invalid refresh token")
	errInvalidTokenPayload     = errors.New("invalid token payload")
	errTokenExpired            = errors.New("token has expired")
	errRefreshTokenExpired     = errors.New("refresh token has expired")
	errInvalidClaims           = errors.New("invalid claims")
	ErrInvalidToken            = errors.New("invalid token")
)
