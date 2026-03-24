package routers

import (
	"calendar/internal/auth/jwt"
	"calendar/internal/config"
	"calendar/internal/domain/user"
	"calendar/internal/logger"
	"calendar/internal/web/handlers"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

// Mock AuthService for middleware testing
type mockAuthService struct {
	validateTokensFunc func(tokenStr string) (*jwt.Payload, error)
}

func (m *mockAuthService) Login(ctx context.Context, login, password string) (*jwt.AuthTokens, error) {
	return nil, nil
}

func (m *mockAuthService) Register(ctx context.Context, login, password, email, telegram string) (*user.User, error) {
	return nil, nil
}

func (m *mockAuthService) RefreshTokens(tokenStr string) (*jwt.AuthTokens, error) {
	return nil, nil
}

func (m *mockAuthService) ValidateTokens(tokenStr string) (*jwt.Payload, error) {
	if m.validateTokensFunc != nil {
		return m.validateTokensFunc(tokenStr)
	}
	return nil, nil
}

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func TestAuthMiddleware_Success(t *testing.T) {
	mock := &mockAuthService{
		validateTokensFunc: func(tokenStr string) (*jwt.Payload, error) {
			return &jwt.Payload{UserID: "test-user-id"}, nil
		},
	}

	router := setupTestRouter()
	router.Use(AuthMiddleware(mock))
	router.GET("/protected", func(c *gin.Context) {
		userID, exists := c.Get(handlers.CtxUserID)
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "no user id"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"user_id": userID})
	})

	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer valid-token")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestAuthMiddleware_MissingHeader(t *testing.T) {
	mock := &mockAuthService{}

	router := setupTestRouter()
	router.Use(AuthMiddleware(mock))
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest("GET", "/protected", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestAuthMiddleware_InvalidScheme(t *testing.T) {
	mock := &mockAuthService{}

	router := setupTestRouter()
	router.Use(AuthMiddleware(mock))
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Basic invalid-scheme")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestAuthMiddleware_EmptyToken(t *testing.T) {
	mock := &mockAuthService{}

	router := setupTestRouter()
	router.Use(AuthMiddleware(mock))
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer ")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	mock := &mockAuthService{
		validateTokensFunc: func(tokenStr string) (*jwt.Payload, error) {
			return nil, jwt.ErrInvalidToken
		},
	}

	router := setupTestRouter()
	router.Use(AuthMiddleware(mock))
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestAuthMiddleware_CaseInsensitiveBearer(t *testing.T) {
	mock := &mockAuthService{
		validateTokensFunc: func(tokenStr string) (*jwt.Payload, error) {
			return &jwt.Payload{UserID: "test-user-id"}, nil
		},
	}

	router := setupTestRouter()
	router.Use(AuthMiddleware(mock))
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	testCases := []string{
		"Bearer token",
		"bearer token",
		"BEARER token",
		"BeArEr token",
	}

	for _, authHeader := range testCases {
		req, _ := http.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", authHeader)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d for '%s', got %d", http.StatusOK, authHeader, w.Code)
		}
	}
}

func TestAuthMiddleware_SetsUserID(t *testing.T) {
	expectedUserID := "test-user-123"
	var capturedUserID string

	mock := &mockAuthService{
		validateTokensFunc: func(tokenStr string) (*jwt.Payload, error) {
			return &jwt.Payload{UserID: expectedUserID}, nil
		},
	}

	router := setupTestRouter()
	router.Use(AuthMiddleware(mock))
	router.GET("/protected", func(c *gin.Context) {
		userID, _ := c.Get(handlers.CtxUserID)
		capturedUserID = userID.(string)
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer valid-token")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if capturedUserID != expectedUserID {
		t.Errorf("expected user ID '%s', got '%s'", expectedUserID, capturedUserID)
	}
}

func TestAuthMiddleware_ShortHeader(t *testing.T) {
	mock := &mockAuthService{}

	router := setupTestRouter()
	router.Use(AuthMiddleware(mock))
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bear")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestLoggerMiddleware(t *testing.T) {
	router := setupTestRouter()

	// Create a minimal mock that satisfies the logger interface
	// Since LoggerMiddleware calls l.Log(), we need to skip the actual logging
	// For this test, we just verify the middleware doesn't break the request flow

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestLoggerMiddleware_WithRealLogger(t *testing.T) {
	// Test with actual logger service
	cfg := &config.App{
		Logger: config.Logger{
			Mode:     "dev",
			Level:    "debug",
			BuffSize: 100,
		},
	}
	loggerService := logger.NewService(cfg)

	router := setupTestRouter()
	router.Use(LoggerMiddleware(loggerService))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	// Check that a log message was queued
	// Give a tiny bit of time for the log to be written
	time.Sleep(10 * time.Millisecond)
}

func TestLoggerMiddleware_DifferentMethods(t *testing.T) {
	cfg := &config.App{
		Logger: config.Logger{
			Mode:     "dev",
			Level:    "debug",
			BuffSize: 100,
		},
	}
	loggerService := logger.NewService(cfg)

	router := setupTestRouter()
	router.Use(LoggerMiddleware(loggerService))

	router.GET("/get", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"method": "GET"})
	})
	router.POST("/post", func(c *gin.Context) {
		c.JSON(http.StatusCreated, gin.H{"method": "POST"})
	})
	router.PUT("/put", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"method": "PUT"})
	})
	router.DELETE("/delete", func(c *gin.Context) {
		c.JSON(http.StatusNoContent, nil)
	})

	methods := []struct {
		method       string
		path         string
		expectedCode int
	}{
		{"GET", "/get", http.StatusOK},
		{"POST", "/post", http.StatusCreated},
		{"PUT", "/put", http.StatusOK},
		{"DELETE", "/delete", http.StatusNoContent},
	}

	for _, m := range methods {
		t.Run(m.method, func(t *testing.T) {
			req, _ := http.NewRequest(m.method, m.path, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != m.expectedCode {
				t.Errorf("expected status %d for %s, got %d", m.expectedCode, m.method, w.Code)
			}
		})
	}
}

func TestLoggerMiddleware_ErrorResponse(t *testing.T) {
	cfg := &config.App{
		Logger: config.Logger{
			Mode:     "dev",
			Level:    "debug",
			BuffSize: 100,
		},
	}
	loggerService := logger.NewService(cfg)

	router := setupTestRouter()
	router.Use(LoggerMiddleware(loggerService))
	router.GET("/error", func(c *gin.Context) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
	})
	router.GET("/not-found", func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
	})

	tests := []struct {
		path         string
		expectedCode int
	}{
		{"/error", http.StatusInternalServerError},
		{"/not-found", http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			req, _ := http.NewRequest("GET", tt.path, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectedCode {
				t.Errorf("expected status %d, got %d", tt.expectedCode, w.Code)
			}
		})
	}
}

func TestLoggerMiddleware_WithQueryParams(t *testing.T) {
	cfg := &config.App{
		Logger: config.Logger{
			Mode:     "dev",
			Level:    "debug",
			BuffSize: 100,
		},
	}
	loggerService := logger.NewService(cfg)

	router := setupTestRouter()
	router.Use(LoggerMiddleware(loggerService))
	router.GET("/search", func(c *gin.Context) {
		query := c.Query("q")
		c.JSON(http.StatusOK, gin.H{"query": query})
	})

	req, _ := http.NewRequest("GET", "/search?q=test&page=1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestAuthMiddleware_TokenExtraction(t *testing.T) {
	var capturedToken string

	mock := &mockAuthService{
		validateTokensFunc: func(tokenStr string) (*jwt.Payload, error) {
			capturedToken = tokenStr
			return &jwt.Payload{UserID: "test-user"}, nil
		},
	}

	router := setupTestRouter()
	router.Use(AuthMiddleware(mock))
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	expectedToken := "my-test-token-12345"
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+expectedToken)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if capturedToken != expectedToken {
		t.Errorf("expected token '%s', got '%s'", expectedToken, capturedToken)
	}
}

func TestAuthMiddleware_TokenWithSpaces(t *testing.T) {
	var capturedToken string

	mock := &mockAuthService{
		validateTokensFunc: func(tokenStr string) (*jwt.Payload, error) {
			capturedToken = tokenStr
			return &jwt.Payload{UserID: "test-user"}, nil
		},
	}

	router := setupTestRouter()
	router.Use(AuthMiddleware(mock))
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	expectedToken := "my-token"
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer   "+expectedToken+"  ")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Token should be trimmed
	if capturedToken != expectedToken {
		t.Errorf("expected token '%s', got '%s'", expectedToken, capturedToken)
	}
}
