package user

import (
	"calendar/internal/auth/jwt"
	"calendar/internal/config"
	"calendar/internal/domain"
	"calendar/internal/domain/user"
	"calendar/internal/logger"
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"unicode"
	"unicode/utf8"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repo   StorageService
	jwt    JwtAuthService
	cfg    *config.App
	logger *logger.Service
}

type JwtAuthService interface {
	GenerateTokens(user *user.User) (*jwt.AuthTokens, error)
	ValidateTokens(tokenStr string) (*jwt.Payload, error)
	RefreshTokens(refreshToken string) (*jwt.AuthTokens, error)
}

type StorageService interface {
	GetUser(login string, ctx context.Context) (*user.User, error)
	SaveUser(user *user.User, ctx context.Context) error
}

func NewService(repo StorageService, jwt JwtAuthService, cfg *config.App, logger *logger.Service) *Service {
	return &Service{
		repo:   repo,
		jwt:    jwt,
		cfg:    cfg,
		logger: logger,
	}
}

func (s *Service) Login(login, password string, ctx context.Context) (*jwt.AuthTokens, error) {
	if login == "" || password == "" {
		s.logger.Log(zapcore.DebugLevel, "empty login or password")
		return nil, domain.ErrEmptyField
	}

	usr, err := s.repo.GetUser(login, ctx)
	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword(usr.Password, []byte(password))
	if err != nil {
		s.logger.Log(zapcore.DebugLevel, "invalid password", zap.Error(err))
		return nil, err
	}

	jwtresp, err := s.jwt.GenerateTokens(usr)
	if err != nil {
		s.logger.Log(zapcore.ErrorLevel, "jwt Generate Tokens Error", zap.Error(err))
		return nil, err
	}
	return jwtresp, nil
}

func (s *Service) Register(login, password, email, telegram string, ctx context.Context) (*user.User, error) {
	if err := s.isValidlogin(login); err != nil {
		s.logger.Log(zapcore.DebugLevel, "invalid login", zap.Error(err))
		return nil, err
	}

	if err := s.isValidpassword(password); err != nil {
		s.logger.Log(zapcore.DebugLevel, "invalid password", zap.Error(err))
		return nil, err
	}

	if err := s.isValidTelegramChatId(telegram); err != nil {
		s.logger.Log(zapcore.DebugLevel, "invalid telegram chat id", zap.Error(err))
		return nil, err
	}

	if err := s.isValidEmail(email); err != nil {
		s.logger.Log(zapcore.DebugLevel, "invalid email", zap.Error(err))
		return nil, err
	}

	ch, err := s.repo.GetUser(login, ctx)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		s.logger.Log(zapcore.ErrorLevel, "error checking existing user", zap.Error(err))
		return nil, err
	}

	if ch != nil {
		s.logger.Log(zapcore.DebugLevel, "user with this login already exists")
		return nil, domain.ErrUserAlreadyExists
	}

	usr, err := user.NewUser(login, password, email, telegram)

	if err != nil {
		s.logger.Log(zapcore.ErrorLevel, "user creation error", zap.Error(err))
		return nil, err
	}

	err = s.repo.SaveUser(usr, ctx)

	if err != nil {
		s.logger.Log(zapcore.ErrorLevel, "save user error", zap.Error(err))
		return nil, err
	}
	return usr, nil
}

func (s *Service) RefreshTokens(refreshToken string) (*jwt.AuthTokens, error) {
	tkns, err := s.jwt.RefreshTokens(refreshToken)
	if err != nil {
		s.logger.Log(zapcore.DebugLevel, "jwt Refresh Tokens Error", zap.Error(err))
		return nil, err
	}
	s.logger.Log(zapcore.DebugLevel, "tokens refreshed successfully")
	return tkns, nil
}

func (s *Service) ValidateTokens(tokenStr string) (*jwt.Payload, error) {
	payload, err := s.jwt.ValidateTokens(tokenStr)
	if err != nil {
		s.logger.Log(zapcore.DebugLevel, "jwt Validate Tokens Error", zap.Error(err))
		return nil, err
	}
	s.logger.Log(zapcore.DebugLevel, "tokens validated successfully")
	return payload, nil
}

func (s *Service) isValidlogin(login string) error {
	cfg := s.cfg.UserValidation
	if utf8.RuneCountInString(login) < cfg.MinLength || utf8.RuneCountInString(login) > cfg.MaxLength {
		return fmt.Errorf("%w: must be between %d and %d characters", domain.ErrInvalidLogin, cfg.MinLength, cfg.MaxLength)
	}

	escapedChars := regexp.QuoteMeta(cfg.AllowedCharacters)
	loginRegexp := regexp.MustCompile(`^[` + escapedChars + `]+$`)
	if !loginRegexp.MatchString(login) {
		return fmt.Errorf("%w: must contain only letters, digits, underscores, or hyphens and must not contain spaces", domain.ErrInvalidLogin)
	}
	return nil
}

func (s *Service) isValidpassword(password string) error {
	cfg := s.cfg.PasswordValidation

	l := utf8.RuneCountInString(password)
	if l < cfg.MinLength || l > cfg.MaxLength {
		return fmt.Errorf(
			"%w: must be %d–%d characters",
			domain.ErrInvalidPassword,
			cfg.MinLength, cfg.MaxLength,
		)
	}

	if !utf8.ValidString(password) {
		return fmt.Errorf("%w: contains invalid UTF-8 characters", domain.ErrInvalidPassword)
	}

	var hasUpper, hasLower, hasDigit bool

	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		}
	}

	if cfg.RequireUpper && !hasUpper {
		return fmt.Errorf("%w: must contain at least one uppercase letter", domain.ErrInvalidPassword)
	}
	if cfg.RequireLower && !hasLower {
		return fmt.Errorf("%w: must contain at least one lowercase letter", domain.ErrInvalidPassword)
	}
	if cfg.RequireDigit && !hasDigit {
		return fmt.Errorf("%w: must contain at least one digit", domain.ErrInvalidPassword)
	}

	return nil
}

func (s *Service) isValidTelegramChatId(chatId string) error {
	if chatId == "" {
		return nil
	}
	_, err := strconv.Atoi(chatId)
	if err != nil {
		return fmt.Errorf("%w: chat_id must be digit", domain.ErrInvalidTelegramChatID)
	}

	return nil
}

func (s *Service) isValidEmail(email string) error {
	l := utf8.RuneCountInString(email)
	if l < 6 { // 1@1.ru от 6 символов
		return fmt.Errorf("%w: must be at least 6 characters", domain.ErrInvalidEmail)
	}

	re := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`) // для базовой проверки формата asda@asdasd.asdads
	if !re.MatchString(email) {
		return fmt.Errorf("%w: invalid email format", domain.ErrInvalidEmail)
	}

	return nil
}
