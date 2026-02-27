package jwt

import (
	"calendar/internal/config"
	"calendar/internal/domain/user"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Service struct {
	accessSecret    string
	refreshSecret   string
	ExpAccessToken  int // в минутах
	ExpRefreshToken int // в часах
}

func NewService(cfg *config.App) *Service {
	return &Service{
		accessSecret:    cfg.Jwt.AccessSecret,
		refreshSecret:   cfg.Jwt.RefreshSecret,
		ExpAccessToken:  cfg.Jwt.ExpAccessToken,
		ExpRefreshToken: cfg.Jwt.ExpRefreshToken,
	}
}

func (s *Service) GenerateTokens(u *user.User) (*AuthTokens, error) {
	access, err := s.generateAccessToken(u)
	if err != nil {
		return nil, err
	}
	refresh, err := s.generateRefreshToken(u)
	if err != nil {
		return nil, err
	}
	return &AuthTokens{
		AccessToken:  access,
		RefreshToken: refresh,
	}, nil
}

func (s *Service) ValidateTokens(tokenStr string) (*Payload, error) {
	claims, err := s.validateAccessToken(tokenStr)
	if err != nil {
		return nil, err
	}

	uuidStr, ok := claims["uuid"].(string)
	if !ok {
		return nil, errors.New("invalid token payload")
	}

	return &Payload{UserID: uuidStr}, nil
}

func (s *Service) RefreshTokens(refreshToken string) (*AuthTokens, error) {
	claims, err := s.validateRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	uuidStr, ok := claims["uuid"].(string)
	if !ok {
		return nil, errors.New("invalid refresh token payload")
	}

	u := &user.User{Id: uuid.MustParse(uuidStr)}

	return s.GenerateTokens(u)
}

func (s *Service) generateAccessToken(u *user.User) (string, error) {
	claims := jwt.MapClaims{
		"uuid": u.Id.String(),
		"exp":  time.Now().Add(time.Minute * time.Duration(s.ExpAccessToken)).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.accessSecret))
}

func (s *Service) generateRefreshToken(u *user.User) (string, error) {
	claims := jwt.MapClaims{
		"uuid": u.Id.String(),
		"exp":  time.Now().Add(time.Hour * time.Duration(s.ExpRefreshToken)).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.refreshSecret))
}

func (s *Service) validateAccessToken(tokenStr string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.accessSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid access token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid claims")
	}

	if exp, ok := claims["exp"].(float64); !ok || time.Unix(int64(exp), 0).Before(time.Now()) {
		return nil, errors.New("token has expired")
	}

	return claims, nil
}

func (s *Service) validateRefreshToken(tokenStr string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.refreshSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid refresh token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid claims")
	}

	if exp, ok := claims["exp"].(float64); !ok || time.Unix(int64(exp), 0).Before(time.Now()) {
		return nil, errors.New("refresh token has expired")
	}

	return claims, nil
}
