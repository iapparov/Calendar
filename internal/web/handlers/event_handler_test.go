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
	saveFunc      func(userId, date, eventName, eventText, reminder string, ctx context.Context) (*event.Event, error)
	updateFunc    func(eventId, userId, date, eventText, eventName, reminder string, ctx context.Context) (*event.Event, error)
	deleteFunc    func(eventId, userId string, ctx context.Context) error
	loadDayFunc   func(userID string, date string, ctx context.Context) ([]*event.Event, error)
	loadWeekFunc  func(userID string, weekStart string, ctx context.Context) ([]*event.Event, error)
	loadMonthFunc func(userID string, monthStart string, ctx context.Context) ([]*event.Event, error)
}

func (m *mockEventService) Save(userId, date, eventName, eventText, reminder string, ctx context.Context) (*event.Event, error) {
	if m.saveFunc != nil {
		return m.saveFunc(userId, date, eventName, eventText, reminder, ctx)
	}
	return nil, nil
}

func (m *mockEventService) Update(eventId, userId, date, eventText, eventName, reminder string, ctx context.Context) (*event.Event, error) {
	if m.updateFunc != nil {
		return m.updateFunc(eventId, userId, date, eventText, eventName, reminder, ctx)
	}
	return nil, nil
}

func (m *mockEventService) Delete(eventId, userId string, ctx context.Context) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(eventId, userId, ctx)
	}
	return nil
}

func (m *mockEventService) LoadDay(userID string, date string, ctx context.Context) ([]*event.Event, error) {
	if m.loadDayFunc != nil {
		return m.loadDayFunc(userID, date, ctx)
	}
	return nil, nil
}

func (m *mockEventService) LoadWeek(userID string, weekStart string, ctx context.Context) ([]*event.Event, error) {
	if m.loadWeekFunc != nil {
		return m.loadWeekFunc(userID, weekStart, ctx)
	}
	return nil, nil
}

func (m *mockEventService) LoadMonth(userID string, monthStart string, ctx context.Context) ([]*event.Event, error) {
	if m.loadMonthFunc != nil {
		return m.loadMonthFunc(userID, monthStart, ctx)
	}
	return nil, nil
}

// Mock AuthService
type mockAuthService struct {
	loginFunc          func(login, password string, ctx context.Context) (*jwt.AuthTokens, error)
	registerFunc       func(login, password, email, telegram string, ctx context.Context) (*user.User, error)
	refreshTokensFunc  func(tokenStr string) (*jwt.AuthTokens, error)
	validateTokensFunc func(tokenStr string) (*jwt.Payload, error)
}

func (m *mockAuthService) Login(login, password string, ctx context.Context) (*jwt.AuthTokens, error) {
	if m.loginFunc != nil {
		return m.loginFunc(login, password, ctx)
	}
	return nil, nil
}

func (m *mockAuthService) Register(login, password, email, telegram string, ctx context.Context) (*user.User, error) {
	if m.registerFunc != nil {
		return m.registerFunc(login, password, email, telegram, ctx)
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
		EventId: uuid.New(),
		UserID:  uuid.New(),
		Date:    time.Now(),
		Name:    "Test Event",
		Text:    "Description",
		Status:  event.StatusActive,
	}

	mock := &mockEventService{
		saveFunc: func(userId, date, eventName, eventText, reminder string, ctx context.Context) (*event.Event, error) {
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
		saveFunc: func(userId, date, eventName, eventText, reminder string, ctx context.Context) (*event.Event, error) {
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
		deleteFunc: func(eventId, userId string, ctx context.Context) error {
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
		deleteFunc: func(eventId, userId string, ctx context.Context) error {
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
		{EventId: uuid.New(), Name: "Event 1"},
		{EventId: uuid.New(), Name: "Event 2"},
	}

	mock := &mockEventService{
		loadDayFunc: func(userID string, date string, ctx context.Context) ([]*event.Event, error) {
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
		EventId: uuid.New(),
		Name:    "Updated Event",
	}

	mock := &mockEventService{
		updateFunc: func(eventId, userId, date, eventText, eventName, reminder string, ctx context.Context) (*event.Event, error) {
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
		{EventId: uuid.New(), Name: "Event 1"},
	}

	mock := &mockEventService{
		loadWeekFunc: func(userID string, weekStart string, ctx context.Context) ([]*event.Event, error) {
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
		{EventId: uuid.New(), Name: "Event 1"},
	}

	mock := &mockEventService{
		loadMonthFunc: func(userID string, monthStart string, ctx context.Context) ([]*event.Event, error) {
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
		loadDayFunc: func(userID string, date string, ctx context.Context) ([]*event.Event, error) {
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
		saveFunc: func(userId, date, eventName, eventText, reminder string, ctx context.Context) (*event.Event, error) {
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
		updateFunc: func(eventId, userId, date, eventText, eventName, reminder string, ctx context.Context) (*event.Event, error) {
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
		updateFunc: func(eventId, userId, date, eventText, eventName, reminder string, ctx context.Context) (*event.Event, error) {
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
		updateFunc: func(eventId, userId, date, eventText, eventName, reminder string, ctx context.Context) (*event.Event, error) {
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
		deleteFunc: func(eventId, userId string, ctx context.Context) error {
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
		deleteFunc: func(eventId, userId string, ctx context.Context) error {
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
