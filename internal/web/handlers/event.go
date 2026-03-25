package handlers

import (
	"calendar/internal/domain"
	"calendar/internal/domain/event"
	"calendar/internal/web/dto"
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type CalendarHandler struct {
	service      EventService
	dayHandler   gin.HandlerFunc
	weekHandler  gin.HandlerFunc
	monthHandler gin.HandlerFunc
	emptyEvents  []*event.Event
}

type EventService interface {
	Save(ctx context.Context, userID, date, eventName, eventText, reminder string) (*event.Event, error)
	Update(ctx context.Context, eventID, userID, date, eventText, eventName, reminder string) (*event.Event, error)
	Delete(ctx context.Context, eventID, userID string) error
	LoadDay(ctx context.Context, userID string, date string) ([]*event.Event, error)
	LoadWeek(ctx context.Context, userID string, weekStart string) ([]*event.Event, error)
	LoadMonth(ctx context.Context, userID string, monthStart string) ([]*event.Event, error)
}

func NewCalendarHandler(service EventService) *CalendarHandler {
	h := &CalendarHandler{
		service: service,
	}
	// Pre-build closures once — avoid creating a new closure per request.
	h.dayHandler = h.buildEventsHandler(service.LoadDay)
	h.weekHandler = h.buildEventsHandler(service.LoadWeek)
	h.monthHandler = h.buildEventsHandler(service.LoadMonth)
	// emptyEvents is a pre-allocated empty slice to avoid allocation when there are no events.
	h.emptyEvents = make([]*event.Event, 0)
	return h
}

// @Summary Create event
// @Description Creates a new event for the authenticated user
// @Tags events
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token" default(Bearer <token>)
// @Param request body dto.EventRequestCreate true "Event data"
// @Success 201 {object} event.Event
// @Failure 400 {object} map[string]string{} "Error message"
// @Failure 401 {object} map[string]string{} "Error message"
// @Failure 500 {object} map[string]string{} "Error message"
// @Router /api/v1/events [post]
func (h *CalendarHandler) CreateEvent(ctx *gin.Context) {
	var er dto.EventRequestCreate
	if err := ctx.ShouldBindJSON(&er); err != nil {
		ctx.JSON(http.StatusBadRequest, &errorResponse{Error: err.Error()})
		return
	}
	userID, exist := ctx.Get(CtxUserID)
	if !exist {
		ctx.JSON(http.StatusBadRequest, errUserIDNotFound)
		return
	}
	evt, err := h.service.Save(ctx.Request.Context(), userID.(string), er.Date, er.EventName, er.EventText, er.ReminderTime)
	if err != nil {
		if isValidationError(err) {
			ctx.JSON(http.StatusBadRequest, &errorResponse{Error: err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, &errorResponse{Error: err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, evt)
}

// @Summary Update event
// @Description Updates an existing event for the authenticated user
// @Tags events
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token" default(Bearer <token>)
// @Param request body dto.EventRequestUpdate true "Event update data"
// @Success 200 {object} event.Event
// @Failure 400 {object} map[string]string{} "Error message"
// @Failure 401 {object} map[string]string{} "Error message"
// @Failure 404 {object} map[string]string{} "Error message"
// @Failure 500 {object} map[string]string{} "Error message"
// @Router /api/v1/events [put]
func (h *CalendarHandler) UpdateEvent(ctx *gin.Context) {
	var er dto.EventRequestUpdate
	if err := ctx.ShouldBindJSON(&er); err != nil {
		ctx.JSON(http.StatusBadRequest, &errorResponse{Error: err.Error()})
		return
	}
	userID, exist := ctx.Get(CtxUserID)
	if !exist {
		ctx.JSON(http.StatusBadRequest, errUserIDNotFound)
		return
	}
	evt, err := h.service.Update(ctx.Request.Context(), er.EventID, userID.(string), er.Date, er.EventText, er.EventName, er.ReminderTime)
	if err != nil {
		if isValidationError(err) {
			ctx.JSON(http.StatusBadRequest, &errorResponse{Error: err.Error()})
			return
		}
		if isNotFound(err) {
			ctx.JSON(http.StatusNotFound, &errorResponse{Error: err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, &errorResponse{Error: err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, evt)
}

// @Summary Delete event
// @Description Deletes an event for the authenticated user
// @Tags events
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token" default(Bearer <token>)
// @Param request body dto.EventRequestDelete true "Event ID to delete"
// @Success 204 "Event deleted"
// @Failure 400 {object} map[string]string{} "Error message"
// @Failure 401 {object} map[string]string{} "Error message"
// @Failure 404 {object} map[string]string{} "Error message"
// @Failure 500 {object} map[string]string{} "Error message"
// @Router /api/v1/events [delete]
func (h *CalendarHandler) DeleteEvent(ctx *gin.Context) {
	var evt dto.EventRequestDelete
	if err := ctx.ShouldBindJSON(&evt); err != nil {
		ctx.JSON(http.StatusBadRequest, &errorResponse{Error: err.Error()})
		return
	}

	userID, exist := ctx.Get(CtxUserID)
	if !exist {
		ctx.JSON(http.StatusBadRequest, errUserIDNotFound)
		return
	}

	if err := h.service.Delete(ctx.Request.Context(), evt.EventID, userID.(string)); err != nil {
		if isValidationError(err) {
			ctx.JSON(http.StatusBadRequest, &errorResponse{Error: err.Error()})
			return
		}
		if isNotFound(err) {
			ctx.JSON(http.StatusNotFound, &errorResponse{Error: err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, &errorResponse{Error: err.Error()})
		return
	}

	ctx.Status(http.StatusNoContent)

}

// @Summary Get events for a day
// @Description Returns a list of user events for the specified day
// @Tags events
// @Produce json
// @Param Authorization header string true "Bearer token" default(Bearer <token>)
// @Param date query string true "Date in YYYY-MM-DD format" example("2026-02-27")
// @Success 200 {array} event.Event
// @Failure 400 {object} map[string]string{} "Error message"
// @Failure 401 {object} map[string]string{} "Error message"
// @Failure 500 {object} map[string]string{} "Error message"
// @Router /api/v1/events/day [get]
func (h *CalendarHandler) EventsForDay(ctx *gin.Context) {
	h.dayHandler(ctx)
}

// @Summary Get events for a week
// @Description Returns a list of user events for the week starting from the specified date
// @Tags events
// @Produce json
// @Param Authorization header string true "Bearer token" default(Bearer <token>)
// @Param date query string true "Week start date in YYYY-MM-DD format" example("2026-02-23")
// @Success 200 {array} event.Event
// @Failure 400 {object} map[string]string{} "Error message"
// @Failure 401 {object} map[string]string{} "Error message"
// @Failure 500 {object} map[string]string{} "Error message"
// @Router /api/v1/events/week [get]
func (h *CalendarHandler) EventsForWeek(ctx *gin.Context) {
	h.weekHandler(ctx)
}

// @Summary Get events for a month
// @Description Returns a list of user events for the month starting from the specified date
// @Tags events
// @Produce json
// @Param Authorization header string true "Bearer token" default(Bearer <token>)
// @Param date query string true "Month start date in YYYY-MM-DD format" example("2026-02-01")
// @Success 200 {array} event.Event
// @Failure 400 {object} map[string]string{} "Error message"
// @Failure 401 {object} map[string]string{} "Error message"
// @Failure 500 {object} map[string]string{} "Error message"
// @Router /api/v1/events/month [get]
func (h *CalendarHandler) EventsForMonth(ctx *gin.Context) {
	h.monthHandler(ctx)
}

func (h *CalendarHandler) buildEventsHandler(loadFunc func(ctx context.Context, userID string, date string) ([]*event.Event, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userID, exist := ctx.Get(CtxUserID)
		if !exist {
			ctx.JSON(http.StatusBadRequest, errUserIDNotFound)
			return
		}

		date := ctx.Query("date")
		if date == "" {
			ctx.JSON(http.StatusBadRequest, errDateRequired)
			return
		}

		events, err := loadFunc(ctx.Request.Context(), userID.(string), date)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, &errorResponse{Error: err.Error()})
			return
		}

		if events == nil {
			events = h.emptyEvents
		}

		ctx.JSON(http.StatusOK, events)
	}
}

func isValidationError(err error) bool {
	return errors.Is(err, domain.ErrValidation)
}

func isNotFound(err error) bool {
	return errors.Is(err, domain.ErrNotFound)
}

func isConflict(err error) bool {
	return errors.Is(err, domain.ErrConflict)
}
