package jwt

import (
	"calendar/internal/config"
	"calendar/internal/domain/user"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Service struct {
	accessSecretBytes  []byte
	refreshSecretBytes []byte
	expAccessToken     int // in minutes
	expRefreshToken    int // in minutes
}

func NewService(cfg *config.App) *Service {
	return &Service{
		accessSecretBytes:  []byte(cfg.Jwt.AccessSecret),
		refreshSecretBytes: []byte(cfg.Jwt.RefreshSecret),
		expAccessToken:     cfg.Jwt.ExpAccessToken,
		expRefreshToken:    cfg.Jwt.ExpRefreshToken,
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
	claims, err := s.validateToken(tokenStr, s.accessSecretBytes, errInvalidAccessToken, errTokenExpired)
	if err != nil {
		return nil, err
	}

	uuidStr, ok := claims["uuid"].(string)
	if !ok {
		return nil, errInvalidTokenPayload
	}

	return &Payload{UserID: uuidStr}, nil
}

func (s *Service) RefreshTokens(refreshToken string) (*AuthTokens, error) {
	claims, err := s.validateToken(refreshToken, s.refreshSecretBytes, errInvalidRefreshToken, errRefreshTokenExpired)
	if err != nil {
		return nil, err
	}

	uuidStr, ok := claims["uuid"].(string)
	if !ok {
		return nil, errInvalidTokenPayload
	}

	u := &user.User{ID: uuid.MustParse(uuidStr)}

	return s.GenerateTokens(u)
}

func (s *Service) generateAccessToken(u *user.User) (string, error) {
	claims := jwt.MapClaims{
		"uuid": u.ID.String(),
		"exp":  time.Now().Add(time.Minute * time.Duration(s.expAccessToken)).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.accessSecretBytes)
}

func (s *Service) generateRefreshToken(u *user.User) (string, error) {
	claims := jwt.MapClaims{
		"uuid": u.ID.String(),
		"exp":  time.Now().Add(time.Minute * time.Duration(s.expRefreshToken)).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.refreshSecretBytes)
}

func (s *Service) validateToken(tokenStr string, secret []byte, errInvalid, errExpired error) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errUnexpectedSigningMethod
		}
		return secret, nil
	})
	if err != nil || !token.Valid {
		return nil, errInvalid
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errInvalidClaims
	}

	if exp, ok := claims["exp"].(float64); !ok || time.Unix(int64(exp), 0).Before(time.Now()) {
		return nil, errExpired
	}

	return claims, nil
}
