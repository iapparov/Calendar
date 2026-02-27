package postgres

import (
	"calendar/internal/domain"
	"calendar/internal/domain/event"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (p *Postgres) Save(event *event.Event, ctx context.Context) error {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, p.cfg.DB.Postgres.WriteTimeout)
	defer cancel()

	query := `
	INSERT INTO events (id, user_id, event_date, event_name, description, status, reminder_time)
	VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := p.db.Exec(
		ctxWithTimeout,
		query,
		event.EventId,
		event.UserID,
		event.Date,
		event.Name,
		event.Text,
		event.Status,
		event.Reminder,
	)

	if err != nil {
		return fmt.Errorf("save event exec: %w", err)
	}
	return nil
}

func (p *Postgres) Delete(eventId, userId uuid.UUID, ctx context.Context) error {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, p.cfg.DB.Postgres.WriteTimeout)
	defer cancel()

	query := `
	UPDATE events
	SET status = $1 WHERE id = $2 AND user_id = $3
	`

	result, err := p.db.Exec(
		ctxWithTimeout,
		query,
		event.StatusArchive,
		eventId,
		userId,
	)
	if err != nil {
		return fmt.Errorf("failed to execute delete event query: %w", err)
	}

	rows := result.RowsAffected()

	if rows == 0 {
		return domain.ErrEventNotFound
	}
	return nil

}

func (p *Postgres) Update(event *event.Event, ctx context.Context) error {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, p.cfg.DB.Postgres.WriteTimeout)
	defer cancel()

	query := `
	UPDATE events
	SET event_date = $1, event_name = $2, description = $3, reminder_time = $4
	WHERE id = $5 AND user_id = $6
	`

	_, err := p.db.Exec(
		ctxWithTimeout,
		query,
		event.Date,
		event.Name,
		event.Text,
		event.Reminder,
		event.EventId,
		event.UserID,
	)

	if err != nil {
		return fmt.Errorf("update event exec: %w", err)
	}
	return nil
}

func (p *Postgres) Get(eventId, userId uuid.UUID, ctx context.Context) (*event.Event, error) {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, p.cfg.DB.Postgres.ReadTimeout)
	defer cancel()

	query := `
	SELECT id, user_id, event_date, event_name, description, status, reminder_time, reminder_sent
	FROM events
	WHERE id = $1 AND user_id = $2 AND status = $3
	`

	var evt event.Event

	err := p.db.QueryRow(
		ctxWithTimeout,
		query,
		eventId,
		userId,
		event.StatusActive,
	).Scan(
		&evt.EventId,
		&evt.UserID,
		&evt.Date,
		&evt.Name,
		&evt.Text,
		&evt.Status,
		&evt.Reminder,
		&evt.ReminderSent,
	)

	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("get event query: %w", err)
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}

	evt.Date = evt.Date.UTC()
	return &evt, nil

}

func (p *Postgres) LoadDay(userID uuid.UUID, date time.Time, ctx context.Context) ([]*event.Event, error) {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, p.cfg.DB.Postgres.ReadTimeout)
	defer cancel()

	// Приводим дату к началу дня в UTC
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	endOfDay := startOfDay.AddDate(0, 0, 1)

	query := `
	SELECT id, user_id, event_date, event_name, description, status, reminder_time, reminder_sent
	FROM events
	WHERE user_id = $1 AND event_date >= $2 AND event_date < $3 AND status = $4
	`

	rows, err := p.db.Query(
		ctxWithTimeout,
		query,
		userID,
		startOfDay,
		endOfDay,
		event.StatusActive,
	)

	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("load day query exec: %w", err)
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return []*event.Event{}, nil
	}
	defer rows.Close()

	result := make([]*event.Event, 0)
	for rows.Next() {
		var evt event.Event
		err := rows.Scan(
			&evt.EventId,
			&evt.UserID,
			&evt.Date,
			&evt.Name,
			&evt.Text,
			&evt.Status,
			&evt.Reminder,
			&evt.ReminderSent,
		)
		if err != nil {
			return nil, fmt.Errorf("load events scan: %w", err)
		}
		evt.Date = evt.Date.UTC()
		result = append(result, &evt)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("load events rows: %w", err)
	}
	return result, nil
}

func (p *Postgres) LoadWeek(userID uuid.UUID, date time.Time, ctx context.Context) ([]*event.Event, error) {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, p.cfg.DB.Postgres.ReadTimeout)
	defer cancel()

	// Приводим дату к полуночи
	date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())

	// Определяем начало недели (понедельник)
	weekday := int(date.Weekday())
	if weekday == 0 { // Sunday
		weekday = 7
	}
	startOfWeek := date.AddDate(0, 0, -(weekday - 1))
	endOfWeek := startOfWeek.AddDate(0, 0, 7)

	query := `
		SELECT id, user_id, event_date, event_name, description, status, reminder_time, reminder_sent
		FROM events
		WHERE user_id = $1
		  AND status = $2
		  AND event_date >= $3
		  AND event_date < $4
		ORDER BY event_date
	`

	rows, err := p.db.Query(
		ctxWithTimeout,
		query,
		userID,
		event.StatusActive,
		startOfWeek,
		endOfWeek,
	)
	if err != nil {
		return nil, fmt.Errorf("load events query: %w", err)
	}
	defer rows.Close()

	events := make([]*event.Event, 0)
	for rows.Next() {
		var e event.Event
		if err := rows.Scan(
			&e.EventId,
			&e.UserID,
			&e.Date,
			&e.Name,
			&e.Text,
			&e.Status,
			&e.Reminder,
			&e.ReminderSent,
		); err != nil {
			return nil, fmt.Errorf("load events scan: %w", err)
		}
		events = append(events, &e)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("load events rows: %w", err)
	}

	return events, nil
}

func (p *Postgres) LoadMonth(userID uuid.UUID, date time.Time, ctx context.Context) ([]*event.Event, error) {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, p.cfg.DB.Postgres.ReadTimeout)
	defer cancel()

	// Начало месяца
	startOfMonth := time.Date(
		date.Year(),
		date.Month(),
		1,
		0, 0, 0, 0,
		date.Location(),
	)

	// Начало следующего месяца
	endOfMonth := startOfMonth.AddDate(0, 1, 0)

	query := `
		SELECT id, user_id, event_date, event_name, description, status, reminder_time, reminder_sent
		FROM events
		WHERE user_id = $1
		  AND status = $2
		  AND event_date >= $3
		  AND event_date < $4
		ORDER BY event_date
	`

	rows, err := p.db.Query(
		ctxWithTimeout,
		query,
		userID,
		event.StatusActive,
		startOfMonth,
		endOfMonth,
	)

	if err != nil {
		return nil, fmt.Errorf("load events: %w", err)
	}
	defer rows.Close()

	events := make([]*event.Event, 0)
	for rows.Next() {
		var e event.Event
		if err := rows.Scan(
			&e.EventId,
			&e.UserID,
			&e.Date,
			&e.Name,
			&e.Text,
			&e.Status,
			&e.Reminder,
			&e.ReminderSent,
		); err != nil {
			return nil, fmt.Errorf("failed to load event: %w", err)
		}
		events = append(events, &e)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan event rows: %w", err)
	}

	return events, nil
}

func (p *Postgres) CleanEvents(ctx context.Context, beforeDate time.Time) (int, error) {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, p.cfg.DB.Postgres.WriteTimeout)
	defer cancel()

	query := `
		UPDATE events
		SET status = $1
		WHERE event_date < $2 AND status = $3
	`

	result, err := p.db.Exec(
		ctxWithTimeout,
		query,
		event.StatusArchive,
		beforeDate,
		event.StatusActive,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to clean events: %w", err)
	}

	rowsAffected := result.RowsAffected()

	return int(rowsAffected), nil
}

// MarkReminderSent отмечает уведомление как отправленное
func (p *Postgres) MarkReminderSent(eventID string, ctx context.Context) error {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, p.cfg.DB.Postgres.WriteTimeout)
	defer cancel()

	query := `UPDATE events SET reminder_sent = true WHERE id = $1`

	_, err := p.db.Exec(ctxWithTimeout, query, eventID)
	if err != nil {
		return fmt.Errorf("mark reminder sent: %w", err)
	}

	return nil
}

// LoadPendingReminders загружает все активные события с напоминаниями в будущем
// Используется для прогрева sender service при перезапуске
func (p *Postgres) LoadPendingReminders(ctx context.Context) ([]*event.Event, error) {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, p.cfg.DB.Postgres.LongTimeout)
	defer cancel()

	query := `
		SELECT id, user_id, event_date, event_name, description, status, reminder_time
		FROM events
		WHERE status = $1
		  AND reminder_sent = false
		  AND reminder_time IS NOT NULL
		  AND reminder_time > $2
		ORDER BY reminder_time
	`

	rows, err := p.db.Query(
		ctxWithTimeout,
		query,
		event.StatusActive,
		time.Now().UTC(),
	)
	if err != nil {
		return nil, fmt.Errorf("load pending reminders: %w", err)
	}
	defer rows.Close()

	events := make([]*event.Event, 0)
	for rows.Next() {
		var e event.Event
		if err := rows.Scan(
			&e.EventId,
			&e.UserID,
			&e.Date,
			&e.Name,
			&e.Text,
			&e.Status,
			&e.Reminder,
		); err != nil {
			return nil, fmt.Errorf("scan pending reminder: %w", err)
		}
		e.Date = e.Date.UTC()
		e.Reminder = e.Reminder.UTC()
		events = append(events, &e)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return events, nil
}
