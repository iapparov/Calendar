package jwt

import (
	"calendar/internal/config"
	"calendar/internal/domain/user"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func createTestService() *Service {
	return &Service{
		accessSecretBytes:  []byte("test-access-secret-key-12345"),
		refreshSecretBytes: []byte("test-refresh-secret-key-12345"),
		expAccessToken:     15,
		expRefreshToken:    24,
	}
}

func createTestUser() *user.User {
	return &user.User{
		ID:    uuid.New(),
		Login: "testuser",
	}
}

func TestNewService(t *testing.T) {
	cfg := &config.App{
		Jwt: config.Jwt{
			AccessSecret:    "access-secret",
			RefreshSecret:   "refresh-secret",
			ExpAccessToken:  30,
			ExpRefreshToken: 48,
		},
	}

	service := NewService(cfg)

	if string(service.accessSecretBytes) != "access-secret" {
		t.Errorf("expected accessSecretBytes 'access-secret', got '%s'", service.accessSecretBytes)
	}
	if string(service.refreshSecretBytes) != "refresh-secret" {
		t.Errorf("expected refreshSecretBytes 'refresh-secret', got '%s'", service.refreshSecretBytes)
	}
	if service.expAccessToken != 30 {
		t.Errorf("expected expAccessToken 30, got %d", service.expAccessToken)
	}
	if service.expRefreshToken != 48 {
		t.Errorf("expected expRefreshToken 48, got %d", service.expRefreshToken)
	}
}

func TestGenerateTokens_Success(t *testing.T) {
	service := createTestService()
	u := createTestUser()

	tokens, err := service.GenerateTokens(u)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tokens == nil {
		t.Fatal("tokens should not be nil")
	}
	if tokens.AccessToken == "" {
		t.Error("access token should not be empty")
	}
	if tokens.RefreshToken == "" {
		t.Error("refresh token should not be empty")
	}
	if tokens.AccessToken == tokens.RefreshToken {
		t.Error("access and refresh tokens should be different")
	}
}

func TestValidateTokens_Success(t *testing.T) {
	service := createTestService()
	u := createTestUser()

	tokens, _ := service.GenerateTokens(u)
	payload, err := service.ValidateTokens(tokens.AccessToken)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payload == nil {
		t.Fatal("payload should not be nil")
	}
	if payload.UserID != u.ID.String() {
		t.Errorf("expected user ID '%s', got '%s'", u.ID.String(), payload.UserID)
	}
}

func TestValidateTokens_InvalidToken(t *testing.T) {
	service := createTestService()

	_, err := service.ValidateTokens("invalid-token")

	if err == nil {
		t.Error("expected error for invalid token")
	}
}

func TestValidateTokens_WrongSecret(t *testing.T) {
	service1 := createTestService()
	service2 := &Service{
		accessSecretBytes:  []byte("different-secret"),
		refreshSecretBytes: []byte("different-secret"),
		expAccessToken:     15,
		expRefreshToken:    24,
	}

	u := createTestUser()
	tokens, _ := service1.GenerateTokens(u)

	_, err := service2.ValidateTokens(tokens.AccessToken)

	if err == nil {
		t.Error("expected error when validating with wrong secret")
	}
}

func TestValidateTokens_ExpiredToken(t *testing.T) {
	service := &Service{
		accessSecretBytes:  []byte("test-secret"),
		refreshSecretBytes: []byte("test-secret"),
		expAccessToken:     -1, // Already expired
		expRefreshToken:    24,
	}

	u := createTestUser()
	tokens, _ := service.GenerateTokens(u)

	_, err := service.ValidateTokens(tokens.AccessToken)

	if err == nil {
		t.Error("expected error for expired token")
	}
}

func TestRefreshTokens_Success(t *testing.T) {
	service := createTestService()
	u := createTestUser()

	originalTokens, _ := service.GenerateTokens(u)
	newTokens, err := service.RefreshTokens(originalTokens.RefreshToken)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if newTokens == nil {
		t.Fatal("new tokens should not be nil")
	}
	if newTokens.AccessToken == "" {
		t.Error("new access token should not be empty")
	}
	if newTokens.RefreshToken == "" {
		t.Error("new refresh token should not be empty")
	}
}

func TestRefreshTokens_InvalidToken(t *testing.T) {
	service := createTestService()

	_, err := service.RefreshTokens("invalid-refresh-token")

	if err == nil {
		t.Error("expected error for invalid refresh token")
	}
}

func TestRefreshTokens_AccessTokenAsRefresh(t *testing.T) {
	service := createTestService()
	u := createTestUser()

	tokens, _ := service.GenerateTokens(u)

	// Try to use access token as refresh token (should fail due to different secret)
	_, err := service.RefreshTokens(tokens.AccessToken)

	if err == nil {
		t.Error("should not accept access token as refresh token")
	}
}

func TestValidateTokens_AlgorithmConfusion(t *testing.T) {
	service := createTestService()
	u := createTestUser()

	// Create a token with "none" algorithm
	claims := jwt.MapClaims{
		"uuid": u.ID.String(),
		"exp":  time.Now().Add(time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	tokenStr, _ := token.SignedString(jwt.UnsafeAllowNoneSignatureType)

	_, err := service.ValidateTokens(tokenStr)

	if err == nil {
		t.Error("should reject tokens with 'none' algorithm")
	}
}

func TestGenerateTokens_ContainsExpectedClaims(t *testing.T) {
	service := createTestService()
	testUser := createTestUser()

	tokens, _ := service.GenerateTokens(testUser)

	// Parse the token to check claims
	parser := jwt.NewParser()
	token, _, err := parser.ParseUnverified(tokens.AccessToken, jwt.MapClaims{})
	if err != nil {
		t.Fatalf("failed to parse token: %v", err)
	}

	claims := token.Claims.(jwt.MapClaims)

	if claims["uuid"] != testUser.ID.String() {
		t.Errorf("expected uuid '%s', got '%v'", testUser.ID.String(), claims["uuid"])
	}

	if _, ok := claims["exp"]; !ok {
		t.Error("token should contain 'exp' claim")
	}
}

func TestAuthTokens_Struct(t *testing.T) {
	tokens := AuthTokens{
		AccessToken:      "access",
		RefreshToken:     "refresh",
		AccessExpiresAt:  123,
		RefreshExpiresAt: 456,
		TokenType:        "Bearer",
	}

	if tokens.AccessToken != "access" {
		t.Error("AccessToken mismatch")
	}
	if tokens.RefreshToken != "refresh" {
		t.Error("RefreshToken mismatch")
	}
}

func TestPayload_Struct(t *testing.T) {
	payload := Payload{
		UserID: "test-user-id",
	}

	if payload.UserID != "test-user-id" {
		t.Error("UserID mismatch")
	}
}
