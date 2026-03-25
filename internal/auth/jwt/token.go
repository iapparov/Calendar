package jwt

type AuthTokens struct {
	AccessToken      string
	RefreshToken     string
	AccessExpiresAt  int64
	RefreshExpiresAt int64
	TokenType        string
}

type Payload struct {
	UserID string
}
