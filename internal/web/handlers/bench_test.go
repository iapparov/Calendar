package handlers

import (
	"bytes"
	"calendar/internal/auth/jwt"
	"calendar/internal/domain/event"
	"calendar/internal/domain/user"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ---------------------------------------------------------------------------
// Shared setup helpers
// ---------------------------------------------------------------------------

func init() { gin.SetMode(gin.TestMode) }

var (
	benchUserID = uuid.New().String()
	benchEvent  = &event.Event{
		EventID: uuid.New(),
		UserID:  uuid.MustParse(benchUserID),
		Date:    time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC),
		Name:    "Team Meeting",
		Text:    "Discuss Q3 roadmap",
		Status:  event.StatusActive,
	}
	benchEvents = []*event.Event{benchEvent, benchEvent, benchEvent}
	benchTokens = &jwt.AuthTokens{AccessToken: "at", RefreshToken: "rt"}
	benchUser   = &user.User{ID: uuid.New(), Login: "bench", Email: "b@b.com"}
)

// stubEventService returns pre-built data with zero overhead so that benchmarks
// measure only the handler + gin serialisation path.
type stubEventService struct{}

func (s *stubEventService) Save(context.Context, string, string, string, string, string) (*event.Event, error) {
	return benchEvent, nil
}
func (s *stubEventService) Update(context.Context, string, string, string, string, string, string) (*event.Event, error) {
	return benchEvent, nil
}
func (s *stubEventService) Delete(context.Context, string, string) error { return nil }
func (s *stubEventService) LoadDay(_ context.Context, _ string, _ string) ([]*event.Event, error) {
	return benchEvents, nil
}
func (s *stubEventService) LoadWeek(_ context.Context, _ string, _ string) ([]*event.Event, error) {
	return benchEvents, nil
}
func (s *stubEventService) LoadMonth(_ context.Context, _ string, _ string) ([]*event.Event, error) {
	return benchEvents, nil
}

type stubAuthService struct{}

func (s *stubAuthService) Login(context.Context, string, string) (*jwt.AuthTokens, error) {
	return benchTokens, nil
}
func (s *stubAuthService) Register(context.Context, string, string, string, string) (*user.User, error) {
	return benchUser, nil
}
func (s *stubAuthService) RefreshTokens(string) (*jwt.AuthTokens, error) { return benchTokens, nil }
func (s *stubAuthService) ValidateTokens(string) (*jwt.Payload, error) {
	return &jwt.Payload{UserID: benchUserID}, nil
}

func benchRouter() *gin.Engine {
	r := gin.New() // no logger/recovery — pure handler speed
	eh := NewCalendarHandler(&stubEventService{})
	uh := NewUserHandler(&stubAuthService{})

	// Auth-free event endpoints with user_id injected via middleware stub.
	r.Use(func(c *gin.Context) { c.Set(CtxUserID, benchUserID); c.Next() })
	r.POST("/events", eh.CreateEvent)
	r.PUT("/events", eh.UpdateEvent)
	r.DELETE("/events", eh.DeleteEvent)
	r.GET("/events/day", eh.EventsForDay)
	r.GET("/events/week", eh.EventsForWeek)
	r.GET("/events/month", eh.EventsForMonth)

	// Auth endpoints.
	r.POST("/auth/login", uh.Login)
	r.POST("/auth/register", uh.Register)
	r.POST("/auth/refresh-token", uh.RefreshToken)
	return r
}

// ---------------------------------------------------------------------------
// Event handler benchmarks
// ---------------------------------------------------------------------------

func BenchmarkCreateEvent(b *testing.B) {
	router := benchRouter()
	body := []byte(`{"event_name":"Meeting","date":"2026-06-15","event":"text","reminder_time":"2026-06-15T09:00"}`)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/events", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
	}
}

func BenchmarkCreateEvent_Parallel(b *testing.B) {
	router := benchRouter()
	body := []byte(`{"event_name":"Meeting","date":"2026-06-15","event":"text","reminder_time":"2026-06-15T09:00"}`)

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodPost, "/events", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)
		}
	})
}

func BenchmarkUpdateEvent(b *testing.B) {
	router := benchRouter()
	body := []byte(`{"event_id":"` + uuid.New().String() + `","event_name":"Upd","date":"2026-06-15","event":"t","reminder_time":"2026-06-15T09:00"}`)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPut, "/events", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
	}
}

func BenchmarkDeleteEvent(b *testing.B) {
	router := benchRouter()
	body := []byte(`{"event_id":"` + uuid.New().String() + `"}`)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodDelete, "/events", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
	}
}

func BenchmarkEventsForDay(b *testing.B) {
	router := benchRouter()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/events/day?date=2026-06-15", nil)
		router.ServeHTTP(w, req)
	}
}

func BenchmarkEventsForDay_Parallel(b *testing.B) {
	router := benchRouter()

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/events/day?date=2026-06-15", nil)
			router.ServeHTTP(w, req)
		}
	})
}

func BenchmarkEventsForWeek(b *testing.B) {
	router := benchRouter()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/events/week?date=2026-06-15", nil)
		router.ServeHTTP(w, req)
	}
}

func BenchmarkEventsForMonth(b *testing.B) {
	router := benchRouter()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/events/month?date=2026-06-15", nil)
		router.ServeHTTP(w, req)
	}
}

// ---------------------------------------------------------------------------
// User / Auth handler benchmarks
// ---------------------------------------------------------------------------

func BenchmarkLogin(b *testing.B) {
	router := benchRouter()
	body := []byte(`{"login":"testuser","password":"Password123"}`)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
	}
}

func BenchmarkLogin_Parallel(b *testing.B) {
	router := benchRouter()
	body := []byte(`{"login":"testuser","password":"Password123"}`)

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)
		}
	})
}

func BenchmarkRegister(b *testing.B) {
	router := benchRouter()
	body := []byte(`{"login":"newuser","password":"Password123","email":"new@example.com","telegram_chat_id":"123"}`)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
	}
}

func BenchmarkRefreshToken(b *testing.B) {
	router := benchRouter()
	body := []byte(`{"refresh_token":"some-refresh-token"}`)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/auth/refresh-token", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
	}
}
