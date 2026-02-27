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
	service EventService
}

type EventService interface {
	Save(userId, date, eventName, eventText, reminder string, ctx context.Context) (*event.Event, error)
	Update(eventId, userId, date, eventText, eventName, reminder string, ctx context.Context) (*event.Event, error)
	Delete(eventId, userId string, ctx context.Context) error
	LoadDay(userID string, date string, ctx context.Context) ([]*event.Event, error)
	LoadWeek(userID string, weekStart string, ctx context.Context) ([]*event.Event, error)
	LoadMonth(userID string, monthStart string, ctx context.Context) ([]*event.Event, error)
}

func NewCalendarHandler(service EventService) *CalendarHandler {
	return &CalendarHandler{
		service: service,
	}
}

// @Summary Создание события
// @Description Создает новое событие для авторизованного пользователя
// @Tags event
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer токен" default(Bearer <token>)
// @Param request body dto.EventRequestCreate true "Данные события"
// @Success 201 {object} event.Event
// @Failure 400 {object} map[string]string{} "Сообщение об ошибке"
// @Failure 401 {object} map[string]string{} "Сообщение об ошибке"
// @Failure 500 {object} map[string]string{} "Сообщение об ошибке"
// @Router /api/v1/event/create_event [post]
func (h *CalendarHandler) CreateEvent(ctx *gin.Context) {
	var er dto.EventRequestCreate
	if err := ctx.ShouldBindJSON(&er); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userId, exist := ctx.Get(CtxUserID)
	if !exist {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "user_id not found in context"})
		return
	}
	evt, err := h.service.Save(userId.(string), er.Date, er.EventName, er.EventText, er.ReminderTime, ctx.Request.Context())
	if err != nil {
		if isValidationError(err) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, evt)
}

// @Summary Обновление события
// @Description Обновляет существующее событие авторизованного пользователя
// @Tags event
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer токен" default(Bearer <token>)
// @Param request body dto.EventRequestUpdate true "Данные для обновления события"
// @Success 200 {object} event.Event
// @Failure 400 {object} map[string]string{} "Сообщение об ошибке"
// @Failure 401 {object} map[string]string{} "Сообщение об ошибке"
// @Failure 404 {object} map[string]string{} "Сообщение об ошибке"
// @Failure 500 {object} map[string]string{} "Сообщение об ошибке"
// @Router /api/v1/event/update_event [put]
func (h *CalendarHandler) UpdateEvent(ctx *gin.Context) {
	var er dto.EventRequestUpdate
	if err := ctx.ShouldBindJSON(&er); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userId, exist := ctx.Get(CtxUserID)
	if !exist {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "user_id not found in context"})
		return
	}
	evt, err := h.service.Update(er.EventId, userId.(string), er.Date, er.EventText, er.EventName, er.ReminderTime, ctx.Request.Context())
	if err != nil {
		if isValidationError(err) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if isNotFound(err) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, evt)
}

// @Summary Удаление события
// @Description Удаляет событие авторизованного пользователя
// @Tags event
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer токен" default(Bearer <token>)
// @Param request body dto.EventRequestDelete true "ID события для удаления"
// @Success 204 "Событие удалено"
// @Failure 400 {object} map[string]string{} "Сообщение об ошибке"
// @Failure 401 {object} map[string]string{} "Сообщение об ошибке"
// @Failure 404 {object} map[string]string{} "Сообщение об ошибке"
// @Failure 500 {object} map[string]string{} "Сообщение об ошибке"
// @Router /api/v1/event/delete_event [delete]
func (h *CalendarHandler) DeleteEvent(ctx *gin.Context) {
	var evt dto.EventRequestDelete
	if err := ctx.ShouldBindJSON(&evt); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userId, exist := ctx.Get(CtxUserID)
	if !exist {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "user_id not found in context"})
		return
	}

	if err := h.service.Delete(evt.EventId, userId.(string), ctx.Request.Context()); err != nil {
		if isValidationError(err) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if isNotFound(err) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusNoContent, nil)

}

// @Summary Получение событий за день
// @Description Возвращает список событий пользователя за указанный день
// @Tags event
// @Produce json
// @Param Authorization header string true "Bearer токен" default(Bearer <token>)
// @Param date query string true "Дата в формате YYYY-MM-DD" example("2026-02-27")
// @Success 200 {array} event.Event
// @Failure 400 {object} map[string]string{} "Сообщение об ошибке"
// @Failure 401 {object} map[string]string{} "Сообщение об ошибке"
// @Failure 500 {object} map[string]string{} "Сообщение об ошибке"
// @Router /api/v1/event/events_for_day [get]
func (h *CalendarHandler) EventsForDay(ctx *gin.Context) {
	h.eventsHandler(h.service.LoadDay)(ctx)
}

// @Summary Получение событий за неделю
// @Description Возвращает список событий пользователя за неделю начиная с указанной даты
// @Tags event
// @Produce json
// @Param Authorization header string true "Bearer токен" default(Bearer <token>)
// @Param date query string true "Начало недели в формате YYYY-MM-DD" example("2026-02-23")
// @Success 200 {array} event.Event
// @Failure 400 {object} map[string]string{} "Сообщение об ошибке"
// @Failure 401 {object} map[string]string{} "Сообщение об ошибке"
// @Failure 500 {object} map[string]string{} "Сообщение об ошибке"
// @Router /api/v1/event/events_for_week [get]
func (h *CalendarHandler) EventsForWeek(ctx *gin.Context) {
	h.eventsHandler(h.service.LoadWeek)(ctx)
}

// @Summary Получение событий за месяц
// @Description Возвращает список событий пользователя за месяц начиная с указанной даты
// @Tags event
// @Produce json
// @Param Authorization header string true "Bearer токен" default(Bearer <token>)
// @Param date query string true "Начало месяца в формате YYYY-MM-DD" example("2026-02-01")
// @Success 200 {array} event.Event
// @Failure 400 {object} map[string]string{} "Сообщение об ошибке"
// @Failure 401 {object} map[string]string{} "Сообщение об ошибке"
// @Failure 500 {object} map[string]string{} "Сообщение об ошибке"
// @Router /api/v1/event/events_for_month [get]
func (h *CalendarHandler) EventsForMonth(ctx *gin.Context) {
	h.eventsHandler(h.service.LoadMonth)(ctx)
}

func (h *CalendarHandler) eventsHandler(loadFunc func(user string, date string, ctx context.Context) ([]*event.Event, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userId, exist := ctx.Get(CtxUserID)
		if !exist {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "user_id not found in context"})
			return
		}

		date := ctx.Query("date")
		if date == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "date query parameter is required"})
			return
		}

		events, err := loadFunc(userId.(string), date, ctx.Request.Context())
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if events == nil {
			events = make([]*event.Event, 0)
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
