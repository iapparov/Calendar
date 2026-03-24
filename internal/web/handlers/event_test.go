package handlers

import (
	"bytes"
	"calendar/internal/auth/jwt"
	"calendar/internal/domain"
	"calendar/internal/domain/event"
	"calendar/internal/domain/user"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Mock EventService
type mockEventService struct {
	saveFunc      func(ctx context.Context, userID, date, eventName, eventText, reminder string) (*event.Event, error)
	updateFunc    func(ctx context.Context, eventID, userID, date, eventText, eventName, reminder string) (*event.Event, error)
	deleteFunc    func(ctx context.Context, eventID, userID string) error
	loadDayFunc   func(ctx context.Context, userID string, date string) ([]*event.Event, error)
	loadWeekFunc  func(ctx context.Context, userID string, weekStart string) ([]*event.Event, error)
	loadMonthFunc func(ctx context.Context, userID string, monthStart string) ([]*event.Event, error)
}

func (m *mockEventService) Save(ctx context.Context, userID, date, eventName, eventText, reminder string) (*event.Event, error) {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, userID, date, eventName, eventText, reminder)
	}
	return nil, nil
}

func (m *mockEventService) Update(ctx context.Context, eventID, userID, date, eventText, eventName, reminder string) (*event.Event, error) {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, eventID, userID, date, eventText, eventName, reminder)
	}
	return nil, nil
}

func (m *mockEventService) Delete(ctx context.Context, eventID, userID string) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, eventID, userID)
	}
	return nil
}

func (m *mockEventService) LoadDay(ctx context.Context, userID string, date string) ([]*event.Event, error) {
	if m.loadDayFunc != nil {
		return m.loadDayFunc(ctx, userID, date)
	}
	return nil, nil
}

func (m *mockEventService) LoadWeek(ctx context.Context, userID string, weekStart string) ([]*event.Event, error) {
	if m.loadWeekFunc != nil {
		return m.loadWeekFunc(ctx, userID, weekStart)
	}
	return nil, nil
}

func (m *mockEventService) LoadMonth(ctx context.Context, userID string, monthStart string) ([]*event.Event, error) {
	if m.loadMonthFunc != nil {
		return m.loadMonthFunc(ctx, userID, monthStart)
	}
	return nil, nil
}

// Mock AuthService
type mockAuthService struct {
	loginFunc          func(ctx context.Context, login, password string) (*jwt.AuthTokens, error)
	registerFunc       func(ctx context.Context, login, password, email, telegram string) (*user.User, error)
	refreshTokensFunc  func(tokenStr string) (*jwt.AuthTokens, error)
	validateTokensFunc func(tokenStr string) (*jwt.Payload, error)
}

func (m *mockAuthService) Login(ctx context.Context, login, password string) (*jwt.AuthTokens, error) {
	if m.loginFunc != nil {
		return m.loginFunc(ctx, login, password)
	}
	return nil, nil
}

func (m *mockAuthService) Register(ctx context.Context, login, password, email, telegram string) (*user.User, error) {
	if m.registerFunc != nil {
		return m.registerFunc(ctx, login, password, email, telegram)
	}
	return nil, nil
}

func (m *mockAuthService) RefreshTokens(tokenStr string) (*jwt.AuthTokens, error) {
	if m.refreshTokensFunc != nil {
		return m.refreshTokensFunc(tokenStr)
	}
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

func TestNewCalendarHandler(t *testing.T) {
	mock := &mockEventService{}
	handler := NewCalendarHandler(mock)

	if handler == nil {
		t.Fatal("handler should not be nil")
	}
	if handler.service == nil {
		t.Error("service should not be nil")
	}
}

func TestCreateEvent_Success(t *testing.T) {
	expectedEvent := &event.Event{
		EventID: uuid.New(),
		UserID:  uuid.New(),
		Date:    time.Now(),
		Name:    "Test Event",
		Text:    "Description",
		Status:  event.StatusActive,
	}

	mock := &mockEventService{
		saveFunc: func(ctx context.Context, userID, date, eventName, eventText, reminder string) (*event.Event, error) {
			return expectedEvent, nil
		},
	}

	handler := NewCalendarHandler(mock)
	router := setupTestRouter()

	router.POST("/create_event", func(c *gin.Context) {
		c.Set(CtxUserID, uuid.New().String())
		handler.CreateEvent(c)
	})

	body := `{"event_name":"Test Event","date":"2026-03-15","event":"Description","reminder_time":"2026-03-14T10:30"}`
	req, _ := http.NewRequest("POST", "/create_event", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, w.Code)
	}
}

func TestCreateEvent_NoUserID(t *testing.T) {
	mock := &mockEventService{}
	handler := NewCalendarHandler(mock)
	router := setupTestRouter()

	router.POST("/create_event", handler.CreateEvent)

	body := `{"event_name":"Test Event","date":"2026-03-15","event":"Description","reminder_time":"2026-03-14T10:30"}`
	req, _ := http.NewRequest("POST", "/create_event", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestCreateEvent_ValidationError(t *testing.T) {
	mock := &mockEventService{
		saveFunc: func(ctx context.Context, userID, date, eventName, eventText, reminder string) (*event.Event, error) {
			return nil, domain.ErrInvalidEventName
		},
	}

	handler := NewCalendarHandler(mock)
	router := setupTestRouter()

	router.POST("/create_event", func(c *gin.Context) {
		c.Set(CtxUserID, uuid.New().String())
		handler.CreateEvent(c)
	})

	body := `{"event_name":"","date":"2026-03-15","event":"Description","reminder_time":"2026-03-14T10:30"}`
	req, _ := http.NewRequest("POST", "/create_event", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestDeleteEvent_Success(t *testing.T) {
	mock := &mockEventService{
		deleteFunc: func(ctx context.Context, eventID, userID string) error {
			return nil
		},
	}

	handler := NewCalendarHandler(mock)
	router := setupTestRouter()

	router.DELETE("/delete_event", func(c *gin.Context) {
		c.Set(CtxUserID, uuid.New().String())
		handler.DeleteEvent(c)
	})

	body := `{"event_id":"` + uuid.New().String() + `"}`
	req, _ := http.NewRequest("DELETE", "/delete_event", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected status %d, got %d", http.StatusNoContent, w.Code)
	}
}

func TestDeleteEvent_NotFound(t *testing.T) {
	mock := &mockEventService{
		deleteFunc: func(ctx context.Context, eventID, userID string) error {
			return domain.ErrEventNotFound
		},
	}

	handler := NewCalendarHandler(mock)
	router := setupTestRouter()

	router.DELETE("/delete_event", func(c *gin.Context) {
		c.Set(CtxUserID, uuid.New().String())
		handler.DeleteEvent(c)
	})

	body := `{"event_id":"` + uuid.New().String() + `"}`
	req, _ := http.NewRequest("DELETE", "/delete_event", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestEventsForDay_Success(t *testing.T) {
	events := []*event.Event{
		{EventID: uuid.New(), Name: "Event 1"},
		{EventID: uuid.New(), Name: "Event 2"},
	}

	mock := &mockEventService{
		loadDayFunc: func(ctx context.Context, userID string, date string) ([]*event.Event, error) {
			return events, nil
		},
	}

	handler := NewCalendarHandler(mock)
	router := setupTestRouter()

	router.GET("/events_for_day", func(c *gin.Context) {
		c.Set(CtxUserID, uuid.New().String())
		handler.EventsForDay(c)
	})

	req, _ := http.NewRequest("GET", "/events_for_day?date=2026-03-15", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var result []*event.Event
	err := json.Unmarshal(w.Body.Bytes(), &result)
	if err != nil {
		t.Errorf("failed to unmarshal response: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 events, got %d", len(result))
	}
}

func TestEventsForDay_MissingDate(t *testing.T) {
	mock := &mockEventService{}
	handler := NewCalendarHandler(mock)
	router := setupTestRouter()

	router.GET("/events_for_day", func(c *gin.Context) {
		c.Set(CtxUserID, uuid.New().String())
		handler.EventsForDay(c)
	})

	req, _ := http.NewRequest("GET", "/events_for_day", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestUpdateEvent_Success(t *testing.T) {
	updatedEvent := &event.Event{
		EventID: uuid.New(),
		Name:    "Updated Event",
	}

	mock := &mockEventService{
		updateFunc: func(ctx context.Context, eventID, userID, date, eventText, eventName, reminder string) (*event.Event, error) {
			return updatedEvent, nil
		},
	}

	handler := NewCalendarHandler(mock)
	router := setupTestRouter()

	router.PUT("/update_event", func(c *gin.Context) {
		c.Set(CtxUserID, uuid.New().String())
		handler.UpdateEvent(c)
	})

	body := `{"event_id":"` + uuid.New().String() + `","event_name":"Updated Event","date":"2026-03-15","event":"Updated desc","reminder_time":"2026-03-14T10:30"}`
	req, _ := http.NewRequest("PUT", "/update_event", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestIsValidationError(t *testing.T) {
	tests := []struct {
		err      error
		expected bool
	}{
		{domain.ErrValidation, true},
		{domain.ErrInvalidLogin, true},
		{domain.ErrNotFound, false},
		{nil, false},
	}

	for _, tt := range tests {
		result := isValidationError(tt.err)
		if result != tt.expected {
			t.Errorf("isValidationError(%v) = %v, want %v", tt.err, result, tt.expected)
		}
	}
}

func TestIsNotFound(t *testing.T) {
	tests := []struct {
		err      error
		expected bool
	}{
		{domain.ErrNotFound, true},
		{domain.ErrEventNotFound, true},
		{domain.ErrValidation, false},
		{nil, false},
	}

	for _, tt := range tests {
		result := isNotFound(tt.err)
		if result != tt.expected {
			t.Errorf("isNotFound(%v) = %v, want %v", tt.err, result, tt.expected)
		}
	}
}

func TestIsConflict(t *testing.T) {
	tests := []struct {
		err      error
		expected bool
	}{
		{domain.ErrConflict, true},
		{domain.ErrUserAlreadyExists, true},
		{domain.ErrValidation, false},
		{nil, false},
	}

	for _, tt := range tests {
		result := isConflict(tt.err)
		if result != tt.expected {
			t.Errorf("isConflict(%v) = %v, want %v", tt.err, result, tt.expected)
		}
	}
}

func TestEventsForWeek_Success(t *testing.T) {
	events := []*event.Event{
		{EventID: uuid.New(), Name: "Event 1"},
	}

	mock := &mockEventService{
		loadWeekFunc: func(ctx context.Context, userID string, weekStart string) ([]*event.Event, error) {
			return events, nil
		},
	}

	handler := NewCalendarHandler(mock)
	router := setupTestRouter()

	router.GET("/events_for_week", func(c *gin.Context) {
		c.Set(CtxUserID, uuid.New().String())
		handler.EventsForWeek(c)
	})

	req, _ := http.NewRequest("GET", "/events_for_week?date=2026-03-15", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestEventsForMonth_Success(t *testing.T) {
	events := []*event.Event{
		{EventID: uuid.New(), Name: "Event 1"},
	}

	mock := &mockEventService{
		loadMonthFunc: func(ctx context.Context, userID string, monthStart string) ([]*event.Event, error) {
			return events, nil
		},
	}

	handler := NewCalendarHandler(mock)
	router := setupTestRouter()

	router.GET("/events_for_month", func(c *gin.Context) {
		c.Set(CtxUserID, uuid.New().String())
		handler.EventsForMonth(c)
	})

	req, _ := http.NewRequest("GET", "/events_for_month?date=2026-03-15", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestEventsForDay_NoUserID(t *testing.T) {
	mock := &mockEventService{}
	handler := NewCalendarHandler(mock)
	router := setupTestRouter()

	router.GET("/events_for_day", handler.EventsForDay)

	req, _ := http.NewRequest("GET", "/events_for_day?date=2026-03-15", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestEventsForDay_ServiceError(t *testing.T) {
	mock := &mockEventService{
		loadDayFunc: func(ctx context.Context, userID string, date string) ([]*event.Event, error) {
			return nil, errors.New("database error")
		},
	}

	handler := NewCalendarHandler(mock)
	router := setupTestRouter()

	router.GET("/events_for_day", func(c *gin.Context) {
		c.Set(CtxUserID, uuid.New().String())
		handler.EventsForDay(c)
	})

	req, _ := http.NewRequest("GET", "/events_for_day?date=2026-03-15", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestCreateEvent_InvalidJSON(t *testing.T) {
	mock := &mockEventService{}
	handler := NewCalendarHandler(mock)
	router := setupTestRouter()

	router.POST("/create_event", func(c *gin.Context) {
		c.Set(CtxUserID, uuid.New().String())
		handler.CreateEvent(c)
	})

	body := `{invalid json}`
	req, _ := http.NewRequest("POST", "/create_event", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestCreateEvent_ServiceError(t *testing.T) {
	mock := &mockEventService{
		saveFunc: func(ctx context.Context, userID, date, eventName, eventText, reminder string) (*event.Event, error) {
			return nil, errors.New("database error")
		},
	}

	handler := NewCalendarHandler(mock)
	router := setupTestRouter()

	router.POST("/create_event", func(c *gin.Context) {
		c.Set(CtxUserID, uuid.New().String())
		handler.CreateEvent(c)
	})

	body := `{"event_name":"Test","date":"2026-03-15","event":"Desc","reminder_time":"2026-03-14T10:30"}`
	req, _ := http.NewRequest("POST", "/create_event", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestUpdateEvent_NoUserID(t *testing.T) {
	mock := &mockEventService{}
	handler := NewCalendarHandler(mock)
	router := setupTestRouter()

	router.PUT("/update_event", handler.UpdateEvent)

	body := `{"event_id":"` + uuid.New().String() + `","event_name":"Updated","date":"2026-03-15","event":"Desc","reminder_time":"2026-03-14T10:30"}`
	req, _ := http.NewRequest("PUT", "/update_event", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestUpdateEvent_NotFound(t *testing.T) {
	mock := &mockEventService{
		updateFunc: func(ctx context.Context, eventID, userID, date, eventText, eventName, reminder string) (*event.Event, error) {
			return nil, domain.ErrEventNotFound
		},
	}

	handler := NewCalendarHandler(mock)
	router := setupTestRouter()

	router.PUT("/update_event", func(c *gin.Context) {
		c.Set(CtxUserID, uuid.New().String())
		handler.UpdateEvent(c)
	})

	body := `{"event_id":"` + uuid.New().String() + `","event_name":"Updated","date":"2026-03-15","event":"Desc","reminder_time":"2026-03-14T10:30"}`
	req, _ := http.NewRequest("PUT", "/update_event", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestUpdateEvent_ValidationError(t *testing.T) {
	mock := &mockEventService{
		updateFunc: func(ctx context.Context, eventID, userID, date, eventText, eventName, reminder string) (*event.Event, error) {
			return nil, domain.ErrInvalidEventName
		},
	}

	handler := NewCalendarHandler(mock)
	router := setupTestRouter()

	router.PUT("/update_event", func(c *gin.Context) {
		c.Set(CtxUserID, uuid.New().String())
		handler.UpdateEvent(c)
	})

	body := `{"event_id":"` + uuid.New().String() + `","event_name":"","date":"2026-03-15","event":"Desc","reminder_time":"2026-03-14T10:30"}`
	req, _ := http.NewRequest("PUT", "/update_event", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestUpdateEvent_ServiceError(t *testing.T) {
	mock := &mockEventService{
		updateFunc: func(ctx context.Context, eventID, userID, date, eventText, eventName, reminder string) (*event.Event, error) {
			return nil, errors.New("database error")
		},
	}

	handler := NewCalendarHandler(mock)
	router := setupTestRouter()

	router.PUT("/update_event", func(c *gin.Context) {
		c.Set(CtxUserID, uuid.New().String())
		handler.UpdateEvent(c)
	})

	body := `{"event_id":"` + uuid.New().String() + `","event_name":"Test","date":"2026-03-15","event":"Desc","reminder_time":"2026-03-14T10:30"}`
	req, _ := http.NewRequest("PUT", "/update_event", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestDeleteEvent_NoUserID(t *testing.T) {
	mock := &mockEventService{}
	handler := NewCalendarHandler(mock)
	router := setupTestRouter()

	router.DELETE("/delete_event", handler.DeleteEvent)

	body := `{"event_id":"` + uuid.New().String() + `"}`
	req, _ := http.NewRequest("DELETE", "/delete_event", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestDeleteEvent_ValidationError(t *testing.T) {
	mock := &mockEventService{
		deleteFunc: func(ctx context.Context, eventID, userID string) error {
			return domain.ErrInvalidUserID
		},
	}

	handler := NewCalendarHandler(mock)
	router := setupTestRouter()

	router.DELETE("/delete_event", func(c *gin.Context) {
		c.Set(CtxUserID, uuid.New().String())
		handler.DeleteEvent(c)
	})

	body := `{"event_id":"invalid"}`
	req, _ := http.NewRequest("DELETE", "/delete_event", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestDeleteEvent_ServiceError(t *testing.T) {
	mock := &mockEventService{
		deleteFunc: func(ctx context.Context, eventID, userID string) error {
			return errors.New("database error")
		},
	}

	handler := NewCalendarHandler(mock)
	router := setupTestRouter()

	router.DELETE("/delete_event", func(c *gin.Context) {
		c.Set(CtxUserID, uuid.New().String())
		handler.DeleteEvent(c)
	})

	body := `{"event_id":"` + uuid.New().String() + `"}`
	req, _ := http.NewRequest("DELETE", "/delete_event", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}
