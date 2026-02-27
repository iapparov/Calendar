package dto

type LoginRequest struct {
	Login    string `json:"login" binding:"required" example:"john_doe"`
	Password string `json:"password" binding:"required" example:"securePassword123"`
}

type RegisterRequest struct {
	Login    string `json:"login" binding:"required" example:"john_doe"`
	Telegram string `json:"telegram_chat_id" example:"123456789"`
	Email    string `json:"email" binding:"required" example:"john@example.com"`
	Password string `json:"password" binding:"required" example:"securePassword123"`
}

type RegisterResponse struct {
	ID       string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Login    string `json:"login" example:"john_doe"`
	Telegram string `json:"telegram_chat_id" example:"123456789"`
	Email    string `json:"email" example:"john@example.com"`
}

type AuthTokens struct {
	AccessToken  string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIs..."`
	RefreshToken string `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIs..."`
}

type TokenRefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required" example:"eyJhbGciOiJIUzI1NiIs..."`
}
