package handlers

import (
	"bytes"
	"calendar/internal/auth/jwt"
	"calendar/internal/domain"
	"calendar/internal/domain/user"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
)

func TestNewUserHandler(t *testing.T) {
	mock := &mockAuthService{}
	handler := NewUserHandler(mock)

	if handler == nil {
		t.Fatal("handler should not be nil")
	}
	if handler.service == nil {
		t.Error("service should not be nil")
	}
}

func TestLogin_Success(t *testing.T) {
	expectedTokens := &jwt.AuthTokens{
		AccessToken:  "access-token",
		RefreshToken: "refresh-token",
	}

	mock := &mockAuthService{
		loginFunc: func(ctx context.Context, login, password string) (*jwt.AuthTokens, error) {
			return expectedTokens, nil
		},
	}

	handler := NewUserHandler(mock)
	router := setupTestRouter()
	router.POST("/login", handler.Login)

	body := `{"login":"testuser","password":"password123"}`
	req, _ := http.NewRequest("POST", "/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var result map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &result)
	if err != nil {
		t.Errorf("failed to unmarshal response: %v", err)
	}

	if result["access_token"] != "access-token" {
		t.Errorf("expected access_token 'access-token', got '%s'", result["access_token"])
	}
	if result["refresh_token"] != "refresh-token" {
		t.Errorf("expected refresh_token 'refresh-token', got '%s'", result["refresh_token"])
	}
}

func TestLogin_InvalidCredentials(t *testing.T) {
	mock := &mockAuthService{
		loginFunc: func(ctx context.Context, login, password string) (*jwt.AuthTokens, error) {
			return nil, errors.New("invalid credentials")
		},
	}

	handler := NewUserHandler(mock)
	router := setupTestRouter()
	router.POST("/login", handler.Login)

	body := `{"login":"testuser","password":"wrongpassword"}`
	req, _ := http.NewRequest("POST", "/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestLogin_InvalidJSON(t *testing.T) {
	mock := &mockAuthService{}
	handler := NewUserHandler(mock)
	router := setupTestRouter()
	router.POST("/login", handler.Login)

	body := `{"login":}`
	req, _ := http.NewRequest("POST", "/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestRegister_Success(t *testing.T) {
	expectedUser := &user.User{
		ID:       uuid.New(),
		Login:    "newuser",
		Email:    "new@example.com",
		Telegram: "123456789",
	}

	mock := &mockAuthService{
		registerFunc: func(ctx context.Context, login, password, email, telegram string) (*user.User, error) {
			return expectedUser, nil
		},
	}

	handler := NewUserHandler(mock)
	router := setupTestRouter()
	router.POST("/register", handler.Register)

	body := `{"login":"newuser","password":"Password123","email":"new@example.com","telegram_chat_id":"123456789"}`
	req, _ := http.NewRequest("POST", "/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var result map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &result)
	if err != nil {
		t.Errorf("failed to unmarshal response: %v", err)
	}

	if result["login"] != "newuser" {
		t.Errorf("expected login 'newuser', got '%s'", result["login"])
	}
}

func TestRegister_ValidationError(t *testing.T) {
	mock := &mockAuthService{
		registerFunc: func(ctx context.Context, login, password, email, telegram string) (*user.User, error) {
			return nil, domain.ErrInvalidLogin
		},
	}

	handler := NewUserHandler(mock)
	router := setupTestRouter()
	router.POST("/register", handler.Register)

	body := `{"login":"x","password":"Password123","email":"new@example.com"}`
	req, _ := http.NewRequest("POST", "/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestRegister_UserAlreadyExists(t *testing.T) {
	mock := &mockAuthService{
		registerFunc: func(ctx context.Context, login, password, email, telegram string) (*user.User, error) {
			return nil, domain.ErrUserAlreadyExists
		},
	}

	handler := NewUserHandler(mock)
	router := setupTestRouter()
	router.POST("/register", handler.Register)

	body := `{"login":"existinguser","password":"Password123","email":"existing@example.com"}`
	req, _ := http.NewRequest("POST", "/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("expected status %d, got %d", http.StatusConflict, w.Code)
	}
}

func TestRefreshToken_Success(t *testing.T) {
	expectedTokens := &jwt.AuthTokens{
		AccessToken:  "new-access-token",
		RefreshToken: "new-refresh-token",
	}

	mock := &mockAuthService{
		refreshTokensFunc: func(tokenStr string) (*jwt.AuthTokens, error) {
			return expectedTokens, nil
		},
	}

	handler := NewUserHandler(mock)
	router := setupTestRouter()
	router.POST("/refresh-token", handler.RefreshToken)

	body := `{"refresh_token":"old-refresh-token"}`
	req, _ := http.NewRequest("POST", "/refresh-token", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var result map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &result)
	if err != nil {
		t.Errorf("failed to unmarshal response: %v", err)
	}

	if result["access_token"] != "new-access-token" {
		t.Errorf("expected access_token 'new-access-token', got '%s'", result["access_token"])
	}
}

func TestRefreshToken_InvalidToken(t *testing.T) {
	mock := &mockAuthService{
		refreshTokensFunc: func(tokenStr string) (*jwt.AuthTokens, error) {
			return nil, errors.New("invalid token")
		},
	}

	handler := NewUserHandler(mock)
	router := setupTestRouter()
	router.POST("/refresh-token", handler.RefreshToken)

	body := `{"refresh_token":"invalid-token"}`
	req, _ := http.NewRequest("POST", "/refresh-token", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestRefreshToken_MissingToken(t *testing.T) {
	mock := &mockAuthService{}
	handler := NewUserHandler(mock)
	router := setupTestRouter()
	router.POST("/refresh-token", handler.RefreshToken)

	body := `{}`
	req, _ := http.NewRequest("POST", "/refresh-token", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestCtxUserID_Constant(t *testing.T) {
	if CtxUserID != "user_id" {
		t.Errorf("expected CtxUserID to be 'user_id', got '%s'", CtxUserID)
	}
}
