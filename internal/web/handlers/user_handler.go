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
	Service AuthService
}

type AuthService interface {
	Login(login, password string, ctx context.Context) (*jwt.AuthTokens, error)
	Register(login, password, email, telegram string, ctx context.Context) (*user.User, error)
	RefreshTokens(tokenStr string) (*jwt.AuthTokens, error)
	ValidateTokens(tokenStr string) (*jwt.Payload, error)
}

func NewUserHandler(service AuthService) *UserHandler {
	return &UserHandler{
		Service: service,
	}
}

const (
	CtxUserID = "user_id"
)

// @Summary Регистрация нового пользователя
// @Description Регистрирует нового пользователя с указанными данными
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "Данные для регистрации"
// @Success 200 {object} dto.RegisterResponse
// @Failure 400 {object} map[string]string{} "Сообщение об ошибке"
// @Failure 409 {object} map[string]string{} "Сообщение об ошибке"
// @Failure 500 {object} map[string]string{} "Сообщение об ошибке"
// @Router /api/v1/auth/register [post]
func (h *UserHandler) Register(ctx *gin.Context) {
	var req dto.RegisterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	usr, err := h.Service.Register(req.Login, req.Password, req.Email, req.Telegram, ctx.Request.Context())
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
		ID:       usr.Id.String(),
		Login:    usr.Login,
		Email:    usr.Email,
		Telegram: usr.Telegram,
	}
	ctx.JSON(http.StatusOK, res)
}

// @Summary Авторизация пользователя
// @Description Выполняет вход пользователя и возвращает JWT токены
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Логин и пароль"
// @Success 200 {object} dto.AuthTokens
// @Failure 400 {object} map[string]string{} "Сообщение об ошибке"
// @Failure 401 {object} map[string]string{} "Сообщение об ошибке"
// @Router /api/v1/auth/login [post]
func (h *UserHandler) Login(ctx *gin.Context) {
	var req dto.LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	jwtResp, err := h.Service.Login(req.Login, req.Password, ctx.Request.Context())
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

// @Summary Обновление токенов
// @Description Обновляет пару access/refresh токенов по refresh токену
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.TokenRefreshRequest true "Refresh токен"
// @Success 200 {object} dto.AuthTokens
// @Failure 400 {object} map[string]string{} "Сообщение об ошибке"
// @Failure 401 {object} map[string]string{} "Сообщение об ошибке"
// @Router /api/v1/auth/refresh-token [post]
func (h *UserHandler) RefreshToken(ctx *gin.Context) {
	var req dto.TokenRefreshRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	jwtResp, err := h.Service.RefreshTokens(req.RefreshToken)
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
