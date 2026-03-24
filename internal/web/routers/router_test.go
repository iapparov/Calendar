package routers

import (
	"bytes"
	"calendar/internal/auth/jwt"
	"calendar/internal/config"
	"calendar/internal/domain/event"
	"calendar/internal/domain/user"
	"calendar/internal/logger"
	"calendar/internal/web/handlers"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Mock EventService
type mockEventService struct{}

func (m *mockEventService) Save(ctx context.Context, userID, date, eventName, eventText, reminder string) (*event.Event, error) {
	return &event.Event{EventID: uuid.New()}, nil
}

func (m *mockEventService) Update(ctx context.Context, eventID, userID, date, eventText, eventName, reminder string) (*event.Event, error) {
	return &event.Event{EventID: uuid.New()}, nil
}

func (m *mockEventService) Delete(ctx context.Context, eventID, userID string) error {
	return nil
}

func (m *mockEventService) LoadDay(ctx context.Context, userID string, date string) ([]*event.Event, error) {
	return []*event.Event{}, nil
}

func (m *mockEventService) LoadWeek(ctx context.Context, userID string, weekStart string) ([]*event.Event, error) {
	return []*event.Event{}, nil
}

func (m *mockEventService) LoadMonth(ctx context.Context, userID string, monthStart string) ([]*event.Event, error) {
	return []*event.Event{}, nil
}

// Mock AuthService for router test
type mockRouterAuthService struct{}

func (m *mockRouterAuthService) Login(ctx context.Context, login, password string) (*jwt.AuthTokens, error) {
	return &jwt.AuthTokens{AccessToken: "test-token", RefreshToken: "refresh-token"}, nil
}

func (m *mockRouterAuthService) Register(ctx context.Context, login, password, email, telegram string) (*user.User, error) {
	return &user.User{ID: uuid.New(), Login: login, Email: email}, nil
}

func (m *mockRouterAuthService) RefreshTokens(tokenStr string) (*jwt.AuthTokens, error) {
	return &jwt.AuthTokens{AccessToken: "new-token", RefreshToken: "new-refresh"}, nil
}

func (m *mockRouterAuthService) ValidateTokens(tokenStr string) (*jwt.Payload, error) {
	return &jwt.Payload{UserID: uuid.New().String()}, nil
}

func createTestLogger() *logger.Service {
	cfg := &config.App{
		Logger: config.Logger{
			Mode:     "dev",
			Level:    "debug",
			BuffSize: 100,
		},
	}
	return logger.NewService(cfg)
}

func TestRegisterRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	eventService := &mockEventService{}
	authService := &mockRouterAuthService{}

	eventHandler := handlers.NewCalendarHandler(eventService)
	userHandler := handlers.NewUserHandler(authService)
	loggerService := createTestLogger()

	// Should not panic
	RegisterRoutes(engine, eventHandler, userHandler, loggerService)

	// Verify routes are registered
	routes := engine.Routes()
	if len(routes) == 0 {
		t.Error("expected routes to be registered")
	}
}

func TestRegisterRoutes_AuthEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	eventService := &mockEventService{}
	authService := &mockRouterAuthService{}

	eventHandler := handlers.NewCalendarHandler(eventService)
	userHandler := handlers.NewUserHandler(authService)
	loggerService := createTestLogger()

	RegisterRoutes(engine, eventHandler, userHandler, loggerService)

	// Test login endpoint
	t.Run("POST /api/v1/auth/login", func(t *testing.T) {
		body := `{"login":"testuser","password":"password123"}`
		req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
		}
	})

	// Test register endpoint
	t.Run("POST /api/v1/auth/register", func(t *testing.T) {
		body := `{"login":"newuser","password":"Password123","email":"test@example.com"}`
		req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
		}
	})

	// Test refresh-token endpoint
	t.Run("POST /api/v1/auth/refresh-token", func(t *testing.T) {
		body := `{"refresh_token":"some-refresh-token"}`
		req, _ := http.NewRequest("POST", "/api/v1/auth/refresh-token", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
		}
	})
}

func TestRegisterRoutes_EventEndpoints_WithAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	eventService := &mockEventService{}
	authService := &mockRouterAuthService{}

	eventHandler := handlers.NewCalendarHandler(eventService)
	userHandler := handlers.NewUserHandler(authService)
	loggerService := createTestLogger()

	RegisterRoutes(engine, eventHandler, userHandler, loggerService)

	// Test create_event endpoint with auth
	t.Run("POST /api/v1/events", func(t *testing.T) {
		body := `{"event_name":"Test Event","date":"2026-03-15","event":"Description","reminder_time":"2026-03-14"}`
		req, _ := http.NewRequest("POST", "/api/v1/events", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer valid-token")

		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("expected status %d, got %d, body: %s", http.StatusCreated, w.Code, w.Body.String())
		}
	})

	// Test events_for_day endpoint with auth
	t.Run("GET /api/v1/events/day", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/events/day?date=2026-03-15", nil)
		req.Header.Set("Authorization", "Bearer valid-token")

		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
		}
	})

	// Test events_for_week endpoint with auth
	t.Run("GET /api/v1/events/week", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/events/week?date=2026-03-15", nil)
		req.Header.Set("Authorization", "Bearer valid-token")

		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
		}
	})

	// Test events_for_month endpoint with auth
	t.Run("GET /api/v1/events/month", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/events/month?date=2026-03-15", nil)
		req.Header.Set("Authorization", "Bearer valid-token")

		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
		}
	})
}

func TestRegisterRoutes_EventEndpoints_WithoutAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	eventService := &mockEventService{}
	authService := &mockRouterAuthService{}

	eventHandler := handlers.NewCalendarHandler(eventService)
	userHandler := handlers.NewUserHandler(authService)
	loggerService := createTestLogger()

	RegisterRoutes(engine, eventHandler, userHandler, loggerService)

	// Test create_event endpoint without auth - should fail
	t.Run("POST /api/v1/events without auth", func(t *testing.T) {
		body := `{"event_name":"Test Event","date":"2026-03-15","event":"Description","reminder_time":"2026-03-14"}`
		req, _ := http.NewRequest("POST", "/api/v1/events", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		// No Authorization header

		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
		}
	})
}

func TestRegisterRoutes_CORS(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	eventService := &mockEventService{}
	authService := &mockRouterAuthService{}

	eventHandler := handlers.NewCalendarHandler(eventService)
	userHandler := handlers.NewUserHandler(authService)
	loggerService := createTestLogger()

	RegisterRoutes(engine, eventHandler, userHandler, loggerService)

	// Test OPTIONS request (CORS preflight)
	t.Run("OPTIONS /api/v1/auth/login", func(t *testing.T) {
		req, _ := http.NewRequest("OPTIONS", "/api/v1/auth/login", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		req.Header.Set("Access-Control-Request-Method", "POST")

		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		// CORS should respond with appropriate headers
		if w.Header().Get("Access-Control-Allow-Origin") == "" {
			t.Error("expected Access-Control-Allow-Origin header")
		}
	})
}

func TestRegisterRoutes_UpdateEvent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	eventService := &mockEventService{}
	authService := &mockRouterAuthService{}

	eventHandler := handlers.NewCalendarHandler(eventService)
	userHandler := handlers.NewUserHandler(authService)
	loggerService := createTestLogger()

	RegisterRoutes(engine, eventHandler, userHandler, loggerService)

	t.Run("PUT /api/v1/events", func(t *testing.T) {
		body := `{"event_id":"` + uuid.New().String() + `","event_name":"Updated Event","date":"2026-03-20","event":"Updated Description","reminder_time":"2026-03-19"}`
		req, _ := http.NewRequest("PUT", "/api/v1/events", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer valid-token")

		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d, body: %s", http.StatusOK, w.Code, w.Body.String())
		}
	})
}

func TestRegisterRoutes_DeleteEvent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	eventService := &mockEventService{}
	authService := &mockRouterAuthService{}

	eventHandler := handlers.NewCalendarHandler(eventService)
	userHandler := handlers.NewUserHandler(authService)
	loggerService := createTestLogger()

	RegisterRoutes(engine, eventHandler, userHandler, loggerService)

	t.Run("DELETE /api/v1/events", func(t *testing.T) {
		body := `{"event_id":"` + uuid.New().String() + `"}`
		req, _ := http.NewRequest("DELETE", "/api/v1/events", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer valid-token")

		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		if w.Code != http.StatusNoContent {
			t.Errorf("expected status %d, got %d", http.StatusNoContent, w.Code)
		}
	})
}
