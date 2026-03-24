package handlers

import (
	"calendar/internal/auth/jwt"
	"calendar/internal/domain/user"
	"calendar/internal/web/dto"
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	service AuthService
}

type AuthService interface {
	Login(ctx context.Context, login, password string) (*jwt.AuthTokens, error)
	Register(ctx context.Context, login, password, email, telegram string) (*user.User, error)
	RefreshTokens(tokenStr string) (*jwt.AuthTokens, error)
	ValidateTokens(tokenStr string) (*jwt.Payload, error)
}

func NewUserHandler(service AuthService) *UserHandler {
	return &UserHandler{
		service: service,
	}
}

// Auth returns the underlying AuthService for use in middleware.
func (h *UserHandler) Auth() AuthService {
	return h.service
}

const (
	CtxUserID = "user_id"
)

// @Summary Register a new user
// @Description Registers a new user with the provided data
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "Registration data"
// @Success 200 {object} dto.RegisterResponse
// @Failure 400 {object} map[string]string{} "Error message"
// @Failure 409 {object} map[string]string{} "Error message"
// @Failure 500 {object} map[string]string{} "Error message"
// @Router /api/v1/auth/register [post]
func (h *UserHandler) Register(ctx *gin.Context) {
	var req dto.RegisterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	usr, err := h.service.Register(ctx.Request.Context(), req.Login, req.Password, req.Email, req.Telegram)
	if err != nil {
		if isValidationError(err) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if isConflict(err) {
			ctx.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	res := dto.RegisterResponse{
		ID:       usr.ID.String(),
		Login:    usr.Login,
		Email:    usr.Email,
		Telegram: usr.Telegram,
	}
	ctx.JSON(http.StatusOK, res)
}

// @Summary User login
// @Description Authenticates the user and returns JWT tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Login credentials"
// @Success 200 {object} dto.AuthTokens
// @Failure 400 {object} map[string]string{} "Error message"
// @Failure 401 {object} map[string]string{} "Error message"
// @Router /api/v1/auth/login [post]
func (h *UserHandler) Login(ctx *gin.Context) {
	var req dto.LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	jwtResp, err := h.service.Login(ctx.Request.Context(), req.Login, req.Password)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	res := dto.AuthTokens{
		AccessToken:  jwtResp.AccessToken,
		RefreshToken: jwtResp.RefreshToken,
	}
	ctx.JSON(http.StatusOK, res)
}

// @Summary Refresh tokens
// @Description Refreshes the access/refresh token pair using a refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.TokenRefreshRequest true "Refresh token"
// @Success 200 {object} dto.AuthTokens
// @Failure 400 {object} map[string]string{} "Error message"
// @Failure 401 {object} map[string]string{} "Error message"
// @Router /api/v1/auth/refresh-token [post]
func (h *UserHandler) RefreshToken(ctx *gin.Context) {
	var req dto.TokenRefreshRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	jwtResp, err := h.service.RefreshTokens(req.RefreshToken)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	res := dto.AuthTokens{
		AccessToken:  jwtResp.AccessToken,
		RefreshToken: jwtResp.RefreshToken,
	}
	ctx.JSON(http.StatusOK, res)
}
